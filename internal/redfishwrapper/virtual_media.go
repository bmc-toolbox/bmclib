package redfishwrapper

import (
	"context"
	"errors"
	"fmt"
	"slices"

	rf "github.com/stmcginnis/gofish/redfish"
)

// getVirtualMedia retrieves virtual media resources by first checking the
// Redfish Manager path and falling back to the System path if none are found.
//
// Some BMC implementations (e.g., Dell iDRAC) expose VirtualMedia under the
// System resource (/redfish/v1/Systems/{SystemId}/VirtualMedia) rather than
// the Manager resource (/redfish/v1/Managers/{ManagerId}/VirtualMedia).
// Both locations are valid per the Redfish specification.
func (c *Client) getVirtualMedia(ctx context.Context) ([]*rf.VirtualMedia, error) {
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

// SetVirtualMedia sets virtual media on the system. If mediaURL is empty,
// matching media may be ejected. When multiple matching virtual media slots
// exist, each slot is tried in order until one succeeds.
func (c *Client) SetVirtualMedia(ctx context.Context, kind string, mediaURL string) (bool, error) {
	var mediaKind rf.VirtualMediaType

	switch kind {
	case "CD":
		mediaKind = rf.CDMediaType
	case "Floppy":
		mediaKind = rf.FloppyMediaType
	case "USBStick":
		mediaKind = rf.USBStickMediaType
	case "DVD":
		mediaKind = rf.DVDMediaType
	default:
		return false, errors.New("invalid media type")
	}

	virtualMedia, err := c.getVirtualMedia(ctx)
	if err != nil {
		return false, err
	}

	supportedMediaTypes := []string{}
	var slotErrors []error

	for _, vm := range virtualMedia {
		if !slices.Contains(vm.MediaTypes, mediaKind) {
			for _, mt := range vm.MediaTypes {
				supportedMediaTypes = append(supportedMediaTypes, string(mt))
			}

			continue
		}

		if mediaURL == "" {
			// Only ejecting the media was requested.
			if vm.Inserted && vm.SupportsMediaEject {
				if err := vm.EjectMedia(); err != nil {
					slotErrors = append(slotErrors, fmt.Errorf("%s: eject: %w", vm.ODataID, err))
					continue
				}
			}

			return true, nil
		}

		// Ejecting the media before inserting new media makes the success rate of inserting the new media higher.
		if vm.Inserted && vm.SupportsMediaEject {
			if err := vm.EjectMedia(); err != nil {
				slotErrors = append(slotErrors, fmt.Errorf("%s: eject before insert: %w", vm.ODataID, err))
				continue
			}
		}

		if !vm.SupportsMediaInsert {
			slotErrors = append(slotErrors, fmt.Errorf("%s: does not support insert", vm.ODataID))
			continue
		}

		if err := vm.InsertMedia(mediaURL, true, true); err != nil {
			// Some BMCs (e.g., Supermicro X11SDV-4C-TLN2F) don't support the
			// Inserted and WriteProtected properties, so retry without them.
			if err := vm.InsertMediaConfig(rf.VirtualMediaConfig{Image: mediaURL}); err != nil {
				slotErrors = append(slotErrors, fmt.Errorf("%s: insert: %w", vm.ODataID, err))
				continue
			}
		}

		return true, nil
	}

	if len(slotErrors) > 0 {
		return false, fmt.Errorf("all matching virtual media slots failed: %w", errors.Join(slotErrors...))
	}

	return false, fmt.Errorf("not a supported media type: %s. supported media types: %v", kind, supportedMediaTypes)
}

func (c *Client) InsertedVirtualMedia(ctx context.Context) ([]string, error) {
	virtualMedia, err := c.getVirtualMedia(ctx)
	if err != nil {
		return nil, err
	}

	var inserted []string

	for _, media := range virtualMedia {
		if media.Inserted {
			inserted = append(inserted, media.ID)
		}
	}

	return inserted, nil
}