package bmc

import (
	"context"
	"fmt"
	"io"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
)

// FloppyImageUploader defines methods to upload a floppy image
type FloppyImageUploader interface {
	UploadFloppyImage(ctx context.Context, image io.Reader) (err error)
}

// floppyImageUploaderProvider is an internal struct to correlate an implementation/provider and its name
type floppyImageUploaderProvider struct {
	name string
	impl FloppyImageUploader
}

// uploadFloppyImage is a wrapper method to invoke methods for the FloppyImageUploader interface
func uploadFloppyImage(ctx context.Context, image io.Reader, p []floppyImageUploaderProvider) (metadata Metadata, err error) {
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
			uploadErr := elem.impl.UploadFloppyImage(ctx, image)
			if uploadErr != nil {
				err = multierror.Append(err, errors.WithMessagef(uploadErr, "provider: %v", elem.name))
				continue
			}

			metadataLocal.SuccessfulProvider = elem.name
			return metadataLocal, nil
		}
	}

	return metadataLocal, multierror.Append(err, errors.New("failed to upload floppy image"))
}

// UploadFloppyImageFromInterfaces identifies implementations of the FloppyImageUploader interface and passes the found implementations to the uploadFloppyImage() wrapper
func UploadFloppyImageFromInterfaces(ctx context.Context, image io.Reader, p []interface{}) (metadata Metadata, err error) {
	providers := make([]floppyImageUploaderProvider, 0)
	for _, elem := range p {
		temp := floppyImageUploaderProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case FloppyImageUploader:
			temp.impl = p
			providers = append(providers, temp)
		default:
			e := fmt.Sprintf("not a FloppyImageUploader implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}

	if len(providers) == 0 {
		return metadata, multierror.Append(err, errors.New("no FloppyImageUploader implementations found"))
	}

	return uploadFloppyImage(ctx, image, providers)
}

// FloppyImageUploader defines methods to unmount a floppy image
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

// UploadFloppyImageFromInterfaces identifies implementations of the FloppyImageUnmounter interface and passes the found implementations to the unmountFloppyImage() wrapper
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
		return metadata, multierror.Append(err, errors.New("no FloppyImageUnmounter implementations found"))
	}

	return unmountFloppyImage(ctx, providers)
}
