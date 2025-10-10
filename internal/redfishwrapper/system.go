package redfishwrapper

import (
	"context"
	"fmt"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"

	"github.com/pkg/errors"
	redfish "github.com/stmcginnis/gofish/redfish"
)

// The methods here should be a thin wrapper so as to only guard the client from authentication failures.

// AccountService gets the Redfish AccountService.d
func (c *Client) AccountService() (*redfish.AccountService, error) {
	if err := c.SessionActive(); err != nil {
		return nil, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	return c.client.Service.AccountService()
}

// UpdateService gets the update service instance.
func (c *Client) UpdateService() (*redfish.UpdateService, error) {
	if err := c.SessionActive(); err != nil {
		return nil, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	return c.client.Service.UpdateService()
}

// System gets the system matching c.systemName or when c.systemName is not set
// and there is only one system it returns that system.
func (c *Client) System() (*redfish.ComputerSystem, error) {
	if err := c.SessionActive(); err != nil {
		return nil, err
	}

	systems, err := c.client.Service.Systems()
	if err != nil {
		return nil, err
	}

	// If no system name is set and there is only one system, return it.
	// This is to handle backwards compatibility where we didn't require passing
	// a system name to the client.
	if c.systemName == "" && len(systems) == 1 && systems[0] != nil {
		return systems[0], nil
	}

	return c.matchingSystem(systems)
}

// Manager gets the manager instances of this service. It matches the manager
// to the system name if one is set in the client. If no system name is set
// and there is only one manager it returns that manager. Otherwise it returns
// an error.
func (c *Client) Manager(ctx context.Context) (*redfish.Manager, error) {
	if err := c.SessionActive(); err != nil {
		return nil, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	ms, err := c.client.Service.Managers()
	if err != nil {
		return nil, err
	}

	// If no system name is set and there is only one manager, return it.
	// This is to handle backwards compatibility where we didn't require passing
	// a system name to the client.
	if c.systemName == "" && len(ms) == 1 && ms[0] != nil {
		return ms[0], nil
	}

	for _, m := range ms {
		sys, err := m.ManagerForServers()
		if err != nil {
			continue
		}
		if _, err := c.matchingSystem(sys); err == nil {
			return m, nil
		}
	}

	return nil, fmt.Errorf("no matching redfish manager found for system: %s", c.systemName)
}

func (c *Client) Managers(ctx context.Context) ([]*redfish.Manager, error) {
	if err := c.SessionActive(); err != nil {
		return nil, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	return c.client.Service.Managers()
}

// Chassis gets the chassis instances managed by this service.
func (c *Client) Chassis(ctx context.Context) ([]*redfish.Chassis, error) {
	if err := c.SessionActive(); err != nil {
		return nil, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	return c.client.Service.Chassis()
}

func (c *Client) matchingSystem(systems []*redfish.ComputerSystem) (*redfish.ComputerSystem, error) {
	for _, s := range systems {
		if s != nil && s.Name == c.systemName {
			return s, nil
		}
	}

	return nil, fmt.Errorf("no matching redfish system found for system: %s", c.systemName)
}
