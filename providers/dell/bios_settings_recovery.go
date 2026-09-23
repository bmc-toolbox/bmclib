package dell

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/stmcginnis/gofish/schemas"

	"github.com/bmc-toolbox/bmclib/v2/internal/redfishwrapper"
)

// dellPendingSettingsMessageRegistry and dellPendingSettingsMessageKey identify the message
// iDRAC returns when a Bios/Settings PATCH is rejected because an earlier PATCH already sealed a
// pending, uncommitted BIOS config job. iDRAC allows only one such job at a time, regardless of
// which attributes are involved or which Redfish resource issued the earlier PATCH (confirmed
// live: even a direct PATCH to a ComputerSystem's SecureBoot resource seals this same job type,
// on a PowerEdge R6715, iDRAC firmware 1.5.3) - so two independent BIOS-affecting calls before
// the next reset fail on the second one, even though both would happily apply together at that
// reset.
//
// The message ID's middle segment is the IDRAC message registry's own schema version, which
// varies by iDRAC firmware generation, not by the error itself - confirmed live with the exact
// same conflict reported as both "IDRAC.2.14.SYS011" (PowerEdge R6715, iDRAC firmware 1.5.3) and
// "IDRAC.2.9.SYS011" (PowerEdge R760xd2). Matching the full message ID as one fixed string
// silently disables recovery on any iDRAC generation not on the exact registry version this was
// first observed against, with no error - the caller just gets the raw SYS011 conflict back.
const (
	dellPendingSettingsMessageRegistry = "IDRAC"
	dellPendingSettingsMessageKey      = "SYS011"
)

// isPendingSettingsConflict reports whether err is iDRAC's rejection of a Bios/Settings PATCH
// because a previous, uncommitted config job already exists.
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

// isDellPendingSettingsMessageID reports whether messageID is some registry-version build of
// dellPendingSettingsMessageRegistry's dellPendingSettingsMessageKey (e.g. "IDRAC.2.14.SYS011",
// "IDRAC.2.9.SYS011"), ignoring the registry schema version segment in the middle.
func isDellPendingSettingsMessageID(messageID string) bool {
	parts := strings.Split(messageID, ".")
	return len(parts) >= 2 &&
		parts[0] == dellPendingSettingsMessageRegistry &&
		parts[len(parts)-1] == dellPendingSettingsMessageKey
}

// recoveringRedfishClient is a redfishwrapper.Client whose SetBiosConfiguration recovers from
// iDRAC's one-pending-job-at-a-time limit instead of failing outright. It deliberately shadows
// the embedded client's SetBiosConfiguration so that every BIOS Setup attribute write in this
// package gets that recovery without having to opt in: a call reading
// c.redfishwrapper.SetBiosConfiguration(...) anywhere in providers/dell resolves here, and
// there is no same-named unrecovered alternative sitting next to it for a new caller to reach
// for by mistake.
//
// Everything else on redfishwrapper.Client is promoted unchanged - only BIOS Setup attribute
// writes need Dell's conflict handling.
//
// Go has no virtual dispatch through embedding, so this only covers calls made through the
// wrapper: naming the embedded field explicitly (c.redfishwrapper.Client.SetBiosConfiguration)
// still reaches the unrecovered version, and redfishwrapper's own internals would too if they
// ever called it (they don't today). TestFeatureSetterInheritsConflictRecovery guards the
// property this type exists for.
type recoveringRedfishClient struct {
	*redfishwrapper.Client
}

// SetBiosConfiguration stages BIOS Setup attribute changes, recovering from iDRAC's
// one-pending-job-at-a-time limit (see recoverFromPendingSettingsConflict) rather than failing
// on the conflict. It shadows redfishwrapper.Client.SetBiosConfiguration, deferring to it for
// the write itself and only adding the recovery.
func (c *recoveringRedfishClient) SetBiosConfiguration(ctx context.Context, attrs map[string]string) error {
	err := c.Client.SetBiosConfiguration(ctx, attrs)
	if err != nil && isPendingSettingsConflict(err) {
		return c.recoverFromPendingSettingsConflict(ctx, attrs)
	}
	return err
}

