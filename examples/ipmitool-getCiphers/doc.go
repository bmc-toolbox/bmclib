/*
getCiphers and getSolCiphers retrieve the ciphers supported by the BMC for IPMI and SOL connections.

	$ go run ./examples/ipmitool-getCiphers/main.go -h
	Usage of /var/folders/nc/f6w6rbq941jcsmkpwr8tgl700000gp/T/go-build3555327710/b001/exe/main:
		-host string
					BMC hostname to connect to
		-password string
					Username to login with
		-user string
					Username to login with
*/
package main
