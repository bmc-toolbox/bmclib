/*
inventory is an example commmand that utilizes the 'v1' bmclib interface
methods to upload and mount, unmount a floppy image.

	    # mount image
		$ go run examples/floppy-image/main.go \
			-host 10.1.2.3 \
			-user ADMIN \
			-password hunter2 \
			-image /tmp/floppy.img

		# un-mount image
		$ go run examples/floppy-image/main.go \
			-host 10.1.2.3 \
			-user ADMIN \
			-password hunter2 \
			-unmount
*/
package main
