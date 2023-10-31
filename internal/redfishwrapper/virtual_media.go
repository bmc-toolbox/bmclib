package redfishwrapper

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	rf "github.com/stmcginnis/gofish/redfish"
)

// Set the virtual media attached to the system, or just eject everything if mediaURL is empty.
func (c *Client) SetVirtualMedia(ctx context.Context, kind string, mediaURL string) (ok bool, err error) {
	managers, err := c.Managers(ctx)
	if err != nil {
		return false, err
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

	for _, manager := range managers {
		virtualMedia, err := manager.VirtualMedia()
		if err != nil {
			return false, err
		}
		for _, media := range virtualMedia {
			if media.Inserted {
				err = media.EjectMedia()
				if err != nil {
					return false, err
				}
			}
		}
	}

	// An empty mediaURL means eject everything, so if that's the case we're done. Otherwise, we
	// need to insert the media.
	if mediaURL != "" {
		setMedia := false
		for _, manager := range managers {
			virtualMedia, err := manager.VirtualMedia()
			if err != nil {
				return false, err
			}

			for _, media := range virtualMedia {
				for _, t := range media.MediaTypes {
					if t == mediaKind {
						err = media.InsertMedia(mediaURL, true, true)
						if err != nil {
							return false, err
						}
						setMedia = true
						break
					}
				}
			}
		}
		if !setMedia {
			return false, fmt.Errorf("media kind %s not supported", kind)
		}
	}

	return true, nil
}

func (c *Client) InsertedVirtualMedia(ctx context.Context) ([]string, error) {
	managers, err := c.Managers(ctx)
	if err != nil {
		return nil, err
	}

	var inserted []string

	for _, manager := range managers {
		virtualMedia, err := manager.VirtualMedia()
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
