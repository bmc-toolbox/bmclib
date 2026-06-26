package bmc

import "context"

// CSRRequest describes a certificate-signing-request to generate.
type CSRRequest struct {
	// CommonName is the fully-qualified domain name of the entity (required).
	CommonName string
	// Country, State, City, Organization, OrganizationalUnit and Email are the
	// optional distinguished-name fields.
	Country            string
	State              string
	City               string
	Organization       string
	OrganizationalUnit string
	Email              string
	// CertificateCollection is the target certificate collection @odata.id. When
	// empty the provider uses its default (the BMC HTTPS certificate collection).
	CertificateCollection string
}

// CertificateInfo describes an installed certificate.
type CertificateInfo struct {
	// ID is the Redfish Certificate Id.
	ID string
	// SubjectCommonName is the certificate subject common name.
	SubjectCommonName string
	// IssuerCommonName is the certificate issuer common name.
	IssuerCommonName string
	// ValidNotBefore / ValidNotAfter are the validity bounds.
	ValidNotBefore string
	ValidNotAfter  string
}

// CertificateManager is implemented by providers that can manage BMC
// certificates.
type CertificateManager interface {
	// CertificateLocations returns the @odata.id locations of installed
	// certificates.
	CertificateLocations(ctx context.Context) ([]string, error)
	// Certificate returns a certificate's properties by its @odata.id location.
	Certificate(ctx context.Context, location string) (CertificateInfo, error)
	// GenerateCSR generates a certificate-signing-request and returns the PEM CSR.
	GenerateCSR(ctx context.Context, req CSRRequest) (csr string, err error)
	// ReplaceCertificate replaces the certificate at targetURI with the given PEM
	// certificate.
	ReplaceCertificate(ctx context.Context, certificatePEM, targetURI string) error
	// RekeyCertificate generates a new key pair and CSR for the certificate at
	// certURI.
	RekeyCertificate(ctx context.Context, certURI string) error
	// RenewCertificate generates a CSR using the existing key for the certificate
	// at certURI.
	RenewCertificate(ctx context.Context, certURI string) error
}
