package redfishwrapper

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/stmcginnis/gofish"
	"github.com/stmcginnis/gofish/schemas"
)

// getVirtualMedia retrieves virtual media resources by first checking the
// Redfish Manager path and falling back to the System path if none are found.
//
// Some BMC implementations (e.g., Dell iDRAC) expose VirtualMedia under the
// System resource (/redfish/v1/Systems/{SystemId}/VirtualMedia) rather than
// the Manager resource (/redfish/v1/Managers/{ManagerId}/VirtualMedia).
// Both locations are valid per the Redfish specification.
func (c *Client) getVirtualMedia(ctx context.Context) ([]*schemas.VirtualMedia, error) {
	// Try Manager path first (standard Redfish location).
	m, err := c.Manager(ctx)
	if err == nil {
		vm, err := m.VirtualMedia()
		if err == nil && len(vm) > 0 {
			return vm, nil
		}
	}

	// Fallback to System path (Dell iDRAC and other implementations that
	// expose VirtualMedia under ComputerSystem per Redfish spec v1.12.0+).
	sys, err := c.System()
	if err == nil {
		vm, err := sys.VirtualMedia()
		if err == nil && len(vm) > 0 {
			return vm, nil
		}
	}

	return nil, errors.New("no virtual media found at Manager or System resource paths")
}

// ejectVirtualMedia ejects the media from every slot of mediaKind that has media inserted. Each
// inserted image is a boot option, so one left in another slot can be booted instead of the next
// image inserted. It succeeds when every inserted slot of the kind was ejected, including when none
// had media, and fails when any eject fails or no slot supports the kind.
func ejectVirtualMedia(virtualMedia []*schemas.VirtualMedia, mediaKind schemas.VirtualMediaType, kind string) (bool, error) {
	supportedMediaTypes := []string{}
	matched := false
	var slotErrors []error

	for _, vm := range virtualMedia {
		if !slices.Contains(vm.MediaTypes, mediaKind) {
			for _, mt := range vm.MediaTypes {
				if !slices.Contains(supportedMediaTypes, string(mt)) {
					supportedMediaTypes = append(supportedMediaTypes, string(mt))
				}
			}

			continue
		}

		matched = true
		if !gofish.Deref(vm.Inserted) {
			continue
		}

		// The gofish library v0.22.0+ falls back to PATCH when the EjectMedia action isn't available,
		// as on some Lenovo XCC implementations.
		if _, err := vm.EjectMedia(); err != nil {
			slotErrors = append(slotErrors, fmt.Errorf("%s: eject: %w", vm.ODataID, err))
		}
	}

	if len(slotErrors) > 0 {
		return false, fmt.Errorf("failed to eject virtual media: %w", errors.Join(slotErrors...))
	}
	if !matched {
		return false, fmt.Errorf("not a supported media type: %s. supported media types: %v", kind, supportedMediaTypes)
	}

	return true, nil
}

// virtualMediaKind returns the Redfish media type for kind.
func virtualMediaKind(kind string) (schemas.VirtualMediaType, error) {
	switch kind {
	case "CD":
		return schemas.CDVirtualMediaType, nil
	case "Floppy":
		return schemas.FloppyVirtualMediaType, nil
	case "USBStick":
		return schemas.USBStickVirtualMediaType, nil
	case "DVD":
		return schemas.DVDVirtualMediaType, nil
	default:
		return "", errors.New("invalid media type")
	}
}

// EjectVirtualMedia ejects the media from every slot of the kind.
func (c *Client) EjectVirtualMedia(ctx context.Context, kind string) (bool, error) {
	mediaKind, err := virtualMediaKind(kind)
	if err != nil {
		return false, err
	}

	virtualMedia, err := c.getVirtualMedia(ctx)
	if err != nil {
		return false, err
	}

	return ejectVirtualMedia(virtualMedia, mediaKind, kind)
}

// SetVirtualMedia sets virtual media on the system. It first ejects the media
// from every slot of the kind, and returns there if mediaURL is empty.
// Otherwise, when multiple matching virtual media slots exist, each slot is
// tried in order until one accepts the media.
func (c *Client) SetVirtualMedia(ctx context.Context, kind, mediaURL string) (bool, error) {
	mediaKind, err := virtualMediaKind(kind)
	if err != nil {
		return false, err
	}

	virtualMedia, err := c.getVirtualMedia(ctx)
	if err != nil {
		return false, err
	}

	// To guarantee reliable boot behavior across all BMCs, ensure no other
	// media of this kind is mounted in secondary slots. Ejecting the media
	// before inserting new media also makes inserting it more likely to succeed.
	_, ejectErr := ejectVirtualMedia(virtualMedia, mediaKind, kind)

	if mediaURL == "" {
		if ejectErr != nil {
			return false, ejectErr
		}

		return true, nil
	}

	supportedMediaTypes := []string{}
	var slotErrors []error

	for _, vm := range virtualMedia {
		if !slices.Contains(vm.MediaTypes, mediaKind) {
			for _, mt := range vm.MediaTypes {
				if !slices.Contains(supportedMediaTypes, string(mt)) {
					supportedMediaTypes = append(supportedMediaTypes, string(mt))
				}
			}

			continue
		}

		insertMediaAttempts := []schemas.VirtualMediaInsertMediaParameters{
			{
				Image:          mediaURL,
				Inserted:       gofish.ToRef(true),
				WriteProtected: gofish.ToRef(true),
			},
			{
				Image:    mediaURL,
				Inserted: gofish.ToRef(true),
			},
			{
				Image: mediaURL,
			},
		}

		// Attempt to insert media using different parameter combinations.
		// The gofish library v0.22.0+ supports fallback to PATCH if standard Actions are not available.
		// This sequence tries:
		// 1. Image + Inserted + WriteProtected (standard)
		// 2. Image + Inserted (for BMCs requiring Inserted but not WriteProtected, e.g., some Lenovo XCC implementations)
		// 3. Image only (for BMCs that don't support Inserted/WriteProtected properties, e.g., Supermicro X11SDV-4C-TLN2F)
		var insertErrors []error

		for _, insertMedia := range insertMediaAttempts {
			if _, err := vm.InsertMedia(&insertMedia); err != nil {
				insertErrors = append(insertErrors, err)
				continue
			}

			return true, nil
		}

		slotErrors = append(slotErrors, fmt.Errorf("%s: insert: %w", vm.ODataID, errors.Join(insertErrors...)))
		continue
	}

	if len(slotErrors) > 0 {
		if ejectErr != nil {
			slotErrors = append([]error{ejectErr}, slotErrors...)
		}

		return false, fmt.Errorf("all matching virtual media slots failed: %w", errors.Join(slotErrors...))
	}

	return false, fmt.Errorf("not a supported media type: %s. supported media types: %v", kind, supportedMediaTypes)
}

// InsertedVirtualMedia returns the IDs of virtual media slots that currently have media inserted.
func (c *Client) InsertedVirtualMedia(ctx context.Context) ([]string, error) {
	virtualMedia, err := c.getVirtualMedia(ctx)
	if err != nil {
		return nil, err
	}

	var inserted []string

	for _, media := range virtualMedia {
		if gofish.Deref(media.Inserted) {
			inserted = append(inserted, media.ID)
		}
	}

	return inserted, nil
}
