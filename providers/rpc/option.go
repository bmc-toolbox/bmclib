package rpc

import (
	"crypto/tls"
	"crypto/x509"
	"net/http"
	"sort"
)

const (
	// SHA256 is the SHA256 algorithm.
	SHA256 Algorithm = "sha256"
	// SHA256Short is the short version of the SHA256 algorithm.
	SHA256Short Algorithm = "256"
	// SHA512 is the SHA512 algorithm.
	SHA512 Algorithm = "sha512"
	// SHA512Short is the short version of the SHA512 algorithm.
	SHA512Short Algorithm = "512"
)

// SetBaseSignatureHeader sets the header name that should contain the signature(s). Example: X-BMCLIB-Signature
func (c *Config) SetBaseSignatureHeader(header string) {
	c.sig.HeaderName = header
}

// SetIncludedPayloadHeaders are headers whose values will be included in the signature payload. Example: X-BMCLIB-Timestamp
func (c *Config) SetIncludedPayloadHeaders(headers []string) {
	c.sig.IncludedPayloadHeaders = append(headers, c.timestampHeader)
}

// IncludeAlgoHeader determines whether to append the algorithm to the signature header or not.
// Example: X-BMCLIB-Signature becomes X-BMCLIB-Signature-256
// When set to false, a header will be added for each algorithm. Example: X-BMCLIB-Signature-256 and X-BMCLIB-Signature-512
func (c *Config) DisableIncludeAlgoHeader() {
	c.sig.AppendAlgoToHeader = false
}

func (c *Config) SetIncludeAlgoPrefix(include bool) {
	c.sig.HMAC.PrefixSig = include
}

// remove an element at index i from a slice of strings.
func remove(s []string, i int) []string {
	s[i] = s[len(s)-1]
	s[len(s)-1] = ""
	return s[:len(s)-1]
}

// SetTimestampHeader sets the header name that should contain the timestamp. Example: X-BMCLIB-Timestamp
func (c *Config) SetTimestampHeader(header string) {
	// update c.IncludedPayloadHeaders with timestamp header
	// remove old timestamp header from c.IncludedPayloadHeaders
	c.sig.IncludedPayloadHeaders = append(c.sig.IncludedPayloadHeaders, header)
	sort.Strings(c.sig.IncludedPayloadHeaders)
	idx := sort.SearchStrings(c.sig.IncludedPayloadHeaders, timestampHeader)
	c.sig.IncludedPayloadHeaders = remove(c.sig.IncludedPayloadHeaders, idx)
	c.timestampHeader = header
}

// addSecrets adds secrets to the Config.
func (c *Config) addSecrets(s Secrets) {
	for algo, secrets := range s {
		switch algo {
		case SHA256, SHA256Short:
			c.sig.HMAC.AddSecretSHA256(secrets...)
		case SHA512, SHA512Short:
			c.sig.HMAC.AddSecretSHA512(secrets...)
		}
	}
}

func (c *Config) AddTLSCert(cert []byte) {
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(cert)
	tp := &http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:    caCertPool,
			MinVersion: tls.VersionTLS12,
		},
	}
	c.httpClient = &http.Client{Transport: tp}
}
