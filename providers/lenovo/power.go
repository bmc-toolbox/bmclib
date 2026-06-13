package lenovo

import (
	"context"
)

// PowerStateGet returns the current power state of the managed system (e.g.
// "On", "Off"), read from the ComputerSystem PowerState property.
//
// Implements bmc.PowerStateGetter.
func (c *Conn) PowerStateGet(ctx context.Context) (state string, err error) {
	return c.redfishwrapper.SystemPowerStatus(ctx)
}

// PowerSet sets the system power state via the ComputerSystem.Reset action.
//
// Accepted states (case-insensitive) and the XCC ResetType they map to:
//
//	on     -> On             (no-op if already on)
//	off    -> ForceOff
//	soft   -> GracefulShutdown
//	reset  -> ForceRestart (with power-off/on fallback)
//	cycle  -> ForceRestart
//
// Unsupported states return an error. Implements bmc.PowerSetter.
func (c *Conn) PowerSet(ctx context.Context, state string) (ok bool, err error) {
	return c.redfishwrapper.PowerSet(ctx, state)
}

// SendNMI issues a non-maskable interrupt to the host via a ComputerSystem.Reset
// action with ResetType "Nmi".
//
// Implements bmc.NMISender.
func (c *Conn) SendNMI(ctx context.Context) error {
	return c.redfishwrapper.SendNMI(ctx)
}
