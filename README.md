# bmclib - board management controller library

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

Hardware      | User accounts | Syslog  |  NTP  | Ldap  | Ldap groups  | SSL  |
:-----------  | :-----------: | :-----: | :---: | :---: | :----------: | :--: |
Dell M1000e   | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | | 
Dell iDRAC8   | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | |
HP c7000      | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | |
HP iLO4       | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | |
Supermicro X10 | :heavy_check_mark: | :heavy_check_mark: | :heavy_check_mark: | | | |

#### Acknowledgment

bmclib was originally developed for [Booking.com](http://www.booking.com).
With approval from [Booking.com](http://www.booking.com), the code and
specification were generalized and published as Open Source on github, for
which the authors would like to express their gratitude.

#### Authors
- Juliano Martinez
- Joel Rebello 
- Guilherme M. Schroeder
