package dell

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	common "github.com/metal-toolbox/bmc-common"
	"github.com/metal-toolbox/bmclib/constants"
	bmcliberrs "github.com/metal-toolbox/bmclib/errors"
	rfw "github.com/metal-toolbox/bmclib/internal/redfishwrapper"
	"github.com/pkg/errors"
	"github.com/stmcginnis/gofish/redfish"
)

// bmc client interface implementations methods
func (c *Conn) FirmwareInstallSteps(ctx context.Context, component string) ([]constants.FirmwareInstallStep, error) {
	if err := c.deviceSupported(); err != nil {
		return nil, bmcliberrs.NewErrUnsupportedHardware(err.Error())
	}

	return []constants.FirmwareInstallStep{
		constants.FirmwareInstallStepUploadInitiateInstall,
		constants.FirmwareInstallStepInstallStatus,
	}, nil
}

func (c *Conn) FirmwareInstallUploadAndInitiate(ctx context.Context, component string, file *os.File) (taskID string, err error) {
	if err := c.deviceSupported(); err != nil {
		return "", bmcliberrs.NewErrUnsupportedHardware(err.Error())
	}

	//	// expect atleast 5 minutes left in the deadline to proceed with the upload
	d, _ := ctx.Deadline()
	if time.Until(d) < 10*time.Minute {
		return "", errors.New("remaining context deadline insufficient to perform update: " + time.Until(d).String())
	}

	// list current tasks on BMC
	tasks, err := c.redfishwrapper.Tasks(ctx)
	if err != nil {
		return "", errors.Wrap(err, "error listing bmc redfish tasks")
	}

	// validate a new firmware install task can be queued
	if err := c.checkQueueability(component, tasks); err != nil {
		return "", errors.Wrap(bmcliberrs.ErrFirmwareInstall, err.Error())
	}

	params := &rfw.RedfishUpdateServiceParameters{
		Targets:            []string{},
		OperationApplyTime: constants.OnReset,
		Oem:                []byte(`{}`),
	}

	return c.redfishwrapper.FirmwareUpload(ctx, file, params)
}

// checkQueueability returns an error if an existing firmware task is in progress for the given component
func (c *Conn) checkQueueability(component string, tasks []*redfish.Task) error {
	errTaskActive := errors.New("A firmware job was found active for component: " + component)

	// Redfish on the Idrac names firmware install tasks in this manner.
	taskNameMap := map[string]string{
		common.SlugBIOS:              "Firmware Update: BIOS",
		common.SlugBMC:               "Firmware Update: iDRAC with Lifecycle Controller",
		common.SlugNIC:               "Firmware Update: Network",
		common.SlugDrive:             "Firmware Update: Serial ATA",
		common.SlugStorageController: "Firmware Update: SAS RAID",
	}

	for _, t := range tasks {
		if t.Name == taskNameMap[strings.ToUpper(component)] {
			// taskInfo returned in error if any.
			taskInfo := fmt.Sprintf("id: %s, state: %s, status: %s", t.ID, t.TaskState, t.TaskStatus)

			// convert redfish task state to bmclib state
			convstate := c.redfishwrapper.ConvertTaskState(string(t.TaskState))
			// check if task is active based on converted state
			active, err := c.redfishwrapper.TaskStateActive(convstate)
			if err != nil {
				return errors.Wrap(err, taskInfo)
			}

			if active {
				return errors.Wrap(errTaskActive, taskInfo)
			}
		}
	}

	return nil
}

