# Changelog
All notable changes to this project goes here.

## [v0.0.6] - 25-10-2018
### Changed
- A new release since we ran go mod init on the repo.
- vendor directory to stay for now.

## [v0.0.5] - 16-10-2018
### Added
- Expose various asset attributes in configuration template.
- Split out bmc, chassis login logic into bmc-toolbox/bmclogin - ensures more reliable logins.
- Add support to "enc" lookup asset inventory, attributes using an external binary - docs/assetLookup.md
- Add documentation for configuration templating - docs/configTemplating.md
- Add documentation for asset/inventory lookup binary, docs/assetLookup.md
- Add support to power up blades in chassis as part of chassis_setup actions.
- Add --servers arg which will obsolete --blades, --discretes (so we end up with --servers/--chassis)

### Changed
- Removes 'setup' flag, merge chassis setup logic into setup_chassis.go
- Metrics collection rewrite, use rcrowley/go-metrics
- Fix flag variable naming consistency - use plural variables.

### Fixed
- Throw an error is there is no configuration to be applied.
- Minor fixes to flexaddress state change.

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
