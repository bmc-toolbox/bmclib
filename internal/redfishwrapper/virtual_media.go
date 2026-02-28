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

// Set the virtual media attached to the system, or just eject everything if mediaURL is empty.
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
					return false, fmt.Errorf("error ejecting media: %v", err)
				}
			}

			return true, nil
		}

		// Ejecting the media before inserting a new new media makes the success rate of inserting the new media higher.
		if vm.Inserted && vm.SupportsMediaEject {
			if err := vm.EjectMedia(); err != nil {
				return false, fmt.Errorf("error ejecting media before inserting media: %v", err)
			}
		}

		if !vm.SupportsMediaInsert {
			return false, fmt.Errorf("BMC does not support inserting virtual media of kind: %s", kind)
		}

		if err := vm.InsertMedia(mediaURL, true, true); err != nil {
			// Some BMC's (Supermicro X11SDV-4C-TLN2F, for example) don't support the "inserted" and "writeProtected" properties,
			// so we try to insert the media without them if the first attempt fails.
			if err := vm.InsertMediaConfig(rf.VirtualMediaConfig{Image: mediaURL}); err != nil {
				return false, err
			}
		}

		return true, nil
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