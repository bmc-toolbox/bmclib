package bmc

import "context"

// StorageControllerInfo describes a storage controller.
type StorageControllerInfo struct {
	// ID is the Redfish Storage resource Id (used to address volumes/drives).
	ID string
	// Name is the controller name.
	Name string
	// Model is the controller model.
	Model string
	// Health is the controller health (e.g. "OK", "Warning", "Critical").
	Health string
	// DriveCount is the number of drives managed by the controller.
	DriveCount int
}

// VolumeInfo describes a storage volume.
type VolumeInfo struct {
	// ID is the Redfish Volume resource Id.
	ID string
	// Name is the volume name.
	Name string
	// RAIDType is the volume RAID type (e.g. "RAID0", "RAID1").
	RAIDType string
	// CapacityBytes is the volume capacity in bytes.
	CapacityBytes int64
	// Health is the volume health (e.g. "OK", "Warning", "Critical").
	Health string
}

// VolumeCreateRequest describes a volume to create on a storage controller.
type VolumeCreateRequest struct {
	// Name is the requested volume name.
	Name string
	// RAIDType is the requested RAID type (e.g. "RAID0", "RAID1", "RAID5").
	RAIDType string
	// CapacityBytes is the requested capacity in bytes. Zero lets the
	// controller choose the maximum available capacity.
	CapacityBytes int64
	// Drives is the list of member drive Redfish @odata.id paths.
	Drives []string
}

// VolumeManager is implemented by providers that can read and manage storage
// volumes on a controller.
//
// Volumes and controllers are addressed by their Redfish resource Id within the
// system's Storage collection.
type VolumeManager interface {
	// StorageControllers lists the storage controllers of the system.
	StorageControllers(ctx context.Context) ([]StorageControllerInfo, error)
	// Volumes lists the volumes managed by the given storage controller.
	Volumes(ctx context.Context, storageID string) ([]VolumeInfo, error)
	// VolumeCreate creates a volume on the given storage controller and returns
	// the new volume Id (or a task Id when the operation is asynchronous).
	VolumeCreate(ctx context.Context, storageID string, req VolumeCreateRequest) (volumeID string, err error)
	// VolumeInitialize initializes a volume. initType is a Redfish
	// InitializeType such as "Fast" or "Slow".
	VolumeInitialize(ctx context.Context, storageID, volumeID, initType string) error
	// VolumeUpdate updates volume settings via a PATCH of the given attributes.
	VolumeUpdate(ctx context.Context, storageID, volumeID string, settings map[string]any) error
	// VolumeDelete deletes a volume.
	VolumeDelete(ctx context.Context, storageID, volumeID string) error
}
