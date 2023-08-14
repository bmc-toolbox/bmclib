package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bmc-toolbox/bmclib/v2/providers/rpc/internal/hmac"
)

const (
	signatureHeader = "X-BMCLIB-Signature"
)

// Signature contains the configuration for signing HTTP requests.
// It wraps the hmac signing functionality for HTTP requests.
type Signature struct {
	// HeaderName is the header name that should contain the signature(s). Example: X-BMCLIB-Signature
	HeaderName string
	// AppendAlgoToHeader decides whether to append the algorithm to the signature header or not.
	// Example: X-BMCLIB-Signature becomes X-BMCLIB-Signature-256
	// When set to true, a header will be added for each algorithm. Example: X-BMCLIB-Signature-256 and X-BMCLIB-Signature-512
	AppendAlgoToHeader bool
	// IncludedPayloadHeaders are headers whose values will be included in the signature payload. Example: X-BMCLIB-My-Custom-Header
	// All headers will be deduplicated.
	IncludedPayloadHeaders []string
	// HMAC holds and handles signing.
	HMAC *hmac.Conf
}

func NewSignature() Signature {
	return Signature{
		HeaderName:         signatureHeader,
		AppendAlgoToHeader: true,
		HMAC:               hmac.NewHMAC(),
	}
}

// deduplicate returns a new slice with duplicates values removed.
func deduplicate(s []string) []string {
	if len(s) <= 1 {
		return s
	}
	result := []string{}
	seen := make(map[string]struct{})
	for _, val := range s {
		val := strings.ToLower(val)
		if _, ok := seen[val]; !ok {
			result = append(result, val)
			seen[val] = struct{}{}
		}
	}
	return result
}

func (s Signature) AddSignature(req *http.Request) error {
	// get the body and reset it as readers can only be read once.
	body, err := io.ReadAll(req.Body)
	if err != nil {
		return err
	}
	req.Body = io.NopCloser(bytes.NewBuffer(body))
	// add headers to signature payload, no space between values.
	for _, h := range deduplicate(s.IncludedPayloadHeaders) {
		if val := req.Header.Get(h); val != "" {
			body = append(body, []byte(val)...)
		}
	}
	signed, err := s.HMAC.Sign(body)
	if err != nil {
		return err
	}

	if s.AppendAlgoToHeader {
		if len(signed[hmac.SHA256]) > 0 {
			req.Header.Add(fmt.Sprintf("%s-%s", s.HeaderName, hmac.SHA256Short), strings.Join(signed[hmac.SHA256], ","))
		}
		if len(signed[hmac.SHA512]) > 0 {
			req.Header.Add(fmt.Sprintf("%s-%s", s.HeaderName, hmac.SHA512Short), strings.Join(signed[hmac.SHA512], ","))
		}
	} else {
		all := signed[hmac.SHA256]
		all = append(all, signed[hmac.SHA512]...)
		req.Header.Add(s.HeaderName, strings.Join(all, ","))
	}

	return nil
}
