package dell

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/go-logr/logr"
	"github.com/stmcginnis/gofish/schemas"

	"github.com/bmc-toolbox/bmclib/v2/internal/redfishwrapper"
)

// iDRAC allows one pending BIOS config job at a time. Any write to Bios/Settings (or to a resource
// that mirrors a BIOS attribute, such as the ComputerSystem SecureBoot resource) seals that job,
// and every further write is rejected with the message below until the job is deleted or runs.
//
// The middle segment of the message ID is the message registry's schema version, which differs
// between iDRAC generations ("IDRAC.2.14.SYS011", "IDRAC.2.9.SYS011"), so only the first and last
// segments are matched.
const (
	dellPendingSettingsMessageRegistry = "IDRAC"
	dellPendingSettingsMessageKey      = "SYS011"
)

// isPendingSettingsConflict reports whether err is iDRAC's rejection of a Bios/Settings write
// because a pending BIOS config job already exists.
func isPendingSettingsConflict(err error) bool {
	var redfishErr *schemas.Error
	if !errors.As(err, &redfishErr) {
		return false
	}

	for i := range redfishErr.ExtendedInfos {
		if isDellPendingSettingsMessageID(redfishErr.ExtendedInfos[i].MessageID) {
			return true
		}
	}

	return false
}

func isDellPendingSettingsMessageID(messageID string) bool {
	parts := strings.Split(messageID, ".")
	return len(parts) >= 2 &&
		parts[0] == dellPendingSettingsMessageRegistry &&
		parts[len(parts)-1] == dellPendingSettingsMessageKey
}

// recoveringRedfishClient is a redfishwrapper.Client whose SetBiosConfiguration copes with
// iDRAC's one-pending-job limit instead of failing on it. Calls made through the wrapper
// (c.redfishwrapper.SetBiosConfiguration) resolve here; every other method is promoted unchanged,
// including the ones that write BIOS state through other resources (the SecureBoot resource,
// ResetBiosConfiguration), and Go has no virtual dispatch, so naming the embedded Client
// explicitly reaches the plain method.
type recoveringRedfishClient struct {
	*redfishwrapper.Client
	log logr.Logger
}

// errNothingToInspect means the system has no Bios resource, or its Bios resource does not
// redirect writes to a separate pending settings resource, so there is nothing to look at before
// writing.
var errNothingToInspect = errors.New("no pending BIOS settings resource to inspect")

// errNoPendingBIOSConfigJob is reported when a write still conflicts although no BIOS config job
// could be found to cancel.
var errNoPendingBIOSConfigJob = errors.New("no pending BIOS config job found to cancel")

// SetBiosConfiguration stages BIOS Setup attribute changes.
//
// It reads the pending attributes before writing, since a pending job holds Bios/Settings locked
// whichever attribute is written:
//
//   - nothing pending: a plain write, as redfishwrapper.Client does it;
//   - something pending: see applyOverPendingJob.
//
// A write rejected with SYS011 despite an empty pending set (for example because another client
// staged a change in between) is returned to the caller as is.
//
// It is not safe against another client writing BIOS settings on the same BMC at the same time:
// the read, the job delete and the write are separate requests, and a change staged by someone
// else in between is lost together with the job.
func (c *recoveringRedfishClient) SetBiosConfiguration(ctx context.Context, attrs map[string]string) error {
	current, err := c.readBIOSAttributes(ctx)
	switch {
	case errors.Is(err, errNothingToInspect):
		// No staging resource to inspect: write as usual.
	case err != nil:
		return fmt.Errorf("reading pending BIOS settings: %w", err)
	case len(current.pending) > 0:
		return c.applyOverPendingJob(ctx, current, attrs)
	}

	return c.Client.SetBiosConfiguration(ctx, attrs)
}

// biosResource is the part of the Bios resource that is needed: its applied attributes and the
// standard @Redfish.Settings block that points writes at a separate pending resource.
type biosResource struct {
	Attributes      map[string]any `json:"Attributes"`
	RedfishSettings struct {
		SettingsObject struct {
			ODataID string `json:"@odata.id"`
		} `json:"SettingsObject"`
	} `json:"@Redfish.Settings"`
}

// biosAttributes are the BIOS attributes as applied, and as staged for the next reset.
type biosAttributes struct {
	applied, pending map[string]any
}

