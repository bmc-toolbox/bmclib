package lenovo

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/bmc-toolbox/bmclib/v2/constants"
)

// withDeadline returns a context with a deadline, required by the wrapper's
// firmware upload (it derives the HTTP client timeout from the context).
func withDeadline(t *testing.T) (context.Context, context.CancelFunc) {
	t.Helper()
	return context.WithTimeout(context.Background(), 30*time.Second)
}

// Requirement: Claim and release the update service + multipart push returns a
// task.
func TestFirmwareInstallClaimsAndPushes(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)
	ctx, cancel := withDeadline(t)
	defer cancel()

	taskID, err := c.FirmwareInstall(ctx, "BMC-Backup", "", false, strings.NewReader("firmware-bytes"))
	if err != nil {
		t.Fatalf("FirmwareInstall: %v", err)
	}
	if taskID != "1" {
		t.Errorf("taskID = %q, want %q", taskID, "1")
	}
	if !ts.didClaimBusy() {
		t.Error("expected the update service to be claimed (HttpPushUriTargetsBusy=true)")
	}
	if !ts.didMultipartPush() {
		t.Error("expected a multipart push to /mfwupdate")
	}
}

// Requirement: Service is busy.
func TestFirmwareInstallServiceBusy(t *testing.T) {
	ts := newTestServer(t, testServerOpts{updateServiceFixture: "updateservice.busy.json"})
	c := ts.openedClient(t)
	ctx, cancel := withDeadline(t)
	defer cancel()

	_, err := c.FirmwareInstall(ctx, "BMC-Backup", "", false, strings.NewReader("fw"))
	if err == nil {
		t.Fatal("expected FirmwareInstall to fail when the update service is busy")
	}
	if ts.didMultipartPush() || ts.didRawPush() {
		t.Error("did not expect a push when the service is busy")
	}
}

// Requirement: Multipart and raw push paths — fallback to raw push.
func TestFirmwareInstallRawFallback(t *testing.T) {
	ts := newTestServer(t, testServerOpts{updateServiceFixture: "updateservice.rawonly.json"})
	c := ts.openedClient(t)
	ctx, cancel := withDeadline(t)
	defer cancel()

	if _, err := c.FirmwareInstall(ctx, "", "", false, strings.NewReader("fw")); err != nil {
		t.Fatalf("FirmwareInstall (raw): %v", err)
	}
	if !ts.didRawPush() {
		t.Error("expected a raw push to /fwupdate when multipart is unavailable")
	}
	if ts.didMultipartPush() {
		t.Error("did not expect a multipart push when MultipartHttpPushUri is absent")
	}
}

// Requirement: Task polling without TaskMonitor GET + completed task maps to a
// terminal state + service released after completion.
func TestFirmwareTaskStatus(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	state, status, err := c.FirmwareTaskStatus(context.Background(), constants.FirmwareInstallStepInstallStatus, "BMC-Backup", "1", "")
	if err != nil {
		t.Fatalf("FirmwareTaskStatus: %v", err)
	}
	if state != constants.Complete {
		t.Errorf("state = %q, want %q", state, constants.Complete)
	}
	if status == "" {
		t.Error("expected a non-empty status string")
	}
	if !ts.didReleaseBusy() {
		t.Error("expected the update service to be released after a terminal task")
	}
}

// Requirement: FirmwareInstallStatus reports the install status.
func TestFirmwareInstallStatus(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	status, err := c.FirmwareInstallStatus(context.Background(), "", "BMC-Backup", "1")
	if err != nil {
		t.Fatalf("FirmwareInstallStatus: %v", err)
	}
	if status != string(constants.Complete) {
		t.Errorf("status = %q, want %q", status, constants.Complete)
	}
}

// Requirement: Firmware install steps.
func TestFirmwareInstallSteps(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	steps, err := c.FirmwareInstallSteps(context.Background(), "BMC-Backup")
	if err != nil {
		t.Fatalf("FirmwareInstallSteps: %v", err)
	}
	if len(steps) == 0 {
		t.Fatal("expected at least one install step")
	}
	if steps[0] != constants.FirmwareInstallStepUploadInitiateInstall {
		t.Errorf("first step = %q, want %q", steps[0], constants.FirmwareInstallStepUploadInitiateInstall)
	}
}

// Requirement: two-phase upload then start update.
func TestFirmwareUploadThenInstallUploaded(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)
	ctx, cancel := withDeadline(t)
	defer cancel()

	f, cleanup, err := tempFileFromReader(strings.NewReader("fw"))
	if err != nil {
		t.Fatalf("temp file: %v", err)
	}
	defer cleanup()

	if _, err := c.FirmwareUpload(ctx, "BMC-Backup", f); err != nil {
		t.Fatalf("FirmwareUpload: %v", err)
	}
	if !ts.didMultipartPush() {
		t.Error("expected the upload to push to /mfwupdate")
	}

	taskID, err := c.FirmwareInstallUploaded(ctx, "BMC-Backup", "1")
	if err != nil {
		t.Fatalf("FirmwareInstallUploaded: %v", err)
	}
	if taskID != "1" {
		t.Errorf("install taskID = %q, want %q", taskID, "1")
	}
	if !ts.didStartUpdate() {
		t.Error("expected FirmwareInstallUploaded to POST UpdateService.StartUpdate")
	}
}

// Requirement: Simple update issues SimpleUpdate.
func TestSimpleUpdate(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	taskID, err := c.SimpleUpdate(context.Background(), "https://images.example.com/fw.uxz", "HTTPS")
	if err != nil {
		t.Fatalf("SimpleUpdate: %v", err)
	}
	if taskID != "1" {
		t.Errorf("taskID = %q, want %q", taskID, "1")
	}
	if !ts.didSimpleUpdate() {
		t.Error("expected a POST to UpdateService.SimpleUpdate")
	}
}
