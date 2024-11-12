package asrockrack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
	"github.com/pkg/errors"
)

type power struct {
	Command int `json:"power_command"`
}

// PowerStateGet gets the power state of a machine
func (a *ASRockRack) PowerStateGet(ctx context.Context) (state string, err error) {
	info, err := a.chassisStatusInfo(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "401") {
			// during a BMC update, only the flash-progress endpoint can be queried
			// and so we cannot determine server power status
			// we don't return an error here because we don't want the bmclib client to retry another provider.
			progress, err := a.flashProgress(ctx, "/api/maintenance/firmware/flash-progress")
			if err == nil && progress.Action != "" {
				a.log.V(2).WithValues(
					"action", progress.Action,
					"progress", progress.Progress,
					"state", progress.State,
				).Info("bmc in flash mode, power status cannot be determined")

				return "", errors.Wrap(
					bmclibErrs.ErrBMCUpdating,
					fmt.Sprintf(
						"action: %s, progress: %s, state: %d",
						progress.Action,
						progress.Progress,
						progress.State,
					),
				)
			}
		}

		return "", errors.Wrap(bmclibErrs.ErrPowerStatusRead, err.Error())
	}

	switch info.PowerStatus {
	case 0:
		return "Off", nil
	case 1:
		return "On", nil
	default:
		return "", errors.Wrap(
			bmclibErrs.ErrPowerStatusRead,
			fmt.Errorf("unknown status: %d", info.PowerStatus).Error(),
		)
	}
}

// PowerSet sets the hardware power state of a machine
func (a *ASRockRack) PowerSet(ctx context.Context, state string) (ok bool, err error) {
	switch strings.ToLower(state) {
	case "on":
		return a.powerAction(ctx, 1)
	case "off":
		return a.powerAction(ctx, 0)
	case "soft":
		return a.powerAction(ctx, 5)
	case "reset":
		return a.powerAction(ctx, 3)
	case "cycle":
		return a.powerAction(ctx, 2)
	default:
		return false, errors.New("requested power state unknown: " + state)
	}
}

func (a *ASRockRack) powerAction(ctx context.Context, action int) (ok bool, err error) {
	endpoint := "/api/actions/power"

	p := power{Command: action}
	payload, err := json.Marshal(p)
	if err != nil {
		return false, err
	}

	headers := map[string]string{"Content-Type": "application/json"}
	_, statusCode, err := a.queryHTTPS(
		ctx,
		endpoint,
		"POST",
		bytes.NewReader(payload),
		headers,
		0,
	)
	if err != nil {
		return false, errors.Wrap(bmclibErrs.ErrPowerStatusSet, err.Error())
	}

	if statusCode != http.StatusOK {
		return false, errors.Wrap(
			bmclibErrs.ErrNon200Response,
			fmt.Errorf("%d", statusCode).Error(),
		)
	}

	return true, nil
}

// BmcReset will reset the BMC - ASRR BMCs only support a cold reset.
func (a *ASRockRack) BmcReset(ctx context.Context, resetType string) (ok bool, err error) {
	err = a.resetBMC(ctx)
	if err != nil {
		return false, err
	}

	return true, nil
}

// 4. reset BMC - performs a cold reset
func (a *ASRockRack) resetBMC(ctx context.Context) error {
	endpoint := "api/maintenance/reset"

	_, statusCode, err := a.queryHTTPS(ctx, endpoint, "POST", nil, nil, 0)
	if err != nil {
		return err
	}

	// The E3C256D4ID BMC returns a 500 status error on the BMC reset request
	if statusCode != http.StatusOK && statusCode != http.StatusInternalServerError {
		return fmt.Errorf("non 200 response: %d", statusCode)
	}

	return nil
}
