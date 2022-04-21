package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

// SecureTLS disables InsecureSkipVerify and adds a cert pool to an HTTP client's
// TLS config
func SecureTLS(c *http.Client, rootCAs *x509.CertPool) {
	tp := DefaultTransport()
	if c.Transport != nil {
		if assertedTransport, ok := c.Transport.(*http.Transport); ok {
			tp = assertedTransport
		}
		// otherwise, we overwrite the transport
	}
	tp.TLSClientConfig.InsecureSkipVerify = false
	tp.TLSClientConfig.RootCAs = rootCAs
	c.Transport = tp
}

// DefaultTransport sets an HTTP Transport
func DefaultTransport() *http.Transport {
	return &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
		Dial: (&net.Dialer{
			Timeout:   120 * time.Second,
			KeepAlive: 120 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   120 * time.Second,
		ResponseHeaderTimeout: 120 * time.Second,
	}
}

// SecureTLSOption disables InsecureSkipVerify and adds a cert pool to an HTTP client's
// TLS config
func SecureTLSOption(rootCAs *x509.CertPool) func(*http.Client) {
	return func(c *http.Client) {
		SecureTLS(c, rootCAs)
	}
}

// Build builds a client session with our default parameters
func Build(opts ...func(*http.Client)) (client *http.Client, err error) {
	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return client, err
	}

	client = &http.Client{
		Timeout:   time.Second * 120,
		Transport: DefaultTransport(),
		Jar:       jar,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(client)
		}
	}

	return client, err
}

// StandardizeProcessorName makes the processor name standard across vendors
func StandardizeProcessorName(name string) string {
	return strings.ToLower(strings.TrimSuffix(strings.TrimSpace(strings.Split(name, "@")[0]), " 0"))
}
