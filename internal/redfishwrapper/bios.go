package redfishwrapper

import (
	"context"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

func (c *Client) GetBiosConfiguration(ctx context.Context) (biosConfig map[string]string, err error) {
	systems, err := c.Systems()
	if err != nil {
		return nil, err
	}

	biosConfig = make(map[string]string)
	for _, sys := range systems {
		if !c.compatibleOdataID(sys.ODataID, knownSystemsOdataIDs) {
			continue
		}

		bios, err := sys.Bios()
		if err != nil {
			return nil, err
		}

		if bios == nil {
			return nil, bmclibErrs.ErrNoBiosAttributes
		}

		for attr := range bios.Attributes {
			biosConfig[attr] = bios.Attributes.String(attr)
		}
	}

	return biosConfig, nil
}
