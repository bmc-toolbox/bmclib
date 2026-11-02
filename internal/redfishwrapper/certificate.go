package redfishwrapper

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"path"

	"github.com/stmcginnis/gofish/schemas"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

const (
	generateCSRAction        = "CertificateService.GenerateCSR"
	replaceCertificateAction = "CertificateService.ReplaceCertificate"
)

// odataLink is a Redfish resource reference as sent in action parameters.
type odataLink struct {
	ODataID string `json:"@odata.id"`
}

// GetManagerHTTPSCertificates returns the certificates installed for the HTTPS endpoint of the
// first manager that exposes an HTTPS certificate collection. Errors raised while collecting are
// returned to the caller.
func (c *Client) GetManagerHTTPSCertificates(ctx context.Context) ([]bmc.Certificate, error) {
	_, certs, err := c.managerHTTPSCertificates(ctx)
	if err != nil {
		return nil, err
	}

	out := make([]bmc.Certificate, 0, len(certs))
	for _, cert := range certs {
		if cert == nil {
			continue
		}
		out = append(out, toCertificate(cert))
	}

	return out, nil
}

// GenerateManagerHTTPSCSR asks the CertificateService for a certificate signing request targeted
// at the manager HTTPS certificate collection and returns the PEM-encoded CSR.
func (c *Client) GenerateManagerHTTPSCSR(ctx context.Context, req *bmc.CSRRequest) (string, error) {
	if req == nil {
		return "", errors.New("nil CSR request")
	}

	collection, _, err := c.managerHTTPSCertificateCollection(ctx)
	if err != nil {
		return "", err
	}

	target, err := c.certificateServiceActionTarget(generateCSRAction)
	if err != nil {
		return "", err
	}

	payload := map[string]any{
		"CertificateCollection": odataLink{ODataID: collection},
	}
	for key, value := range map[string]string{
		"CommonName":         req.CommonName,
		"Organization":       req.Organization,
		"OrganizationalUnit": req.OrganizationalUnit,
		"City":               req.City,
		"State":              req.State,
		"Country":            req.Country,
		"Email":              req.Email,
	} {
		if value != "" {
			payload[key] = value
		}
	}
	if len(req.AlternativeNames) > 0 {
		payload["AlternativeNames"] = req.AlternativeNames
	}
	if req.KeyBitLength > 0 {
		payload["KeyBitLength"] = req.KeyBitLength
	}

	// gofish returns an error for any non-2xx response, carrying the BMC's error body.
	resp, err := c.PostWithHeaders(ctx, target, payload, nil)
	if err != nil {
		return "", fmt.Errorf("%s: %w", generateCSRAction, err)
	}
	defer func() { _ = resp.Body.Close() }()

	var result struct {
		CSRString string `json:"CSRString"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || result.CSRString == "" {
		return "", fmt.Errorf("%s: response (status %d) carries no CSRString; task monitor: %q",
			generateCSRAction, resp.StatusCode, resp.Header.Get("Location"))
	}

	return result.CSRString, nil
}

// ReplaceManagerHTTPSCertificate replaces the certificate installed for the manager HTTPS endpoint
// through CertificateService.ReplaceCertificate. Any 2xx reply is taken as success; a task the BMC
// may start to apply the certificate is not awaited.
func (c *Client) ReplaceManagerHTTPSCertificate(ctx context.Context, certificatePEM string) error {
	if certificatePEM == "" {
		return errors.New("empty certificate")
	}

	_, certs, err := c.managerHTTPSCertificates(ctx)
	if err != nil {
		return err
	}
	if len(certs) == 0 || certs[0] == nil {
		return fmt.Errorf("%w: the HTTPS certificate collection is empty", bmclibErrs.ErrManagerHTTPSCertificatesNotFound)
	}

	target, err := c.certificateServiceActionTarget(replaceCertificateAction)
	if err != nil {
		return err
	}

	payload := map[string]any{
		"CertificateString": certificatePEM,
		"CertificateType":   string(schemas.PEMCertificateType),
		"CertificateUri":    odataLink{ODataID: certs[0].ODataID},
	}

	resp, err := c.PostWithHeaders(ctx, target, payload, nil)
	if err != nil {
		return fmt.Errorf("%s: %w", replaceCertificateAction, err)
	}
	_ = resp.Body.Close()

	return nil
}

// managerHTTPSCertificateCollection returns the @odata.id of the HTTPS certificate collection of
// the first manager that links one, together with that manager's NetworkProtocol resource.
//
// gofish keeps the HTTPS.Certificates link private on ManagerNetworkProtocol and does not keep
// the raw document, so the link is read from the resource JSON; the collection members are then
// read through gofish.
func (c *Client) managerHTTPSCertificateCollection(ctx context.Context) (string, *schemas.ManagerNetworkProtocol, error) {
	managers, err := c.Managers(ctx)
	if err != nil {
		return "", nil, err
	}

	for _, m := range managers {
		if m == nil {
			continue
		}

		np, err := m.NetworkProtocol()
		if err != nil {
			if isNotFound(err) {
				continue
			}
			return "", nil, fmt.Errorf("manager %s NetworkProtocol: %w", m.ID, err)
		}
		if np == nil {
			continue
		}

		link, err := c.httpsCertificatesLink(np.ODataID)
		if err != nil {
			return "", nil, fmt.Errorf("manager %s NetworkProtocol: %w", m.ID, err)
		}
		if link == "" {
			continue
		}

		return link, np, nil
	}

	return "", nil, bmclibErrs.ErrManagerHTTPSCertificatesNotFound
}

// managerHTTPSCertificates returns the HTTPS certificate collection @odata.id and its members.
func (c *Client) managerHTTPSCertificates(ctx context.Context) (string, []*schemas.Certificate, error) {
	collection, np, err := c.managerHTTPSCertificateCollection(ctx)
	if err != nil {
		return "", nil, err
	}

	certs, err := schemas.ListReferencedCertificates(np.GetClient(), collection)
	if err != nil {
		return "", nil, fmt.Errorf("HTTPS certificates %s: %w", collection, err)
	}

	return collection, certs, nil
}

// httpsCertificatesLink reads the HTTPS.Certificates link of a ManagerNetworkProtocol resource.
func (c *Client) httpsCertificatesLink(networkProtocolURI string) (string, error) {
	resp, err := c.Get(networkProtocolURI)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d reading %s", resp.StatusCode, networkProtocolURI)
	}

	var doc struct {
		HTTPS struct {
			Certificates odataLink `json:"Certificates"`
		} `json:"HTTPS"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return "", fmt.Errorf("decoding %s: %w", networkProtocolURI, err)
	}

	return doc.HTTPS.Certificates.ODataID, nil
}

