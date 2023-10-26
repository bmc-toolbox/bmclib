package redfishwrapper

import (
	"context"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
	"github.com/pkg/errors"
	gofishrf "github.com/stmcginnis/gofish/redfish"
)

func (c *Client) Task(ctx context.Context, taskID string) (*gofishrf.Task, error) {
	c.client.Service.Tasks()
	tasks, err := c.Tasks(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "error querying redfish tasks")
	}

	for _, t := range tasks {
		if t.ID != taskID {
			continue
		}

		return t, nil
	}

	return nil, bmclibErrs.ErrTaskNotFound
}
