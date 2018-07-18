### bmcbutler - A BMC configuration tool

Bmcbutler is a tool to configure BMCs using [bmclib](https://github.com/ncode/bmclib),
assets - BMCs to configure are read from an inventory source defined in `bmcbutler.yml` configuration file.

Multiple butler processes are spawned to configure BMCs based on configuration declared in [configuration.yml sample](../master/cfg/configuration.yml), [setup.yml sample](../master/cfg/setup.yml).

For supported BMCs see the bmclib page.

##### Build
`go get github.com/bmc-toolbox/bmcbutler`

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
[bmcbutler.yml sample](../master/bmcbutler.yml.sample)


###### BMC configuration
Configuration to be applied to BMCs.

BMC configuration is split into two types,

* configuration - configuration to be applied periodically.
* setup - one time setup configuration.

```
# create a directory for BMC config
mkdir ~/.bmcbutler/cfg
```

Create a directory /etc/bmcbutler/cfg/
add the BMC yaml config definitions in there,

[configuration.yml sample](../master/cfg/configuration.yml)
[setup.yml sample](../master/cfg/setup.yml)

###### inventory
Bmcbutler was written with the intent of sourcing inventory assets and configuring their bmcs,
a csv inventory example is provided to play with.

[inventory.csv sample](../master/inventory.csv.sample)

The 'inventory' parameter points Bmcbutler to the inventory source.

##### Run

Configure Blades/Chassis/Discretes

```
#configure all BMCs in inventory
bmcbutler configure --all

#configure BMCs identified by IP(s)
bmcbutler configure --iplist 192.168.0.1,192.168.0.2,192.168.0.2

#configure BMCs identified by serial(s)
bmcbutler configure --chassis --serial <serial> --verbose
bmcbutler configure --blade --serial <serial1>,<serial2> --verbose
bmcbutler configure --discrete --serial <serial> --verbose
```

#### Acknowledgment

bmcbutler was originally developed for [Booking.com](http://www.booking.com).
With approval from [Booking.com](http://www.booking.com), the code and
specification were generalized and published as Open Source on github, for
which the authors would like to express their gratitude.
