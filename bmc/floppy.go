package bmc

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/go-multierror"
	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
	"github.com/pkg/errors"
)

// FloppyImageMounter defines methods to upload a floppy image
type FloppyImageMounter interface {
	MountFloppyImage(ctx context.Context, image io.Reader) (err error)
}

// floppyImageUploaderProvider is an internal struct to correlate an implementation/provider and its name
type floppyImageUploaderProvider struct {
	name string
	impl FloppyImageMounter
}

// mountFloppyImage is a wrapper method to invoke methods for the FloppyImageMounter interface
func mountFloppyImage(ctx context.Context, image io.Reader, p []floppyImageUploaderProvider) (metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range p {
		if elem.impl == nil {
			continue
		}

		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			uploadErr := elem.impl.MountFloppyImage(ctx, image)
			if uploadErr != nil {
				err = multierror.Append(err, errors.WithMessagef(uploadErr, "provider: %v", elem.name))
				continue
			}

			metadataLocal.SuccessfulProvider = elem.name
			return metadataLocal, nil
		}
	}

	return metadataLocal, multierror.Append(err, errors.New("failed to mount floppy image"))
}

// MountFloppyImageFromInterfaces identifies implementations of the FloppyImageMounter interface and passes the found implementations to the mountFloppyImage() wrapper
func MountFloppyImageFromInterfaces(ctx context.Context, image io.Reader, p []interface{}) (metadata Metadata, err error) {
	providers := make([]floppyImageUploaderProvider, 0)
	for _, elem := range p {
		temp := floppyImageUploaderProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case FloppyImageMounter:
			temp.impl = p
			providers = append(providers, temp)
		default:
			e := fmt.Sprintf("not a FloppyImageMounter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}

	if len(providers) == 0 {
		return metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				"no FloppyImageMounter implementations found",
			),
		)

	}

	return mountFloppyImage(ctx, image, providers)
}

// FloppyImageMounter defines methods to unmount a floppy image
type FloppyImageUnmounter interface {
	UnmountFloppyImage(ctx context.Context) (err error)
}

// floppyImageUnmounterProvider is an internal struct to correlate an implementation/provider and its name
type floppyImageUnmounterProvider struct {
	name string
	impl FloppyImageUnmounter
}

// unmountFloppyImage is a wrapper method to invoke methods for the FloppyImageUnmounter interface
func unmountFloppyImage(ctx context.Context, p []floppyImageUnmounterProvider) (metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range p {
		if elem.impl == nil {
			continue
		}

		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			uploadErr := elem.impl.UnmountFloppyImage(ctx)
			if uploadErr != nil {
				err = multierror.Append(err, errors.WithMessagef(uploadErr, "provider: %v", elem.name))
				continue
			}

			metadataLocal.SuccessfulProvider = elem.name
			return metadataLocal, nil
		}
	}

	return metadataLocal, multierror.Append(err, errors.New("failed to unmount floppy image"))
}

// MountFloppyImageFromInterfaces identifies implementations of the FloppyImageUnmounter interface and passes the found implementations to the unmountFloppyImage() wrapper
func UnmountFloppyImageFromInterfaces(ctx context.Context, p []interface{}) (metadata Metadata, err error) {
	providers := make([]floppyImageUnmounterProvider, 0)
	for _, elem := range p {
		temp := floppyImageUnmounterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case FloppyImageUnmounter:
			temp.impl = p
			providers = append(providers, temp)
		default:
			e := fmt.Sprintf("not a FloppyImageUnmounter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}

	if len(providers) == 0 {
		return metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				"no FloppyImageUnmounter implementations found",
			),
		)
	}

	return unmountFloppyImage(ctx, providers)
}
