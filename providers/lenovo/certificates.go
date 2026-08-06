package lenovo

import (
	"context"
	"net/http"
	"net/url"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

const (
	certificateServiceURI   = "/redfish/v1/CertificateService"
	certificateLocationsURI = certificateServiceURI + "/CertificateLocations"
	generateCSRURI          = certificateServiceURI + "/Actions/CertificateService.GenerateCSR"
	replaceCertificateURI   = certificateServiceURI + "/Actions/CertificateService.ReplaceCertificate"
	// defaultCertificateCollection is the XCC BMC HTTPS certificate collection.
	defaultCertificateCollection = "/redfish/v1/Managers/1/NetworkProtocol/HTTPS/Certificates"
)

// compile-time assertion that the provider implements the interface.
var _ bmc.CertificateManager = (*Conn)(nil)

// CertificateLocations returns the @odata.id locations of installed certificates.
//
// Implements bmc.CertificateManager.
func (c *Conn) CertificateLocations(ctx context.Context) ([]string, error) {
	var doc struct {
		Links struct {
			Certificates []odataID `json:"Certificates"`
		} `json:"Links"`
	}
	if err := c.getJSON(certificateLocationsURI, &doc); err != nil {
		return nil, err
	}

	out := make([]string, 0, len(doc.Links.Certificates))
	for _, cert := range doc.Links.Certificates {
		out = append(out, cert.ODataID)
	}

	return out, nil
}

// Certificate returns a certificate's properties by its @odata.id location.
//
// Implements bmc.CertificateManager.
func (c *Conn) Certificate(ctx context.Context, location string) (bmc.CertificateInfo, error) {
	var doc struct {
		ID      string `json:"Id"`
		Subject struct {
			CommonName string `json:"CommonName"`
		} `json:"Subject"`
		Issuer struct {
			CommonName string `json:"CommonName"`
		} `json:"Issuer"`
		ValidNotBefore string `json:"ValidNotBefore"`
		ValidNotAfter  string `json:"ValidNotAfter"`
	}
	if err := c.getJSON(location, &doc); err != nil {
		return bmc.CertificateInfo{}, err
	}

	return bmc.CertificateInfo{
		ID:                doc.ID,
		SubjectCommonName: doc.Subject.CommonName,
		IssuerCommonName:  doc.Issuer.CommonName,
		ValidNotBefore:    doc.ValidNotBefore,
		ValidNotAfter:     doc.ValidNotAfter,
	}, nil
}

// GenerateCSR generates a CSR via CertificateService.GenerateCSR and returns the
// PEM CSR string.
//
// Implements bmc.CertificateManager.
func (c *Conn) GenerateCSR(ctx context.Context, req bmc.CSRRequest) (string, error) {
	collection := req.CertificateCollection
	if collection == "" {
		collection = defaultCertificateCollection
	}

	payload := map[string]any{
		"CommonName":            req.CommonName,
		"CertificateCollection": map[string]string{"@odata.id": collection},
	}
	addIfSet(payload, "Country", req.Country)
	addIfSet(payload, "State", req.State)
	addIfSet(payload, "City", req.City)
	addIfSet(payload, "Organization", req.Organization)
	addIfSet(payload, "OrganizationalUnit", req.OrganizationalUnit)
	addIfSet(payload, "Email", req.Email)

	resp, err := c.redfishwrapper.PostWithHeaders(ctx, generateCSRURI, payload, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", parseRedfishError(resp)
	}

	var out struct {
		CSRString string `json:"CSRString"`
	}
	if err := decodeJSONBody(resp, &out); err != nil {
		return "", err
	}

	return out.CSRString, nil
}

// ReplaceCertificate replaces the certificate at targetURI with the given PEM.
//
// Implements bmc.CertificateManager.
func (c *Conn) ReplaceCertificate(ctx context.Context, certificatePEM, targetURI string) error {
	payload := map[string]any{
		"CertificateString": certificatePEM,
		"CertificateType":   "PEM",
		"CertificateUri":    map[string]string{"@odata.id": targetURI},
	}

	return checkResponse(c.redfishwrapper.PostWithHeaders(ctx, replaceCertificateURI, payload, nil))
}

// RekeyCertificate generates a new key pair and CSR for the certificate at
// certURI via Certificate.Rekey.
//
// Implements bmc.CertificateManager.
func (c *Conn) RekeyCertificate(ctx context.Context, certURI string) error {
	target, err := url.JoinPath(certURI, "Actions/Certificate.Rekey")
	if err != nil {
		return err
	}
	return checkResponse(c.redfishwrapper.PostWithHeaders(ctx, target, map[string]any{}, nil))
}

// RenewCertificate generates a CSR using the existing key for the certificate at
// certURI via Certificate.Renew.
//
// Implements bmc.CertificateManager.
func (c *Conn) RenewCertificate(ctx context.Context, certURI string) error {
	target, err := url.JoinPath(certURI, "Actions/Certificate.Renew")
	if err != nil {
		return err
	}
	return checkResponse(c.redfishwrapper.PostWithHeaders(ctx, target, map[string]any{}, nil))
}

// addIfSet adds k=v to m when v is non-empty.
func addIfSet(m map[string]any, k, v string) {
	if v != "" {
		m[k] = v
	}
}
