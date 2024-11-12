package bmc

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
	"github.com/pkg/errors"
)

// ScreenshotGetter interface provides methods to query for a BMC screen capture.
type ScreenshotGetter interface {
	Screenshot(ctx context.Context) (image []byte, fileType string, err error)
}

type screenshotGetterProvider struct {
	name string
	ScreenshotGetter
}

// screenshot returns an image capture of the video output.
func screenshot(ctx context.Context, generic []screenshotGetterProvider) (image []byte, fileType string, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range generic {
		if elem.ScreenshotGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return image, fileType, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			image, fileType, vErr := elem.Screenshot(ctx)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue

			}
			metadataLocal.SuccessfulProvider = elem.name
			return image, fileType, metadataLocal, nil
		}
	}

	return image, fileType, metadataLocal, multierror.Append(err, errors.New("failed to capture screenshot"))
}

// ScreenshotFromInterfaces identifies implementations of the ScreenshotGetter interface and passes the found implementations to the screenshot() wrapper method.
func ScreenshotFromInterfaces(ctx context.Context, generic []interface{}) (image []byte, fileType string, metadata Metadata, err error) {
	implementations := make([]screenshotGetterProvider, 0)
	for _, elem := range generic {
		temp := screenshotGetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case ScreenshotGetter:
			temp.ScreenshotGetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a ScreenshotGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return image, fileType, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no ScreenshotGetter implementations found"),
			),
		)
	}

	return screenshot(ctx, implementations)
}
