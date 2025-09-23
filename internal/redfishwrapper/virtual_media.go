package redfishwrapper

import (
	"context"
	"errors"
	"fmt"
	"slices"

	rf "github.com/stmcginnis/gofish/redfish"
)

// Set the virtual media attached to the system, or just eject everything if mediaURL is empty.
func (c *Client) SetVirtualMedia(ctx context.Context, kind string, mediaURL string) (bool, error) {
	managers, err := c.Managers(ctx)
	if err != nil {
		return false, err
	}
	if len(managers) == 0 {
		return false, errors.New("no redfish managers found")
	}

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

	supportedMediaTypes := []string{}
	for _, m := range managers {
		virtualMedia, err := m.VirtualMedia()
		if err != nil {
			return false, err
		}
		if len(virtualMedia) == 0 {
			return false, errors.New("no virtual media found")
		}

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
	}

	return false, fmt.Errorf("not a supported media type: %s. supported media types: %v", kind, supportedMediaTypes)
}

func (c *Client) InsertedVirtualMedia(ctx context.Context) ([]string, error) {
	managers, err := c.Managers(ctx)
	if err != nil {
		return nil, err
	}

	var inserted []string

	for _, m := range managers {
		virtualMedia, err := m.VirtualMedia()
		if err != nil {
			return nil, err
		}

		for _, media := range virtualMedia {
			if media.Inserted {
				inserted = append(inserted, media.ID)
			}
		}
	}

	return inserted, nil
}
