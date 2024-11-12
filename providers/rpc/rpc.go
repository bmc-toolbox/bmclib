package rpc

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"hash"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
	"github.com/metal-toolbox/bmclib/internal/httpclient"
	"github.com/metal-toolbox/bmclib/providers"
)

const (
	// ProviderName for the RPC implementation.
	ProviderName = "rpc"
	// ProviderProtocol for the rpc implementation.
	ProviderProtocol = "http"

	// defaults
	timestampHeader      = "X-BMCLIB-Timestamp"
	signatureHeader      = "X-BMCLIB-Signature"
	contentType          = "application/json"
	maxContentLenAllowed = 512 << (10 * 1) // 512KB

	// SHA256 is the SHA256 algorithm.
	SHA256 Algorithm = "sha256"
	// SHA256Short is the short version of the SHA256 algorithm.
	SHA256Short Algorithm = "256"
	// SHA512 is the SHA512 algorithm.
	SHA512 Algorithm = "sha512"
	// SHA512Short is the short version of the SHA512 algorithm.
	SHA512Short Algorithm = "512"
)

// Features implemented by the RPC provider.
var Features = registrar.Features{
	providers.FeaturePowerSet,
	providers.FeaturePowerState,
	providers.FeatureBootDeviceSet,
}

// Algorithm is the type for HMAC algorithms.
type Algorithm string

// Secrets hold per algorithm slice secrets.
// These secrets will be used to create HMAC signatures.
type Secrets map[Algorithm][]string

// Signatures hold per algorithm slice of signatures.
type Signatures map[Algorithm][]string

// Provider defines the configuration for sending rpc notifications.
type Provider struct {
	// ConsumerURL is the URL where an rpc consumer/listener is running
	// and to which we will send and receive all notifications.
	ConsumerURL string
	// Host is the BMC ip address or hostname or identifier.
	Host string
	// HTTPClient is the http client used for all HTTP calls.
	HTTPClient *http.Client
	// Logger is the logger to use for logging.
	Logger logr.Logger
	// LogNotificationsDisabled determines whether responses from rpc consumer/listeners will be logged or not.
	LogNotificationsDisabled bool
	// Opts are the options for the rpc provider.
	Opts Opts

	// listenerURL is the URL of the rpc consumer/listener.
	listenerURL *url.URL
}

type Opts struct {
	// Request is the options used to create the rpc HTTP request.
	Request RequestOpts
	// Signature is the options used for adding an HMAC signature to an HTTP request.
	Signature SignatureOpts
	// HMAC is the options used to create a HMAC signature.
	HMAC HMACOpts
	// Experimental options.
	Experimental Experimental
}

type RequestOpts struct {
	// HTTPContentType is the content type to use for the rpc request notification.
	HTTPContentType string
	// HTTPMethod is the HTTP method to use for the rpc request notification.
	HTTPMethod string
	// StaticHeaders are predefined headers that will be added to every request.
	StaticHeaders http.Header
	// TimestampFormat is the time format for the timestamp header.
	TimestampFormat string
	// TimestampHeader is the header name that should contain the timestamp. Example: X-BMCLIB-Timestamp
	TimestampHeader string
}

type SignatureOpts struct {
	// HeaderName is the header name that should contain the signature(s). Example: X-BMCLIB-Signature
	HeaderName string
	// AppendAlgoToHeaderDisabled decides whether to append the algorithm to the signature header or not.
	// Example: X-BMCLIB-Signature becomes X-BMCLIB-Signature-256
	// When set to true, a header will be added for each algorithm. Example: X-BMCLIB-Signature-256 and X-BMCLIB-Signature-512
	AppendAlgoToHeaderDisabled bool
	// IncludedPayloadHeaders are headers whose values will be included in the signature payload. Example: given these headers in a request:
	// X-My-Header=123,X-Another=456, and IncludedPayloadHeaders := []string{"X-Another"}, the value of "X-Another" will be included in the signature payload.
	// All headers will be deduplicated.
	IncludedPayloadHeaders []string
}

