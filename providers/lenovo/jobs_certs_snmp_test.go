package lenovo

import (
	"context"
	"strings"
	"testing"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// Requirement: Read job service.
func TestJobService(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	info, err := c.JobService(context.Background())
	if err != nil {
		t.Fatalf("JobService: %v", err)
	}
	if !info.ServiceEnabled {
		t.Error("expected the job service to be enabled")
	}
}

// Requirement: Read a job.
func TestJob(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	job, err := c.Job(context.Background(), "Restart")
	if err != nil {
		t.Fatalf("Job: %v", err)
	}
	if job.ID != "Restart" || job.JobState != "Scheduled" {
		t.Errorf("unexpected job: %+v", job)
	}
}

// Requirement: Update a job schedule.
func TestJobUpdateSchedule(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if err := c.JobUpdateSchedule(context.Background(), "Restart", map[string]any{"RecurrenceInterval": "P1D"}); err != nil {
		t.Fatalf("JobUpdateSchedule: %v", err)
	}
	if !ts.didUpdateJobSchedule() {
		t.Error("expected a PATCH of the job schedule")
	}
}

// Requirement: Unknown job errors.
func TestJobUnknown(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	if _, err := c.Job(context.Background(), "DoesNotExist"); err == nil {
		t.Fatal("expected an error for an unknown job")
	}
}

// Requirement: Read certificate locations.
func TestCertificateLocations(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	locs, err := c.CertificateLocations(context.Background())
	if err != nil {
		t.Fatalf("CertificateLocations: %v", err)
	}
	if len(locs) != 1 || !strings.Contains(locs[0], "Certificates/1") {
		t.Fatalf("unexpected locations: %v", locs)
	}
}

// Requirement: Read a certificate.
func TestCertificate(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	cert, err := c.Certificate(context.Background(), "/redfish/v1/Managers/1/NetworkProtocol/HTTPS/Certificates/1")
	if err != nil {
		t.Fatalf("Certificate: %v", err)
	}
	if cert.SubjectCommonName != "XCC-7Z60-SN" || cert.ValidNotAfter == "" {
		t.Errorf("unexpected certificate: %+v", cert)
	}
}

// Requirement: Generate a CSR.
func TestGenerateCSR(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	csr, err := c.GenerateCSR(context.Background(), bmc.CSRRequest{CommonName: "XCC-test", Country: "US"})
	if err != nil {
		t.Fatalf("GenerateCSR: %v", err)
	}
	if !strings.Contains(csr, "BEGIN CERTIFICATE REQUEST") {
		t.Errorf("unexpected CSR: %q", csr)
	}
	if !ts.didGenerateCSR() {
		t.Error("expected a GenerateCSR action")
	}
}

// Requirement: Replace a certificate.
func TestReplaceCertificate(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	err := c.ReplaceCertificate(context.Background(), "-----BEGIN CERTIFICATE-----\nx\n-----END CERTIFICATE-----",
		"/redfish/v1/Managers/1/NetworkProtocol/HTTPS/Certificates/1")
	if err != nil {
		t.Fatalf("ReplaceCertificate: %v", err)
	}
	if !ts.didReplaceCertificate() {
		t.Error("expected a ReplaceCertificate action")
	}
}

// Requirement: Rekey and renew a certificate.
func TestRekeyRenewCertificate(t *testing.T) {
	certURI := "/redfish/v1/Managers/1/NetworkProtocol/HTTPS/Certificates/1"

	t.Run("rekey", func(t *testing.T) {
		ts := newTestServer(t, testServerOpts{})
		c := ts.openedClient(t)
		if err := c.RekeyCertificate(context.Background(), certURI); err != nil {
			t.Fatalf("RekeyCertificate: %v", err)
		}
		if !ts.didRekeyCertificate() {
			t.Error("expected a Certificate.Rekey action")
		}
	})

	t.Run("renew", func(t *testing.T) {
		ts := newTestServer(t, testServerOpts{})
		c := ts.openedClient(t)
		if err := c.RenewCertificate(context.Background(), certURI); err != nil {
			t.Fatalf("RenewCertificate: %v", err)
		}
		if !ts.didRenewCertificate() {
			t.Error("expected a Certificate.Renew action")
		}
	})
}

// Requirement: Read SNMP config.
func TestSNMP(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	cfg, err := c.SNMP(context.Background())
	if err != nil {
		t.Fatalf("SNMP: %v", err)
	}
	// CommunityNames lives at the top level of the XCC OEM SNMP resource (not
	// inside SNMPTraps); assert it is read from there.
	if cfg.TrapPort != 162 || len(cfg.CommunityNames) != 1 || cfg.CommunityNames[0] != "public" {
		t.Errorf("unexpected SNMP config: %+v", cfg)
	}
}

// Requirement: Configure the alert filter + enable v1/v3 traps.
func TestSNMPConfigure(t *testing.T) {
	t.Run("alert filter", func(t *testing.T) {
		ts := newTestServer(t, testServerOpts{})
		c := ts.openedClient(t)
		if err := c.SetSNMPAlertFilter(context.Background(), map[string]any{"SNMPv1TrapEnabled": true}); err != nil {
			t.Fatalf("SetSNMPAlertFilter: %v", err)
		}
		if !ts.didPatchSNMP() {
			t.Error("expected a PATCH of the SNMP resource")
		}
	})

	t.Run("enable v1 trap", func(t *testing.T) {
		ts := newTestServer(t, testServerOpts{})
		c := ts.openedClient(t)
		if err := c.EnableSNMPv1Trap(context.Background(), true); err != nil {
			t.Fatalf("EnableSNMPv1Trap: %v", err)
		}
		if !ts.didPatchSNMP() {
			t.Error("expected a PATCH enabling the SNMPv1 trap")
		}
	})

	t.Run("enable v3 trap", func(t *testing.T) {
		ts := newTestServer(t, testServerOpts{})
		c := ts.openedClient(t)
		if err := c.EnableSNMPv3Trap(context.Background(), true); err != nil {
			t.Fatalf("EnableSNMPv3Trap: %v", err)
		}
		if !ts.didPatchSNMP() {
			t.Error("expected a PATCH enabling the SNMPv3 trap")
		}
	})
}
