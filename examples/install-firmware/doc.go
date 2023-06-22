/*
install-firmware is an example commmand that utilizes the 'v1' bmclib interface
methods to flash a firmware image to a BMC.

Note: The example installs the firmware and polls until the status until the install is complete,
and if required by the install process - power cycles the host.

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

						   # install bios firmware on a supermicro X11
						   #
						   $ go run .  -host 192.168.1.1 -user ADMIN -password hunter2 -component bios -firmware BIOS_X11DPH-0981_20220208_3.6_STD.bin
						   INFO[0007] set firmware install mode                     component=BIOS ip="https://192.168.1.1" model=X11DPH-T
						   INFO[0011] uploading firmware                            component=BIOS ip="https://192.168.1.1" model=X11DPH-T
						   INFO[0091] verifying uploaded firmware                   component=BIOS ip="https://192.168.1.1" model=X11DPH-T
						   INFO[0105] initiating firmware install                   component=BIOS ip="https://192.168.1.1" model=X11DPH-T
						   INFO[0115] firmware install running                      component=bios state=running
						   INFO[0132] firmware install running                      component=bios state=running
						   ...
						   ...
	                       INFO[0628] firmware install running                      component=bios state=running
	                       INFO[0635] host powercycle required                      component=bios state=powercycle-host
	                       INFO[0637] host power cycled, all done!                  component=bios state=powercycle-host



							# install bmc firmware on a supermicro X11
							#
							$ go run .  -host 192.168.1.1 -user ADMIN -password hunter2 -component bmc -firmware BMC_X11AST2500-4101MS_20220225_01.74.02_STD.bin
			                INFO[0007] setting device to firmware install mode       component=BMC ip="https://192.168.1.1"
			                INFO[0009] uploading firmware                            ip="https://192.168.1.1"
			                INFO[0045] verifying uploaded firmware                   ip="https://192.168.1.1"
			                INFO[0047] initiating firmware install                   ip="https://192.168.1.1"
			                INFO[0079] firmware install running                      component=bmc state=running
			                INFO[0085] firmware install running                      component=bmc state=running
			                ...
							...
							INFO[0233] firmware install running                      component=bmc state=running
		                    INFO[0238] firmware install completed                    component=bmc state=complete
*/
package main
