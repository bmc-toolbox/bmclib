package lenovo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bmc-toolbox/bmclib/v2/constants"
	"github.com/bmc-toolbox/bmclib/v2/internal/redfishwrapper"
)

const (
	// updateServicePath is the XCC Redfish UpdateService resource.
	updateServicePath = "/redfish/v1/UpdateService"
	// firmwareInventoryBase is the base path for XCC firmware-inventory targets.
	firmwareInventoryBase = "/redfish/v1/UpdateService/FirmwareInventory/"
)

// updateServiceDoc is a partial model of the XCC UpdateService used to drive the
// firmware push protocol. Only the fields the protocol needs are modelled.
type updateServiceDoc struct {
	ServiceEnabled         bool   `json:"ServiceEnabled"`
	HTTPPushURI            string `json:"HttpPushUri"`
	MultipartHTTPPushURI   string `json:"MultipartHttpPushUri"`
	HTTPPushURITargetsBusy bool   `json:"HttpPushUriTargetsBusy"`
}

// readUpdateService GETs and parses the XCC UpdateService.
func (c *Conn) readUpdateService(ctx context.Context) (*updateServiceDoc, error) {
	resp, err := c.redfishwrapper.Get(updateServicePath)
	if err != nil {
		return nil, fmt.Errorf("reading XCC update service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, parseRedfishError(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errFailedToParseResponse, err)
	}

	doc := &updateServiceDoc{}
	if err := json.Unmarshal(body, doc); err != nil {
		return nil, fmt.Errorf("%w: %v", errFailedToParseResponse, err)
	}

	return doc, nil
}

// componentToTargets maps a component slug to the XCC FirmwareInventory
// target(s) referenced by HttpPushUriTargets / UpdateParameters.Targets.
//
// An empty component yields no targets, letting the XCC auto-detect the target
// from the firmware image. A component that already looks like a Redfish path is
// used verbatim; otherwise it is treated as a FirmwareInventory member id (e.g.
// "BMC-Backup", "UEFI").
func componentToTargets(component string) []string {
	switch {
	case component == "":
		return nil
	case len(component) > 0 && component[0] == '/':
		return []string{component}
	default:
		return []string{firmwareInventoryBase + component}
	}
}

// claimUpdateService claims the XCC update service for a firmware push.
//
// Per the XCC protocol the client must verify HttpPushUriTargetsBusy is false,
// then set it to true to claim the service (and set HttpPushUriTargets to the
// component target(s)). It errors if the service is already busy.
func (c *Conn) claimUpdateService(ctx context.Context, targets []string) error {
	doc, err := c.readUpdateService(ctx)
	if err != nil {
		return err
	}

	if doc.HTTPPushURITargetsBusy {
		return fmt.Errorf("XCC update service is busy (HttpPushUriTargetsBusy=true); another firmware update is in progress")
	}

	payload := map[string]any{"HttpPushUriTargetsBusy": true}
	if len(targets) > 0 {
		payload["HttpPushUriTargets"] = targets
	}

	return checkResponse(c.redfishwrapper.PatchWithHeaders(ctx, updateServicePath, payload, nil))
}

// releaseUpdateService releases the claim taken by claimUpdateService by setting
// HttpPushUriTargetsBusy back to false. When clearTargets is true (e.g. for a
// BMC backup target) it also clears HttpPushUriTargets.
func (c *Conn) releaseUpdateService(ctx context.Context, clearTargets bool) error {
	payload := map[string]any{"HttpPushUriTargetsBusy": false}
	if clearTargets {
		payload["HttpPushUriTargets"] = []string{}
	}

	return checkResponse(c.redfishwrapper.PatchWithHeaders(ctx, updateServicePath, payload, nil))
}

// pushFirmware runs the XCC push protocol: claim the service, push the image
// (multipart to MultipartHttpPushUri, falling back to a raw push to HttpPushUri
// — the selection is made by the underlying wrapper), and return the created
// task id. On any error after the claim the service is released so the busy flag
// is not leaked.
func (c *Conn) pushFirmware(ctx context.Context, component string, file *os.File, applyTime constants.OperationApplyTime) (taskID string, err error) {
	targets := componentToTargets(component)

	if err := c.claimUpdateService(ctx, targets); err != nil {
		return "", err
	}

	params := &redfishwrapper.RedfishUpdateServiceParameters{
		Targets:            targets,
		OperationApplyTime: applyTime,
	}

	taskID, err = c.redfishwrapper.FirmwareUpload(ctx, file, params)
	if err != nil {
		// Release the claim so a failed push does not leave the service busy.
		_ = c.releaseUpdateService(ctx, len(targets) > 0)
		return "", fmt.Errorf("xcc firmware push: %w", err)
	}

	return taskID, nil
}

// operationApplyTimeOrDefault converts the bmclib operationApplyTime string into
// a constants.OperationApplyTime, defaulting to Immediate when empty.
func operationApplyTimeOrDefault(s string) constants.OperationApplyTime {
	if s == "" {
		return constants.Immediate
	}
	return constants.OperationApplyTime(s)
}
