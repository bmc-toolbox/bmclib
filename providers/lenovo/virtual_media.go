package lenovo

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// compile-time assertion that the provider implements the interface.
var _ bmc.VirtualMediaSetter = (*Conn)(nil)

// validVirtualMediaKinds are the Redfish VirtualMediaType values accepted by
// SetVirtualMedia.
var validVirtualMediaKinds = map[string]bool{
	"CD": true, "DVD": true, "Floppy": true, "USBStick": true,
}

// vmSlot is the subset of a VirtualMedia resource the provider needs.
type vmSlot struct {
	id         string
	odataID    string
	mediaTypes []string
	inserted   bool
	image      string
}

// occupied reports whether the slot currently holds media. Some XCC firmware
// reports a mounted image via a non-empty Image while Inserted briefly lags (or
// vice versa), so both are considered.
func (s vmSlot) occupied() bool {
	return s.inserted || strings.TrimSpace(s.image) != ""
}

// ejectPayload is the Redfish body that ejects media from a VirtualMedia slot.
func ejectPayload() map[string]any {
	return map[string]any{"Image": nil, "Inserted": false}
}

// SetVirtualMedia inserts or ejects remote virtual media on the XCC.
//
// XCC does NOT implement the Redfish VirtualMedia.InsertMedia/EjectMedia
// *actions* on its media slots; it expects the client to PATCH the VirtualMedia
// resource directly (Image/Inserted/WriteProtected/TransferProtocolType), as
// documented in the XCC REST API guide. This method therefore drives the slots
// with PATCH rather than the gofish action helpers (which fail on XCC with
// "does not support insert").
//
// kind is a Redfish virtual-media type: "CD", "DVD", "Floppy" or "USBStick".
// A non-empty mediaURL inserts the image (TransferProtocolType is derived from
// the URL scheme); an empty mediaURL ejects matching media. Slots are selected
// by their advertised MediaTypes, preferring free, remote-capable slots — so a
// remote ISO lands in a "Remote" slot rather than an upload-only "RDOC" slot.
// Implements bmc.VirtualMediaSetter.
func (c *Conn) SetVirtualMedia(ctx context.Context, kind string, mediaURL string) (ok bool, err error) {
	if !validVirtualMediaKinds[kind] {
		return false, fmt.Errorf("invalid virtual media kind %q (want CD|DVD|Floppy|USBStick)", kind)
	}

	slots, err := c.virtualMediaSlots(ctx)
	if err != nil {
		return false, err
	}

	candidates := slotsForKind(slots, kind)
	if len(candidates) == 0 {
		// Some XCC firmware does not populate MediaTypes; fall back to trying
		// every slot rather than failing outright.
		candidates = slots
	}
	if len(candidates) == 0 {
		return false, fmt.Errorf("no virtual media slots found")
	}

	if mediaURL == "" {
		return c.ejectVirtualMedia(ctx, candidates)
	}

	return c.insertVirtualMedia(ctx, candidates, mediaURL)
}

// insertVirtualMedia mounts mediaURL into a matching slot.
//
// To avoid creating a second mount (e.g. media already in EXT2, a new mount
// landing in EXT3), it first ejects every matching slot that is already
// occupied, then inserts into the best-ranked now-free slot. This makes
// repeated inserts idempotent: the device ends up with a single mount.
func (c *Conn) insertVirtualMedia(ctx context.Context, candidates []vmSlot, mediaURL string) (bool, error) {
	proto := transferProtocolType(mediaURL)

	// 1) Clear any media already present in matching slots.
	for i := range candidates {
		if !candidates[i].occupied() {
			continue
		}
		if err := c.patchVirtualMedia(ctx, candidates[i].odataID, ejectPayload()); err == nil {
			candidates[i].inserted = false
			candidates[i].image = ""
		}
	}

	// 2) Insert into the best now-free slot (free + remote-capable first).
	rankSlots(candidates)

	var slotErrs []string
	for _, s := range candidates {
		payload := map[string]any{
			"Image":          mediaURL,
			"Inserted":       true,
			"WriteProtected": true,
		}
		if proto != "" {
			payload["TransferProtocolType"] = proto
		}

		if err := c.patchVirtualMedia(ctx, s.odataID, payload); err == nil {
			return true, nil
		} else {
			// Retry with a minimal payload — some XCC firmware rejects the
			// Inserted/WriteProtected/TransferProtocolType properties.
			minimal := map[string]any{"Image": mediaURL, "Inserted": true}
			if err2 := c.patchVirtualMedia(ctx, s.odataID, minimal); err2 == nil {
				return true, nil
			}
			slotErrs = append(slotErrs, fmt.Sprintf("%s: %v", s.odataID, err))
		}
	}

	return false, fmt.Errorf("failed to insert virtual media into any matching slot:\n%s", strings.Join(slotErrs, "\n"))
}

