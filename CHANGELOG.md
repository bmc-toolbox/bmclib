# Changelog
All notable changes to this project goes here.

## [v0.0.4] - 03-09-2018
### Added
- Adds --dry-run flag to just run without taking actions.
- Adds --location, --butlers to overide config params.

### Changed
- Update config handling, read in all config before run.
- rewrite butler go routine handling
 - butler manager handles interrupts and notifies butler go routines over chan.
 - butler manager waits for each butler routine to notify when its done over chan.
- Inventory/asset filters flags updated,
 - --chassis/--blades/--discretes can be invoked to configure just those asset types.
 - --serials, --ips can be used without --chassis/--blades/--discretes args.
 - use plurals for all flags
- Inventory is fetched by invoking inventory type -> AssetRetrieve() which returns the approprate method.
- Merge code among cmd/(configure,setup,execute) into pre(), post() methods.

### Fixed
- Ensure metrics channel is closed after all metrics are sent.
- Skip assets from dora inventory with 0.0.0.0 as IPs.
- Vendor default password lookup uses vendor instead of model.

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