// readBIOSAttributes returns the applied BIOS attributes and those already staged in the Bios
// resource's pending settings object (Bios/Settings on iDRAC), which is empty if nothing is staged.
// Numbers are kept as json.Number so that resubmitting them does not alter large values.
func (c *recoveringRedfishClient) readBIOSAttributes(ctx context.Context) (biosAttributes, error) {
	biosURL, err := c.SystemsBIOSOdataID(ctx)
	if err != nil {
		return biosAttributes{}, fmt.Errorf("%w: finding Bios resource: %v", errNothingToInspect, err)
	}

	resp, err := c.Get(biosURL)
	if err != nil {
		return biosAttributes{}, fmt.Errorf("reading Bios resource: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var bios biosResource
	dec := json.NewDecoder(resp.Body)
	dec.UseNumber()
	if err := dec.Decode(&bios); err != nil {
		return biosAttributes{}, fmt.Errorf("decoding Bios resource: %w", err)
	}

	settingsURL := bios.RedfishSettings.SettingsObject.ODataID
	if settingsURL == "" {
		return biosAttributes{}, fmt.Errorf("%w: Bios resource has no @Redfish.Settings.SettingsObject", errNothingToInspect)
	}
	if settingsURL == biosURL {
		// Writes go to the Bios resource itself, whose attributes are the applied ones, not pending.
		return biosAttributes{}, fmt.Errorf("%w: SettingsObject is the Bios resource itself", errNothingToInspect)
	}

	settingsResp, err := c.Get(settingsURL)
	if err != nil {
		return biosAttributes{}, fmt.Errorf("reading %s: %w", settingsURL, err)
	}
	defer func() { _ = settingsResp.Body.Close() }()

	var settings struct {
		Attributes map[string]any `json:"Attributes"`
	}
	dec = json.NewDecoder(settingsResp.Body)
	dec.UseNumber()
	if err := dec.Decode(&settings); err != nil {
		return biosAttributes{}, fmt.Errorf("decoding %s: %w", settingsURL, err)
	}

	return biosAttributes{applied: bios.Attributes, pending: settings.Attributes}, nil
}

// applyOverPendingJob handles a write while BIOS attributes are already staged (pending is
// non-empty), which means a pending job is holding Bios/Settings locked.
//
// If every requested attribute already has the requested value, staged or, when nothing is staged
// for it, applied, and a live job carries the staged ones, there is nothing to do: the job is left
// alone, so repeating a call is idempotent.
//
// Otherwise the job is deleted (only if it has not started applying), the requested attributes
// are merged over the staged ones, and the merged set is written once, so that a sequence of
// BIOS-affecting calls ends up as one job and one reboot. The merge carries over attributes that
// an earlier caller staged and this caller did not ask for: iDRAC keeps a single job, and
// deleting it discards what it staged.
func (c *recoveringRedfishClient) applyOverPendingJob(ctx context.Context, current biosAttributes, newAttrs map[string]string) error {
	jobs, err := c.pendingBIOSConfigJobs(ctx)
	if err != nil {
		return fmt.Errorf("finding pending BIOS config job: %w", err)
	}

	if len(jobs) > 0 && alreadyHolds(current, newAttrs) {
		return nil
	}

	return c.replacePendingJobs(ctx, jobs, current.pending, newAttrs)
}

// replacePendingJobs deletes jobs, merges newAttrs over the pending attributes, and writes the
// merged set. If anything fails after a job was deleted, it re-stages the attributes the jobs
// held so the earlier caller's change is not lost along with the failed one; the returned error
// says whether that worked.
func (c *recoveringRedfishClient) replacePendingJobs(ctx context.Context, jobs []*schemas.Job, pending map[string]any, newAttrs map[string]string) error {
	if len(jobs) > 0 {
		ids := make([]string, 0, len(jobs))
		for _, j := range jobs {
			ids = append(ids, j.ID)
		}
		c.log.Info("deleting pending BIOS config job(s) to merge a new write into them",
			"jobs", ids, "stagedAttributes", sortedKeys(pending), "requestedAttributes", sortedKeys(newAttrs))
	}

	deleted := 0
	for _, j := range jobs {
		resp, err := c.Delete(j.ODataID)
		if err != nil {
			return c.restoreStaged(ctx, deleted, pending, fmt.Errorf("deleting job %s: %w", j.ID, err))
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		deleted++
	}

	merged := make(schemas.SettingsAttributes, len(pending)+len(newAttrs))
	for k, v := range pending {
		merged[k] = v
	}
	for k, v := range newAttrs {
		merged[k] = v
	}

	err := c.ApplyBiosAttributes(ctx, merged)
	if err != nil && deleted == 0 && isPendingSettingsConflict(err) {
		return fmt.Errorf("%w: %w", errNoPendingBIOSConfigJob, err)
	}

	return c.restoreStaged(ctx, deleted, pending, err)
}

// restoreStaged turns cause into the error to return after a failed attempt, first re-staging
// pending if jobs were deleted along the way. It returns nil if cause is nil.
func (c *recoveringRedfishClient) restoreStaged(ctx context.Context, deleted int, pending map[string]any, cause error) error {
	if cause == nil || deleted == 0 {
		return cause
	}

	restore := make(schemas.SettingsAttributes, len(pending))
	for k, v := range pending {
		restore[k] = v
	}
	if err := c.ApplyBiosAttributes(ctx, restore); err != nil {
		return fmt.Errorf("%w; the deleted job(s) held %v, which could not be re-staged either (%v)", cause, sortedKeys(pending), err)
	}

	return fmt.Errorf("%w; the deleted job(s) held %v, which were re-staged", cause, sortedKeys(pending))
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	return keys
}

// dellBIOSConfigJobType is the Oem.Dell.JobType of the job a Bios/Settings write creates. Not
// every iDRAC generation returns an Oem block for jobs in the JobService collection, so jobs are
// also recognized by name: JobService/Jobs names them "ConfigBIOS:<attribute registry path>".
const (
	dellBIOSConfigJobType       = "BIOSConfiguration"
	dellBIOSConfigJobNamePrefix = "ConfigBIOS:"
)

// JobState values iDRAC reports that gofish's standard schemas.JobState set does not define.
// A BIOS config job that has not started shows "Scheduled" or "Starting" depending on the iDRAC
// generation.
const (
	dellScheduledJobState           schemas.JobState = "Scheduled"
	dellSchedulingJobState          schemas.JobState = "Scheduling"
	dellFailedJobState              schemas.JobState = "Failed"
	dellCompletedWithErrorsJobState schemas.JobState = "CompletedWithErrors"
)

// dellJobTerminalStates are the states of a finished job, which cannot be holding the lock.
var dellJobTerminalStates = map[schemas.JobState]bool{
	schemas.CompletedJobState:       true,
	schemas.CancelledJobState:       true,
	schemas.ExceptionJobState:       true,
	dellFailedJobState:              true,
	dellCompletedWithErrorsJobState: true,
}

// dellJobCancelableStates are the states in which a job has not started executing. It is an
// allow-list on purpose: Suspended and Interrupted mean a job started and merely paused, and an
// unrecognized state is refused rather than assumed safe to delete.
var dellJobCancelableStates = map[schemas.JobState]bool{
	schemas.NewJobState:        true,
	schemas.PendingJobState:    true,
	schemas.StartingJobState:   true,
	schemas.ValidatingJobState: true,
	dellScheduledJobState:      true,
	dellSchedulingJobState:     true,
}

// pendingBIOSConfigJobs returns the live BIOS config jobs, which can be deleted. It returns an
// error if any of them has started, or may have started, applying: deleting a running job is far
// more disruptive than the conflict being worked around.
//
// Candidates are picked from the JobService by name or Oem block, which not every iDRAC
// generation fills in. Whether a candidate has started is then read from Dell's own job resource
// (see dellJob), which reports it on every generation.
func (c *recoveringRedfishClient) pendingBIOSConfigJobs(ctx context.Context) ([]*schemas.Job, error) {
	jobs, err := c.Jobs(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing jobs: %w", err)
	}

	var live []*schemas.Job
	for _, j := range jobs {
		if dellJobTerminalStates[j.JobState] {
			continue
		}

		// A missing or malformed Oem block leaves o empty, and the job is then recognized by name.
		var o oem
		_ = json.Unmarshal(j.OEM, &o)

		if o.JobType != dellBIOSConfigJobType && !strings.HasPrefix(j.Name, dellBIOSConfigJobNamePrefix) {
			continue
		}

		if !dellJobCancelableStates[j.JobState] {
			return nil, fmt.Errorf("BIOS config job %s is in state %q, which is not known to be safe to delete", j.ID, j.JobState)
		}

		dj, err := c.dellJob(j.ID)
		if err != nil {
			return nil, fmt.Errorf("checking whether BIOS config job %s has started: %w", j.ID, err)
		}
		if dj.ActualRunningStartTime != "" {
			return nil, fmt.Errorf("BIOS config job %s has already started applying; refusing to delete it", j.ID)
		}

		live = append(live, j)
	}

	return live, nil
}

// alreadyHolds reports whether every attribute in want already has the requested value: the
// staged one if something is staged for it, otherwise the applied one.
func alreadyHolds(current biosAttributes, want map[string]string) bool {
	for name, wantValue := range want {
		have, ok := current.pending[name]
		if !ok {
			have, ok = current.applied[name]
		}
		if !ok || attributeString(have) != wantValue {
			return false
		}
	}

	return true
}

// attributeString spells a staged attribute value the way bmclib's string-typed API would have.
func attributeString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case json.Number:
		return t.String()
	default:
		return fmt.Sprint(v)
	}
}