// pendingSettingsRedirect holds the standard Redfish @Redfish.Settings block that points a
// resource's settings updates at a separate "pending" resource.
type pendingSettingsRedirect struct {
	RedfishSettings struct {
		SettingsObject struct {
			ODataID string `json:"@odata.id"`
		} `json:"SettingsObject"`
	} `json:"@Redfish.Settings"`
}

// pendingSettings is the subset of a pending settings resource (e.g. .../Bios/Settings) that
// recoverFromPendingSettingsConflict needs: the attributes already staged there.
type pendingSettings struct {
	Attributes map[string]any `json:"Attributes"`
}

// dellBIOSConfigJobType is the Oem.Dell.JobType value iDRAC reports for the BIOS Setup config
// job that Bios/Settings PATCHes create - used to find the specific job blocking a pending
// settings conflict among all jobs the JobService is tracking, without assuming it's the only
// non-completed one (a firmware install, for example, might also be in flight).
//
// Not every iDRAC generation populates Oem.Dell on a Job resource returned from the standard
// JobService/Jobs collection, though - confirmed live on a PowerEdge R760xd2 that the pending
// job has no Oem property at all, just Id/Name/JobState. dellBIOSConfigJobNamePrefix is the
// fallback signal for that case: Dell names every BIOS Setup config job "ConfigBIOS:<attribute
// registry path>" (e.g. "ConfigBIOS:BIOS.Setup.1-1") regardless of iDRAC generation, so
// cancelPendingBIOSConfigJob accepts either signal rather than requiring the Oem one.
const (
	dellBIOSConfigJobType       = "BIOSConfiguration"
	dellBIOSConfigJobNamePrefix = "ConfigBIOS:"
)

// dellJobCancelableStates is the allow-list of schemas.JobState values cancelPendingBIOSConfigJob
// will delete a job in: states that, per the Redfish spec, mean the operation has not yet begun
// executing. Deliberately an allow-list, not a deny-list of just "Running" - Suspended and
// Interrupted both mean a job already started executing and merely paused (the Redfish spec's
// own wording: "expected to restart"), Continue means the same thing mid-resume, and Stopping/
// UserIntervention/Service aren't "not yet started" either. oem.Dell.ActualRunningStartTime,
// the other signal cancelPendingBIOSConfigJob checks, is Dell-specific and empty whenever
// Oem.Dell is absent (see isBIOSConfigJob) - relying on a single JobState value to cover for it
// would miss any of those. A job in a state this allow-list doesn't recognize - including a
// future JobState value neither this list nor Dell's Oem extension anticipates - is refused
// rather than assumed safe.
var dellJobCancelableStates = map[schemas.JobState]bool{
	schemas.NewJobState:        true,
	schemas.PendingJobState:    true,
	schemas.StartingJobState:   true,
	schemas.ValidatingJobState: true,
}

// dellJobOem is the Oem.Dell subset of a Job resource that cancelPendingBIOSConfigJob needs.
type dellJobOem struct {
	Dell struct {
		JobType string `json:"JobType"`
		// ActualRunningStartTime is set once iDRAC actually starts applying the job (during
		// POST on the next reset), as opposed to merely holding it pending. Empty/absent means
		// the job is still just staged and safe to cancel.
		ActualRunningStartTime string `json:"ActualRunningStartTime"`
	} `json:"Dell"`
}