type HMACOpts struct {
	// Hashes is a map of algorithms to a slice of hash.Hash (these are the hashed secrets). The slice is used to support multiple secrets.
	Hashes map[Algorithm][]hash.Hash
	// PrefixSigDisabled determines whether the algorithm will be prefixed to the signature. Example: sha256=abc123
	PrefixSigDisabled bool
	// Secrets are a map of algorithms to secrets used for signing.
	Secrets Secrets
}

type Experimental struct {
	// CustomRequestPayload must be in json.
	CustomRequestPayload []byte
	// DotPath is the path to where the bmclib RequestPayload{} will be embedded. For example: object.data.body
	DotPath string
}

// New returns a new Config containing all the defaults for the rpc provider.
func New(consumerURL string, host string, secrets Secrets) *Provider {
	// defaults
	c := &Provider{
		Host:        host,
		ConsumerURL: consumerURL,
		HTTPClient:  httpclient.Build(),
		Logger:      logr.Discard(),
		Opts: Opts{
			Request: RequestOpts{
				HTTPContentType: contentType,
				HTTPMethod:      http.MethodPost,
				TimestampFormat: time.RFC3339,
				TimestampHeader: timestampHeader,
			},
			Signature: SignatureOpts{
				HeaderName:             signatureHeader,
				IncludedPayloadHeaders: []string{},
			},
			HMAC: HMACOpts{
				Hashes:  map[Algorithm][]hash.Hash{},
				Secrets: secrets,
			},
			Experimental: Experimental{},
		},
	}

	if len(secrets) > 0 {
		c.Opts.HMAC.Hashes = CreateHashes(secrets)
	}

	return c
}

// Name returns the name of this rpc provider.
// Implements bmc.Provider interface
func (p *Provider) Name() string {
	return ProviderName
}

// Open a connection to the rpc consumer.
// For the rpc provider, Open means validating the Config and
// that communication with the rpc consumer can be established.
func (p *Provider) Open(ctx context.Context) error {
	// 1. validate consumerURL is a properly formatted URL.
	// 2. validate that we can communicate with the rpc consumer.
	u, err := url.Parse(p.ConsumerURL)
	if err != nil {
		return err
	}
	p.listenerURL = u

	if _, err = p.process(ctx, RequestPayload{
		ID:     time.Now().UnixNano(),
		Host:   p.Host,
		Method: PingMethod,
	}); err != nil {
		return err
	}

	return nil
}

// Close a connection to the rpc consumer.
func (p *Provider) Close(_ context.Context) (err error) {
	return nil
}

// BootDeviceSet sends a next boot device rpc notification.
func (p *Provider) BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	rp := RequestPayload{
		ID:     time.Now().UnixNano(),
		Host:   p.Host,
		Method: BootDeviceMethod,
		Params: BootDeviceParams{
			Device:     bootDevice,
			Persistent: setPersistent,
			EFIBoot:    efiBoot,
		},
	}
	resp, err := p.process(ctx, rp)
	if err != nil {
		return false, err
	}
	if resp.Error != nil && resp.Error.Code != 0 {
		return false, fmt.Errorf("error from rpc consumer: %v", resp.Error)
	}

	return true, nil
}

// PowerSet sets the power state of a BMC machine.
func (p *Provider) PowerSet(ctx context.Context, state string) (ok bool, err error) {
	switch strings.ToLower(state) {
	case "on", "off", "cycle":
		rp := RequestPayload{
			ID:     time.Now().UnixNano(),
			Host:   p.Host,
			Method: PowerSetMethod,
			Params: PowerSetParams{
				State: strings.ToLower(state),
			},
		}
		resp, err := p.process(ctx, rp)
		if err != nil {
			return ok, err
		}
		if resp.Error != nil && resp.Error.Code != 0 {
			return ok, fmt.Errorf("error from rpc consumer: %v", resp.Error)
		}

		return true, nil
	}

	return false, errors.New("requested power state is not supported")
}

