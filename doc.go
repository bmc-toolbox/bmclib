// Copyright 2019 The bmclib Authors. All rights reserved.
// Use of this source code is governed by an Apache that can be found in the LICENSE file.

/*

Package bmclib abstracts various vendor/models of Baseboard Management controllers.

ENV vars
========
export DEBUG_BMCLIB=1 for bmclib to verbose log
export BMCLIB_TEST=1 to run on a dummy bmc (dry run).

Scan and connect
----------------

Connect to a BMC - "discover" its model, vendor, for list of supported BMCs see README.md.


	connection, err = discover.ScanAndConnect(ip, user, pass)
	if err != nil {
		return connection, errors.New("ScanAndConnect attempt unsuccessful.")
	}


Once a connection is setup, the connection needs to be type asserted, to either a 'Bmc' or 'BmcChassis'.

	switch connection.(type) {
	case devices.Bmc:
		bmc := connection.(devices.Bmc)

		// invoke Bmc interface methods here
		...

		bmc.Close()
	case devices.BmcChassis:
		chassis := connection.(devices.BmcChassis)

		// invoke BmcChassis interface methods here
		...

		chassis.Close()
	default:
		log.Error("Unknown device")
	}


*/
package bmclib