// recoverFromPendingSettingsConflict merges newAttrs into whatever BIOS attributes are already
// staged in the pending settings resource, cancels the job blocking further changes, and
// resubmits the merged superset - so a sequence of BIOS-affecting calls collapses into a single
// pending job and a single reboot, instead of every call after the first failing with
// isPendingSettingsConflict.
//
// Already-pending values are carried over as decoded from JSON (bool/float64/string, whatever
// iDRAC reports), not coerced to string: Dell's BIOS attribute registry is strict about
// attribute types, and resubmitting e.g. "42" for an integer attribute risks iDRAC rejecting the
// retry too. Only newAttrs, built from bmclib's string-typed public API, are ever string-valued.
func (c *recoveringRedfishClient) recoverFromPendingSettingsConflict(ctx context.Context, newAttrs map[string]string) error {
	biosURL, err := c.SystemsBIOSOdataID(ctx)
	if err != nil {
		return fmt.Errorf("finding Bios resource to locate pending settings: %w", err)
	}

	resp, err := c.Get(biosURL)
	if err != nil {
		return fmt.Errorf("reading Bios resource to find pending settings: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var redirect pendingSettingsRedirect
	if err := json.NewDecoder(resp.Body).Decode(&redirect); err != nil {
		return fmt.Errorf("decoding Bios resource: %w", err)
	}

	settingsURL := redirect.RedfishSettings.SettingsObject.ODataID
	if settingsURL == "" {
		return errors.New("iDRAC has no @Redfish.Settings.SettingsObject on its Bios resource; cannot recover from a pending settings conflict")
	}

	pendingResp, err := c.Get(settingsURL)
	if err != nil {
		return fmt.Errorf("reading pending BIOS settings: %w", err)
	}
	defer func() { _ = pendingResp.Body.Close() }()

	var pending pendingSettings
	if err := json.NewDecoder(pendingResp.Body).Decode(&pending); err != nil {
		return fmt.Errorf("decoding pending BIOS settings: %w", err)
	}

	if err := c.cancelPendingBIOSConfigJob(ctx); err != nil {
		return fmt.Errorf("canceling pending BIOS config job: %w", err)
	}

	merged := make(schemas.SettingsAttributes, len(pending.Attributes)+len(newAttrs))
	for k, v := range pending.Attributes {
		merged[k] = v
	}
	for k, v := range newAttrs {
		merged[k] = v
	}

	// A single retry, not a loop: if this also conflicts (e.g. a concurrent operator staged
	// something else in the interim), that's a genuine race to surface, not something to keep
	// retrying blindly.
	return c.ApplyBiosAttributes(ctx, merged)
}

// cancelPendingBIOSConfigJob finds the BIOS Setup config job iDRAC has sealed from an earlier
// Bios/Settings PATCH and cancels it via a standard Redfish DELETE, so a merged superset of
// attributes can be resubmitted into a fresh job.
//
// iDRAC's own DellManager.ClearPending OEM action - advertised right alongside the pending
// settings this function reads - is not a safe substitute: confirmed live, it is itself rejected
// with the exact same pending-settings conflict it would be used to resolve. Deleting the job
// directly is the mechanism iDRAC's own conflict message points at ("delete the configuration
// jobs before attempting more set attribute operations"), and is what resolved this same
// conflict manually, repeatedly, while developing this fix.
func (c *recoveringRedfishClient) cancelPendingBIOSConfigJob(ctx context.Context) error {
	jobs, err := c.Jobs(ctx)
	if err != nil {
		return fmt.Errorf("listing jobs: %w", err)
	}

	for _, j := range jobs {
		switch j.JobState {
		case schemas.CompletedJobState, schemas.CancelledJobState, schemas.ExceptionJobState:
			// Already finished, one way or another - not a live blocker. Keep scanning: a stale
			// job matching isBIOSConfigJob below must not stop the search for the actual one.
			continue
		}

		// Best-effort: a malformed or absent Oem.Dell block just leaves oem at its zero value,
		// so isBIOSConfigJob falls through to the Name-based signal instead of skipping the job
		// outright.
		var oem dellJobOem
		_ = json.Unmarshal(j.OEM, &oem)

		isBIOSConfigJob := oem.Dell.JobType == dellBIOSConfigJobType ||
			strings.HasPrefix(j.Name, dellBIOSConfigJobNamePrefix)
		if !isBIOSConfigJob {
			continue
		}

		if !dellJobCancelableStates[j.JobState] {
			return fmt.Errorf("BIOS config job %s is in state %q, not safe to cancel", j.ID, j.JobState)
		}

		// oem.Dell.ActualRunningStartTime is Dell's own signal for "actually applying", checked
		// in addition to dellJobCancelableStates above (not instead of it): it's only ever
		// populated when Oem.Dell is present at all (see isBIOSConfigJob above), so on a job
		// identified only via its Name it's always empty and could never fire alone.
		if oem.Dell.ActualRunningStartTime != "" {
			return fmt.Errorf("BIOS config job %s has already started applying; refusing to cancel it", j.ID)
		}

		delResp, err := c.Delete(j.ODataID)
		if err != nil {
			return fmt.Errorf("deleting job %s: %w", j.ID, err)
		}
		if delResp != nil {
			_ = delResp.Body.Close()
		}

		return nil
	}

	return errors.New("no pending BIOS config job found to cancel")
}
