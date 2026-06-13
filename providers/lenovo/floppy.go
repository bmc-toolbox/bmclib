package lenovo

import (
	"context"
	"errors"
	"io"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
)

// compile-time assertions that the provider implements the interfaces.
var (
	_ bmc.FloppyImageMounter   = (*Conn)(nil)
	_ bmc.FloppyImageUnmounter = (*Conn)(nil)
)

// errFloppyUploadUnsupported is returned by MountFloppyImage: the XCC mounts
// virtual media by URL and its documented Redfish API exposes no endpoint to
// upload a raw floppy image.
var errFloppyUploadUnsupported = errors.New(
	"lenovo XCC mounts virtual media by URL and does not support uploading a raw floppy image; " +
		"host the image and use SetVirtualMedia(ctx, \"Floppy\", url) instead")

// MountFloppyImage is not supported on the XCC.
//
// Unlike vendors that accept a multipart image upload, the XCC's Redfish API
// only mounts virtual media referenced by a URL. This method therefore returns a
// descriptive error directing the caller to host the image and use
// [Conn.SetVirtualMedia] with a Floppy media URL. Implements
// bmc.FloppyImageMounter.
func (c *Conn) MountFloppyImage(ctx context.Context, image io.Reader) error {
	return errFloppyUploadUnsupported
}

// UnmountFloppyImage ejects the media mounted in the XCC Floppy virtual-media
// slot.
//
// Implements bmc.FloppyImageUnmounter.
func (c *Conn) UnmountFloppyImage(ctx context.Context) error {
	// Use the provider's PATCH-based SetVirtualMedia (XCC does not expose the
	// EjectMedia action), not the gofish action-based wrapper.
	_, err := c.SetVirtualMedia(ctx, "Floppy", "")
	return err
}
