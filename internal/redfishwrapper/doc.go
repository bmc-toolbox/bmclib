/*
Package redfishwrapper provides a wrapper around the gofish library to interact with Redfish API.

It abstracts common BMC operations like setting virtual media, getting inventory, etc.

## Virtual Media

bmclib supports attaching and ejecting virtual media through the Redfish `gofish` provider.

		cl := bmclib.NewClient("192.168.1.1", "admin", "hunter2")

		if err := cl.Open(ctx); err != nil {
	  	return(err)
		}
		defer cl.Close(ctx)

		// Attach an ISO image as CD/DVD virtual media:
		ok, err := cl.SetVirtualMedia(ctx, "CD", "https://example.com/installer.iso")
		if err != nil {
	  	return(err)
		}
		if !ok {
	  	return(fmt.Errorf("failed to set virtual media"))
		}

		// To eject currently attached virtual media, pass an empty media URL:
		ok, err := cl.SetVirtualMedia(ctx, "CD", "")
		if err != nil {
	  	return(err)
		}
		if !ok {
	  	return(fmt.Errorf("failed to eject virtual media"))
		}

Supported media kinds are: CD, DVD, Floppy, USBStick.

The Redfish provider first looks for VirtualMedia under the Manager resource (/redfish/v1/Managers/{ManagerId}/VirtualMedia) and
falls back to the System resource (/redfish/v1/Systems/{SystemId}/VirtualMedia) for implementations that expose virtual media there.
When the BMC does not expose standard VirtualMedia.InsertMedia or VirtualMedia.EjectMedia actions, bmclib relies on gofish
to fall back to PATCHing the VirtualMedia resource directly. This allows virtual media support on implementations such as Lenovo XCC
that manage virtual media through PATCH instead of Redfish Actions.
*/
package redfishwrapper
