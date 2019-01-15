### bmcbutler

[![Status](https://api.travis-ci.org/bmc-toolbox/bmcbutler.svg?branch=master)](https://travis-ci.org/bmc-toolbox/bmcbutler)
[![Go Report Card](https://goreportcard.com/badge/github.com/bmc-toolbox/bmcbutler)](https://goreportcard.com/report/github.com/bmc-toolbox/bmcbutler)

##### About

Bmcbutler is a BMC (Baseboard Management Controller) configuration management tool that uses [bmclib](https://github.com/ncode/bmclib).

For list of supported BMCs and configuration options supported, see [supported hardware](https://github.com/bmc-toolbox/bmclib/blob/master/README.md)

##### Build
`go get github.com/bmc-toolbox/bmcbutler`

###### Build with vendored modules (go 1.11)
`GO111MODULE=on go build -mod vendor -v`

###### Notes on working with go mod
To pick a specific bmclib SHA.

`GO111MODULE=on go get github.com/bmc-toolbox/bmclib@2d1bd1cb`

To add/update the vendor dir.

`GO111MODULE=on go mod vendor`

##### Setup
Theres two parts to setting up configuration for bmcbutler,

* Bmcbutler configuration
* Configuration for BMCs

This document assumes the Bmcbutler configuration directory is ~/.bmcbutler.

###### Bmcbutler configuration
Setup configuration Bmcbutler requires to run.

```
# create a configuration directory for ~/.bmcbutler
mkdir ~/.bmcbutler/
```
Copy the sample config into ~/.bmcbutler/
[bmcbutler.yml sample](../master/samples/bmcbutler.yml.sample)

###### BMC configuration
Configuration to be applied to BMCs.

```
# create a directory for BMC config
mkdir ~/.bmcbutler/cfg
```
add the BMC yaml config definitions in there, for sample config see [configuration.yml sample](../master/cfg/configuration.yml)

###### bmc configuration templating
configuration.yml supports templating, for details see [configTemplating](../master/docs/configTemplating.md)

###### inventory
Bmcbutler was written with the intent of sourcing inventory assets and configuring their bmcs,
a csv inventory example is provided to play with.

[inventory.csv sample](../master/samples/inventory.csv.sample)

The 'inventory' parameter points Bmcbutler to the inventory source.


##### Run

Configure Blades/Chassis/Discretes

```
#configure all BMCs in inventory, dry run with verbose output
bmcbutler configure --all --dryrun -v

#configure all servers in given locations
bmcbutler configure --servers --locations ams2

#configure all chassis in given locations
bmcbutler configure --chassis --locations ams2,lhr3 

#configure all servers in given location, spawning given butlers
bmcbutler configure --servers --locations lhr5 --butlers 200

#configure one or more BMCs identified by IP(s)
bmcbutler configure --ips 192.168.0.1,192.168.0.2,192.168.0.2

#configure one or more BMCs identified by serial(s)
bmcbutler configure --serials <serial1>,<serial2>

bmcbutler configure --serial <serial1>,<serial2> --verbose
bmcbutler configure  --serial <serial> --verbose

#Apply specific configuration resource(s)
bmcbutler configure --ips 192.168.1.4 --resources ntp,syslog,user
```

#### Acknowledgment

bmcbutler was originally developed for [Booking.com](http://www.booking.com).
With approval from [Booking.com](http://www.booking.com), the code and
specification were generalized and published as Open Source on github, for
which the authors would like to express their gratitude.
