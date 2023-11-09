package supermicro

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

var (
	errFloppyImageMounted = errors.New("floppy image is currently mounted")
)

func (c *Client) floppyImageMounted(ctx context.Context) (bool, error) {
	if err := c.openRedfish(ctx); err != nil {
		return false, err
	}

	inserted, err := c.redfish.InsertedVirtualMedia(ctx)
	if err != nil {
		return false, err
	}

	for _, media := range inserted {
		if strings.Contains(strings.ToLower(media), "floppy") {
			return true, nil
		}
	}

	return false, nil
}

func (c *Client) MountFloppyImage(ctx context.Context, image io.Reader) error {
	mounted, err := c.floppyImageMounted(ctx)
	if err != nil {
		return err
	}

	if mounted {
		return errFloppyImageMounted
	}

	var payloadBuffer bytes.Buffer

	type form struct {
		name string
		data io.Reader
	}

	formParts := []form{
		{
			name: "img_file",
			data: image,
		},
	}

	if c.csrfToken != "" {
		formParts = append(formParts, form{
			name: "csrf-token",
			data: bytes.NewBufferString(c.csrfToken),
		})
	}

	payloadWriter := multipart.NewWriter(&payloadBuffer)

	for _, part := range formParts {
		var partWriter io.Writer

		switch part.name {
		case "img_file":
			file, ok := part.data.(*os.File)
			if !ok {
				return errors.Wrap(ErrMultipartForm, "expected io.Reader for a floppy image file")
			}

			if partWriter, err = payloadWriter.CreateFormFile(part.name, filepath.Base(file.Name())); err != nil {
				return errors.Wrap(ErrMultipartForm, err.Error())
			}

		case "csrf-token":
			// Add csrf token field
			h := make(textproto.MIMEHeader)
			// BMCs with newer firmware (>=1.74.09) accept the form with this name value
			// h.Set("Content-Disposition", `form-data; name="CSRF-TOKEN"`)
			//
			// the BMCs running older firmware (<=1.23.06) versions expects the name value in this format
			// and the newer firmware (>=1.74.09) seem to be backwards compatible with this name value format.
			h.Set("Content-Disposition", `form-data; name="CSRF_TOKEN"`)

			if partWriter, err = payloadWriter.CreatePart(h); err != nil {
				return errors.Wrap(ErrMultipartForm, err.Error())
			}
		default:
			return errors.Wrap(ErrMultipartForm, "unexpected form part: "+part.name)
		}

		if _, err = io.Copy(partWriter, part.data); err != nil {
			return err
		}
	}
	payloadWriter.Close()

	resp, statusCode, err := c.query(
		ctx,
		"cgi/uimapin.cgi",
		http.MethodPost,
		bytes.NewReader(payloadBuffer.Bytes()),
		map[string]string{"Content-Type": payloadWriter.FormDataContentType()},
		0,
	)

	if err != nil {
		return errors.Wrap(ErrMultipartForm, err.Error())
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("non 200 response: %d %s", statusCode, resp)
	}

	return nil
}

func (c *Client) UnmountFloppyImage(ctx context.Context) error {
	mounted, err := c.floppyImageMounted(ctx)
	if err != nil {
		return err
	}

	if !mounted {
		return nil
	}

	resp, statusCode, err := c.query(
		ctx,
		"cgi/uimapout.cgi",
		http.MethodPost,
		nil,
		nil,
		0,
	)

	if err != nil {
		return err
	}

	if statusCode != http.StatusOK {
		return fmt.Errorf("non 200 response: %d %s", statusCode, resp)
	}

	return nil
}
