/*
Virtual Media is an example command to mount and umount virtual media (ISO) on a BMC.

	# mount an ISO
	$ go run examples/virtualmedia/main.go \
		-host 10.1.2.3 \
		-user root \
		-password calvin \
		-iso http://example.com/image.iso

	# unmount an ISO
	$ go run examples/virtualmedia/main.go \
		-host 10.1.2.3 \
		-user root \
		-password calvin \
		-iso ""
*/
package main