// FirmwareTaskStatus returns the status of a firmware related task queued on the BMC.
func (c *Conn) FirmwareTaskStatus(ctx context.Context, kind constants.FirmwareInstallStep, component, taskID, installVersion string) (state constants.TaskState, status string, err error) {
	if err := c.deviceSupported(); err != nil {
		return "", "", bmcliberrs.NewErrUnsupportedHardware(err.Error())
	}

	// Dell jobs are turned into Redfish tasks on the idrac
	// once the Redfish task completes successfully, the Redfish task is purged,
	// and the dell Job stays around.
	task, err := c.redfishwrapper.Task(ctx, taskID)
	if err != nil {
		if errors.Is(err, bmcliberrs.ErrTaskNotFound) {
			return c.statusFromJob(taskID)
		}

		return "", "", err
	}

	return c.statusFromTaskOem(taskID, task.Oem)
}

func (c *Conn) statusFromJob(taskID string) (constants.TaskState, string, error) {
	job, err := c.job(taskID)
	if err != nil {
		return "", "", err
	}

	s := strings.ToLower(job.JobState)
	state := c.redfishwrapper.ConvertTaskState(s)

	status := fmt.Sprintf(
		"id: %s, state: %s, status: %s, progress: %d%%",
		taskID,
		job.JobState,
		job.Message,
		job.PercentComplete,
	)

	return state, status, nil
}

func (c *Conn) statusFromTaskOem(taskID string, oem json.RawMessage) (constants.TaskState, string, error) {
	data, err := convFirmwareTaskOem(oem)
	if err != nil {
		return "", "", err
	}

	s := strings.ToLower(data.Dell.JobState)
	state := c.redfishwrapper.ConvertTaskState(s)

	status := fmt.Sprintf(
		"id: %s, state: %s, status: %s, progress: %d%%",
		taskID,
		data.Dell.JobState,
		data.Dell.Message,
		data.Dell.PercentComplete,
	)

	return state, status, nil
}

func (c *Conn) job(jobID string) (*Dell, error) {
	errLookup := errors.New("error querying dell job: " + jobID)

	endpoint := "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/Jobs/" + jobID
	resp, err := c.redfishwrapper.Get(endpoint)
	if err != nil {
		return nil, errors.Wrap(errLookup, err.Error())
	}

	if resp.StatusCode != 200 {
		return nil, errors.Wrap(errLookup, "unexpected status code: "+resp.Status)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(errLookup, err.Error())
	}

	dell := &Dell{}
	err = json.Unmarshal(body, &dell)
	if err != nil {
		return nil, errors.Wrap(errLookup, err.Error())
	}

	return dell, nil
}

type oem struct {
	Dell `json:"Dell"`
}

type Dell struct {
	OdataType         string        `json:"@odata.type"`
	CompletionTime    interface{}   `json:"CompletionTime"`
	Description       string        `json:"Description"`
	EndTime           string        `json:"EndTime"`
	ID                string        `json:"Id"`
	JobState          string        `json:"JobState"`
	JobType           string        `json:"JobType"`
	Message           string        `json:"Message"`
	MessageArgs       []interface{} `json:"MessageArgs"`
	MessageID         string        `json:"MessageId"`
	Name              string        `json:"Name"`
	PercentComplete   int           `json:"PercentComplete"`
	StartTime         string        `json:"StartTime"`
	TargetSettingsURI interface{}   `json:"TargetSettingsURI"`
}

func convFirmwareTaskOem(oemdata json.RawMessage) (oem, error) {
	oem := oem{}

	errTaskOem := errors.New("error in Task Oem data: " + string(oemdata))

	if len(oemdata) == 0 || string(oemdata) == `{}` {
		return oem, errors.Wrap(errTaskOem, "empty oem data")
	}

	if err := json.Unmarshal(oemdata, &oem); err != nil {
		return oem, errors.Wrap(errTaskOem, "failed to unmarshal: "+err.Error())
	}

	if oem.Dell.Description == "" || oem.Dell.JobState == "" {
		return oem, errors.Wrap(errTaskOem, "invalid oem data")
	}

	if oem.Dell.JobType != "FirmwareUpdate" {
		return oem, errors.Wrap(errTaskOem, "unexpected job type: "+oem.Dell.JobType)
	}

	return oem, nil
}
