package discover

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	_ "github.com/bmc-toolbox/bmclib/logging" // this make possible to setup logging and properties at any stage
	"github.com/bmc-toolbox/bmclib/providers/dummy/ibmc"
)

type Options struct {
	// Hint is an opaque integer that hints which probe should be probed first.
	Hint int

	// HintCallBack is a function that is called back with an opaque hint that might be used
	// for the next ScanAndConnect attempt.  The callback is called only on successful scan.
	// If your code persists the hint as "best effort", always return a nil error.  Callback is
	// synchronous.
	HintCallback func(int) error
}

// Option is part of the functional options pattern, see the `With*` functions and
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type Option func(*Options)

// WithProbeHint sets the Options.Hint option.
func WithProbeHint(hint int) Option { return func(args *Options) { args.Hint = hint } }

// WithHintCallBack sets the Options.HintCallback option.
func WithHintCallBack(fn func(int) error) Option {
	return func(args *Options) { args.HintCallback = fn }
}

// ScanAndConnect will scan the bmc trying to learn the device type and return a working connection.
func ScanAndConnect(host string, username string, password string, options ...Option) (bmcConnection interface{}, err error) {
	log.WithFields(log.Fields{"step": "ScanAndConnect", "host": host}).Debug("detecting vendor")

	opts := &Options{Hint: 0, HintCallback: func(_ int) error { return nil }}
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
	var devices = []func() (interface{}, error){
		probe.hpIlo,
		probe.idrac8,
		probe.idrac9,
		probe.supermicrox,
		probe.hpC7000,
		probe.m1000e,
		probe.quanta,
		probe.hpCl100,
	}

	// Check if probe hint is in bounds, revert to default if not.  Code below relies on this.
	if opts.Hint < 0 || opts.Hint >= len(devices) {
		opts.Hint = 0
	}

	// Swap the hinted probe with the first probe. Relies on bounds check above.
	devices[0], devices[opts.Hint] = devices[opts.Hint], devices[0]

	for i, probeDevice := range devices {
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

		// Success.  Figure out which probe it was.  We need to do some reconstruction
		// because of the swapping above.  Relies on bounds check above.
		var hint int
		switch i {
		case 0:
			hint = opts.Hint
		case opts.Hint:
			hint = 0
		default:
			hint = i
		}

		if err := opts.HintCallback(hint); err != nil {
			return nil, err
		}

		// return a bmcConnection
		return bmcConnection, nil
	}

	return nil, errors.ErrVendorUnknown
}
