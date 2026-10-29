package lenovo

import (
	"context"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// compile-time assertions that the provider implements the interfaces.
var (
	_ bmc.ManagerHTTPSCertificatesGetter  = (*Conn)(nil)
	_ bmc.ManagerHTTPSCSRGenerator        = (*Conn)(nil)
	_ bmc.ManagerHTTPSCertificateReplacer = (*Conn)(nil)
)

// GetManagerHTTPSCertificates returns the certificates installed for the manager HTTPS endpoint.
//
// Implements bmc.ManagerHTTPSCertificatesGetter.
func (c *Conn) GetManagerHTTPSCertificates(ctx context.Context) (certificates []bmc.Certificate, err error) {
	return c.redfishwrapper.GetManagerHTTPSCertificates(ctx)
}

// GenerateManagerHTTPSCSR generates a certificate signing request for the manager HTTPS endpoint.
//
// Implements bmc.ManagerHTTPSCSRGenerator.
func (c *Conn) GenerateManagerHTTPSCSR(ctx context.Context, req *bmc.CSRRequest) (csrPEM string, err error) {
	return c.redfishwrapper.GenerateManagerHTTPSCSR(ctx, req)
}

// ReplaceManagerHTTPSCertificate replaces the certificate installed for the manager HTTPS endpoint.
//
// Implements bmc.ManagerHTTPSCertificateReplacer.
func (c *Conn) ReplaceManagerHTTPSCertificate(ctx context.Context, certificatePEM string) (err error) {
	return c.redfishwrapper.ReplaceManagerHTTPSCertificate(ctx, certificatePEM)
}
