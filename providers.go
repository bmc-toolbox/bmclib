package bmclib

import (
	"slices"
	"strings"

	"github.com/bmc-toolbox/bmclib/v2/providers/asrockrack"
	"github.com/bmc-toolbox/bmclib/v2/providers/compal"
	"github.com/bmc-toolbox/bmclib/v2/providers/dell"
	"github.com/bmc-toolbox/bmclib/v2/providers/intelamt"
	"github.com/bmc-toolbox/bmclib/v2/providers/ipmi"
	"github.com/bmc-toolbox/bmclib/v2/providers/ipmitool"
	"github.com/bmc-toolbox/bmclib/v2/providers/lenovo"
	"github.com/bmc-toolbox/bmclib/v2/providers/openbmc"
	"github.com/bmc-toolbox/bmclib/v2/providers/redfish"
	"github.com/bmc-toolbox/bmclib/v2/providers/supermicro"
)

// defaultProvider is a provider NewClient registers unless WithProviders leaves it out.
type defaultProvider struct {
	name     string
	register func(*Client) error
}

// noError adapts a register method that can't fail to the defaultProvider signature.
func noError(register func(*Client)) func(*Client) error {
	return func(c *Client) error {
		register(c)
		return nil
	}
}

// defaultProviderList holds the providers NewClient registers by default, in the order it
// registers them. registerProviders and ProviderNames both read it, so they can't differ.
var defaultProviderList = []defaultProvider{
	{ipmi.ProviderName, (*Client).registerIPMIProvider},
	{ipmitool.ProviderName, (*Client).registerIpmitoolProvider},
	{asrockrack.ProviderName, noError((*Client).registerASRRProvider)},
	{redfish.ProviderName, noError((*Client).registerGofishProvider)},
	{intelamt.ProviderName, noError((*Client).registerIntelAMTProvider)},
	{dell.ProviderName, noError((*Client).registerDellProvider)},
	{lenovo.ProviderName, noError((*Client).registerLenovoProvider)},
	{compal.ProviderName, noError((*Client).registerCompalProvider)},
	{supermicro.ProviderName, noError((*Client).registerSupermicroProvider)},
	{openbmc.ProviderName, noError((*Client).registerOpenBMCProvider)},
}

// ProviderNames returns the names of the providers NewClient registers by default, in the
// order it registers them. The rpc and homeassistant providers are not included: they are
// registered only when configured, and then alone. A new slice is returned on each call.
func ProviderNames() []string {
	names := make([]string, 0, len(defaultProviderList))
	for _, p := range defaultProviderList {
		names = append(names, p.name)
	}

	return names
}

// providerAllowed reports whether NewClient registers the provider name: always when
// WithProviders named none, otherwise when it is one of them, compared without regard to case.
func (c *Client) providerAllowed(name string) bool {
	if len(c.providers) == 0 {
		return true
	}

	return slices.ContainsFunc(c.providers, func(p string) bool { return strings.EqualFold(p, name) })
}

// unknownProviders returns the names given to WithProviders that are not ProviderNames.
func (c *Client) unknownProviders() []string {
	known := ProviderNames()
	var unknown []string
	for _, p := range c.providers {
		if !slices.ContainsFunc(known, func(k string) bool { return strings.EqualFold(k, p) }) {
			unknown = append(unknown, p)
		}
	}

	return unknown
}
