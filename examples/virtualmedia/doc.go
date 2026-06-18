/*
Virtual Media is an example command to mount and unmount virtual media on a BMC.

	# mount an ISO as CD/DVD virtual media
	$ go run examples/virtualmedia/main.go \
		-host 10.1.2.3 \
		-user root \
		-password calvin \
		-media-kind CD \
		-iso http://example.com/image.iso

	# unmount/eject virtual media
	$ go run examples/virtualmedia/main.go \
		-host 10.1.2.3 \
		-user root \
		-password calvin \
		-media-kind CD \
		-eject

Supported media kinds are: CD, DVD, Floppy, USBStick.

The Redfish gofish provider supports BMCs that expose VirtualMedia through
standard InsertMedia/EjectMedia actions and BMCs that require PATCHing the
VirtualMedia resource directly.
*/
package main
