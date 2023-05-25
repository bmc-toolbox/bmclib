package redfishwrapper

import (
	"context"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"

	"github.com/pkg/errors"
	gofishrf "github.com/stmcginnis/gofish/redfish"
)

// The methods here should be a thin wrapper so as to only guard the client from authentication failures.

// AccountService gets the Redfish AccountService.d
func (c *Client) AccountService() (*gofishrf.AccountService, error) {
	if err := c.SessionActive(); err != nil {
		return nil, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	return c.client.Service.AccountService()
}

// UpdateService gets the update service instance.
func (c *Client) UpdateService() (*gofishrf.UpdateService, error) {
	if err := c.SessionActive(); err != nil {
		return nil, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	return c.client.Service.UpdateService()
}

// Systems get the system instances from the service.
func (c *Client) Systems() ([]*gofishrf.ComputerSystem, error) {
	if err := c.SessionActive(); err != nil {
		return nil, err
	}

	return c.client.Service.Systems()
}

// Managers gets the manager instances of this service.
func (c *Client) Managers(ctx context.Context) ([]*gofishrf.Manager, error) {
	if err := c.SessionActive(); err != nil {
		return nil, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	return c.client.Service.Managers()
}

// Chassis gets the chassis instances managed by this service.
func (c *Client) Chassis(ctx context.Context) ([]*gofishrf.Chassis, error) {
	if err := c.SessionActive(); err != nil {
		return nil, errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	return c.client.Service.Chassis()
}
