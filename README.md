# bmclib v2 - board management controller library

[![Status](https://github.com/metal-toolbox/bmclib/actions/workflows/ci.yaml/badge.svg)](https://github.com/metal-toolbox/bmclib/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/metal-toolbox/bmclib)](https://goreportcard.com/report/github.com/metal-toolbox/bmclib)
[![GoDoc](https://godoc.org/github.com/metal-toolbox/bmclib?status.svg)](https://godoc.org/github.com/metal-toolbox/bmclib)

bmclib v2 is a library to abstract interacting with baseboard management controllers.

## Supported BMC interfaces.

 - [Redfish](https://github.com/metal-toolbox/bmclib/tree/main/providers/redfish)
 - [IPMItool](https://github.com/metal-toolbox/bmclib/tree/main/providers/ipmitool)
 - [Intel AMT](https://github.com/metal-toolbox/bmclib/tree/main/providers/intelamt)
 - [Asrockrack](https://github.com/metal-toolbox/bmclib/tree/main/providers/asrockrack)
 - [RPC](providers/rpc/)

## Installation

```bash
go get github.com/metal-toolbox/bmclib
```

## Import

```go
import (
  bmclib "github.com/metal-toolbox/bmclib"
)
```

### Usage

The snippet below connects to a BMC and retrieves the device hardware, firmware inventory.

```go
import (
  bmclib "github.com/metal-toolbox/bmclib"
)

    // setup logger
    l := logrus.New()
    l.Level = logrus.DebugLevel
    logger := logrusr.New(l)

    clientOpts := []bmclib.Option{bmclib.WithLogger(logger)}

    // init client
    client := bmclib.NewClient(*host, "admin", "hunter2", clientOpts...)

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

## BMC connections

bmclib performs queries on BMCs using [multiple `drivers`](https://github.com/metal-toolbox/bmclib/blob/main/bmc/connection.go#L30),
these `drivers` are the various services exposed by a BMC - `redfish` `IPMI` `SSH` and `vendor API` which is basically a custom vendor API endpoint.

The bmclib client determines which driver to use for an action like `Power cycle` or `Create user`
based on its availability or through a compatibility test (when enabled).

When querying multiple BMCs through bmclib its often useful to to limit the BMCs and
drivers that bmclib will attempt to use to connect, the options to limit or filter
out BMCs are described below,

Query just using the `redfish` endpoint.

```go
cl := bmclib.NewClient("192.168.1.1", "admin", "hunter2")
cl.Registry.Drivers = cl.Registry.Using("redfish")
```

Query using the `redfish` endpoint and fall back to `IPMI`

```go
client := bmclib.NewClient("192.168.1.1", "admin", "hunter2")

// overwrite registered drivers by appending Redfish, IPMI drivers in order
drivers := append(registrar.Drivers{}, bmcClient.Registry.Using("redfish")...)
drivers = append(drivers, bmcClient.Registry.Using("ipmi")...)
client.Registry.Drivers = driver
```

Filter drivers to query based on compatibility, this will attempt to check if the driver is
[compatible](https://github.com/metal-toolbox/bmclib/blob/main/providers/redfish/redfish.go#L70)
ideally, this method should be invoked when the client is ready to perform a BMC action.

```go
client := bmclib.NewClient("192.168.1.1", "admin", "hunter2")
client.Registry.Drivers = cl.Registry.FilterForCompatible(ctx)
```

Ignore the Redfish endpoint completely on BMCs running a specific Redfish version.

Note: this version should match the one returned through `curl -k  "https://<BMC IP>/redfish/v1" | jq .RedfishVersion`

```go
opt := bmclib.WithRedfishVersionsNotCompatible([]string{"1.5.0"})

client := bmclib.NewClient("192.168.1.1", "admin", "hunter2", opt...)
cl.Registry.Drivers = cl.Registry.FilterForCompatible(ctx)
```

## Timeouts

bmclib can be configured to apply timeouts to BMC interactions. The following options are available.

**Total max timeout only** - The total time bmclib will wait for all BMC interactions to complete. This is specified using a single `context.WithTimeout` or `context.WithDeadline` that is passed to all method call. With this option, the per provider; per interaction timeout is calculated by the total max timeout divided by the number of providers (currently there are 4 providers).

```go
cl := bmclib.NewClient(host, user, pass, bmclib.WithLogger(log))

ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

if err = cl.Open(ctx); err != nil {
  return(err)
}
defer cl.Close(ctx)

state, err := cl.GetPowerState(ctx)
```

**Total max timeout and a per provider; per interaction timeout** - The total time bmclib will wait for all BMC interactions to complete. This is specified using a single `context.WithTimeout` or `context.WithDeadline` that is passed to all method call. This is honored above all timeouts. The per provider; per interaction timeout is specified using `bmclib.WithPerProviderTimeout` in the Client constructor.

```go
cl := bmclib.NewClient(host, user, pass, bmclib.WithLogger(log), bmclib.WithPerProviderTimeout(15*time.Second))

ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
defer cancel()

if err = cl.Open(ctx); err != nil {
  return(err)
}
defer cl.Close(ctx)

state, err := cl.GetPowerState(ctx)
```

**Per provider; per interaction timeout. No total max timeout** - The time bmclib will wait for a specific provider to complete. This is specified using `bmclib.WithPerProviderTimeout` in the Client constructor.

```go
cl := bmclib.NewClient(host, user, pass, bmclib.WithLogger(log), bmclib.WithPerProviderTimeout(15*time.Second))

ctx := context.Background()

if err = cl.Open(ctx); err != nil {
  return(err)
}
defer cl.Close(ctx)

state, err := cl.GetPowerState(ctx)
```

**Default timeout** - If no timeout is specified with a context or with `bmclib.WithPerProviderTimeout` the default is used. 30 seconds per provider; per interaction.

```go
cl := bmclib.NewClient(host, user, pass, bmclib.WithLogger(log))

ctx := context.Background()

if err = cl.Open(ctx); err != nil {
  return(err)
}
defer cl.Close(ctx)

state, err := cl.GetPowerState(ctx)
```

## Filtering

The `bmclib.Client` can be configured to filter BMC calls based on a few different criteria. Filtering modifies the order and/or the number of providers for BMC calls. This filtering can be permanent or on a one-time basis.

All providers are stored in a registry (see [`Client.Registry`](https://github.com/metal-toolbox/bmclib/blob/b5cdfa3ffe026d3cc3257953abe3234b278ca20a/client.go#L29)) and the default order for providers in the registry is `ipmitool`, `asrockrack`, `gofish`, `IntelAMT`. The default order is defined [here](https://github.com/metal-toolbox/bmclib/blob/b5cdfa3ffe026d3cc3257953abe3234b278ca20a/client.go#L152).

### Permanent Filtering

Permanent filtering modifies the order and/or the number of providers for BMC calls for all client methods (for example: `Open`, `SetPowerState`, etc) calls.

```go
cl := bmclib.NewClient(host, user, pass)
// This will modify the order for all subsequent BMC calls
cl.Registry.Drivers = cl.Registry.PreferDriver("gofish")
if err := cl.Open(ctx); err != nil {
  return(err)
}
```

The following permanent filters are available:

- `cl.Registry.PreferDriver("gofish")` - This moves the `gofish` provider to be the first provider in the registry.
- `cl.Registry.Supports(providers.FeaturePowerSet)` - This removes any provider from the registry that does not support the setting the power state.
- `cl.Registry.Using("redfish")` - This removes any provider from the registry that does not support the `redfish` protocol.
- `cl.Registry.For("gofish")` - This removes any provider from the registry that is not the `gofish` provider.
- `cl.Registry.PreferProtocol("redfish")` - This moves any provider that implements the `redfish` protocol to the beginning of the registry.

### One-time Filtering

One-time filtering modifies the order and/or the number of providers for BMC calls only for a single method call.

```Go
cl := bmclib.NewClient(host, user, pass)
// This will modify the order for only this BMC call
if err := cl.PreferProvider("gofish").Open(ctx); err != nil {
  return(err)
}
```

The following one-time filters are available:

- `cl.PreferProtocol("gofish").GetPowerState(ctx)` - This moves the `gofish` provider to be the first provider in the registry.
- `cl.Supports(providers.FeaturePowerSet).GetPowerState(ctx)` - This removes any provider from the registry that does not support the setting the power state.
- `cl.Using("redfish").GetPowerState(ctx)` - This removes any provider from the registry that does not support the `redfish` protocol.
- `cl.For("gofish").GetPowerState(ctx)` - This removes any provider from the registry that is not the `gofish` provider.
- `cl.PreferProtocol("redfish").GetPowerState(ctx)` - This moves any provider that implements the `redfish` protocol to the beginning of the registry.

### Tracing

To collect trace telemetry, set the `WithTraceProvider()` option on the client
which results in trace spans being collected for each client method.

```go
cl := bmclib.NewClient(
          host,
          user,
          pass,
          bmclib.WithLogger(log),
          bmclib.WithTracerProvider(otel.GetTracerProvider()),
      )
```
