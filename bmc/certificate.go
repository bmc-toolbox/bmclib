package bmc

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// Certificate describes a certificate installed on the BMC.
type Certificate struct {
	// ID is the Redfish Certificate resource ID, e.g. "1" or "SecurityCertificate.1".
	ID string
	// SubjectCommonName is the subject common name, usually the BMC host name.
	SubjectCommonName string
	// IssuerCommonName is the issuer common name; equal to the subject for a self-signed certificate.
	IssuerCommonName string
	// SerialNumber is the certificate serial number as reported by the BMC; empty when not reported.
	SerialNumber string
	// ValidNotBefore and ValidNotAfter are the validity bounds as reported by the BMC.
	ValidNotBefore string
	ValidNotAfter  string
	// KeyUsage lists the certificate key usage values, e.g. "KeyEncipherment", "DigitalSignature".
	KeyUsage []string
}

// CSRRequest holds the subject of a certificate signing request.
//
// Only CommonName is required by the Redfish standard; most BMCs also require Organization,
// OrganizationalUnit, City, State and Country. Empty fields are left out of the request and
// the BMC reports the ones it misses.
type CSRRequest struct {
	// CommonName is the subject common name, usually the BMC host name.
	CommonName string
	// Organization, OrganizationalUnit, City, State and Country are the subject distinguished name fields.
	Organization       string
	OrganizationalUnit string
	City               string
	State              string
	Country            string
	// Email is the subject e-mail address.
	Email string
	// AlternativeNames are the subject alternative names (DNS names or IP addresses).
	AlternativeNames []string
	// KeyBitLength is the length of the key the BMC generates for the request; 0 leaves it to the BMC.
	KeyBitLength int
}

// ManagerHTTPSCertificatesGetter provides retrieval of the certificates installed for the
// manager HTTPS endpoint (the BMC web server).
type ManagerHTTPSCertificatesGetter interface {
	GetManagerHTTPSCertificates(ctx context.Context) (certificates []Certificate, err error)
}

// ManagerHTTPSCSRGenerator provides generation of a certificate signing request for the
// manager HTTPS endpoint; the BMC generates and keeps the private key.
type ManagerHTTPSCSRGenerator interface {
	GenerateManagerHTTPSCSR(ctx context.Context, req *CSRRequest) (csrPEM string, err error)
}

// ManagerHTTPSCertificateReplacer provides replacement of the certificate installed for the
// manager HTTPS endpoint with a PEM-encoded certificate.
type ManagerHTTPSCertificateReplacer interface {
	ReplaceManagerHTTPSCertificate(ctx context.Context, certificatePEM string) (err error)
}

type managerHTTPSCertificatesGetterProvider struct {
	name string
	ManagerHTTPSCertificatesGetter
}

type managerHTTPSCSRGeneratorProvider struct {
	name string
	ManagerHTTPSCSRGenerator
}

type managerHTTPSCertificateReplacerProvider struct {
	name string
	ManagerHTTPSCertificateReplacer
}

func getManagerHTTPSCertificates(ctx context.Context, generic []managerHTTPSCertificatesGetterProvider) (certificates []Certificate, metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.ManagerHTTPSCertificatesGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			certificates, vErr := elem.GetManagerHTTPSCertificates(ctx)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return certificates, metadata, nil
		}
	}

	return nil, metadata, multierror.Append(err, errors.New("failure to get manager HTTPS certificates"))
}

func generateManagerHTTPSCSR(ctx context.Context, req *CSRRequest, generic []managerHTTPSCSRGeneratorProvider) (csrPEM string, metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.ManagerHTTPSCSRGenerator == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			csrPEM, vErr := elem.GenerateManagerHTTPSCSR(ctx, req)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return csrPEM, metadata, nil
		}
	}

	return "", metadata, multierror.Append(err, errors.New("failure to generate manager HTTPS CSR"))
}

func replaceManagerHTTPSCertificate(ctx context.Context, certificatePEM string, generic []managerHTTPSCertificateReplacerProvider) (metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.ManagerHTTPSCertificateReplacer == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			vErr := elem.ReplaceManagerHTTPSCertificate(ctx, certificatePEM)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return metadata, nil
		}
	}

	return metadata, multierror.Append(err, errors.New("failure to replace manager HTTPS certificate"))
}

// GetManagerHTTPSCertificatesFromInterfaces identifies implementations of the
// ManagerHTTPSCertificatesGetter interface and passes them to the getManagerHTTPSCertificates() wrapper.
func GetManagerHTTPSCertificatesFromInterfaces(ctx context.Context, generic []interface{}) (certificates []Certificate, metadata Metadata, err error) {
	implementations := make([]managerHTTPSCertificatesGetterProvider, 0)
	for _, elem := range generic {
		temp := managerHTTPSCertificatesGetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case ManagerHTTPSCertificatesGetter:
			temp.ManagerHTTPSCertificatesGetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a ManagerHTTPSCertificatesGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return nil, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				"no ManagerHTTPSCertificatesGetter implementations found",
			),
		)
	}

	return getManagerHTTPSCertificates(ctx, implementations)
}

// GenerateManagerHTTPSCSRFromInterfaces identifies implementations of the
// ManagerHTTPSCSRGenerator interface and passes them to the generateManagerHTTPSCSR() wrapper.
func GenerateManagerHTTPSCSRFromInterfaces(ctx context.Context, req *CSRRequest, generic []interface{}) (csrPEM string, metadata Metadata, err error) {
	implementations := make([]managerHTTPSCSRGeneratorProvider, 0)
	for _, elem := range generic {
		temp := managerHTTPSCSRGeneratorProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case ManagerHTTPSCSRGenerator:
			temp.ManagerHTTPSCSRGenerator = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a ManagerHTTPSCSRGenerator implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return "", metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				"no ManagerHTTPSCSRGenerator implementations found",
			),
		)
	}

	return generateManagerHTTPSCSR(ctx, req, implementations)
}

// ReplaceManagerHTTPSCertificateFromInterfaces identifies implementations of the
// ManagerHTTPSCertificateReplacer interface and passes them to the replaceManagerHTTPSCertificate() wrapper.
func ReplaceManagerHTTPSCertificateFromInterfaces(ctx context.Context, certificatePEM string, generic []interface{}) (metadata Metadata, err error) {
	implementations := make([]managerHTTPSCertificateReplacerProvider, 0)
	for _, elem := range generic {
		temp := managerHTTPSCertificateReplacerProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case ManagerHTTPSCertificateReplacer:
			temp.ManagerHTTPSCertificateReplacer = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a ManagerHTTPSCertificateReplacer implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				"no ManagerHTTPSCertificateReplacer implementations found",
			),
		)
	}

	return replaceManagerHTTPSCertificate(ctx, certificatePEM, implementations)
}