// PowerStateGet gets the power state of a BMC machine.
func (p *Provider) PowerStateGet(ctx context.Context) (state string, err error) {
	rp := RequestPayload{
		ID:     time.Now().UnixNano(),
		Host:   p.Host,
		Method: PowerGetMethod,
	}
	resp, err := p.process(ctx, rp)
	if err != nil {
		return "", err
	}
	if resp.Error != nil && resp.Error.Code != 0 {
		return "", fmt.Errorf("error from rpc consumer: %v", resp.Error)
	}

	s, ok := resp.Result.(string)
	if !ok {
		return "", fmt.Errorf("expected result equal to type string, got: %T", resp.Result)
	}

	return s, nil
}

// process is the main function for the roundtrip of rpc calls to the ConsumerURL.
func (p *Provider) process(ctx context.Context, rp RequestPayload) (ResponsePayload, error) {
	// 1. create the HTTP request.
	// 2. create the signature payload.
	// 3. sign the signature payload.
	// 4. add signatures to the request as headers.
	// 5. request/response round trip.
	// 6. handle the response.
	req, err := p.createRequest(ctx, rp)
	if err != nil {
		return ResponsePayload{}, err
	}

	// create the signature payload
	reqBuf := new(bytes.Buffer)
	reqBody, err := req.GetBody()
	if err != nil {
		return ResponsePayload{}, fmt.Errorf("failed to get request body: %w", err)
	}
	if _, err := io.Copy(reqBuf, reqBody); err != nil {
		return ResponsePayload{}, fmt.Errorf("failed to read request body: %w", err)
	}

	headersForSig := http.Header{}
	for _, h := range p.Opts.Signature.IncludedPayloadHeaders {
		if val := req.Header.Get(h); val != "" {
			headersForSig.Add(h, val)
		}
	}
	sigPay := createSignaturePayload(reqBuf.Bytes(), headersForSig)

	// sign the signature payload
	sigs, err := sign(sigPay, p.Opts.HMAC.Hashes, p.Opts.HMAC.PrefixSigDisabled)
	if err != nil {
		return ResponsePayload{}, err
	}

	// add signatures to the request as headers.
	for algo, keys := range sigs {
		if len(sigs) > 0 {
			h := p.Opts.Signature.HeaderName
			if !p.Opts.Signature.AppendAlgoToHeaderDisabled {
				h = fmt.Sprintf("%s-%s", h, algo.ToShort())
			}
			req.Header.Add(h, strings.Join(keys, ","))
		}
	}

	// request/response round trip.
	kvs := requestKVS(req.Method, req.URL.String(), req.Header, reqBuf)
	kvs = append(kvs, []interface{}{"host", p.Host, "method", rp.Method, "consumerURL", p.ConsumerURL}...)
	if rp.Params != nil {
		kvs = append(kvs, []interface{}{"params", rp.Params}...)
	}

	resp, err := p.HTTPClient.Do(req)
	if err != nil {
		p.Logger.Error(err, "failed to send rpc notification", kvs...)
		return ResponsePayload{}, err
	}
	defer resp.Body.Close()

	// handle the response
	if resp.ContentLength > maxContentLenAllowed || resp.ContentLength < 0 {
		return ResponsePayload{}, fmt.Errorf("response body is too large: %d bytes, max allowed: %d bytes", resp.ContentLength, maxContentLenAllowed)
	}
	respBuf := new(bytes.Buffer)
	if _, err := io.CopyN(respBuf, resp.Body, resp.ContentLength); err != nil {
		return ResponsePayload{}, fmt.Errorf("failed to read response body: %w", err)
	}
	respPayload, err := p.handleResponse(resp.StatusCode, resp.Header, respBuf, kvs)
	if err != nil {
		return ResponsePayload{}, err
	}

	return respPayload, nil
}

// Transformer implements the mergo interfaces for merging custom types.
func (p *Provider) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
	switch typ {
	case reflect.TypeOf(logr.Logger{}):
		return func(dst, src reflect.Value) error {
			if dst.CanSet() {
				isZero := dst.MethodByName("GetSink")
				result := isZero.Call(nil)
				if result[0].IsNil() {
					dst.Set(src)
				}
			}
			return nil
		}
	}
	return nil
}
