package lenovo

import (
	"context"
	"testing"

	"github.com/bmc-toolbox/bmclib/v2/constants"
)

// Requirement: Tasks are listable.
func TestGetTasks(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	tasks, err := c.GetTasks(context.Background())
	if err != nil {
		t.Fatalf("GetTasks: %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("got %d tasks, want 1", len(tasks))
	}
	task := tasks[0]
	if task.ID != "1" || task.Name != "Task 1" || task.TaskState != "Completed" || task.TaskStatus != "OK" ||
		task.State != constants.Complete || task.PercentComplete == nil || *task.PercentComplete != 100 ||
		len(task.Messages) != 1 || task.Messages[0] != "Device firmware updated successfully." {
		t.Errorf("unexpected task: %+v", task)
	}
}

// Requirement: Jobs are listable.
func TestGetJobs(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	jobs, err := c.GetJobs(context.Background())
	if err != nil {
		t.Fatalf("GetJobs: %v", err)
	}
	if len(jobs) != 1 {
		t.Fatalf("got %d jobs, want 1", len(jobs))
	}
	job := jobs[0]
	if job.ID != "Restart" || job.Name != "Scheduled Restart" || job.JobState != "Scheduled" ||
		job.PercentComplete == nil || *job.PercentComplete != 0 || job.StartTime != "2024-06-01T02:00:00Z" {
		t.Errorf("unexpected job: %+v", job)
	}
}
