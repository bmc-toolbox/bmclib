package lenovo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/schemas"
)

// compile-time assertion that the provider implements the interface.
var _ bmc.VolumeManager = (*Conn)(nil)

// StorageControllers lists the system's storage controllers.
//
// Implements bmc.VolumeManager.
func (c *Conn) StorageControllers(ctx context.Context) ([]bmc.StorageControllerInfo, error) {
	storages, err := c.storageControllers()
	if err != nil {
		return nil, err
	}

	out := make([]bmc.StorageControllerInfo, 0, len(storages))
	for _, s := range storages {
		info := bmc.StorageControllerInfo{
			ID:         s.ID,
			Name:       s.Name,
			Health:     string(s.Status.Health),
			DriveCount: s.DrivesCount,
		}
		// XCC populates the deprecated inline StorageControllers array rather
		// than the Controllers collection, so read the controller model here.
		if sc := s.StorageControllers; len(sc) > 0 { //nolint:staticcheck
			info.Model = sc[0].Model
		}
		out = append(out, info)
	}

	return out, nil
}

// Volumes lists the volumes managed by the given storage controller.
//
// Implements bmc.VolumeManager.
func (c *Conn) Volumes(ctx context.Context, storageID string) ([]bmc.VolumeInfo, error) {
	storage, err := c.storageByID(storageID)
	if err != nil {
		return nil, err
	}

	volumes, err := storage.Volumes()
	if err != nil {
		return nil, fmt.Errorf("listing volumes for controller %q: %w", storageID, err)
	}

	out := make([]bmc.VolumeInfo, 0, len(volumes))
	for _, v := range volumes {
		out = append(out, bmc.VolumeInfo{
			ID:            v.ID,
			Name:          v.Name,
			RAIDType:      string(v.RAIDType),
			CapacityBytes: int64(gofish.Deref(v.CapacityBytes)),
			Health:        string(v.Status.Health),
		})
	}

	return out, nil
}

// VolumeCreate creates a volume on the given storage controller by POSTing to
// its Volumes collection, returning the new volume Id (from the Location header
// or response body).
//
// Implements bmc.VolumeManager.
func (c *Conn) VolumeCreate(ctx context.Context, storageID string, req bmc.VolumeCreateRequest) (string, error) {
	storage, err := c.storageByID(storageID)
	if err != nil {
		return "", err
	}

	payload := map[string]any{
		"Name":     req.Name,
		"RAIDType": req.RAIDType,
	}
	if req.CapacityBytes > 0 {
		payload["CapacityBytes"] = req.CapacityBytes
	}
	if len(req.Drives) > 0 {
		drives := make([]map[string]string, 0, len(req.Drives))
		for _, d := range req.Drives {
			drives = append(drives, map[string]string{"@odata.id": d})
		}
		payload["Links"] = map[string]any{"Drives": drives}
	}

	volumesURL, err := url.JoinPath(storage.ODataID, "Volumes")
	if err != nil {
		return "", err
	}

	resp, err := c.redfishwrapper.PostWithHeaders(ctx, volumesURL, payload, nil)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return "", parseRedfishError(resp)
	}

	// Prefer the created resource location; fall back to the body Id.
	if loc := resp.Header.Get("Location"); loc != "" {
		return path.Base(strings.TrimRight(loc, "/")), nil
	}

	var created struct {
		ID string `json:"Id"`
	}
	if body, rerr := io.ReadAll(resp.Body); rerr == nil {
		_ = json.Unmarshal(body, &created)
	}

	return created.ID, nil
}

// VolumeInitialize initializes a volume via the Volume.Initialize action.
//
// The Redfish action target is derived from the volume path; only the
// InitializeType is sent so the controller's default InitializeMethod applies.
// Implements bmc.VolumeManager.
func (c *Conn) VolumeInitialize(ctx context.Context, storageID, volumeID, initType string) error {
	volume, err := c.volumeByID(storageID, volumeID)
	if err != nil {
		return err
	}

	target, err := url.JoinPath(volume.ODataID, "Actions/Volume.Initialize")
	if err != nil {
		return err
	}
	payload := map[string]any{"InitializeType": initType}

	return checkResponse(c.redfishwrapper.PostWithHeaders(ctx, target, payload, nil))
}

// VolumeUpdate updates volume settings via a PATCH of the given attributes.
//
// Implements bmc.VolumeManager.
func (c *Conn) VolumeUpdate(ctx context.Context, storageID, volumeID string, settings map[string]any) error {
	volume, err := c.volumeByID(storageID, volumeID)
	if err != nil {
		return err
	}

	return checkResponse(c.redfishwrapper.PatchWithHeaders(ctx, volume.ODataID, settings, nil))
}

// VolumeDelete deletes a volume.
//
// Implements bmc.VolumeManager.
func (c *Conn) VolumeDelete(ctx context.Context, storageID, volumeID string) error {
	volume, err := c.volumeByID(storageID, volumeID)
	if err != nil {
		return err
	}

	return checkResponse(c.redfishwrapper.Delete(volume.ODataID))
}

// storageControllers returns the system's gofish Storage resources.
func (c *Conn) storageControllers() ([]*schemas.Storage, error) {
	system, err := c.redfishwrapper.System()
	if err != nil {
		return nil, err
	}

	storages, err := system.Storage()
	if err != nil {
		return nil, fmt.Errorf("listing storage controllers: %w", err)
	}

	return storages, nil
}

// storageByID resolves a gofish Storage resource by its Id.
func (c *Conn) storageByID(storageID string) (*schemas.Storage, error) {
	storages, err := c.storageControllers()
	if err != nil {
		return nil, err
	}

	for _, s := range storages {
		if s.ID == storageID {
			return s, nil
		}
	}

	return nil, fmt.Errorf("storage controller %q not found", storageID)
}

// volumeByID resolves a gofish Volume resource by controller Id and volume Id.
func (c *Conn) volumeByID(storageID, volumeID string) (*schemas.Volume, error) {
	storage, err := c.storageByID(storageID)
	if err != nil {
		return nil, err
	}

	volumes, err := storage.Volumes()
	if err != nil {
		return nil, fmt.Errorf("listing volumes for controller %q: %w", storageID, err)
	}

	for _, v := range volumes {
		if v.ID == volumeID {
			return v, nil
		}
	}

	return nil, fmt.Errorf("volume %q not found on controller %q", volumeID, storageID)
}
