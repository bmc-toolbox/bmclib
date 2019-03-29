package httpclient

import (
	"crypto/tls"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"golang.org/x/net/publicsuffix"
)

// Build builds a client session with our default parameters
func Build() (client *http.Client, err error) {
	tr := &http.Transport{
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true,
		Dial: (&net.Dialer{
			Timeout:   120 * time.Second,
			KeepAlive: 120 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   120 * time.Second,
		ResponseHeaderTimeout: 120 * time.Second,
	}

	jar, err := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	if err != nil {
		return client, err
	}

	client = &http.Client{
		Timeout:   time.Second * 120,
		Transport: tr,
		Jar:       jar,
	}

	return client, err
}

// StandardizeProcessorName makes the processor name standard across vendors
func StandardizeProcessorName(name string) string {
	return strings.ToLower(strings.TrimSuffix(strings.TrimSpace(strings.Split(name, "@")[0]), " 0"))
}
