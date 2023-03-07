package redfish

import (
	"context"

	gofishrf "github.com/stmcginnis/gofish/redfish"
)

func (c *Conn) GetBiosConfiguration(ctx context.Context) (biosConfig map[string]string, err error) {
	// TODO: replace uri parameter with configurable value
	b, err := gofishrf.GetBios(c.redfishwrapper.Client, "/redfish/v1/Systems/System.Embedded.1/Bios")
	if err != nil {
		return nil, err
	}

	biosConfig = make(map[string]string)
	for k := range b.Attributes {
		biosConfig[k] = b.Attributes.String(k)
	}
	return biosConfig, nil
}
