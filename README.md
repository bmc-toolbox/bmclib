### bmcbutler - A BMC configuration tool, based on [bmclib](https://github.com/ncode/bmclib)

For supported hardware see the bmclib page.

##### Build
go get github.com/bmc-toolbox/bmcbutler

##### Setup
Theres two parts to setting up configuration for bmcbutler,
the butler configuration and the configuration for bmcs.

###### bmcbutler configuration
The configuration bmcbutler requires to run.

Add a config under /etc/bmcbutler/bmcbutler.yml
[bmcbutler.yml sample](../master/bmcbutler.yml.sample)

###### bmc configuration
The configuration to be applied to bmcs.

bmc configuration is split into two types,

* configuration - configuration to be applied periodically.
* setup - one time setup configuration.

Create a directory /etc/bmcbutler/cfg/
copy the yaml config definitions in there,

[configuration.yml sample](../master/cfg/configuration.yml)
[setup.yml sample](../master/cfg/setup.yml)

###### inventory
bmcbutler was written with the intent of sourcing inventory assets and configuring their bmcs,
a csv inventory example is provided to play with.

[inventory.csv sample](../master/inventory.csv.sample)

The config file points bmcbutler to the right inventory source.

##### Run

Configure blade/chassis/discrete, this expects the csv file to be in place.

```
bmcbutler configure --chassis --serial <serial> --verbose
bmcbutler configure --blade --serial <serial> --verbose
bmcbutler configure --discrete --serial <serial> --verbose
```

#### Acknowledgment

bmcbutler was originally developed for [Booking.com](http://www.booking.com).
With approval from [Booking.com](http://www.booking.com), the code and
specification were generalized and published as Open Source on github, for
which the authors would like to express their gratitude.
