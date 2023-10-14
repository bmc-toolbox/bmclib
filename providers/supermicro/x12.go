package supermicro

import (
	"context"

	"github.com/bmc-toolbox/common"
	"github.com/pkg/errors"
)

type x12 struct{ *Client }

func (c *x12) deviceModel(ctx context.Context) (string, error) {
	if err := c.redfishSession(ctx); err != nil {
		return "", err
	}

	_, model, err := c.redfish.DeviceVendorModel(ctx)
	if err != nil {
		return "", err
	}

	if model == "" {
		return "", errors.Wrap(ErrModelUnknown, "empty value")
	}

	return common.FormatProductName(model), nil
}
