package bmclib

import "github.com/jacobweinstock/registrar"

// PreferProvider reorders the registry to have the given provider first.
// This is a one time/temporary reordering of the providers in the registry.
// This reorder is not preserved. It is only used for the call that uses the returned Client.
// Update the Client.Registry to make the change permanent. For example, `cl.Registry.Drivers = cl.Registry.PreferDriver("ipmitool")`
func (c *Client) PreferProvider(name string) *Client {
	c.oneTimeRegistry.Drivers = c.Registry.PreferDriver(name)
	c.oneTimeRegistryEnabled = true

	return c
}

// Supports removes any provider from the registry that does not support the given features.
// This is a one time/temporary reordering of the providers in the registry.
// This reorder is not preserved. It is only used for the call that uses the returned Client.
func (c *Client) Supports(features ...registrar.Feature) *Client {
	c.oneTimeRegistry.Drivers = c.Registry.Supports(features...)
	c.oneTimeRegistryEnabled = true

	return c
}

// Using removes any provider from the registry that does not support the given protocol.
// This is a one time/temporary reordering of the providers in the registry.
// This reorder is not preserved. It is only used for the call that uses the returned Client.
func (c *Client) Using(protocol string) *Client {
	c.oneTimeRegistry.Drivers = c.Registry.Using(protocol)
	c.oneTimeRegistryEnabled = true

	return c
}

// For removes any provider from the registry that is not the given provider.
// This is a one time/temporary reordering of the providers in the registry.
// This reorder is not preserved. It is only used for the call that uses the returned Client.
func (c *Client) For(provider string) *Client {
	c.oneTimeRegistry.Drivers = c.Registry.For(provider)
	c.oneTimeRegistryEnabled = true

	return c
}

// PreferProtocol reorders the providers in the registry to have the given protocol first. Matching providers order is preserved.
// This is a one time/temporary reordering of the providers in the registry.
// This reorder is not preserved. It is only used for the call that uses the returned Client.
func (c *Client) PreferProtocol(protocols ...string) *Client {
	c.oneTimeRegistry.Drivers = c.Registry.PreferProtocol(protocols...)
	c.oneTimeRegistryEnabled = true

	return c
}
