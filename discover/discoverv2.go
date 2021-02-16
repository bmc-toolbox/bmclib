package discover

import (
	"context"
	"fmt"

	"github.com/bmc-toolbox/bmclib"
	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/logging"
	"github.com/bmc-toolbox/bmclib/providers/asrockrack"
	"github.com/jacobweinstock/registrar"
)

var (
	// order the devices are to be probed in
	ProbeOrder = []string{
		ProbeASRockRack,
	}
)

// Connect identifies the bmc, registers the bmc connector, features and returns a bmclib client
func ScanAndConnectv2(host string, username string, password string, options ...Option) (client *bmclib.Client, err error) {

	opts := &Options{
		HintCallback: func(_ string) error { return nil },
		Username:     username,
		Password:     password,
		Host:         host,
	}

	// set options
	for _, optFn := range options {
		optFn(opts)
	}

	// default logger
	if opts.Logger == nil {
		opts.Logger = logging.DefaultLogger()
	}

	// default context
	if opts.Context == nil {
		opts.Context = context.Background()
	}
	opts.Logger.V(0).Info("detecting vendor", "step", "ScanAndConnect", "host", host)

	// setup probe
	probe, err := NewProbev2(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to probe device: %s", err.Error())
	}

	// devices maps device name/models to a probing method
	var devices = map[string]func(context.Context) (interface{}, error){
		ProbeASRockRack: probe.asRockRack,
	}

	// when a hint is given, move matching probe to top of the list
	if opts.Hint != "" {
		swapProbe(ProbeOrder, opts.Hint)
	}

	for _, probeID := range ProbeOrder {

		probeDevice := devices[probeID]
		opts.Logger.V(0).Info("probing to identify device", "step", "ScanAndConnect", "host", host, "vendor", probeID)

		device, err := probeDevice(opts.Context)
		if err != nil {
			// log error if probe is not successful
			opts.Logger.V(0).Info("probe failed", "step", "ScanAndConnect", "host", host, "vendor", probeID, "error", err)
			continue
		}

		// invoke hint callback if any
		if hintErr := opts.HintCallback(probeID); hintErr != nil {
			return nil, hintErr
		}

		// return a bmclib client
		if device != nil {
			return clientForProvider(device, opts)
		}
	}

	return nil, errors.ErrVendorUnknown
}

// setup registry, assert provider and return client
func clientForProvider(provider interface{}, opts *Options) (*bmclib.Client, error) {

	// setup driver registry
	registry := registrar.NewRegistry()

	// type assert driver
	switch device := provider.(type) {
	case *asrockrack.ASRockRack:
		// If the Registry interface is extended to include the Provider(), ProviderProtocol() and Features() methods
		// we can eliminate having to pass all of these in here
		registry.Register(device.Vendor(), device.ProviderProtocol(), device.Features(), nil, device)
	default:
		return nil, errors.ErrVendorUnknown
	}

	// setup registry options
	regOptions := bmclib.WithRegistry(registry)

	// setup client based on options
	client := bmclib.NewClient(opts.Host, "", opts.Username, opts.Password, regOptions)
	client.Registry.Logger = opts.Logger

	return client, nil

}
