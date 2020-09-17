package discover

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	_ "github.com/bmc-toolbox/bmclib/logging" // this make possible to setup logging and properties at any stage
	"github.com/bmc-toolbox/bmclib/providers/dummy/ibmc"
)

const (
	ProbeHpIlo       = "hpilo"
	ProbeIdrac8      = "idrac8"
	ProbeIdrac9      = "idrac9"
	ProbeSupermicrox = "supermicrox"
	ProbeHpC7000     = "hpc7000"
	ProbeM1000e      = "m1000e"
	ProbeQuanta      = "quanta"
	ProbeHpCl100     = "hpcl100"
)

// ScanAndConnect will scan the bmc trying to learn the device type and return a working connection.
func ScanAndConnect(host string, username string, password string, options ...Option) (bmcConnection interface{}, err error) {
	log.WithFields(log.Fields{"step": "ScanAndConnect", "host": host}).Debug("detecting vendor")

	opts := &Options{HintCallback: func(_ string) error { return nil }}
	for _, optFn := range options {
		optFn(opts)
	}

	// return a connection to our dummy device.
	if os.Getenv("BMCLIB_TEST") == "1" {
		log.WithFields(log.Fields{"step": "ScanAndConnect", "host": host}).Debug("returning connection to dummy ibmc device.")
		bmc, err := ibmc.New(host, username, password)
		return bmc, err
	}

	client, err := httpclient.Build()
	if err != nil {
		return nil, err
	}

	var probe = Probe{client: client, username: username, password: password, host: host}

	var devices = map[string]func() (interface{}, error){
		ProbeHpIlo:       probe.hpIlo,
		ProbeIdrac8:      probe.idrac8,
		ProbeIdrac9:      probe.idrac9,
		ProbeSupermicrox: probe.supermicrox,
		ProbeHpC7000:     probe.hpC7000,
		ProbeM1000e:      probe.m1000e,
		ProbeQuanta:      probe.quanta,
		ProbeHpCl100:     probe.hpCl100,
	}

	order := []string{ProbeHpIlo,
		ProbeIdrac8,
		ProbeIdrac9,
		ProbeSupermicrox,
		ProbeHpC7000,
		ProbeM1000e,
		ProbeQuanta,
		ProbeHpCl100,
	}

	if opts.Hint != "" {
		swapProbe(order, opts.Hint)
	}

	for _, probeID := range order {
		probeDevice := devices[probeID]

		log.WithFields(log.Fields{"step": "ScanAndConnect", "host": host}).Debug("probing to identify device")

		bmcConnection, err := probeDevice()

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
