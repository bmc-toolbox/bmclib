# Changelog
All notable changes to this project goes here.
## [v0.0.3] - 20-08-2018
### Added
- This CHANGELOG file.
- Adds metrics forwarder, and collect various metrics with support for Graphite.
- Adds --iplist parameter to configure, setup a given bmc IP
- Adds templating support based on plush for for bmc configuration.
- Adds basic execute action, to execute commands on bmcs.
- Adds interrupt handling for butlers to exit gracefully, and not leave connections open.

### Changed
- move code into pkg/, sample files into samples
- Login attempts into bmc to try primary, secondary, default accounts, in that order.
- Split out and merge bmc connection setup from configure, setup (connectionSetup.go)
- Ensure bmc connections are closed after configure/setup.
- Fixes to error handling on connection setup.
