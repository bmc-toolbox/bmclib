# bmclib - board management controller library

[![Status](https://api.travis-ci.org/bmc-toolbox/bmclib.svg?branch=master)](https://travis-ci.org/bmc-toolbox/bmclib)
[![Go Report Card](https://goreportcard.com/badge/github.com/bmc-toolbox/bmclib)](https://goreportcard.com/report/github.com/bmc-toolbox/bmclib)
[![GoDoc](https://godoc.org/github.com/bmc-toolbox/bmclib?status.svg)](https://godoc.org/github.com/bmc-toolbox/bmclib)
[![Development/Support](https://img.shields.io/badge/chat-on%20freenode-brightgreen.svg)](https://kiwiirc.com/client/irc.freenode.net/##bmc-toolbox)

A library to interact with BMCs of different vendors

## Data collection support

Hardware      | Supported | Partially Supported  |
:-----------  | :-------: | :------------------: |
Dell M1000e   | :heavy_check_mark: | |
Dell iDRAC8   | :heavy_check_mark: | |
Dell iDRAC9   | :heavy_check_mark: | |
HP c7000      | :heavy_check_mark: | |
HP iLO3       | | :heavy_check_mark: |
HP iLO4       | :heavy_check_mark: | |
HP iLO5       | :heavy_check_mark: | |
Supermicro X10 | :heavy_check_mark: | |

## Firmware update support

Hardware      | Supported |
:-----------  | :-------: |
Dell M1000e   | :heavy_check_mark: |
Dell iDRAC8   | |
Dell iDRAC9   | |
HP c7000      | :heavy_check_mark: |
HP iLO3       | |
HP iLO4       | |
HP iLO5       | |
Supermicro X10 | |

## Configuration support

Hardware      | User accounts | Syslog  |  NTP  | Ldap  | Ldap groups  | BIOS  | HTTPS Cert  |
:-----------  | :-----------: | :-----: | :---: | :---: | :----------: | :--: | :---: |
Dell M1000e   | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | - | |
Dell iDRAC8   | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | | :heavy_check_mark: |
Dell iDRAC9   | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: |
HP c7000      | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | - | |
HP iLO4       | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | | :heavy_check_mark: |
HP iLO5       | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | | :heavy_check_mark: |
Supermicro X10 | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | | :heavy_check_mark: |


## Debugging

export DEBUG_BMCLIB=1 for bmclib to verbose log

export BMCLIB_TEST=1 to run on a dummy bmc (dry run).

## bmc-toolbox

All of the tooling that uses bmclib is part of the [bmc-toolbox](https://github.com/bmc-toolbox)


#### Inventory collection

[dora](https://github.com/bmc-toolbox/dora) uses bmclib to identify and inventorize assets.

#### Configuration

[bmcbutler](https://github.com/bmc-toolbox/bmcbutler) uses bmclib to manage configuration on BMCs.

#### Web API

[actor](https://github.com/bmc-toolbox/actor) uses bmclib to abstract away various BMCs and expose a consistent web API to interact with them.


#### Acknowledgment

bmclib was originally developed for [Booking.com](http://www.booking.com).
With approval from [Booking.com](http://www.booking.com), the code and
specification were generalized and published as Open Source on github, for
which the authors would like to express their gratitude.

#### Authors
- Juliano Martinez
- Joel Rebello 
- Guilherme M. Schroeder
