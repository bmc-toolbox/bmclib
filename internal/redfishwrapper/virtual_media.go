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

// SetVirtualMedia sets virtual media on the system. If mediaURL is empty,
// matching media may be ejected. When multiple matching virtual media slots
// exist, each slot is tried in order until one succeeds.
func (c *Client) SetVirtualMedia(ctx context.Context, kind string, mediaURL string) (bool, error) {
	var mediaKind schemas.VirtualMediaType

	switch kind {
	case "CD":
		mediaKind = schemas.CDVirtualMediaType
	case "Floppy":
		mediaKind = schemas.FloppyVirtualMediaType
	case "USBStick":
		mediaKind = schemas.USBStickVirtualMediaType
	case "DVD":
		mediaKind = schemas.DVDVirtualMediaType
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
			// Attempt to eject media.
			// The gofish library v0.22.0+ supports fallback to PATCH if standard Actions are not available.
			// This logic first checks if media is inserted, then attempts eject via Action or PATCH.
			// This is relevant for BMCs like Lenovo XCC that may not expose standard EjectMedia actions.
			if gofish.Deref(vm.Inserted) {
				if _, err := vm.EjectMedia(); err != nil {
					slotErrors = append(slotErrors, fmt.Errorf("%s: eject: %w", vm.ODataID, err))
					continue
				}
			}

			return true, nil
		}

		// Ejecting the media before inserting new media makes the success rate of inserting the new media higher.
		if gofish.Deref(vm.Inserted) {
			// Attempt to eject media. If it fails, record the error but continue to attempt insertion,
			// as some BMCs might allow insert even if eject fails or auto-replace media.
			if _, err := vm.EjectMedia(); err != nil {
				slotErrors = append(slotErrors, fmt.Errorf("%s: eject before insert failed: %w. Attempting insert anyway.", vm.ODataID, err))
			}
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
		if gofish.Deref(media.Inserted) {
			inserted = append(inserted, media.ID)
		}
	}

	return inserted, nil
}
