## bmclib v2 - board management controller library

[![Status](https://github.com/bmc-toolbox/bmclib/actions/workflows/ci.yaml/badge.svg)](https://github.com/bmc-toolbox/bmclib/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/bmc-toolbox/bmclib)](https://goreportcard.com/report/github.com/bmc-toolbox/bmclib/v2)
[![GoDoc](https://godoc.org/github.com/bmc-toolbox/bmclib/v2?status.svg)](https://godoc.org/github.com/bmc-toolbox/bmclib/v2)

bmclib v2 is a library to abstract interacting with baseboard management controllers.

### Supported BMC interfaces.

 - [Redfish](https://github.com/bmc-toolbox/bmclib/tree/main/providers/redfish)
 - [IPMItool](https://github.com/bmc-toolbox/bmclib/tree/main/providers/ipmitool)
 - [Intel AMT](https://github.com/bmc-toolbox/bmclib/tree/main/providers/intelamt)
 - [Asrockrack](https://github.com/bmc-toolbox/bmclib/tree/main/providers/asrockrack)

### Installation

```bash
go get github.com/bmc-toolbox/bmclib/v2
```

### Import 

```go
import (
  bmclib "github.com/bmc-toolbox/bmclib/v2"
)
```

### Usage

The snippet below connects to a BMC and retrieves the device hardware, firmware inventory.

```go
import (
  bmclib "github.com/bmc-toolbox/bmclib/v2"
)

    // setup logger
    l := logrus.New()
    l.Level = logrus.DebugLevel
    logger := logrusr.New(l)

    clientOpts := []bmclib.Option{bmclib.WithLogger(logger)}

    // init client
    client := bmclib.NewClient(*host, "", "admin", "hunter2", clientOpts...)

    // open BMC session
    err := client.Open(ctx)
    if err != nil {
        log.Fatal(err, "bmc login failed")
    }

    defer client.Close(ctx)

    // retrieve inventory data
    inventory, err := client.Inventory(ctx)
    if err != nil {
        l.Error(err)
    }

    b, err := json.MarshalIndent(inventory, "", "  ")
    if err != nil {
        l.Error(err)
    }

    fmt.Println(string(b))
```

More sample code can be found in [examples](./examples/)

### BMC connections

bmclib performs queries on BMCs using [multiple `drivers`](https://github.com/bmc-toolbox/bmclib/blob/main/bmc/connection.go#L30),
these `drivers` are the various services exposed by a BMC - `redfish` `IPMI` `SSH` and `vendor API` which is basically a custom vendor API endpoint.

The bmclib client determines which driver to use for an action like `Power cycle` or `Create user`
based on its availability or through a compatibility test (when enabled).

When querying multiple BMCs through bmclib its often useful to to limit the BMCs and
drivers that bmclib will attempt to use to connect, the options to limit or filter
out BMCs are described below,

Query just using the `redfish` endpoint.
```
cl := bmclib.NewClient("192.168.1.1", "", "admin", "hunter2")
cl.Registry.Drivers = cl.Registry.Using("redfish")
```

Query using the `redfish` endpoint and fall back to `IPMI`
```
client := bmclib.NewClient("192.168.1.1", "", "admin", "hunter2")

// overwrite registered drivers by appending Redfish, IPMI drivers in order
drivers := append(registrar.Drivers{}, bmcClient.Registry.Using("redfish")...)
drivers = append(drivers, bmcClient.Registry.Using("ipmi")...)
client.Registry.Drivers = driver
```

Filter drivers to query based on compatibility, this will attempt to check if the driver is
[compatible](https://github.com/bmc-toolbox/bmclib/blob/main/providers/redfish/redfish.go#L70)
ideally, this method should be invoked when the client is ready to perform a BMC action.
```
client := bmclib.NewClient("192.168.1.1", "", "admin", "hunter2")
client.Registry.Drivers = cl.Registry.FilterForCompatible(ctx)
```

Ignore the Redfish endpoint completely on BMCs running a specific Redfish version.

Note: this version should match the one returned through `curl -k  "https://<BMC IP>/redfish/v1" | jq .RedfishVersion`
```
opt := bmclib.WithRedfishVersionsNotCompatible([]string{"1.5.0"})

client := bmclib.NewClient("192.168.1.1", "", "admin", "hunter2", opt...)
cl.Registry.Drivers = cl.Registry.FilterForCompatible(ctx)
```


### bmclib versions

The current bmclib version is `v2` and is being developed on the `main` branch.

The previous bmclib version is in maintenance mode and can be found here [v1](https://github.com/bmc-toolbox/bmclib/v1).

### Acknowledgments

bmclib v2 interfaces with Redfish on BMCs through the Gofish library https://github.com/stmcginnis/gofish

bmclib was originally developed for [Booking.com](http://www.booking.com). With approval from [Booking.com](http://www.booking.com), 
the code and specification were generalized and published as Open Source on github, for which the authors would like to express their gratitude.

### Authors
- [Joel Rebello](https://github.com/joelrebel) 
- [Jacob Weinstock](https://github.com/jacobweinstock)

