package bmc

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type mockManagerHTTPSCertificatesGetter struct {
	certificates []Certificate
	err          error
}

func (m *mockManagerHTTPSCertificatesGetter) GetManagerHTTPSCertificates(ctx context.Context) ([]Certificate, error) {
	return m.certificates, m.err
}

func (m *mockManagerHTTPSCertificatesGetter) Name() string {
	return "mock"
}

type mockManagerHTTPSCSRGenerator struct {
	csr string
	req *CSRRequest
	err error
}

func (m *mockManagerHTTPSCSRGenerator) GenerateManagerHTTPSCSR(ctx context.Context, req *CSRRequest) (string, error) {
	m.req = req
	return m.csr, m.err
}

func (m *mockManagerHTTPSCSRGenerator) Name() string {
	return "mock"
}

type mockManagerHTTPSCertificateReplacer struct {
	pem string
	err error
}

func (m *mockManagerHTTPSCertificateReplacer) ReplaceManagerHTTPSCertificate(ctx context.Context, certificatePEM string) error {
	m.pem = certificatePEM
	return m.err
}

func (m *mockManagerHTTPSCertificateReplacer) Name() string {
	return "mock"
}

func TestGetManagerHTTPSCertificatesFromInterfaces(t *testing.T) {
	certificates := []Certificate{{ID: "1", SubjectCommonName: "bmc.example.com", IssuerCommonName: "Example CA"}}

	testCases := []struct {
		name                 string
		generic              []interface{}
		errMsg               string
		expectedCertificates []Certificate
	}{
		{
			name:                 "success",
			generic:              []interface{}{&mockManagerHTTPSCertificatesGetter{certificates: certificates}},
			expectedCertificates: certificates,
		},
		{
			name:    "not an implementation",
			generic: []interface{}{"foo"},
			errMsg:  "no ManagerHTTPSCertificatesGetter implementations found",
		},
		{
			name:    "no implementations",
			generic: []interface{}{},
			errMsg:  "no ManagerHTTPSCertificatesGetter implementations found",
		},
		{
			name:    "error from getter",
			generic: []interface{}{&mockManagerHTTPSCertificatesGetter{err: errors.New("foobar")}},
			errMsg:  "foobar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := GetManagerHTTPSCertificatesFromInterfaces(context.Background(), tt.generic)

			if tt.errMsg == "" {
				assert.NoError(t, err)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}

			assert.Equal(t, tt.expectedCertificates, got)
		})
	}
}

func TestGenerateManagerHTTPSCSRFromInterfaces(t *testing.T) {
	req := &CSRRequest{CommonName: "bmc.example.com", Organization: "Example"}
	csr := "-----BEGIN CERTIFICATE REQUEST-----\nMIIC\n-----END CERTIFICATE REQUEST-----"

	testCases := []struct {
		name        string
		generic     []interface{}
		errMsg      string
		expectedCSR string
	}{
		{
			name:        "success",
			generic:     []interface{}{&mockManagerHTTPSCSRGenerator{csr: csr}},
			expectedCSR: csr,
		},
		{
			name:    "not an implementation",
			generic: []interface{}{"foo"},
			errMsg:  "no ManagerHTTPSCSRGenerator implementations found",
		},
		{
			name:    "no implementations",
			generic: []interface{}{},
			errMsg:  "no ManagerHTTPSCSRGenerator implementations found",
		},
		{
			name:    "error from generator",
			generic: []interface{}{&mockManagerHTTPSCSRGenerator{err: errors.New("foobar")}},
			errMsg:  "foobar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got, metadata, err := GenerateManagerHTTPSCSRFromInterfaces(context.Background(), req, tt.generic)

			if tt.errMsg == "" {
				assert.NoError(t, err)
				assert.Equal(t, "mock", metadata.SuccessfulProvider)
				assert.Same(t, req, tt.generic[0].(*mockManagerHTTPSCSRGenerator).req)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}

			assert.Equal(t, tt.expectedCSR, got)
		})
	}
}

func TestReplaceManagerHTTPSCertificateFromInterfaces(t *testing.T) {
	pem := "-----BEGIN CERTIFICATE-----\nMIIE\n-----END CERTIFICATE-----"

	testCases := []struct {
		name    string
		generic []interface{}
		errMsg  string
	}{
		{
			name:    "success",
			generic: []interface{}{&mockManagerHTTPSCertificateReplacer{}},
		},
		{
			name:    "not an implementation",
			generic: []interface{}{"foo"},
			errMsg:  "no ManagerHTTPSCertificateReplacer implementations found",
		},
		{
			name:    "no implementations",
			generic: []interface{}{},
			errMsg:  "no ManagerHTTPSCertificateReplacer implementations found",
		},
		{
			name:    "error from replacer",
			generic: []interface{}{&mockManagerHTTPSCertificateReplacer{err: errors.New("foobar")}},
			errMsg:  "foobar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := ReplaceManagerHTTPSCertificateFromInterfaces(context.Background(), pem, tt.generic)

			if tt.errMsg == "" {
				assert.NoError(t, err)
				assert.Equal(t, "mock", metadata.SuccessfulProvider)
				assert.Equal(t, pem, tt.generic[0].(*mockManagerHTTPSCertificateReplacer).pem)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}
		})
	}
}