// certificateServiceActionTarget returns the target URI of a CertificateService action, read from
// the service's Actions; when the firmware omits the target it is derived from the resource
// @odata.id by the Redfish convention.
//
// The actions are POSTed explicitly rather than through gofish's typed GenerateCSR and
// ReplaceCertificate: their CertificateCollection and CertificateUri parameters are plain strings
// in gofish, while the DMTF CertificateService schema (v1_0 onward) defines them as resource links
// ({"@odata.id": ...}), which is what BMCs validate (confirmed on AMI MegaRAC, Dell iDRAC and
// Lenovo XCC).
func (c *Client) certificateServiceActionTarget(action string) (string, error) {
	service, err := c.ServiceRoot()
	if err != nil {
		return "", err
	}

	cs, err := service.CertificateService()
	if err != nil {
		return "", fmt.Errorf("CertificateService: %w", err)
	}
	if cs == nil {
		return "", bmclibErrs.ErrCertificateServiceNotFound
	}

	resp, err := c.Get(cs.ODataID)
	if err != nil {
		return "", fmt.Errorf("CertificateService: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	var doc struct {
		Actions map[string]struct {
			Target string `json:"target"`
		} `json:"Actions"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return "", fmt.Errorf("decoding CertificateService: %w", err)
	}

	if target := doc.Actions["#"+action].Target; target != "" {
		return target, nil
	}

	return path.Join(cs.ODataID, "Actions", action), nil
}

// toCertificate maps a Redfish Certificate to the bmclib representation.
func toCertificate(cert *schemas.Certificate) bmc.Certificate {
	usage := make([]string, 0, len(cert.KeyUsage))
	for _, u := range cert.KeyUsage {
		usage = append(usage, string(u))
	}
	if len(usage) == 0 {
		usage = nil
	}

	return bmc.Certificate{
		ID:                cert.ID,
		SubjectCommonName: cert.Subject.CommonName,
		IssuerCommonName:  cert.Issuer.CommonName,
		SerialNumber:      cert.SerialNumber,
		ValidNotBefore:    cert.ValidNotBefore,
		ValidNotAfter:     cert.ValidNotAfter,
		KeyUsage:          usage,
	}
}
