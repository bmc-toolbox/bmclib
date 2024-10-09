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

	for _, m := range managers {
		virtualMedia, err := m.VirtualMedia()
		if err != nil {
			return false, err
		}
		if len(virtualMedia) == 0 {
			return false, errors.New("no virtual media found")
		}

		for _, vm := range virtualMedia {
			if vm.Inserted {
				if err := vm.EjectMedia(); err != nil {
					return false, err
				}
			}
			if mediaURL == "" {
				// Only ejecting the media was requested.
				return true, nil
			}
			if !slices.Contains(vm.MediaTypes, mediaKind) {
				return false, fmt.Errorf("media kind %s not supported by BMC, supported media kinds %q", kind, vm.MediaTypes)
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

	// If we actual get here, then something very unexpected happened as there isn't a known code path that would cause this error to be returned.
	return false, errors.New("unexpected error setting virtual media")
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
