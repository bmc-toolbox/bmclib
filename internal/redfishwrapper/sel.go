package redfishwrapper

import (
	"context"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/pkg/errors"
)

// Clear SEL clears all of the LogServices logs
func (c *Client) ClearSEL(ctx context.Context) (err error) {
	if err := c.SessionActive(); err != nil {
		return errors.Wrap(bmclibErrs.ErrNotAuthenticated, err.Error())
	}

	chassis, err := c.client.Service.Chassis()
	if err != nil {
		return err
	}

	for _, c := range chassis {
		logServices, err := c.LogServices()
		if err != nil {
			return err
		}

		for _, logService := range logServices {
			err = logService.ClearLog()
			if err != nil {
				return err
			}
		}
	}

	return nil
}
