package main

import (
	"context"
	"time"

	"github.com/bmc-toolbox/bmclib/v2"
	"github.com/bmc-toolbox/bmclib/v2/logging"
	"github.com/bmc-toolbox/bmclib/v2/providers/homeassistant"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	// Start the test consumer
	time.Sleep(100 * time.Millisecond)

	log := logging.ZeroLogger("info")
	opts := []bmclib.Option{
		bmclib.WithLogger(log),
		//bmclib.WithPerProviderTimeout(5 * time.Second),
		bmclib.WithHomeAssistantOpt(homeassistant.Config{
			SwitchEntityID:             "switch.shellypstripg4_98a3167b747c_switch_0",
			PowerOperationDelaySeconds: 2,
		}),
	}
	host := "http://some.homeassistant.instance:8123"
	user := "notuseduser"
	pass := "ey.....hk"
	c := bmclib.NewClient(host, user, pass, opts...)
	if err := c.Open(ctx); err != nil {
		panic(err)
	}
	defer c.Close(ctx)

	ok3, err := c.SetBootDevice(ctx, "pxe", false, false)
	if err != nil {
		panic(err)
	}
	log.Info("set boot device", "ok3", ok3)

	state, err := c.GetPowerState(ctx)
	if err != nil {
		panic(err)
	}
	log.Info("power state", "state", state)
	log.Info("metadata for GetPowerState", "metadata", c.GetMetadata())

	ok, err := c.SetPowerState(ctx, "on")
	if err != nil {
		panic(err)
	}
	log.Info("set power state ON", "ok", ok)
	log.Info("metadata for SetPowerState ON", "metadata", c.GetMetadata())

	ok2, err := c.SetPowerState(ctx, "off")
	if err != nil {
		panic(err)
	}
	log.Info("set power state OFF", "ok2", ok2)
	log.Info("metadata for SetPowerState OFF", "metadata", c.GetMetadata())

	<-ctx.Done()
}
