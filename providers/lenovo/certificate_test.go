package lenovo

import (
	"context"
	"testing"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// Requirement: Manager HTTPS certificates are readable.
func TestGetManagerHTTPSCertificates(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	certificates, err := c.GetManagerHTTPSCertificates(context.Background())
	if err != nil {
		t.Fatalf("GetManagerHTTPSCertificates: %v", err)
	}
	if len(certificates) != 1 {
		t.Fatalf("got %d certificates, want 1", len(certificates))
	}
	cert := certificates[0]
	if cert.ID != "1" || cert.SubjectCommonName != "XCC-7Z60-SN" || cert.IssuerCommonName != "XCC-7Z60-SN" ||
		cert.ValidNotBefore != "2023-01-01T00:00:00Z" || cert.ValidNotAfter != "2033-01-01T00:00:00Z" {
		t.Errorf("unexpected certificate: %+v", cert)
	}
}

// Requirement: A CSR can be generated for the manager HTTPS certificate.
func TestGenerateManagerHTTPSCSR(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	csr, err := c.GenerateManagerHTTPSCSR(context.Background(), &bmc.CSRRequest{
		CommonName:   "xcc-sr650.example.com",
		Organization: "Example",
		Country:      "DE",
	})
	if err != nil {
		t.Fatalf("GenerateManagerHTTPSCSR: %v", err)
	}
	if csr == "" {
		t.Fatal("GenerateManagerHTTPSCSR returned an empty CSR")
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()
	if !ts.csrGenerated {
		t.Fatal("CertificateService.GenerateCSR was not posted")
	}
	collection, _ := ts.csrPayload["CertificateCollection"].(map[string]any)
	if collection["@odata.id"] != "/redfish/v1/Managers/1/NetworkProtocol/HTTPS/Certificates" {
		t.Errorf("CertificateCollection = %v, want the manager HTTPS collection link", ts.csrPayload["CertificateCollection"])
	}
	if ts.csrPayload["CommonName"] != "xcc-sr650.example.com" || ts.csrPayload["Organization"] != "Example" || ts.csrPayload["Country"] != "DE" {
		t.Errorf("unexpected subject in payload: %v", ts.csrPayload)
	}
	if _, ok := ts.csrPayload["City"]; ok {
		t.Errorf("empty subject field City must be omitted: %v", ts.csrPayload)
	}
}

// Requirement: The manager HTTPS certificate can be replaced.
func TestReplaceManagerHTTPSCertificate(t *testing.T) {
	ts := newTestServer(t, testServerOpts{})
	c := ts.openedClient(t)

	const pem = "-----BEGIN CERTIFICATE-----\nMIIE\n-----END CERTIFICATE-----\n"
	if err := c.ReplaceManagerHTTPSCertificate(context.Background(), pem); err != nil {
		t.Fatalf("ReplaceManagerHTTPSCertificate: %v", err)
	}

	ts.mu.Lock()
	defer ts.mu.Unlock()
	if !ts.certReplaced {
		t.Fatal("CertificateService.ReplaceCertificate was not posted")
	}
	target, _ := ts.replacePayload["CertificateUri"].(map[string]any)
	if target["@odata.id"] != "/redfish/v1/Managers/1/NetworkProtocol/HTTPS/Certificates/1" {
		t.Errorf("CertificateUri = %v, want the installed HTTPS certificate link", ts.replacePayload["CertificateUri"])
	}
	if ts.replacePayload["CertificateString"] != pem || ts.replacePayload["CertificateType"] != "PEM" {
		t.Errorf("unexpected replace payload: %v", ts.replacePayload)
	}
}
