package redfish

import (
	"encoding/json"
	"io"
	"strconv"
	"strings"

	bmcliberrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/bmc-toolbox/common"
	"github.com/pkg/errors"

	rfcommon "github.com/stmcginnis/gofish/common"
	rf "github.com/stmcginnis/gofish/redfish"
)

// Dell specific redfish methods

var (
	componentSlugDellJobName = map[string]string{
		common.SlugBIOS:              "Firmware Update: BIOS",
		common.SlugBMC:               "Firmware Update: iDRAC with Lifecycle Controller",
		common.SlugNIC:               "Firmware Update: Network",
		common.SlugDrive:             "Firmware Update: Serial ATA",
		common.SlugStorageController: "Firmware Update: SAS RAID",
	}
)

type dellJob struct {
	PercentComplete int
	OdataID         string `json:"@odata.id"`
	StartTime       string
	CompletionTime  string
	ID              string
	Message         string
	Name            string
	JobState        string
	JobType         string
}

type dellJobResponseData struct {
	Members []*dellJob
}

// dellJobID formats and returns taskID as a Dell Job ID
func dellJobID(id string) string {
	if !strings.HasPrefix(id, "JID") {
		return "JID_" + id
	}

	return id
}

func (c *Conn) getDellFirmwareInstallTaskScheduled(slug string) (*rf.Task, error) {
	// get tasks by state
	tasks, err := c.dellJobs("scheduled")
	if err != nil {
		return nil, err
	}

	// filter to match the task Name based on the component slug
	for _, task := range tasks {
		if task.Name == componentSlugDellJobName[strings.ToUpper(slug)] {
			return task, nil
		}
	}

	return nil, nil
}

func (c *Conn) dellPurgeScheduledFirmwareInstallJob(slug string) error {
	// get tasks by state
	tasks, err := c.dellJobs("scheduled")
	if err != nil {
		return err
	}

	// filter to match the task Name based on the component slug
	for _, task := range tasks {
		if task.Name == componentSlugDellJobName[strings.ToUpper(slug)] {
			err = c.dellPurgeJob(task.ID)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Conn) dellPurgeJob(id string) error {
	id = dellJobID(id)

	endpoint := "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/Jobs/" + id

	resp, err := c.conn.Delete(endpoint)
	if err != nil {
		return errors.Wrap(bmcliberrs.ErrTaskPurge, err.Error())
	}

	if resp.StatusCode != 200 {
		return errors.Wrap(bmcliberrs.ErrTaskPurge, "response code: "+resp.Status)
	}

	return nil
}

// dellFirmwareUpdateTaskStatus looks up the Dell Job and returns it as a redfish task object
func (c *Conn) dellJobAsRedfishTask(jobID string) (*rf.Task, error) {
	jobID = dellJobID(jobID)

	tasks, err := c.dellJobs("")
	if err != nil {
		return nil, err
	}

	for _, task := range tasks {
		if task.ID == jobID {
			return task, nil
		}
	}

	return nil, errors.Wrap(bmcliberrs.ErrTaskNotFound, "task with ID not found: "+jobID)
}

// dellJobs returns all dell jobs as redfish task objects
// state: optional
func (c *Conn) dellJobs(state string) ([]*rf.Task, error) {
	endpoint := "/redfish/v1/Managers/iDRAC.Embedded.1/Oem/Dell/Jobs?$expand=*($levels=1)"

	resp, err := c.conn.Get(endpoint)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("dell jobs endpoint returned unexpected status code: " + strconv.Itoa(resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	data := dellJobResponseData{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	tasks := []*rf.Task{}
	for _, job := range data.Members {
		if state != "" && !strings.EqualFold(job.JobState, state) {
			continue
		}

		tasks = append(tasks, &rf.Task{
			Entity: rfcommon.Entity{
				ID:      job.ID,
				ODataID: job.OdataID,
				Name:    job.Name,
				Client:  c.conn,
			},
			Description:     job.Name,
			PercentComplete: job.PercentComplete,
			StartTime:       job.StartTime,
			EndTime:         job.CompletionTime,
			TaskState:       rf.TaskState(job.JobState),
			TaskStatus:      rfcommon.Health(job.Message), // abuse the TaskStatus to include any status message
		})
	}

	return tasks, nil
}
