/*
install-firmware is an example commmand that utilizes the 'v1' bmclib interface
methods to flash a firmware image to a BMC.

	$ go run ./examples/v1/install-firmware/main.go -h
	Usage of /tmp/go-build2950657412/b001/exe/main:
		-cert-pool string
					Path to an file containing x509 CAs. An empty string uses the system CAs. Only takes effect when --secure-tls=true
		-firmware string
					The local path of the firmware to install
		-host string
					BMC hostname to connect to
		-password string
					Username to login with
		-port int
					BMC port to connect to (default 443)
		-secure-tls
					Enable secure TLS
		-user string
					Username to login with
		-version string
					The firmware version being installed
*/
package main
