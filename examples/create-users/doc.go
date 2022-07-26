/*
create-users is an example commmand that utilizes the 'v1' bmclib interface
methods to create user entries in a BMC using the redfish driver.

	$ go run ./examples/v1/create-users/main.go -h
	Usage of /tmp/go-build440589615/b001/exe/main:
		-cert-pool string
					Path to an file containing x509 CAs. An empty string uses the system CAs. Only takes effect when --secure-tls=true
		-dry-run
					Connect to the BMC but do not create users
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
		-user-csv string
					A CSV file of users to create containing 3 columns: username, password, role
*/
package main
