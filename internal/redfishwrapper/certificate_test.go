package redfishwrapper

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

const (
	dellNetworkProtocolPath    = "/redfish/v1/Managers/iDRAC.Embedded.1/NetworkProtocol"
	dellHTTPSCertificatesPath  = dellNetworkProtocolPath + "/HTTPS/Certificates"
	dellHTTPSCertificatePath   = dellHTTPSCertificatesPath + "/SecurityCertificate.1"
	dellCertificateServicePath = "/redfish/v1/CertificateService"
	dellGenerateCSRPath        = dellCertificateServicePath + "/Actions/CertificateService.GenerateCSR"
	dellReplaceCertificatePath = dellCertificateServicePath + "/Actions/CertificateService.ReplaceCertificate"
)

// newDellManagerClient serves the Dell service root, manager collection and manager with the given
// NetworkProtocol handler, plus the HTTPS certificate collection and its member, and returns an
// opened client.
func newDellManagerClient(t *testing.T, mux *http.ServeMux, networkProtocol http.HandlerFunc) *Client {
	t.Helper()

	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "dell/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Managers", endpointFunc(t, "dell/managers.json"))
	mux.HandleFunc("/redfish/v1/Managers/iDRAC.Embedded.1", endpointFunc(t, "dell/manager.idrac.embedded.1.json"))
	mux.HandleFunc(dellNetworkProtocolPath, networkProtocol)
	mux.HandleFunc(dellHTTPSCertificatesPath, endpointFunc(t, "dell/https_certificates.json"))
	mux.HandleFunc(dellHTTPSCertificatePath, endpointFunc(t, "dell/https_certificate_1.json"))
	mux.HandleFunc(dellCertificateServicePath, endpointFunc(t, "dell/certificateservice.json"))

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))

	ctx := context.Background()
	require.NoError(t, client.Open(ctx))
	t.Cleanup(func() { _ = client.Close(ctx) })

	return client
}

// actionRecorder records the JSON body of one POSTed action and answers with status and body.
func actionRecorder(t *testing.T, status int, body string, got *map[string]any) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(got))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}
}

func TestGetManagerHTTPSCertificates(t *testing.T) {
	mux := http.NewServeMux()
	client := newDellManagerClient(t, mux, endpointFunc(t, "dell/networkprotocol.json"))

	certificates, err := client.GetManagerHTTPSCertificates(context.Background())
	require.NoError(t, err)
	require.Len(t, certificates, 1)

	assert.Equal(t, bmc.Certificate{
		ID:                "SecurityCertificate.1",
		SubjectCommonName: "idrac-7XK9Q33.example.com",
		IssuerCommonName:  "Example Root CA",
		SerialNumber:      "7332B15134C39E3217",
		ValidNotBefore:    "2021-04-16T15:40:38+00:00",
		ValidNotAfter:     "2031-04-14T15:40:38+00:00",
		KeyUsage:          []string{"KeyEncipherment", "DigitalSignature"},
	}, certificates[0])
}

func TestGetManagerHTTPSCertificatesNotFound(t *testing.T) {
	mux := http.NewServeMux()
	// A NetworkProtocol without the HTTPS.Certificates link (firmware predating ManagerNetworkProtocol v1.4).
	client := newDellManagerClient(t, mux, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"@odata.id": "` + dellNetworkProtocolPath + `", "Id": "NetworkProtocol", "HTTPS": {"Port": 443, "ProtocolEnabled": true}}`))
	})

	_, err := client.GetManagerHTTPSCertificates(context.Background())
	assert.ErrorIs(t, err, bmclibErrs.ErrManagerHTTPSCertificatesNotFound)

	_, err = client.GenerateManagerHTTPSCSR(context.Background(), &bmc.CSRRequest{CommonName: "bmc.example.com"})
	assert.ErrorIs(t, err, bmclibErrs.ErrManagerHTTPSCertificatesNotFound)

	err = client.ReplaceManagerHTTPSCertificate(context.Background(), "-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----")
	assert.ErrorIs(t, err, bmclibErrs.ErrManagerHTTPSCertificatesNotFound)
}

func TestGetManagerHTTPSCertificatesError(t *testing.T) {
	mux := http.NewServeMux()
	client := newDellManagerClient(t, mux, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": {"message": "internal error"}}`))
	})

	_, err := client.GetManagerHTTPSCertificates(context.Background())
	require.Error(t, err)
	assert.NotErrorIs(t, err, bmclibErrs.ErrManagerHTTPSCertificatesNotFound, "a failing NetworkProtocol read must be reported, not skipped")
	assert.ErrorContains(t, err, "iDRAC.Embedded.1")
}

func TestGenerateManagerHTTPSCSR(t *testing.T) {
	const csr = "-----BEGIN CERTIFICATE REQUEST-----\nMIICFDCC\n-----END CERTIFICATE REQUEST-----"
	var posted map[string]any

	mux := http.NewServeMux()
	body, err := json.Marshal(map[string]any{"CSRString": csr, "CertificateCollection": map[string]string{"@odata.id": dellHTTPSCertificatesPath}})
	require.NoError(t, err)
	mux.HandleFunc(dellGenerateCSRPath, actionRecorder(t, http.StatusOK, string(body), &posted))
	client := newDellManagerClient(t, mux, endpointFunc(t, "dell/networkprotocol.json"))

	got, err := client.GenerateManagerHTTPSCSR(context.Background(), &bmc.CSRRequest{
		CommonName:         "idrac-7XK9Q33.example.com",
		Organization:       "Example Corp",
		OrganizationalUnit: "Datacenter",
		City:               "Round Rock",
		State:              "Texas",
		Country:            "US",
		AlternativeNames:   []string{"idrac-7XK9Q33", "10.0.0.5"},
		KeyBitLength:       2048,
	})
	require.NoError(t, err)
	assert.Equal(t, csr, got)

	assert.Equal(t, map[string]any{
		"CertificateCollection": map[string]any{"@odata.id": dellHTTPSCertificatesPath},
		"CommonName":            "idrac-7XK9Q33.example.com",
		"Organization":          "Example Corp",
		"OrganizationalUnit":    "Datacenter",
		"City":                  "Round Rock",
		"State":                 "Texas",
		"Country":               "US",
		"AlternativeNames":      []any{"idrac-7XK9Q33", "10.0.0.5"},
		"KeyBitLength":          float64(2048),
	}, posted, "empty subject fields must be omitted and links sent as objects")
}

