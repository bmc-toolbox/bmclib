package supermicro

import (
	"context"
	"strings"

	"github.com/pkg/errors"
)

type x11 struct{ *Client }



func (c *x11) deviceModel(ctx context.Context) (string, error) {
	errBoardPartNumUnknown := errors.New("baseboard part number unknown")
	data, err := c.fruInfo(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			return "", ErrXMLAPIUnsupported
		}

		return "", err
	}

	partNum := strings.TrimSpace(data.Board.PartNum)

	if data.Board == nil || partNum == "" {
		return "", errors.Wrap(errBoardPartNumUnknown, "baseboard part number empty")
	}

	return partNum, nil
}
