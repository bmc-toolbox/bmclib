# Changelog
All notable changes to this project goes here.

## [v0.2.2] - 25-10-2018
### Added
- Add DEBUG_BMCLIB var to verbose log.
- Add screen preview capture support for iLO5, Idrac8,9, Supermicrosx10.
- Add support to configure license keys on ILOs
- Add support to configure Idrac9's - BMC, BIOS.
- Expose PowerCycleBmc() method.
- Add support to configure network config on Idrac8's.
- Add support to remove ldap groups on.

### Changed
- Update chassis configuration resources.
- Bump various httpclient timeouts - TLSHandshakeTimeout, ResponseHeaderTimeout, KeepAlive (for Idrac, ILOs)
- Switch to Debug logging instead of Info to reduce logging spam.
- Reorder configuration resources so NTP config is applied last.

### Fixed
- c7000 login panic.
- c7000 fix err checks.
- Fix validation checks for ldap groups.
- Minor fixes to flexaddress state change.


