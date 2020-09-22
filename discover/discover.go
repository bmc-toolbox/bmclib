package discover

import (
	"context"
	"os"

	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/logging"

	"github.com/bmc-toolbox/bmclib/providers/dummy/ibmc"
	"github.com/go-logr/logr"
)

const (
	Probe_hpIlo       = "hpilo"
	Probe_idrac8      = "idrac8"
	Probe_idrac9      = "idrac9"
	Probe_supermicrox = "supermicrox"
	Probe_hpC7000     = "hpc7000"
	Probe_m1000e      = "m1000e"
	Probe_quanta      = "quanta"
	Probe_hpCl100     = "hpcl100"
)

// ScanAndConnect will scan the bmc trying to learn the device type and return a working connection.
func ScanAndConnect(host string, username string, password string, options ...Option) (bmcConnection interface{}, err error) {
	opts := &Options{HintCallback: func(_ string) error { return nil }}
	for _, optFn := range options {
		optFn(opts)
	}
	if opts.Log == nil {
		// create a default logger
		opts.Log = logging.DefaultLogger()
	}
	opts.Log.V(1).Info("detecting vendor", "step", "ScanAndConnect", "host", host)

	// return a connection to our dummy device.
	if os.Getenv("BMCLIB_TEST") == "1" {
		opts.Log.V(1).Info("returning connection to dummy ibmc device.", "step", "ScanAndConnect", "host", host)
		bmc, err := ibmc.New(host, username, password)
		return bmc, err
	}

	client, err := httpclient.Build()
	if err != nil {
		return nil, err
	}

	var probe = Probe{client: client, username: username, password: password, host: host}

	var devices = map[string]func(context.Context, logr.Logger) (interface{}, error){
		Probe_hpIlo:       probe.hpIlo,
		Probe_idrac8:      probe.idrac8,
		Probe_idrac9:      probe.idrac9,
		Probe_supermicrox: probe.supermicrox,
		Probe_hpC7000:     probe.hpC7000,
		Probe_m1000e:      probe.m1000e,
		Probe_quanta:      probe.quanta,
		Probe_hpCl100:     probe.hpCl100,
	}

	order := []string{Probe_hpIlo,
		Probe_idrac8,
		Probe_idrac9,
		Probe_supermicrox,
		Probe_hpC7000,
		Probe_m1000e,
		Probe_quanta,
		Probe_hpCl100,
	}

	if opts.Hint != "" {
		swapProbe(order, opts.Hint)
	}

	for _, probeID := range order {
		probeDevice := devices[probeID]

		opts.Log.V(1).Info("probing to identify device", "step", "ScanAndConnect", "host", host)

		bmcConnection, err := probeDevice(opts.Context, opts.Log)

		// if the device didn't match continue to probe
		if err != nil && (err == errors.ErrDeviceNotMatched) {
			continue
		}

		// at this point it could be a connection error or a errors.ErrUnsupportedHardware
		if err != nil {
			return nil, err
		}

		if err := opts.HintCallback(probeID); err != nil {
			return nil, err
		}

		// return a bmcConnection
		return bmcConnection, nil
	}

	return nil, errors.ErrVendorUnknown
}

type Options struct {
	// Hint is a probe ID that hints which probe should be probed first.
	Hint string

	// HintCallBack is a function that is called back with a probe ID that might be used
	// for the next ScanAndConnect attempt.  The callback is called only on successful scan.
	// If your code persists the hint as "best effort", always return a nil error.  Callback is
	// synchronous.
	HintCallback func(string) error
	Log          logr.Logger
	Context      context.Context
}

// Option is part of the functional options pattern, see the `With*` functions and
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type Option func(*Options)

// WithProbeHint sets the Options.Hint option.
func WithProbeHint(hint string) Option { return func(args *Options) { args.Hint = hint } }

// WithHintCallBack sets the Options.HintCallback option.
func WithHintCallBack(fn func(string) error) Option {
	return func(args *Options) { args.HintCallback = fn }
}

// WithLog sets the Options.Log option
func WithLog(userLog logr.Logger) Option { return func(args *Options) { args.Log = userLog } }

// WithContext sets the Options.Context option
func WithContext(ctx context.Context) Option { return func(args *Options) { args.Context = ctx } }

func swapProbe(order []string, hint string) {
	// With so few elements and since `==` uses SIMD,
	// looping is faster than having yet another hash map.
	for i := range order {
		if order[i] == hint {
			order[0], order[i] = order[i], order[0]
			break
		}
	}
}