func TestGenerateManagerHTTPSCSRRejected(t *testing.T) {
	var posted map[string]any

	mux := http.NewServeMux()
	mux.HandleFunc(dellGenerateCSRPath, actionRecorder(t, http.StatusBadRequest, `{"error": {"message": "The property Country is a required property"}}`, &posted))
	client := newDellManagerClient(t, mux, endpointFunc(t, "dell/networkprotocol.json"))

	got, err := client.GenerateManagerHTTPSCSR(context.Background(), &bmc.CSRRequest{CommonName: "idrac-7XK9Q33.example.com"})
	assert.Empty(t, got)
	assert.ErrorContains(t, err, "Country is a required property")
	assert.Equal(t, map[string]any{
		"CertificateCollection": map[string]any{"@odata.id": dellHTTPSCertificatesPath},
		"CommonName":            "idrac-7XK9Q33.example.com",
	}, posted)
}

func TestGenerateManagerHTTPSCSRNoCSRString(t *testing.T) {
	var posted map[string]any

	mux := http.NewServeMux()
	mux.HandleFunc(dellGenerateCSRPath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "/redfish/v1/TaskService/Tasks/JID_1")
		actionRecorder(t, http.StatusAccepted, ``, &posted)(w, r)
	})
	client := newDellManagerClient(t, mux, endpointFunc(t, "dell/networkprotocol.json"))

	_, err := client.GenerateManagerHTTPSCSR(context.Background(), &bmc.CSRRequest{CommonName: "idrac-7XK9Q33.example.com"})
	assert.ErrorContains(t, err, "no CSRString")
	assert.ErrorContains(t, err, "JID_1")
}

func TestReplaceManagerHTTPSCertificate(t *testing.T) {
	const pem = "-----BEGIN CERTIFICATE-----\nMIIEEzCCAvugAwIBAgIUczKxUTTDnjIX\n-----END CERTIFICATE-----\n"
	var posted map[string]any

	mux := http.NewServeMux()
	mux.HandleFunc(dellReplaceCertificatePath, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", dellHTTPSCertificatePath)
		actionRecorder(t, http.StatusAccepted, ``, &posted)(w, r)
	})
	client := newDellManagerClient(t, mux, endpointFunc(t, "dell/networkprotocol.json"))

	require.NoError(t, client.ReplaceManagerHTTPSCertificate(context.Background(), pem))
	assert.Equal(t, map[string]any{
		"CertificateString": pem,
		"CertificateType":   "PEM",
		"CertificateUri":    map[string]any{"@odata.id": dellHTTPSCertificatePath},
	}, posted)
}

func TestReplaceManagerHTTPSCertificateRejected(t *testing.T) {
	var posted map[string]any

	mux := http.NewServeMux()
	mux.HandleFunc(dellReplaceCertificatePath, actionRecorder(t, http.StatusBadRequest, `{"error": {"message": "Unable to parse the certificate"}}`, &posted))
	client := newDellManagerClient(t, mux, endpointFunc(t, "dell/networkprotocol.json"))

	err := client.ReplaceManagerHTTPSCertificate(context.Background(), "not a certificate")
	assert.ErrorContains(t, err, "Unable to parse the certificate")
}

func TestCertificateServiceNotFound(t *testing.T) {
	mux := http.NewServeMux()
	// A service root without CertificateService.
	mux.HandleFunc("/redfish/v1/", func(w http.ResponseWriter, _ *http.Request) {
		var root map[string]any
		require.NoError(t, json.Unmarshal(mustReadFile(t, "dell/serviceroot.json"), &root))
		delete(root, "CertificateService")
		_ = json.NewEncoder(w).Encode(root)
	})
	mux.HandleFunc("/redfish/v1/Managers", endpointFunc(t, "dell/managers.json"))
	mux.HandleFunc("/redfish/v1/Managers/iDRAC.Embedded.1", endpointFunc(t, "dell/manager.idrac.embedded.1.json"))
	mux.HandleFunc(dellNetworkProtocolPath, endpointFunc(t, "dell/networkprotocol.json"))
	mux.HandleFunc(dellHTTPSCertificatesPath, endpointFunc(t, "dell/https_certificates.json"))
	mux.HandleFunc(dellHTTPSCertificatePath, endpointFunc(t, "dell/https_certificate_1.json"))

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)
	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)
	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))
	require.NoError(t, client.Open(context.Background()))
	t.Cleanup(func() { _ = client.Close(context.Background()) })

	_, err = client.GenerateManagerHTTPSCSR(context.Background(), &bmc.CSRRequest{CommonName: "idrac-7XK9Q33.example.com"})
	assert.ErrorIs(t, err, bmclibErrs.ErrCertificateServiceNotFound)

	err = client.ReplaceManagerHTTPSCertificate(context.Background(), "-----BEGIN CERTIFICATE-----\n-----END CERTIFICATE-----")
	assert.ErrorIs(t, err, bmclibErrs.ErrCertificateServiceNotFound)
}
