/*
status is an example commmand that utilizes the 'v1' bmclib interface methods
to capture a screenshot.

	$ go run ./examples/v1/status/main.go -h
	Usage of /tmp/go-build1941100323/b001/exe/main:
		-cert-pool string
					Path to an file containing x509 CAs. An empty string uses the system CAs. Only takes effect when --secure-tls=true
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
*/
package main