// ejectVirtualMedia PATCHes any inserted candidate slot to eject its media.
func (c *Conn) ejectVirtualMedia(ctx context.Context, candidates []vmSlot) (bool, error) {
	var slotErrs []string
	var attempted bool

	for _, s := range candidates {
		if !s.occupied() {
			continue
		}
		attempted = true
		if err := c.patchVirtualMedia(ctx, s.odataID, ejectPayload()); err != nil {
			slotErrs = append(slotErrs, fmt.Sprintf("%s: %v", s.odataID, err))
		}
	}

	// Nothing was inserted — ejecting is idempotent, so report success.
	if !attempted {
		return true, nil
	}
	if len(slotErrs) > 0 {
		return false, fmt.Errorf("failed to eject virtual media:\n%s", strings.Join(slotErrs, "\n"))
	}

	return true, nil
}

// patchVirtualMedia PATCHes a VirtualMedia resource and maps the response.
func (c *Conn) patchVirtualMedia(ctx context.Context, odataID string, payload map[string]any) error {
	return checkResponse(c.redfishwrapper.PatchWithHeaders(ctx, odataID, payload, nil))
}

// virtualMediaSlots resolves and reads the VirtualMedia collection, trying the
// Manager path first and falling back to the System path.
func (c *Conn) virtualMediaSlots(ctx context.Context) ([]vmSlot, error) {
	var collectionURL string
	if manager, err := c.redfishwrapper.Manager(ctx); err == nil {
		collectionURL = singleTrailingSlashJoin(manager.ODataID, "VirtualMedia")
	}

	members, err := c.collectionMembersOrEmpty(collectionURL)
	if len(members) == 0 {
		if system, serr := c.redfishwrapper.System(); serr == nil {
			collectionURL = singleTrailingSlashJoin(system.ODataID, "VirtualMedia")
			members, err = c.collectionMembersOrEmpty(collectionURL)
		}
	}
	if len(members) == 0 {
		if err != nil {
			return nil, fmt.Errorf("listing virtual media: %w", err)
		}
		return nil, fmt.Errorf("no virtual media slots found at Manager or System paths")
	}

	slots := make([]vmSlot, 0, len(members))
	for _, m := range members {
		var doc struct {
			ID         string   `json:"Id"`
			ODataID    string   `json:"@odata.id"`
			MediaTypes []string `json:"MediaTypes"`
			Inserted   *bool    `json:"Inserted"`
			Image      string   `json:"Image"`
		}
		if err := c.getJSON(m.ODataID, &doc); err != nil {
			return nil, err
		}

		odataID := doc.ODataID
		if odataID == "" {
			odataID = m.ODataID
		}
		slots = append(slots, vmSlot{
			id:         doc.ID,
			odataID:    odataID,
			mediaTypes: doc.MediaTypes,
			inserted:   doc.Inserted != nil && *doc.Inserted,
			image:      doc.Image,
		})
	}

	return slots, nil
}

// collectionMembersOrEmpty returns a collection's members, or an empty slice
// (and the error) when the URL is empty or the GET fails.
func (c *Conn) collectionMembersOrEmpty(url string) ([]odataID, error) {
	if url == "" {
		return nil, nil
	}
	return c.collectionMembers(url)
}

// slotsForKind returns the slots whose MediaTypes advertise the requested kind.
func slotsForKind(slots []vmSlot, kind string) []vmSlot {
	var out []vmSlot
	for _, s := range slots {
		for _, mt := range s.mediaTypes {
			if mt == kind {
				out = append(out, s)
				break
			}
		}
	}
	return out
}

// rankSlots orders slots best-first for an insert: free slots before occupied
// ones, and remote-capable ("Remote*") slots before generic ones before
// upload-only ("RDOC*") slots.
func rankSlots(slots []vmSlot) {
	score := func(s vmSlot) int {
		sc := 0
		if !s.inserted {
			sc += 2
		}
		id := strings.ToLower(s.id)
		switch {
		case strings.HasPrefix(id, "remote"):
			sc += 4
		case strings.Contains(id, "rdoc"):
			sc += 0
		default:
			sc += 1
		}
		return sc
	}

	sort.SliceStable(slots, func(i, j int) bool { return score(slots[i]) > score(slots[j]) })
}

// transferProtocolType derives the Redfish TransferProtocolType from a media URL
// scheme. It returns "" when the scheme is unknown (the XCC then infers it).
func transferProtocolType(mediaURL string) string {
	lower := strings.ToLower(mediaURL)
	switch {
	case strings.HasPrefix(lower, "https://"):
		return "HTTPS"
	case strings.HasPrefix(lower, "http://"):
		return "HTTP"
	case strings.HasPrefix(lower, "nfs://"):
		return "NFS"
	case strings.HasPrefix(lower, "cifs://"), strings.HasPrefix(lower, "smb://"):
		return "CIFS"
	default:
		return ""
	}
}
