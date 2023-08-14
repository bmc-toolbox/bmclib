package bmclib

import (
	"reflect"

	"dario.cat/mergo"
	"github.com/bmc-toolbox/bmclib/v2/providers/rpc"
	"github.com/go-logr/logr"
)

type RPCOpts struct {
	Secrets rpc.Secrets
	// ConsumerURL is the URL where a rpc consumer/listener is running and to which we will send notifications.
	ConsumerURL string
	// BaseSignatureHeader is the header name that should contain the signature(s). Example: X-BMCLIB-Signature
	BaseSignatureHeader string
	// IncludedPayloadHeaders are headers whose values will be included in the signature payload. Example: X-BMCLIB-Timestamp
	IncludedPayloadHeaders []string
	// HTTPContentType is the content type to use for the rpc request notification.
	HTTPContentType string
	// HTTPMethod is the HTTP method to use for the rpc request notification.
	HTTPMethod string
	// TimestampHeader is the header name that should contain the timestamp. Example: X-BMCLIB-Timestamp
	TimestampHeader string

	// includeAlgoHeader determines whether to append the algorithm to the signature header or not.
	// Example: X-BMCLIB-Signature becomes X-BMCLIB-Signature-256
	// When set to false, a header will be added for each algorithm. Example: X-BMCLIB-Signature-256 and X-BMCLIB-Signature-512
	includeAlgoHeader bool
	// includeAlgoPrefix will prepend the algorithm and an equal sign to the signature. Example: sha256=abc123
	includeAlgoPrefix bool
	// logger is the logger to use for logging.
	logger logr.Logger
	// logNotifications determines whether responses from rpc consumer/listeners will be logged or not.
	logNotifications bool
}

func registerRPC(c *Client) {
	driverRPC := rpc.New(c.providerConfig.rpc.ConsumerURL, c.Auth.Host, c.providerConfig.rpc.Secrets)
	c.providerConfig.rpc.logger = c.Logger
	c.providerConfig.rpc.translate(driverRPC)
	c.Registry.Register(rpc.ProviderName, rpc.ProviderProtocol, rpc.Features, nil, driverRPC)
}

func (w *RPCOpts) translate(wc *rpc.Config) {
	if w.BaseSignatureHeader != "" {
		wc.SetBaseSignatureHeader(w.BaseSignatureHeader)
	}
	if len(w.IncludedPayloadHeaders) > 0 {
		wc.SetIncludedPayloadHeaders(w.IncludedPayloadHeaders)
	}
	if !w.includeAlgoHeader {
		wc.DisableIncludeAlgoHeader()
	}
	wc.SetIncludeAlgoPrefix(w.includeAlgoPrefix)
	wc.Logger = w.logger

	wc.LogNotifications = w.logNotifications
	wc.HTTPContentType = w.HTTPContentType
	wc.HTTPMethod = w.HTTPMethod

	if w.TimestampHeader != "" {
		wc.SetTimestampHeader(w.TimestampHeader)
	}
}

// Transformer for merging the netip.IPPort and logr.Logger structs.
func (r *RPCOpts) Transformer(typ reflect.Type) func(dst, src reflect.Value) error {
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

func WithRPCOpts(opts RPCOpts) Option {
	return func(args *Client) {
		// TODO(jacobweinstock): figure out if ignoring the error is ok.
		// Maybe write a test to validate the merge will never error.
		// Add code comment on the behavior.
		mergoOpts := []func(*mergo.Config){
			mergo.WithAppendSlice,
			mergo.WithTransformers(&RPCOpts{}),
		}
		_ = mergo.Merge(&args.providerConfig.rpc, opts, mergoOpts...)
	}
}

func RPCDisableIncludeAlgoHeader() Option {
	return func(args *Client) {
		args.providerConfig.rpc.includeAlgoHeader = false
	}
}

func RPCDisableIncludeAlgoPrefix() Option {
	return func(args *Client) {
		args.providerConfig.rpc.includeAlgoPrefix = false
	}
}

func RPCDisableLogNotifications() Option {
	return func(args *Client) {
		args.providerConfig.rpc.logNotifications = false
	}
}
