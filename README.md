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

### Versions

The current bmclib version is `v2` and is being developed on the `main` branch.

The previous bmclib version is in maintenance mode and can be found here [v1](https://github.com/bmc-toolbox/bmclib/v1).

### Acknowledgments

bmclib v2 interfaces with Redfish on BMCs through the Gofish library https://github.com/stmcginnis/gofish

bmclib was originally developed for [Booking.com](http://www.booking.com). With approval from [Booking.com](http://www.booking.com), 
the code and specification were generalized and published as Open Source on github, for which the authors would like to express their gratitude.

### Authors
- [Joel Rebello](https://github.com/joelrebel) 
- [Jacob Weinstock](https://github.com/jacobweinstock)

