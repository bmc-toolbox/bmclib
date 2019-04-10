package c7000

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/spf13/viper"
)

var (
	mux     *http.ServeMux
	server  *httptest.Server
	answers = map[string][]byte{
		"/hpoa": []byte(`<?xml version="1.0" encoding="UTF-8"?>
			<SOAP-ENV:Envelope xmlns:SOAP-ENV="http://www.w3.org/2003/05/soap-envelope" xmlns:SOAP-ENC="http://www.w3.org/2003/05/soap-encoding" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:wsu="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-utility-1.0.xsd" xmlns:wsse="http://docs.oasis-open.org/wss/2004/01/oasis-200401-wss-wssecurity-secext-1.0.xsd" xmlns:hpoa="hpoa.xsd">
				<SOAP-ENV:Body>
					<hpoa:userLogInResponse>
						<hpoa:HpOaSessionKeyToken>
							<hpoa:oaSessionKey>a8223b7caad9ea0e</hpoa:oaSessionKey>
						</hpoa:HpOaSessionKeyToken>
					</hpoa:userLogInResponse>
				</SOAP-ENV:Body>
			</SOAP-ENV:Envelope>`),
		"/xmldata": []byte(`
			<RIMP>
				<MP>
					<ST>1</ST>
					<PRIM>true</PRIM>
					<PN>BladeSystem c7000 DDR2 Onboard Administrator with KVM</PN>
					<FWRI>4.70</FWRI>
					<HWRI>65.49</HWRI>
					<SN>OB6BCP1616    </SN>
					<UUID>09OB6BCP1616    </UUID>
					<STE>false</STE>
					<USESTE>false</USESTE>
					<SSO>false</SSO>
					<CIMOM>false</CIMOM>
					<ERS>0</ERS>
				</MP>
				<INFRA2>
					<RACK>UnnamedRack</RACK>
					<ENCL>spare-cz372137h3</ENCL>
					<DATETIME>2017-11-01T09:14:55-05:00</DATETIME>
					<TIMEZONE>CST6CDT</TIMEZONE>
					<PN>BladeSystem c7000 Enclosure G3</PN>
					<ASSET></ASSET>
					<STATUS>OK</STATUS>
					<DIAG>
						<FRU>NO_ERROR</FRU>
						<MgmtProc>NOT_RELEVANT</MgmtProc>
						<thermalWarning>NOT_RELEVANT</thermalWarning>
						<thermalDanger>NOT_RELEVANT</thermalDanger>
						<Keying>NOT_RELEVANT</Keying>
						<Power>NOT_RELEVANT</Power>
						<Cooling>NOT_RELEVANT</Cooling>
						<Location>NOT_RELEVANT</Location>
						<Failure>NOT_TESTED</Failure>
						<Degraded>NOT_TESTED</Degraded>
						<AC>NOT_RELEVANT</AC>
						<i2c>NOT_RELEVANT</i2c>
						<oaRedundancy>NO_ERROR</oaRedundancy>
					</DIAG>
					<ENCL_SN>CZ372137H3</ENCL_SN>
					<PART>681844-B21</PART>
					<UUID>09CZ372137H3</UUID>
					<UIDSTATUS>OFF</UIDSTATUS>
					<ADDR>A9FE01F0</ADDR>
					<SOLUTIONSID>0000000000000000</SOLUTIONSID>
					<DIM>
						<mmHeight>445</mmHeight>
						<mmWidth>444</mmWidth>
						<mmDepth>756</mmDepth>
					</DIM>
					<BLADES>
						<BAYS>
							<BAY NAME="1">
								<SIDE>FRONT</SIDE>
								<mmHeight>181</mmHeight>
								<mmWidth>56</mmWidth>
								<mmDepth>480</mmDepth>
								<mmXOffset>0</mmXOffset>
								<mmYOffset>7</mmYOffset>
							</BAY>
							<BAY NAME="2">
								<SIDE>FRONT</SIDE>
								<mmHeight>181</mmHeight>
								<mmWidth>56</mmWidth>
								<mmDepth>480</mmDepth>
								<mmXOffset>56</mmXOffset>
								<mmYOffset>7</mmYOffset>
							</BAY>
							<BAY NAME="3">
								<SIDE>FRONT</SIDE>
								<mmHeight>181</mmHeight>
								<mmWidth>56</mmWidth>
								<mmDepth>480</mmDepth>
								<mmXOffset>112</mmXOffset>
								<mmYOffset>7</mmYOffset>
							</BAY>
							<BAY NAME="4">
								<SIDE>FRONT</SIDE>
								<mmHeight>181</mmHeight>
								<mmWidth>56</mmWidth>
								<mmDepth>480</mmDepth>
								<mmXOffset>168</mmXOffset>
								<mmYOffset>7</mmYOffset>
							</BAY>
							<BAY NAME="5">
								<SIDE>FRONT</SIDE>
								<mmHeight>181</mmHeight>
								<mmWidth>56</mmWidth>
								<mmDepth>480</mmDepth>
								<mmXOffset>224</mmXOffset>
								<mmYOffset>7</mmYOffset>
							</BAY>
							<BAY NAME="6">
								<SIDE>FRONT</SIDE>
								<mmHeight>181</mmHeight>
								<mmWidth>56</mmWidth>
								<mmDepth>480</mmDepth>
								<mmXOffset>280</mmXOffset>
								<mmYOffset>7</mmYOffset>
							</BAY>
							<BAY NAME="7">
								<SIDE>FRONT</SIDE>
								<mmHeight>181</mmHeight>
								<mmWidth>56</mmWidth>
								<mmDepth>480</mmDepth>
								<mmXOffset>336</mmXOffset>
								<mmYOffset>7</mmYOffset>
							</BAY>
							<BAY NAME="8">
								<SIDE>FRONT</SIDE>
								<mmHeight>181</mmHeight>
								<mmWidth>56</mmWidth>
								<mmDepth>480</mmDepth>
								<mmXOffset>392</mmXOffset>
								<mmYOffset>7</mmYOffset>
							</BAY>
							<BAY NAME="9">
								<SIDE>FRONT</SIDE>
								<mmHeight>181</mmHeight>
								<mmWidth>56</mmWidth>
								<mmDepth>480</mmDepth>
								<mmXOffset>0</mmXOffset>
								<mmYOffset>188</mmYOffset>
							</BAY>
							<BAY NAME="10">
								<SIDE>FRONT</SIDE>
								<mmHeight>181</mmHeight>
								<mmWidth>56</mmWidth>
								<mmDepth>480</mmDepth>
								<mmXOffset>56</mmXOffset>
								<mmYOffset>188</mmYOffset>
							</BAY>
							<BAY NAME="11">
								<SIDE>FRONT</SIDE>
								<mmHeight>181</mmHeight>
								<mmWidth>56</mmWidth>
								<mmDepth>480</mmDepth>
								<mmXOffset>112</mmXOffset>
								<mmYOffset>188</mmYOffset>
							</BAY>
							<BAY NAME="12">
								<SIDE>FRONT</SIDE>
								<mmHeight>181</mmHeight>
								<mmWidth>56</mmWidth>
								<mmDepth>480</mmDepth>
								<mmXOffset>168</mmXOffset>
								<mmYOffset>188</mmYOffset>
							</BAY>
							<BAY NAME="13">
								<SIDE>FRONT</SIDE>
								<mmHeight>181</mmHeight>
								<mmWidth>56</mmWidth>
								<mmDepth>480</mmDepth>
								<mmXOffset>224</mmXOffset>
								<mmYOffset>188</mmYOffset>
							</BAY>
							<BAY NAME="14">
								<SIDE>FRONT</SIDE>
								<mmHeight>181</mmHeight>
								<mmWidth>56</mmWidth>
								<mmDepth>480</mmDepth>
								<mmXOffset>280</mmXOffset>
								<mmYOffset>188</mmYOffset>
							</BAY>
							<BAY NAME="15">
								<SIDE>FRONT</SIDE>
								<mmHeight>181</mmHeight>
								<mmWidth>56</mmWidth>
								<mmDepth>480</mmDepth>
								<mmXOffset>336</mmXOffset>
								<mmYOffset>188</mmYOffset>
							</BAY>
							<BAY NAME="16">
								<SIDE>FRONT</SIDE>
								<mmHeight>181</mmHeight>
								<mmWidth>56</mmWidth>
								<mmDepth>480</mmDepth>
								<mmXOffset>392</mmXOffset>
								<mmYOffset>188</mmYOffset>
							</BAY>
						</BAYS>
						<BLADE>
							<BAY>
								<CONNECTION>1</CONNECTION>
							</BAY>
							<MGMTIPADDR>0.0.0.0</MGMTIPADDR>
							<MGMTIPV6ADDR_LL>fe80::5265:f3ff:fe66:3902/64</MGMTIPV6ADDR_LL>
							<MGMTDNSNAME>.machine.example.com</MGMTDNSNAME>
							<MGMTPN>iLO4</MGMTPN>
							<MGMTFWVERSION>2.55 Aug 16 2017</MGMTFWVERSION>
							<PN>727021-B21</PN>
							<BLADEROMVER>I36 02/17/2017</BLADEROMVER>
							<NAME>bbmi</NAME>
							<PWRM>1.0.9</PWRM>
							<VLAN>1</VLAN>
							<SPN>ProLiant BL460c Gen9</SPN>
							<BSN>CZ3521YAEK</BSN>
							<UUID>727021CZ3521YAEK</UUID>
							<TYPE>SERVER</TYPE>
							<MANUFACTURER>HP</MANUFACTURER>
							<STATUS>OK</STATUS>
							<DIAG>
								<FRU>NO_ERROR</FRU>
								<MgmtProc>NO_ERROR</MgmtProc>
								<thermalWarning>NOT_TESTED</thermalWarning>
								<thermalDanger>NOT_TESTED</thermalDanger>
								<Keying>NO_ERROR</Keying>
								<Power>NO_ERROR</Power>
								<Cooling>NO_ERROR</Cooling>
								<Location>NOT_TESTED</Location>
								<Failure>NO_ERROR</Failure>
								<Degraded>NO_ERROR</Degraded>
								<AC>NOT_RELEVANT</AC>
								<i2c>NOT_RELEVANT</i2c>
								<oaRedundancy>NOT_RELEVANT</oaRedundancy>
							</DIAG>
							<UIDSTATUS>OFF</UIDSTATUS>
							<PORTMAP>
								<STATUS>OK</STATUS>
								<MEZZ>
									<NUMBER>1</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_ONE</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>3</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>1</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>4</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>1</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
								</MEZZ>
								<MEZZ>
									<NUMBER>2</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_TWO</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>5</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>1</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>6</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>1</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>3</NUMBER>
											<TRAYBAYNUMBER>7</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>1</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>4</NUMBER>
											<TRAYBAYNUMBER>8</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>1</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
								</MEZZ>
								<MEZZ>
									<NUMBER>9</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_ONE</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>1</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>1</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>2</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>1</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
									<DEVICE>
										<NAME>HP FlexFabric 10Gb 2-port 536FLB Adapter</NAME>
										<TYPE>MEZZ_DEV_TYPE_ONE</TYPE>
										<STATUS>OK</STATUS>
										<PORT>
											<NUMBER>1</NUMBER>
											<WWPN>8C:DC:D4:1C:5A:B0</WWPN>
											<TYPE>INTERCONNECT_TYPE_ETH</TYPE>
											<STATUS>OK</STATUS>
											<GUIDS>
												<GUID>
													<TYPE>C</TYPE>
													<FUNCTION>a</FUNCTION>
													<GUID_STRING>8C:DC:D4:1C:5A:B0</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>H</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>8C:DC:D4:1C:5A:B1</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>G</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>20:00:8C:DC:D4:1C:5A:B1</GUID_STRING>
													</GUID>
												</GUIDS>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<WWPN>8C:DC:D4:1C:5A:B8</WWPN>
											<TYPE>INTERCONNECT_TYPE_ETH</TYPE>
											<STATUS>UNKNOWN</STATUS>
											<GUIDS>
												<GUID>
													<TYPE>C</TYPE>
													<FUNCTION>a</FUNCTION>
													<GUID_STRING>8C:DC:D4:1C:5A:B8</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>H</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>8C:DC:D4:1C:5A:B9</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>G</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>20:00:8C:DC:D4:1C:5A:B9</GUID_STRING>
													</GUID>
												</GUIDS>
										</PORT>
									</DEVICE>
								</MEZZ>
								<MEZZ>
									<NUMBER>13</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_FIXED</TYPE>
									</SLOT>
								</MEZZ>
							</PORTMAP>
							<TEMPS>
								<TEMP>
									<LOCATION>14</LOCATION>
									<DESC>AMBIENT</DESC>
									<C>16</C>
									<THRESHOLD>
										<DESC>CAUTION</DESC>
										<C>42</C>
										<STATUS>Degraded</STATUS>
									</THRESHOLD>
									<THRESHOLD>
										<DESC>CRITICAL</DESC>
										<C>46</C>
										<STATUS>Non-Recoverable Error</STATUS>
									</THRESHOLD>
								</TEMP>
							</TEMPS>
							<POWER>
								<POWERSTATE>ON</POWERSTATE>
								<POWERMODE>UNKNOWN</POWERMODE>
								<POWER_CONSUMED>198</POWER_CONSUMED>
							</POWER>
							<VMSTAT>
								<SUPPORT>VM_SUPPORTED</SUPPORT>
								<CDROMSTAT>VM_DEV_STATUS_DISCONNECTED</CDROMSTAT>
								<CDROMURL></CDROMURL>
								<FLOPPYSTAT>VM_DEV_STATUS_DISCONNECTED</FLOPPYSTAT>
								<FLOPPYURL></FLOPPYURL>
							</VMSTAT>
							<cUUID>30373237-3132-5A43-3335-32315941454B</cUUID>
							<CONJOINABLE>false</CONJOINABLE>
						</BLADE>
						<BLADE>
							<BAY>
								<CONNECTION>2</CONNECTION>
							</BAY>
							<MGMTIPADDR>0.0.0.0</MGMTIPADDR>
							<MGMTIPV6ADDR_LL>fe80::7210:6fff:feba:2d16/64</MGMTIPV6ADDR_LL>
							<MGMTDNSNAME>.machine.example.com</MGMTDNSNAME>
							<MGMTPN>iLO4</MGMTPN>
							<MGMTFWVERSION>2.54 Jun 15 2017</MGMTFWVERSION>
							<PN>813198-B21</PN>
							<BLADEROMVER>I36 09/12/2016</BLADEROMVER>
							<NAME>bbmi</NAME>
							<PWRM>1.0.9</PWRM>
							<VLAN>1</VLAN>
							<SPN>ProLiant BL460c Gen9</SPN>
							<BSN>CZ36527E98</BSN>
							<UUID>813198CZ36527E98</UUID>
							<TYPE>SERVER</TYPE>
							<MANUFACTURER>HP</MANUFACTURER>
							<STATUS>OK</STATUS>
							<DIAG>
								<FRU>NO_ERROR</FRU>
								<MgmtProc>NO_ERROR</MgmtProc>
								<thermalWarning>NOT_TESTED</thermalWarning>
								<thermalDanger>NOT_TESTED</thermalDanger>
								<Keying>NO_ERROR</Keying>
								<Power>NO_ERROR</Power>
								<Cooling>NO_ERROR</Cooling>
								<Location>NOT_TESTED</Location>
								<Failure>NO_ERROR</Failure>
								<Degraded>NO_ERROR</Degraded>
								<AC>NOT_RELEVANT</AC>
								<i2c>NOT_RELEVANT</i2c>
								<oaRedundancy>NOT_RELEVANT</oaRedundancy>
							</DIAG>
							<UIDSTATUS>OFF</UIDSTATUS>
							<PORTMAP>
								<STATUS>OK</STATUS>
								<MEZZ>
									<NUMBER>1</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_ONE</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>3</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>2</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>4</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>2</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
								</MEZZ>
								<MEZZ>
									<NUMBER>2</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_TWO</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>5</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>2</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>6</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>2</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>3</NUMBER>
											<TRAYBAYNUMBER>7</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>2</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>4</NUMBER>
											<TRAYBAYNUMBER>8</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>2</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
								</MEZZ>
								<MEZZ>
									<NUMBER>9</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_ONE</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>1</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>2</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>2</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>2</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
									<DEVICE>
										<NAME>HP FlexFabric 10Gb 2-port 536FLB Adapter</NAME>
										<TYPE>MEZZ_DEV_TYPE_ONE</TYPE>
										<STATUS>OK</STATUS>
										<PORT>
											<NUMBER>1</NUMBER>
											<WWPN>9C:DC:71:64:E8:20</WWPN>
											<TYPE>INTERCONNECT_TYPE_ETH</TYPE>
											<STATUS>OK</STATUS>
											<GUIDS>
												<GUID>
													<TYPE>C</TYPE>
													<FUNCTION>a</FUNCTION>
													<GUID_STRING>9C:DC:71:64:E8:20</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>H</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>9C:DC:71:64:E8:21</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>G</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>20:00:9C:DC:71:64:E8:21</GUID_STRING>
													</GUID>
												</GUIDS>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<WWPN>9C:DC:71:64:E8:28</WWPN>
											<TYPE>INTERCONNECT_TYPE_ETH</TYPE>
											<STATUS>UNKNOWN</STATUS>
											<GUIDS>
												<GUID>
													<TYPE>C</TYPE>
													<FUNCTION>a</FUNCTION>
													<GUID_STRING>9C:DC:71:64:E8:28</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>H</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>9C:DC:71:64:E8:29</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>G</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>20:00:9C:DC:71:64:E8:29</GUID_STRING>
													</GUID>
												</GUIDS>
										</PORT>
									</DEVICE>
								</MEZZ>
								<MEZZ>
									<NUMBER>13</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_FIXED</TYPE>
									</SLOT>
								</MEZZ>
							</PORTMAP>
							<TEMPS>
								<TEMP>
									<LOCATION>14</LOCATION>
									<DESC>AMBIENT</DESC>
									<C>17</C>
									<THRESHOLD>
										<DESC>CAUTION</DESC>
										<C>42</C>
										<STATUS>Degraded</STATUS>
									</THRESHOLD>
									<THRESHOLD>
										<DESC>CRITICAL</DESC>
										<C>46</C>
										<STATUS>Non-Recoverable Error</STATUS>
									</THRESHOLD>
								</TEMP>
							</TEMPS>
							<POWER>
								<POWERSTATE>ON</POWERSTATE>
								<POWERMODE>UNKNOWN</POWERMODE>
								<POWER_CONSUMED>391</POWER_CONSUMED>
							</POWER>
							<VMSTAT>
								<SUPPORT>VM_SUPPORTED</SUPPORT>
								<CDROMSTAT>VM_DEV_STATUS_DISCONNECTED</CDROMSTAT>
								<CDROMURL></CDROMURL>
								<FLOPPYSTAT>VM_DEV_STATUS_DISCONNECTED</FLOPPYSTAT>
								<FLOPPYURL></FLOPPYURL>
							</VMSTAT>
							<cUUID>31333138-3839-5A43-3336-353237453938</cUUID>
							<CONJOINABLE>false</CONJOINABLE>
						</BLADE>
						<BLADE>
							<BAY>
								<CONNECTION>3</CONNECTION>
							</BAY>
							<MGMTIPADDR>0.0.0.0</MGMTIPADDR>
							<MGMTIPV6ADDR_LL>fe80::9eb6:54ff:fe94:2905/64</MGMTIPV6ADDR_LL>
							<MGMTDNSNAME>.machine.example.com</MGMTDNSNAME>
							<MGMTPN>iLO4</MGMTPN>
							<MGMTFWVERSION>2.55 Aug 16 2017</MGMTFWVERSION>
							<PN>641016-B21     </PN>
							<BLADEROMVER>I31 06/01/2015</BLADEROMVER>
							<NAME>bbmi</NAME>
							<PWRM>3.3.0</PWRM>
							<VLAN>1</VLAN>
							<SPN>ProLiant BL460c Gen8</SPN>
							<BSN>CZ3403YLMV      </BSN>
							<UUID>641016CZ3403YLMV</UUID>
							<TYPE>SERVER</TYPE>
							<MANUFACTURER>HP</MANUFACTURER>
							<STATUS>OK</STATUS>
							<DIAG>
								<FRU>NO_ERROR</FRU>
								<MgmtProc>NO_ERROR</MgmtProc>
								<thermalWarning>NOT_TESTED</thermalWarning>
								<thermalDanger>NOT_TESTED</thermalDanger>
								<Keying>NO_ERROR</Keying>
								<Power>NO_ERROR</Power>
								<Cooling>NO_ERROR</Cooling>
								<Location>NOT_TESTED</Location>
								<Failure>NO_ERROR</Failure>
								<Degraded>NO_ERROR</Degraded>
								<AC>NOT_RELEVANT</AC>
								<i2c>NOT_RELEVANT</i2c>
								<oaRedundancy>NOT_RELEVANT</oaRedundancy>
							</DIAG>
							<UIDSTATUS>OFF</UIDSTATUS>
							<PORTMAP>
								<STATUS>OK</STATUS>
								<MEZZ>
									<NUMBER>1</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_ONE</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>3</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>3</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>4</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>3</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
								</MEZZ>
								<MEZZ>
									<NUMBER>2</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_TWO</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>5</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>3</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>6</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>3</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>3</NUMBER>
											<TRAYBAYNUMBER>7</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>3</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>4</NUMBER>
											<TRAYBAYNUMBER>8</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>3</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
								</MEZZ>
								<MEZZ>
									<NUMBER>9</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_ONE</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>1</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>3</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>2</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>3</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
									<DEVICE>
										<NAME>HP FlexFabric 10Gb 2-port 554FLB Adapter</NAME>
										<TYPE>MEZZ_DEV_TYPE_ONE</TYPE>
										<STATUS>OK</STATUS>
										<PORT>
											<NUMBER>1</NUMBER>
											<WWPN>9C:B6:54:88:FC:B0</WWPN>
											<TYPE>INTERCONNECT_TYPE_ETH</TYPE>
											<STATUS>OK</STATUS>
											<GUIDS>
												<GUID>
													<TYPE>C</TYPE>
													<FUNCTION>a</FUNCTION>
													<GUID_STRING>9C:B6:54:88:FC:B0</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>H</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>9C:B6:54:88:FC:B1</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>G</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>10:00:9C:B6:54:88:FC:B1</GUID_STRING>
													</GUID>
												</GUIDS>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<WWPN>9C:B6:54:88:FC:B4</WWPN>
											<TYPE>INTERCONNECT_TYPE_ETH</TYPE>
											<STATUS>UNKNOWN</STATUS>
											<GUIDS>
												<GUID>
													<TYPE>C</TYPE>
													<FUNCTION>a</FUNCTION>
													<GUID_STRING>9C:B6:54:88:FC:B4</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>H</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>9C:B6:54:88:FC:B5</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>G</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>10:00:9C:B6:54:88:FC:B5</GUID_STRING>
													</GUID>
												</GUIDS>
										</PORT>
									</DEVICE>
								</MEZZ>
								<MEZZ>
									<NUMBER>13</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_FIXED</TYPE>
									</SLOT>
								</MEZZ>
							</PORTMAP>
							<TEMPS>
								<TEMP>
									<LOCATION>14</LOCATION>
									<DESC>AMBIENT</DESC>
									<C>15</C>
									<THRESHOLD>
										<DESC>CAUTION</DESC>
										<C>42</C>
										<STATUS>Degraded</STATUS>
									</THRESHOLD>
									<THRESHOLD>
										<DESC>CRITICAL</DESC>
										<C>46</C>
										<STATUS>Non-Recoverable Error</STATUS>
									</THRESHOLD>
								</TEMP>
							</TEMPS>
							<POWER>
								<POWERSTATE>ON</POWERSTATE>
								<POWERMODE>UNKNOWN</POWERMODE>
								<POWER_CONSUMED>282</POWER_CONSUMED>
							</POWER>
							<VMSTAT>
								<SUPPORT>VM_SUPPORTED</SUPPORT>
								<CDROMSTAT>VM_DEV_STATUS_DISCONNECTED</CDROMSTAT>
								<CDROMURL></CDROMURL>
								<FLOPPYSTAT>VM_DEV_STATUS_DISCONNECTED</FLOPPYSTAT>
								<FLOPPYURL></FLOPPYURL>
							</VMSTAT>
							<cUUID>30313436-3631-5A43-3334-3033594C4D56</cUUID>
							<CONJOINABLE>false</CONJOINABLE>
						</BLADE>
						<BLADE>
							<BAY>
								<CONNECTION>4</CONNECTION>
							</BAY>
							<MGMTIPADDR>0.0.0.0</MGMTIPADDR>
							<MGMTIPV6ADDR_LL>fe80::7646:a0ff:fef7:57c3/64</MGMTIPV6ADDR_LL>
							<MGMTDNSNAME>.machine.example.com</MGMTDNSNAME>
							<MGMTPN>iLO4</MGMTPN>
							<MGMTFWVERSION>2.55 Aug 16 2017</MGMTFWVERSION>
							<PN>641016-B21     </PN>
							<BLADEROMVER>I31 06/01/2015</BLADEROMVER>
							<NAME>bbmi</NAME>
							<PWRM>3.3.0</PWRM>
							<VLAN>1</VLAN>
							<SPN>ProLiant BL460c Gen8</SPN>
							<BSN>CZ3337N5JC      </BSN>
							<UUID>641016CZ3337N5JC</UUID>
							<TYPE>SERVER</TYPE>
							<MANUFACTURER>HP</MANUFACTURER>
							<STATUS>OK</STATUS>
							<DIAG>
								<FRU>NO_ERROR</FRU>
								<MgmtProc>NO_ERROR</MgmtProc>
								<thermalWarning>NOT_TESTED</thermalWarning>
								<thermalDanger>NOT_TESTED</thermalDanger>
								<Keying>NO_ERROR</Keying>
								<Power>NO_ERROR</Power>
								<Cooling>NO_ERROR</Cooling>
								<Location>NOT_TESTED</Location>
								<Failure>NO_ERROR</Failure>
								<Degraded>NO_ERROR</Degraded>
								<AC>NOT_RELEVANT</AC>
								<i2c>NOT_RELEVANT</i2c>
								<oaRedundancy>NOT_RELEVANT</oaRedundancy>
							</DIAG>
							<UIDSTATUS>OFF</UIDSTATUS>
							<PORTMAP>
								<STATUS>OK</STATUS>
								<MEZZ>
									<NUMBER>1</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_ONE</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>3</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>4</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>4</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>4</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
								</MEZZ>
								<MEZZ>
									<NUMBER>2</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_TWO</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>5</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>4</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>6</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>4</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>3</NUMBER>
											<TRAYBAYNUMBER>7</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>4</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>4</NUMBER>
											<TRAYBAYNUMBER>8</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>4</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
								</MEZZ>
								<MEZZ>
									<NUMBER>9</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_ONE</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>1</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>4</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>2</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>4</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
									<DEVICE>
										<NAME>HP FlexFabric 10Gb 2-port 554FLB Adapter</NAME>
										<TYPE>MEZZ_DEV_TYPE_ONE</TYPE>
										<STATUS>OK</STATUS>
										<PORT>
											<NUMBER>1</NUMBER>
											<WWPN>F0:92:1C:0D:F4:E8</WWPN>
											<TYPE>INTERCONNECT_TYPE_ETH</TYPE>
											<STATUS>OK</STATUS>
											<GUIDS>
												<GUID>
													<TYPE>C</TYPE>
													<FUNCTION>a</FUNCTION>
													<GUID_STRING>F0:92:1C:0D:F4:E8</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>H</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>F0:92:1C:0D:F4:E9</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>G</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>10:00:F0:92:1C:0D:F4:E9</GUID_STRING>
													</GUID>
												</GUIDS>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<WWPN>F0:92:1C:0D:F4:EC</WWPN>
											<TYPE>INTERCONNECT_TYPE_ETH</TYPE>
											<STATUS>UNKNOWN</STATUS>
											<GUIDS>
												<GUID>
													<TYPE>C</TYPE>
													<FUNCTION>a</FUNCTION>
													<GUID_STRING>F0:92:1C:0D:F4:EC</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>H</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>F0:92:1C:0D:F4:ED</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>G</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>10:00:F0:92:1C:0D:F4:ED</GUID_STRING>
													</GUID>
												</GUIDS>
										</PORT>
									</DEVICE>
								</MEZZ>
								<MEZZ>
									<NUMBER>13</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_FIXED</TYPE>
									</SLOT>
								</MEZZ>
							</PORTMAP>
							<TEMPS>
								<TEMP>
									<LOCATION>14</LOCATION>
									<DESC>AMBIENT</DESC>
									<C>16</C>
									<THRESHOLD>
										<DESC>CAUTION</DESC>
										<C>42</C>
										<STATUS>Degraded</STATUS>
									</THRESHOLD>
									<THRESHOLD>
										<DESC>CRITICAL</DESC>
										<C>46</C>
										<STATUS>Non-Recoverable Error</STATUS>
									</THRESHOLD>
								</TEMP>
							</TEMPS>
							<POWER>
								<POWERSTATE>ON</POWERSTATE>
								<POWERMODE>UNKNOWN</POWERMODE>
								<POWER_CONSUMED>291</POWER_CONSUMED>
							</POWER>
							<VMSTAT>
								<SUPPORT>VM_SUPPORTED</SUPPORT>
								<CDROMSTAT>VM_DEV_STATUS_DISCONNECTED</CDROMSTAT>
								<CDROMURL></CDROMURL>
								<FLOPPYSTAT>VM_DEV_STATUS_DISCONNECTED</FLOPPYSTAT>
								<FLOPPYURL></FLOPPYURL>
							</VMSTAT>
							<cUUID>30313436-3631-5A43-3333-33374E354A43</cUUID>
							<CONJOINABLE>false</CONJOINABLE>
						</BLADE>
						<BLADE>
							<BAY>
								<CONNECTION>5</CONNECTION>
							</BAY>
							<MGMTIPADDR>0.0.0.0</MGMTIPADDR>
							<MGMTIPV6ADDR_LL>fe80::5265:f3ff:fe66:310c/64</MGMTIPV6ADDR_LL>
							<MGMTDNSNAME>.machine.example.com</MGMTDNSNAME>
							<MGMTPN>iLO4</MGMTPN>
							<MGMTFWVERSION>2.54 Jun 15 2017</MGMTFWVERSION>
							<PN>727021-B21</PN>
							<BLADEROMVER>I36 02/17/2017</BLADEROMVER>
							<NAME>bbmi</NAME>
							<PWRM>1.0.9</PWRM>
							<VLAN>1</VLAN>
							<SPN>ProLiant BL460c Gen9</SPN>
							<BSN>CZ3521Y6XJ</BSN>
							<UUID>727021CZ3521Y6XJ</UUID>
							<TYPE>SERVER</TYPE>
							<MANUFACTURER>HP</MANUFACTURER>
							<STATUS>OK</STATUS>
							<DIAG>
								<FRU>NO_ERROR</FRU>
								<MgmtProc>NO_ERROR</MgmtProc>
								<thermalWarning>NOT_TESTED</thermalWarning>
								<thermalDanger>NOT_TESTED</thermalDanger>
								<Keying>NO_ERROR</Keying>
								<Power>NO_ERROR</Power>
								<Cooling>NO_ERROR</Cooling>
								<Location>NO_ERROR</Location>
								<Failure>NO_ERROR</Failure>
								<Degraded>NO_ERROR</Degraded>
								<AC>NOT_RELEVANT</AC>
								<i2c>NOT_RELEVANT</i2c>
								<oaRedundancy>NOT_RELEVANT</oaRedundancy>
							</DIAG>
							<UIDSTATUS>OFF</UIDSTATUS>
							<PORTMAP>
								<STATUS>OK</STATUS>
								<MEZZ>
									<NUMBER>1</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_ONE</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>3</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>5</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>4</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>5</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
								</MEZZ>
								<MEZZ>
									<NUMBER>2</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_TWO</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>5</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>5</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>6</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>5</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>3</NUMBER>
											<TRAYBAYNUMBER>7</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>5</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>4</NUMBER>
											<TRAYBAYNUMBER>8</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>5</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
								</MEZZ>
								<MEZZ>
									<NUMBER>9</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_ONE</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>1</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>5</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>2</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>5</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
									<DEVICE>
										<NAME>HP FlexFabric 10Gb 2-port 536FLB Adapter</NAME>
										<TYPE>MEZZ_DEV_TYPE_ONE</TYPE>
										<STATUS>OK</STATUS>
										<PORT>
											<NUMBER>1</NUMBER>
											<WWPN>8C:DC:D4:1D:6F:D0</WWPN>
											<TYPE>INTERCONNECT_TYPE_ETH</TYPE>
											<STATUS>OK</STATUS>
											<GUIDS>
												<GUID>
													<TYPE>C</TYPE>
													<FUNCTION>a</FUNCTION>
													<GUID_STRING>8C:DC:D4:1D:6F:D0</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>H</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>8C:DC:D4:1D:6F:D1</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>G</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>20:00:8C:DC:D4:1D:6F:D1</GUID_STRING>
													</GUID>
												</GUIDS>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<WWPN>8C:DC:D4:1D:6F:D8</WWPN>
											<TYPE>INTERCONNECT_TYPE_ETH</TYPE>
											<STATUS>UNKNOWN</STATUS>
											<GUIDS>
												<GUID>
													<TYPE>C</TYPE>
													<FUNCTION>a</FUNCTION>
													<GUID_STRING>8C:DC:D4:1D:6F:D8</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>H</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>8C:DC:D4:1D:6F:D9</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>G</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>20:00:8C:DC:D4:1D:6F:D9</GUID_STRING>
													</GUID>
												</GUIDS>
										</PORT>
									</DEVICE>
								</MEZZ>
								<MEZZ>
									<NUMBER>13</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_FIXED</TYPE>
									</SLOT>
								</MEZZ>
							</PORTMAP>
							<TEMPS>
								<TEMP>
									<LOCATION>14</LOCATION>
									<DESC>AMBIENT</DESC>
									<C>17</C>
									<THRESHOLD>
										<DESC>CAUTION</DESC>
										<C>42</C>
										<STATUS>Degraded</STATUS>
									</THRESHOLD>
									<THRESHOLD>
										<DESC>CRITICAL</DESC>
										<C>46</C>
										<STATUS>Non-Recoverable Error</STATUS>
									</THRESHOLD>
								</TEMP>
							</TEMPS>
							<POWER>
								<POWERSTATE>ON</POWERSTATE>
								<POWERMODE>UNKNOWN</POWERMODE>
								<POWER_CONSUMED>208</POWER_CONSUMED>
							</POWER>
							<VMSTAT>
								<SUPPORT>VM_SUPPORTED</SUPPORT>
								<CDROMSTAT>VM_DEV_STATUS_DISCONNECTED</CDROMSTAT>
								<CDROMURL></CDROMURL>
								<FLOPPYSTAT>VM_DEV_STATUS_DISCONNECTED</FLOPPYSTAT>
								<FLOPPYURL></FLOPPYURL>
							</VMSTAT>
							<cUUID>30373237-3132-5A43-3335-32315936584A</cUUID>
							<CONJOINABLE>false</CONJOINABLE>
						</BLADE>
						<BLADE>
							<BAY>
								<CONNECTION>6</CONNECTION>
							</BAY>
							<MGMTIPADDR>0.0.0.0</MGMTIPADDR>
							<MGMTIPV6ADDR_LL>fe80::9eb6:54ff:fe79:21cc/64</MGMTIPV6ADDR_LL>
							<MGMTDNSNAME>.machine.example.com</MGMTDNSNAME>
							<MGMTPN>iLO4</MGMTPN>
							<MGMTFWVERSION>2.55 Aug 16 2017</MGMTFWVERSION>
							<PN>727021-B21</PN>
							<BLADEROMVER>I36 02/17/2017</BLADEROMVER>
							<NAME>bbmi</NAME>
							<PWRM>1.0.9</PWRM>
							<VLAN>1</VLAN>
							<SPN>ProLiant BL460c Gen9</SPN>
							<BSN>CZ35230JE3</BSN>
							<UUID>727021CZ35230JE3</UUID>
							<TYPE>SERVER</TYPE>
							<MANUFACTURER>HP</MANUFACTURER>
							<STATUS>OK</STATUS>
							<DIAG>
								<FRU>NO_ERROR</FRU>
								<MgmtProc>NO_ERROR</MgmtProc>
								<thermalWarning>NOT_TESTED</thermalWarning>
								<thermalDanger>NOT_TESTED</thermalDanger>
								<Keying>NO_ERROR</Keying>
								<Power>NO_ERROR</Power>
								<Cooling>NO_ERROR</Cooling>
								<Location>NOT_TESTED</Location>
								<Failure>NO_ERROR</Failure>
								<Degraded>NO_ERROR</Degraded>
								<AC>NOT_RELEVANT</AC>
								<i2c>NOT_RELEVANT</i2c>
								<oaRedundancy>NOT_RELEVANT</oaRedundancy>
							</DIAG>
							<UIDSTATUS>OFF</UIDSTATUS>
							<PORTMAP>
								<STATUS>OK</STATUS>
								<MEZZ>
									<NUMBER>1</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_ONE</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>3</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>6</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>4</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>6</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
								</MEZZ>
								<MEZZ>
									<NUMBER>2</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_TWO</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>5</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>6</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>6</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>6</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>3</NUMBER>
											<TRAYBAYNUMBER>7</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>6</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>4</NUMBER>
											<TRAYBAYNUMBER>8</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>6</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
								</MEZZ>
								<MEZZ>
									<NUMBER>9</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_ONE</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>1</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>6</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>2</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>6</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
									<DEVICE>
										<NAME>HP FlexFabric 10Gb 2-port 536FLB Adapter</NAME>
										<TYPE>MEZZ_DEV_TYPE_ONE</TYPE>
										<STATUS>OK</STATUS>
										<PORT>
											<NUMBER>1</NUMBER>
											<WWPN>8C:DC:D4:1A:8F:A0</WWPN>
											<TYPE>INTERCONNECT_TYPE_ETH</TYPE>
											<STATUS>OK</STATUS>
											<GUIDS>
												<GUID>
													<TYPE>C</TYPE>
													<FUNCTION>a</FUNCTION>
													<GUID_STRING>8C:DC:D4:1A:8F:A0</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>H</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>8C:DC:D4:1A:8F:A1</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>G</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>20:00:8C:DC:D4:1A:8F:A1</GUID_STRING>
													</GUID>
												</GUIDS>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<WWPN>8C:DC:D4:1A:8F:A8</WWPN>
											<TYPE>INTERCONNECT_TYPE_ETH</TYPE>
											<STATUS>UNKNOWN</STATUS>
											<GUIDS>
												<GUID>
													<TYPE>C</TYPE>
													<FUNCTION>a</FUNCTION>
													<GUID_STRING>8C:DC:D4:1A:8F:A8</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>H</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>8C:DC:D4:1A:8F:A9</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>G</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>20:00:8C:DC:D4:1A:8F:A9</GUID_STRING>
													</GUID>
												</GUIDS>
										</PORT>
									</DEVICE>
								</MEZZ>
								<MEZZ>
									<NUMBER>13</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_FIXED</TYPE>
									</SLOT>
								</MEZZ>
							</PORTMAP>
							<TEMPS>
								<TEMP>
									<LOCATION>14</LOCATION>
									<DESC>AMBIENT</DESC>
									<C>16</C>
									<THRESHOLD>
										<DESC>CAUTION</DESC>
										<C>42</C>
										<STATUS>Degraded</STATUS>
									</THRESHOLD>
									<THRESHOLD>
										<DESC>CRITICAL</DESC>
										<C>46</C>
										<STATUS>Non-Recoverable Error</STATUS>
									</THRESHOLD>
								</TEMP>
							</TEMPS>
							<POWER>
								<POWERSTATE>ON</POWERSTATE>
								<POWERMODE>UNKNOWN</POWERMODE>
								<POWER_CONSUMED>193</POWER_CONSUMED>
							</POWER>
							<VMSTAT>
								<SUPPORT>VM_SUPPORTED</SUPPORT>
								<CDROMSTAT>VM_DEV_STATUS_DISCONNECTED</CDROMSTAT>
								<CDROMURL></CDROMURL>
								<FLOPPYSTAT>VM_DEV_STATUS_DISCONNECTED</FLOPPYSTAT>
								<FLOPPYURL></FLOPPYURL>
							</VMSTAT>
							<cUUID>30373237-3132-5A43-3335-3233304A4533</cUUID>
							<CONJOINABLE>false</CONJOINABLE>
						</BLADE>
						<BLADE>
							<BAY>
								<CONNECTION>7</CONNECTION>
							</BAY>
							<MGMTIPADDR>0.0.0.0</MGMTIPADDR>
							<MGMTIPV6ADDR_LL>fe80::b6b5:2fff:fe58:ed10/64</MGMTIPV6ADDR_LL>
							<MGMTDNSNAME>.machine.example.com</MGMTDNSNAME>
							<MGMTPN>iLO4</MGMTPN>
							<MGMTFWVERSION>2.55 Aug 16 2017</MGMTFWVERSION>
							<PN>641016-B21     </PN>
							<BLADEROMVER>I31 06/01/2015</BLADEROMVER>
							<NAME>bbmi</NAME>
							<PWRM>3.3.0</PWRM>
							<VLAN>1</VLAN>
							<SPN>ProLiant BL460c Gen8</SPN>
							<BSN>CZ33067KDV      </BSN>
							<UUID>641016CZ33067KDV</UUID>
							<TYPE>SERVER</TYPE>
							<MANUFACTURER>HP</MANUFACTURER>
							<STATUS>OK</STATUS>
							<DIAG>
								<FRU>NO_ERROR</FRU>
								<MgmtProc>NO_ERROR</MgmtProc>
								<thermalWarning>NOT_TESTED</thermalWarning>
								<thermalDanger>NOT_TESTED</thermalDanger>
								<Keying>NO_ERROR</Keying>
								<Power>NO_ERROR</Power>
								<Cooling>NO_ERROR</Cooling>
								<Location>NOT_TESTED</Location>
								<Failure>NO_ERROR</Failure>
								<Degraded>NO_ERROR</Degraded>
								<AC>NOT_RELEVANT</AC>
								<i2c>NOT_RELEVANT</i2c>
								<oaRedundancy>NOT_RELEVANT</oaRedundancy>
							</DIAG>
							<UIDSTATUS>OFF</UIDSTATUS>
							<PORTMAP>
								<STATUS>OK</STATUS>
								<MEZZ>
									<NUMBER>1</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_ONE</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>3</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>7</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>4</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>7</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
								</MEZZ>
								<MEZZ>
									<NUMBER>2</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_TWO</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>5</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>7</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>6</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>7</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>3</NUMBER>
											<TRAYBAYNUMBER>7</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>7</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>4</NUMBER>
											<TRAYBAYNUMBER>8</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>7</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
								</MEZZ>
								<MEZZ>
									<NUMBER>9</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_ONE</TYPE>
										<PORT>
											<NUMBER>1</NUMBER>
											<TRAYBAYNUMBER>1</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>7</TRAYPORTNUMBER>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<TRAYBAYNUMBER>2</TRAYBAYNUMBER>
											<TRAYPORTNUMBER>7</TRAYPORTNUMBER>
										</PORT>
									</SLOT>
									<DEVICE>
										<NAME>HP FlexFabric 10Gb 2-port 554FLB Adapter</NAME>
										<TYPE>MEZZ_DEV_TYPE_ONE</TYPE>
										<STATUS>OK</STATUS>
										<PORT>
											<NUMBER>1</NUMBER>
											<WWPN>AC:16:2D:AB:14:98</WWPN>
											<TYPE>INTERCONNECT_TYPE_ETH</TYPE>
											<STATUS>OK</STATUS>
											<GUIDS>
												<GUID>
													<TYPE>C</TYPE>
													<FUNCTION>a</FUNCTION>
													<GUID_STRING>AC:16:2D:AB:14:98</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>H</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>AC:16:2D:AB:14:99</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>G</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>10:00:AC:16:2D:AB:14:99</GUID_STRING>
													</GUID>
												</GUIDS>
										</PORT>
										<PORT>
											<NUMBER>2</NUMBER>
											<WWPN>AC:16:2D:AB:14:9C</WWPN>
											<TYPE>INTERCONNECT_TYPE_ETH</TYPE>
											<STATUS>UNKNOWN</STATUS>
											<GUIDS>
												<GUID>
													<TYPE>C</TYPE>
													<FUNCTION>a</FUNCTION>
													<GUID_STRING>AC:16:2D:AB:14:9C</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>H</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>AC:16:2D:AB:14:9D</GUID_STRING>
													</GUID>
												<GUID>
													<TYPE>G</TYPE>
													<FUNCTION>b</FUNCTION>
													<GUID_STRING>10:00:AC:16:2D:AB:14:9D</GUID_STRING>
													</GUID>
												</GUIDS>
										</PORT>
									</DEVICE>
								</MEZZ>
								<MEZZ>
									<NUMBER>13</NUMBER>
									<SLOT>
										<TYPE>MEZZ_SLOT_TYPE_FIXED</TYPE>
									</SLOT>
								</MEZZ>
							</PORTMAP>
							<TEMPS>
								<TEMP>
									<LOCATION>14</LOCATION>
									<DESC>AMBIENT</DESC>
									<C>15</C>
									<THRESHOLD>
										<DESC>CAUTION</DESC>
										<C>42</C>
										<STATUS>Degraded</STATUS>
									</THRESHOLD>
									<THRESHOLD>
										<DESC>CRITICAL</DESC>
										<C>46</C>
										<STATUS>Non-Recoverable Error</STATUS>
									</THRESHOLD>
								</TEMP>
							</TEMPS>
							<POWER>
								<POWERSTATE>ON</POWERSTATE>
								<POWERMODE>UNKNOWN</POWERMODE>
								<POWER_CONSUMED>268</POWER_CONSUMED>
							</POWER>
							<VMSTAT>
								<SUPPORT>VM_SUPPORTED</SUPPORT>
								<CDROMSTAT>VM_DEV_STATUS_DISCONNECTED</CDROMSTAT>
								<CDROMURL></CDROMURL>
								<FLOPPYSTAT>VM_DEV_STATUS_DISCONNECTED</FLOPPYSTAT>
								<FLOPPYURL></FLOPPYURL>
							</VMSTAT>
							<cUUID>30313436-3631-5A43-3333-3036374B4456</cUUID>
							<CONJOINABLE>false</CONJOINABLE>
						</BLADE>
					</BLADES>
					<SWITCHES>
						<BAYS>
							<BAY NAME="1">
								<SIDE>REAR</SIDE>
								<mmHeight>28</mmHeight>
								<mmWidth>193</mmWidth>
								<mmDepth>268</mmDepth>
								<mmXOffset>0</mmXOffset>
								<mmYOffset>95</mmYOffset>
							</BAY>
							<BAY NAME="2">
								<SIDE>REAR</SIDE>
								<mmHeight>28</mmHeight>
								<mmWidth>193</mmWidth>
								<mmDepth>268</mmDepth>
								<mmXOffset>193</mmXOffset>
								<mmYOffset>95</mmYOffset>
							</BAY>
							<BAY NAME="3">
								<SIDE>REAR</SIDE>
								<mmHeight>28</mmHeight>
								<mmWidth>193</mmWidth>
								<mmDepth>268</mmDepth>
								<mmXOffset>0</mmXOffset>
								<mmYOffset>123</mmYOffset>
							</BAY>
							<BAY NAME="4">
								<SIDE>REAR</SIDE>
								<mmHeight>28</mmHeight>
								<mmWidth>193</mmWidth>
								<mmDepth>268</mmDepth>
								<mmXOffset>193</mmXOffset>
								<mmYOffset>123</mmYOffset>
							</BAY>
							<BAY NAME="5">
								<SIDE>REAR</SIDE>
								<mmHeight>28</mmHeight>
								<mmWidth>193</mmWidth>
								<mmDepth>268</mmDepth>
								<mmXOffset>0</mmXOffset>
								<mmYOffset>151</mmYOffset>
							</BAY>
							<BAY NAME="6">
								<SIDE>REAR</SIDE>
								<mmHeight>28</mmHeight>
								<mmWidth>193</mmWidth>
								<mmDepth>268</mmDepth>
								<mmXOffset>193</mmXOffset>
								<mmYOffset>151</mmYOffset>
							</BAY>
							<BAY NAME="7">
								<SIDE>REAR</SIDE>
								<mmHeight>28</mmHeight>
								<mmWidth>193</mmWidth>
								<mmDepth>268</mmDepth>
								<mmXOffset>0</mmXOffset>
								<mmYOffset>179</mmYOffset>
							</BAY>
							<BAY NAME="8">
								<SIDE>REAR</SIDE>
								<mmHeight>28</mmHeight>
								<mmWidth>193</mmWidth>
								<mmDepth>268</mmDepth>
								<mmXOffset>193</mmXOffset>
								<mmYOffset>179</mmYOffset>
							</BAY>
						</BAYS>
						<SWITCH>
							<BAY>
								<CONNECTION>1</CONNECTION>
							</BAY>
							<MGMTIPADDR>0.0.0.0</MGMTIPADDR>
							<MGMTURL></MGMTURL>
							<BSN>Y2H70101N0</BSN>
							<PN>538113-B21</PN>
							<FWRI>[Unknown]</FWRI>
							<FABRICTYPE>INTERCONNECT_TYPE_ETH</FABRICTYPE>
							<SPN>HP 10GbE Pass-Thru Module</SPN>
							<MANUFACTURER>HP</MANUFACTURER>
							<STATUS>OK</STATUS>
							<DIAG>
								<FRU>NO_ERROR</FRU>
								<MgmtProc>NOT_TESTED</MgmtProc>
								<thermalWarning>NO_ERROR</thermalWarning>
								<thermalDanger>NO_ERROR</thermalDanger>
								<Keying>NO_ERROR</Keying>
								<Power>NO_ERROR</Power>
								<Cooling>NOT_RELEVANT</Cooling>
								<Location>NOT_RELEVANT</Location>
								<Failure>NO_ERROR</Failure>
								<Degraded>NO_ERROR</Degraded>
								<AC>NOT_RELEVANT</AC>
								<i2c>NOT_RELEVANT</i2c>
								<oaRedundancy>NOT_RELEVANT</oaRedundancy>
							</DIAG>
							<UIDSTATUS>OFF</UIDSTATUS>
							<PORTMAP>
								<STATUS>OK</STATUS>
								<PASSTHRU_MODE_ENABLED>ENABLED</PASSTHRU_MODE_ENABLED>
								<SLOT>
									<NUMBER>1</NUMBER>
									<TYPE>INTERCONNECT_TYPE_ETH</TYPE>
									<PORT>
										<NUMBER>1</NUMBER>
										<BLADEBAYNUMBER>1</BLADEBAYNUMBER>
										<BLADEMEZZNUMBER>9</BLADEMEZZNUMBER>
										<BLADEMEZZPORTNUMBER>1</BLADEMEZZPORTNUMBER>
										<STATUS>OK</STATUS>
										<ENABLED>UNKNOWN</ENABLED>
										<UID_STATUS>UNKNOWN</UID_STATUS>
										<LINK_LED_STATUS>UNKNOWN</LINK_LED_STATUS>
									</PORT>
									<PORT>
										<NUMBER>2</NUMBER>
										<BLADEBAYNUMBER>2</BLADEBAYNUMBER>
										<BLADEMEZZNUMBER>9</BLADEMEZZNUMBER>
										<BLADEMEZZPORTNUMBER>1</BLADEMEZZPORTNUMBER>
										<STATUS>OK</STATUS>
										<ENABLED>UNKNOWN</ENABLED>
										<UID_STATUS>UNKNOWN</UID_STATUS>
										<LINK_LED_STATUS>UNKNOWN</LINK_LED_STATUS>
									</PORT>
									<PORT>
										<NUMBER>3</NUMBER>
										<BLADEBAYNUMBER>3</BLADEBAYNUMBER>
										<BLADEMEZZNUMBER>9</BLADEMEZZNUMBER>
										<BLADEMEZZPORTNUMBER>1</BLADEMEZZPORTNUMBER>
										<STATUS>OK</STATUS>
										<ENABLED>UNKNOWN</ENABLED>
										<UID_STATUS>UNKNOWN</UID_STATUS>
										<LINK_LED_STATUS>UNKNOWN</LINK_LED_STATUS>
									</PORT>
									<PORT>
										<NUMBER>4</NUMBER>
										<BLADEBAYNUMBER>4</BLADEBAYNUMBER>
										<BLADEMEZZNUMBER>9</BLADEMEZZNUMBER>
										<BLADEMEZZPORTNUMBER>1</BLADEMEZZPORTNUMBER>
										<STATUS>OK</STATUS>
										<ENABLED>UNKNOWN</ENABLED>
										<UID_STATUS>UNKNOWN</UID_STATUS>
										<LINK_LED_STATUS>UNKNOWN</LINK_LED_STATUS>
									</PORT>
									<PORT>
										<NUMBER>5</NUMBER>
										<BLADEBAYNUMBER>5</BLADEBAYNUMBER>
										<BLADEMEZZNUMBER>9</BLADEMEZZNUMBER>
										<BLADEMEZZPORTNUMBER>1</BLADEMEZZPORTNUMBER>
										<STATUS>OK</STATUS>
										<ENABLED>UNKNOWN</ENABLED>
										<UID_STATUS>UNKNOWN</UID_STATUS>
										<LINK_LED_STATUS>UNKNOWN</LINK_LED_STATUS>
									</PORT>
									<PORT>
										<NUMBER>6</NUMBER>
										<BLADEBAYNUMBER>6</BLADEBAYNUMBER>
										<BLADEMEZZNUMBER>9</BLADEMEZZNUMBER>
										<BLADEMEZZPORTNUMBER>1</BLADEMEZZPORTNUMBER>
										<STATUS>OK</STATUS>
										<ENABLED>UNKNOWN</ENABLED>
										<UID_STATUS>UNKNOWN</UID_STATUS>
										<LINK_LED_STATUS>UNKNOWN</LINK_LED_STATUS>
									</PORT>
									<PORT>
										<NUMBER>7</NUMBER>
										<BLADEBAYNUMBER>7</BLADEBAYNUMBER>
										<BLADEMEZZNUMBER>9</BLADEMEZZNUMBER>
										<BLADEMEZZPORTNUMBER>1</BLADEMEZZPORTNUMBER>
										<STATUS>OK</STATUS>
										<ENABLED>UNKNOWN</ENABLED>
										<UID_STATUS>UNKNOWN</UID_STATUS>
										<LINK_LED_STATUS>UNKNOWN</LINK_LED_STATUS>
									</PORT>
									<PORT>
										<NUMBER>8</NUMBER>
										<BLADEBAYNUMBER>0</BLADEBAYNUMBER>
										<BLADEMEZZNUMBER>0</BLADEMEZZNUMBER>
										<BLADEMEZZPORTNUMBER>0</BLADEMEZZPORTNUMBER>
										<STATUS>UNKNOWN</STATUS>
										<ENABLED>UNKNOWN</ENABLED>
										<UID_STATUS>UNKNOWN</UID_STATUS>
										<LINK_LED_STATUS>UNKNOWN</LINK_LED_STATUS>
									</PORT>
									<PORT>
										<NUMBER>9</NUMBER>
										<BLADEBAYNUMBER>0</BLADEBAYNUMBER>
										<BLADEMEZZNUMBER>0</BLADEMEZZNUMBER>
										<BLADEMEZZPORTNUMBER>0</BLADEMEZZPORTNUMBER>
										<STATUS>UNKNOWN</STATUS>
										<ENABLED>UNKNOWN</ENABLED>
										<UID_STATUS>UNKNOWN</UID_STATUS>
										<LINK_LED_STATUS>UNKNOWN</LINK_LED_STATUS>
									</PORT>
									<PORT>
										<NUMBER>10</NUMBER>
										<BLADEBAYNUMBER>0</BLADEBAYNUMBER>
										<BLADEMEZZNUMBER>0</BLADEMEZZNUMBER>
										<BLADEMEZZPORTNUMBER>0</BLADEMEZZPORTNUMBER>
										<STATUS>UNKNOWN</STATUS>
										<ENABLED>UNKNOWN</ENABLED>
										<UID_STATUS>UNKNOWN</UID_STATUS>
										<LINK_LED_STATUS>UNKNOWN</LINK_LED_STATUS>
									</PORT>
									<PORT>
										<NUMBER>11</NUMBER>
										<BLADEBAYNUMBER>0</BLADEBAYNUMBER>
										<BLADEMEZZNUMBER>0</BLADEMEZZNUMBER>
										<BLADEMEZZPORTNUMBER>0</BLADEMEZZPORTNUMBER>
										<STATUS>UNKNOWN</STATUS>
										<ENABLED>UNKNOWN</ENABLED>
										<UID_STATUS>UNKNOWN</UID_STATUS>
										<LINK_LED_STATUS>UNKNOWN</LINK_LED_STATUS>
									</PORT>
									<PORT>
										<NUMBER>12</NUMBER>
										<BLADEBAYNUMBER>0</BLADEBAYNUMBER>
										<BLADEMEZZNUMBER>0</BLADEMEZZNUMBER>
										<BLADEMEZZPORTNUMBER>0</BLADEMEZZPORTNUMBER>
										<STATUS>UNKNOWN</STATUS>
										<ENABLED>UNKNOWN</ENABLED>
										<UID_STATUS>UNKNOWN</UID_STATUS>
										<LINK_LED_STATUS>UNKNOWN</LINK_LED_STATUS>
									</PORT>
									<PORT>
										<NUMBER>13</NUMBER>
										<BLADEBAYNUMBER>0</BLADEBAYNUMBER>
										<BLADEMEZZNUMBER>0</BLADEMEZZNUMBER>
										<BLADEMEZZPORTNUMBER>0</BLADEMEZZPORTNUMBER>
										<STATUS>UNKNOWN</STATUS>
										<ENABLED>UNKNOWN</ENABLED>
										<UID_STATUS>UNKNOWN</UID_STATUS>
										<LINK_LED_STATUS>UNKNOWN</LINK_LED_STATUS>
									</PORT>
									<PORT>
										<NUMBER>14</NUMBER>
										<BLADEBAYNUMBER>0</BLADEBAYNUMBER>
										<BLADEMEZZNUMBER>0</BLADEMEZZNUMBER>
										<BLADEMEZZPORTNUMBER>0</BLADEMEZZPORTNUMBER>
										<STATUS>UNKNOWN</STATUS>
										<ENABLED>UNKNOWN</ENABLED>
										<UID_STATUS>UNKNOWN</UID_STATUS>
										<LINK_LED_STATUS>UNKNOWN</LINK_LED_STATUS>
									</PORT>
									<PORT>
										<NUMBER>15</NUMBER>
										<BLADEBAYNUMBER>0</BLADEBAYNUMBER>
										<BLADEMEZZNUMBER>0</BLADEMEZZNUMBER>
										<BLADEMEZZPORTNUMBER>0</BLADEMEZZPORTNUMBER>
										<STATUS>UNKNOWN</STATUS>
										<ENABLED>UNKNOWN</ENABLED>
										<UID_STATUS>UNKNOWN</UID_STATUS>
										<LINK_LED_STATUS>UNKNOWN</LINK_LED_STATUS>
									</PORT>
									<PORT>
										<NUMBER>16</NUMBER>
										<BLADEBAYNUMBER>0</BLADEBAYNUMBER>
										<BLADEMEZZNUMBER>0</BLADEMEZZNUMBER>
										<BLADEMEZZPORTNUMBER>0</BLADEMEZZPORTNUMBER>
										<STATUS>UNKNOWN</STATUS>
										<ENABLED>UNKNOWN</ENABLED>
										<UID_STATUS>UNKNOWN</UID_STATUS>
										<LINK_LED_STATUS>UNKNOWN</LINK_LED_STATUS>
									</PORT>
								</SLOT>
							</PORTMAP>
							<TEMPS>
								<TEMP>
									<LOCATION>13</LOCATION>
									<DESC>AMBIENT</DESC>
									<C>36</C>
									<THRESHOLD>
										<DESC>CAUTION</DESC>
										<C>79</C>
										<STATUS>Degraded</STATUS>
									</THRESHOLD>
									<THRESHOLD>
										<DESC>CRITICAL</DESC>
										<C>81</C>
										<STATUS>Non-Recoverable Error</STATUS>
									</THRESHOLD>
								</TEMP>
							</TEMPS>
							<THERMAL>OK</THERMAL>
							<POWER>
								<POWERSTATE>ON</POWERSTATE>
								<POWER_ON_WATTAGE>75</POWER_ON_WATTAGE>
								<POWER_OFF_WATTAGE>3</POWER_OFF_WATTAGE>
							</POWER>
						</SWITCH>
					</SWITCHES>
					<MANAGERS>
						<BAYS>
							<BAY NAME="1">
								<SIDE>REAR</SIDE>
								<mmHeight>21</mmHeight>
								<mmWidth>160</mmWidth>
								<mmDepth>177</mmDepth>
								<mmXOffset>0</mmXOffset>
								<mmYOffset>225</mmYOffset>
							</BAY>
							<BAY NAME="2">
								<SIDE>REAR</SIDE>
								<mmHeight>21</mmHeight>
								<mmWidth>160</mmWidth>
								<mmDepth>177</mmDepth>
								<mmXOffset>255</mmXOffset>
								<mmYOffset>225</mmYOffset>
							</BAY>
						</BAYS>
						<MANAGER>
							<BAY>
								<CONNECTION>1</CONNECTION>
							</BAY>
							<MGMTIPADDR>127.0.0.1</MGMTIPADDR>
							<NAME>OA-1C98EC1F8273</NAME>
							<ROLE>ACTIVE</ROLE>
							<STATUS>OK</STATUS>
							<FWRI>4.70</FWRI>
							<DIAG>
								<FRU>NO_ERROR</FRU>
								<MgmtProc>NOT_TESTED</MgmtProc>
								<thermalWarning>NOT_RELEVANT</thermalWarning>
								<thermalDanger>NOT_RELEVANT</thermalDanger>
								<Keying>NOT_RELEVANT</Keying>
								<Power>NOT_RELEVANT</Power>
								<Cooling>NOT_RELEVANT</Cooling>
								<Location>NOT_RELEVANT</Location>
								<Failure>NOT_TESTED</Failure>
								<Degraded>NOT_TESTED</Degraded>
								<AC>NOT_RELEVANT</AC>
								<i2c>NOT_RELEVANT</i2c>
								<oaRedundancy>NOT_TESTED</oaRedundancy>
							</DIAG>
							<UIDSTATUS>OFF</UIDSTATUS>
							<WIZARDSTATUS>WIZARD_SETUP_COMPLETE</WIZARDSTATUS>
							<YOUAREHERE>true</YOUAREHERE>
							<BSN>OB6BCP1616    </BSN>
							<UUID>09OB6BCP1616    </UUID>
							<SPN>BladeSystem c7000 DDR2 Onboard Administrator with KVM</SPN>
							<MANUFACTURER>HP</MANUFACTURER>
							<TEMPS>
								<TEMP>
									<LOCATION>17</LOCATION>
									<DESC>AMBIENT</DESC>
									<C>34</C>
									<THRESHOLD>
										<DESC>CAUTION</DESC>
										<C>75</C>
										<STATUS>Degraded</STATUS>
									</THRESHOLD>
									<THRESHOLD>
										<DESC>CRITICAL</DESC>
										<C>80</C>
										<STATUS>Non-Recoverable Error</STATUS>
									</THRESHOLD>
								</TEMP>
							</TEMPS>
							<POWER>
								<POWERSTATE>ON</POWERSTATE>
							</POWER>
							<MACADDR>1C:98:EC:1F:82:73</MACADDR>
							<IPV6STATUS>ENABLED</IPV6STATUS>
							<MGMTIPv6ADDR1>fe80::1e98:ecff:fe1f:8273/64</MGMTIPv6ADDR1>
						</MANAGER>
						<MANAGER>
							<BAY>
								<CONNECTION>2</CONNECTION>
							</BAY>
							<MGMTIPADDR>10.193.251.23</MGMTIPADDR>
							<NAME>OA-94188272E9F5</NAME>
							<ROLE>STANDBY</ROLE>
							<STATUS>OK</STATUS>
							<FWRI>4.70</FWRI>
							<DIAG>
								<FRU>NO_ERROR</FRU>
								<MgmtProc>NO_ERROR</MgmtProc>
								<thermalWarning>NOT_RELEVANT</thermalWarning>
								<thermalDanger>NOT_RELEVANT</thermalDanger>
								<Keying>NOT_RELEVANT</Keying>
								<Power>NOT_RELEVANT</Power>
								<Cooling>NOT_RELEVANT</Cooling>
								<Location>NOT_RELEVANT</Location>
								<Failure>NOT_TESTED</Failure>
								<Degraded>NOT_TESTED</Degraded>
								<AC>NOT_RELEVANT</AC>
								<i2c>NOT_RELEVANT</i2c>
								<oaRedundancy>NOT_TESTED</oaRedundancy>
							</DIAG>
							<UIDSTATUS>OFF</UIDSTATUS>
							<WIZARDSTATUS>WIZARD_SETUP_COMPLETE</WIZARDSTATUS>
							<YOUAREHERE>false</YOUAREHERE>
							<BSN>OB73CP2812    </BSN>
							<UUID>09OB73CP2812    </UUID>
							<SPN>BladeSystem c7000 DDR2 Onboard Administrator with KVM</SPN>
							<MANUFACTURER>HP</MANUFACTURER>
							<TEMPS>
								<TEMP>
									<LOCATION>17</LOCATION>
									<DESC>AMBIENT</DESC>
									<C>34</C>
									<THRESHOLD>
										<DESC>CAUTION</DESC>
										<C>75</C>
										<STATUS>Degraded</STATUS>
									</THRESHOLD>
									<THRESHOLD>
										<DESC>CRITICAL</DESC>
										<C>80</C>
										<STATUS>Non-Recoverable Error</STATUS>
									</THRESHOLD>
								</TEMP>
							</TEMPS>
							<POWER>
								<POWERSTATE>ON</POWERSTATE>
							</POWER>
							<MACADDR>94:18:82:72:E9:F5</MACADDR>
							<IPV6STATUS>ENABLED</IPV6STATUS>
							<MGMTIPv6ADDR2>fe80::9618:82ff:fe72:e9f5/64</MGMTIPv6ADDR2>
						</MANAGER>
					</MANAGERS>
					<LCDS>
						<BAYS>
							<BAY NAME="1">
								<SIDE>FRONT</SIDE>
								<mmHeight>55</mmHeight>
								<mmWidth>92</mmWidth>
								<mmDepth>15</mmDepth>
								<mmXOffset>145</mmXOffset>
								<mmYOffset>365</mmYOffset>
							</BAY>
						</BAYS>
						<LCD>
							<BAY>
								<CONNECTION>1</CONNECTION>
							</BAY>
							<STATUS>OK</STATUS>
							<DIAG>
								<FRU>NO_ERROR</FRU>
								<MgmtProc>NOT_RELEVANT</MgmtProc>
								<thermalWarning>NOT_RELEVANT</thermalWarning>
								<thermalDanger>NOT_RELEVANT</thermalDanger>
								<Keying>NOT_RELEVANT</Keying>
								<Power>NOT_RELEVANT</Power>
								<Cooling>NOT_RELEVANT</Cooling>
								<Location>NOT_RELEVANT</Location>
								<Failure>NOT_TESTED</Failure>
								<Degraded>NOT_TESTED</Degraded>
								<AC>NOT_RELEVANT</AC>
								<i2c>NOT_RELEVANT</i2c>
								<oaRedundancy>NOT_RELEVANT</oaRedundancy>
							</DIAG>
							<SPN>BladeSystem c7000 Insight Display</SPN>
							<MANUFACTURER>HP</MANUFACTURER>
							<FWRI>2.a.3</FWRI>
							<IMAGE_URL>/cgi-bin/getLCDImage?oaSessionKey=</IMAGE_URL>
							<PIN_ENABLED>false</PIN_ENABLED>
							<BUTTON_LOCK_ENABLED>false</BUTTON_LOCK_ENABLED>
							<USERNOTES>Upload up to^six lines of^text information and your^320x240 bitmap using the^Onboard Administrator^web user interface</USERNOTES>
							<PN>519349-001</PN>
						</LCD>
					</LCDS>
					<FANS>
						<STATUS>OK</STATUS>
						<REDUNDANCY>REDUNDANT</REDUNDANCY>
						<WANTED_FANS>10</WANTED_FANS>
						<NEEDED_FANS>9</NEEDED_FANS>
						<BAYS>
							<BAY NAME="1">
								<SIDE>REAR</SIDE>
								<mmHeight>93</mmHeight>
								<mmWidth>78</mmWidth>
								<mmDepth>194</mmDepth>
								<mmXOffset>20</mmXOffset>
								<mmYOffset>0</mmYOffset>
							</BAY>
							<BAY NAME="2">
								<SIDE>REAR</SIDE>
								<mmHeight>93</mmHeight>
								<mmWidth>78</mmWidth>
								<mmDepth>194</mmDepth>
								<mmXOffset>98</mmXOffset>
								<mmYOffset>0</mmYOffset>
							</BAY>
							<BAY NAME="3">
								<SIDE>REAR</SIDE>
								<mmHeight>93</mmHeight>
								<mmWidth>78</mmWidth>
								<mmDepth>194</mmDepth>
								<mmXOffset>176</mmXOffset>
								<mmYOffset>0</mmYOffset>
							</BAY>
							<BAY NAME="4">
								<SIDE>REAR</SIDE>
								<mmHeight>93</mmHeight>
								<mmWidth>78</mmWidth>
								<mmDepth>194</mmDepth>
								<mmXOffset>254</mmXOffset>
								<mmYOffset>0</mmYOffset>
							</BAY>
							<BAY NAME="5">
								<SIDE>REAR</SIDE>
								<mmHeight>93</mmHeight>
								<mmWidth>78</mmWidth>
								<mmDepth>194</mmDepth>
								<mmXOffset>332</mmXOffset>
								<mmYOffset>0</mmYOffset>
							</BAY>
							<BAY NAME="6">
								<SIDE>REAR</SIDE>
								<mmHeight>93</mmHeight>
								<mmWidth>78</mmWidth>
								<mmDepth>194</mmDepth>
								<mmXOffset>20</mmXOffset>
								<mmYOffset>261</mmYOffset>
							</BAY>
							<BAY NAME="7">
								<SIDE>REAR</SIDE>
								<mmHeight>93</mmHeight>
								<mmWidth>78</mmWidth>
								<mmDepth>194</mmDepth>
								<mmXOffset>98</mmXOffset>
								<mmYOffset>261</mmYOffset>
							</BAY>
							<BAY NAME="8">
								<SIDE>REAR</SIDE>
								<mmHeight>93</mmHeight>
								<mmWidth>78</mmWidth>
								<mmDepth>194</mmDepth>
								<mmXOffset>176</mmXOffset>
								<mmYOffset>261</mmYOffset>
							</BAY>
							<BAY NAME="9">
								<SIDE>REAR</SIDE>
								<mmHeight>93</mmHeight>
								<mmWidth>78</mmWidth>
								<mmDepth>194</mmDepth>
								<mmXOffset>254</mmXOffset>
								<mmYOffset>261</mmYOffset>
							</BAY>
							<BAY NAME="10">
								<SIDE>REAR</SIDE>
								<mmHeight>93</mmHeight>
								<mmWidth>78</mmWidth>
								<mmDepth>194</mmDepth>
								<mmXOffset>332</mmXOffset>
								<mmYOffset>261</mmYOffset>
							</BAY>
						</BAYS>
						<FAN>
							<BAY>
								<CONNECTION>1</CONNECTION>
							</BAY>
							<STATUS>OK</STATUS>
							<PN>412140-B21</PN>
							<PRODUCTNAME>Active Cool 200 Fan</PRODUCTNAME>
							<PWR_USED>7</PWR_USED>
							<RPM_CUR>5502</RPM_CUR>
							<RPM_MAX>18000</RPM_MAX>
							<RPM_MIN>600</RPM_MIN>
						</FAN>
						<FAN>
							<BAY>
								<CONNECTION>2</CONNECTION>
							</BAY>
							<STATUS>OK</STATUS>
							<PN>412140-B21</PN>
							<PRODUCTNAME>Active Cool 200 Fan</PRODUCTNAME>
							<PWR_USED>9</PWR_USED>
							<RPM_CUR>5500</RPM_CUR>
							<RPM_MAX>18000</RPM_MAX>
							<RPM_MIN>600</RPM_MIN>
						</FAN>
						<FAN>
							<BAY>
								<CONNECTION>3</CONNECTION>
							</BAY>
							<STATUS>OK</STATUS>
							<PN>412140-B21</PN>
							<PRODUCTNAME>Active Cool 200 Fan</PRODUCTNAME>
							<PWR_USED>9</PWR_USED>
							<RPM_CUR>5500</RPM_CUR>
							<RPM_MAX>18000</RPM_MAX>
							<RPM_MIN>600</RPM_MIN>
						</FAN>
						<FAN>
							<BAY>
								<CONNECTION>4</CONNECTION>
							</BAY>
							<STATUS>OK</STATUS>
							<PN>412140-B21</PN>
							<PRODUCTNAME>Active Cool 200 Fan</PRODUCTNAME>
							<PWR_USED>7</PWR_USED>
							<RPM_CUR>5499</RPM_CUR>
							<RPM_MAX>18000</RPM_MAX>
							<RPM_MIN>600</RPM_MIN>
						</FAN>
						<FAN>
							<BAY>
								<CONNECTION>5</CONNECTION>
							</BAY>
							<STATUS>OK</STATUS>
							<PN>412140-B21</PN>
							<PRODUCTNAME>Active Cool 200 Fan</PRODUCTNAME>
							<PWR_USED>9</PWR_USED>
							<RPM_CUR>5499</RPM_CUR>
							<RPM_MAX>18000</RPM_MAX>
							<RPM_MIN>600</RPM_MIN>
						</FAN>
						<FAN>
							<BAY>
								<CONNECTION>6</CONNECTION>
							</BAY>
							<STATUS>OK</STATUS>
							<PN>412140-B21</PN>
							<PRODUCTNAME>Active Cool 200 Fan</PRODUCTNAME>
							<PWR_USED>9</PWR_USED>
							<RPM_CUR>5499</RPM_CUR>
							<RPM_MAX>18000</RPM_MAX>
							<RPM_MIN>600</RPM_MIN>
						</FAN>
						<FAN>
							<BAY>
								<CONNECTION>7</CONNECTION>
							</BAY>
							<STATUS>OK</STATUS>
							<PN>412140-B21</PN>
							<PRODUCTNAME>Active Cool 200 Fan</PRODUCTNAME>
							<PWR_USED>9</PWR_USED>
							<RPM_CUR>5500</RPM_CUR>
							<RPM_MAX>18000</RPM_MAX>
							<RPM_MIN>600</RPM_MIN>
						</FAN>
						<FAN>
							<BAY>
								<CONNECTION>8</CONNECTION>
							</BAY>
							<STATUS>OK</STATUS>
							<PN>412140-B21</PN>
							<PRODUCTNAME>Active Cool 200 Fan</PRODUCTNAME>
							<PWR_USED>7</PWR_USED>
							<RPM_CUR>5500</RPM_CUR>
							<RPM_MAX>18000</RPM_MAX>
							<RPM_MIN>600</RPM_MIN>
						</FAN>
						<FAN>
							<BAY>
								<CONNECTION>9</CONNECTION>
							</BAY>
							<STATUS>OK</STATUS>
							<PN>412140-B21</PN>
							<PRODUCTNAME>Active Cool 200 Fan</PRODUCTNAME>
							<PWR_USED>7</PWR_USED>
							<RPM_CUR>5498</RPM_CUR>
							<RPM_MAX>18000</RPM_MAX>
							<RPM_MIN>600</RPM_MIN>
						</FAN>
						<FAN>
							<BAY>
								<CONNECTION>10</CONNECTION>
							</BAY>
							<STATUS>OK</STATUS>
							<PN>412140-B21</PN>
							<PRODUCTNAME>Active Cool 200 Fan</PRODUCTNAME>
							<PWR_USED>7</PWR_USED>
							<RPM_CUR>5500</RPM_CUR>
							<RPM_MAX>18000</RPM_MAX>
							<RPM_MIN>600</RPM_MIN>
						</FAN>
					</FANS>
					<POWER>
						<TYPE>INTERNAL_DC</TYPE>
						<STATUS>OK</STATUS>
						<CAPACITY>5300</CAPACITY>
						<OUTPUT_POWER>9546</OUTPUT_POWER>
						<POWER_CONSUMED>2406</POWER_CONSUMED>
						<REDUNDANT_CAPACITY>2894</REDUNDANT_CAPACITY>
						<REDUNDANCY>REDUNDANT</REDUNDANCY>
						<REDUNDANCYMODE>AC_REDUNDANT</REDUNDANCYMODE>
						<WANTED_PS>2</WANTED_PS>
						<NEEDED_PS>1</NEEDED_PS>
						<DYNAMICPOWERSAVER>false</DYNAMICPOWERSAVER>
						<POWERONFLAG>false</POWERONFLAG>
						<BAYS>
							<BAY NAME="1">
								<SIDE>FRONT</SIDE>
								<mmHeight>56</mmHeight>
								<mmWidth>70</mmWidth>
								<mmDepth>700</mmDepth>
								<mmXOffset>0</mmXOffset>
								<mmYOffset>365</mmYOffset>
							</BAY>
							<BAY NAME="2">
								<SIDE>FRONT</SIDE>
								<mmHeight>56</mmHeight>
								<mmWidth>70</mmWidth>
								<mmDepth>700</mmDepth>
								<mmXOffset>70</mmXOffset>
								<mmYOffset>365</mmYOffset>
							</BAY>
							<BAY NAME="3">
								<SIDE>FRONT</SIDE>
								<mmHeight>56</mmHeight>
								<mmWidth>70</mmWidth>
								<mmDepth>700</mmDepth>
								<mmXOffset>140</mmXOffset>
								<mmYOffset>365</mmYOffset>
							</BAY>
							<BAY NAME="4">
								<SIDE>FRONT</SIDE>
								<mmHeight>56</mmHeight>
								<mmWidth>70</mmWidth>
								<mmDepth>700</mmDepth>
								<mmXOffset>210</mmXOffset>
								<mmYOffset>365</mmYOffset>
							</BAY>
							<BAY NAME="5">
								<SIDE>FRONT</SIDE>
								<mmHeight>56</mmHeight>
								<mmWidth>70</mmWidth>
								<mmDepth>700</mmDepth>
								<mmXOffset>280</mmXOffset>
								<mmYOffset>365</mmYOffset>
							</BAY>
							<BAY NAME="6">
								<SIDE>FRONT</SIDE>
								<mmHeight>56</mmHeight>
								<mmWidth>70</mmWidth>
								<mmDepth>700</mmDepth>
								<mmXOffset>350</mmXOffset>
								<mmYOffset>365</mmYOffset>
							</BAY>
						</BAYS>
						<POWERSUPPLY>
							<BAY>
								<CONNECTION>1</CONNECTION>
							</BAY>
							<STATUS>OK</STATUS>
							<DIAG>
								<FRU>NO_ERROR</FRU>
								<MgmtProc>NOT_RELEVANT</MgmtProc>
								<thermalWarning>NOT_RELEVANT</thermalWarning>
								<thermalDanger>NOT_RELEVANT</thermalDanger>
								<Keying>NOT_RELEVANT</Keying>
								<Power>NOT_RELEVANT</Power>
								<Cooling>NOT_RELEVANT</Cooling>
								<Location>NOT_TESTED</Location>
								<Failure>NO_ERROR</Failure>
								<Degraded>NOT_TESTED</Degraded>
								<AC>NO_ERROR</AC>
								<i2c>NOT_RELEVANT</i2c>
								<oaRedundancy>NOT_RELEVANT</oaRedundancy>
							</DIAG>
							<ACINPUT>OK</ACINPUT>
							<ACTUALOUTPUT>263</ACTUALOUTPUT>
							<CAPACITY>2650</CAPACITY>
							<SN>5DRCA0AHL610QJ</SN>
							<FWRI>0.00</FWRI>
							<PN>733459-B21</PN>
						</POWERSUPPLY>
						<POWERSUPPLY>
							<BAY>
								<CONNECTION>2</CONNECTION>
							</BAY>
							<STATUS>OK</STATUS>
							<DIAG>
								<FRU>NO_ERROR</FRU>
								<MgmtProc>NOT_RELEVANT</MgmtProc>
								<thermalWarning>NOT_RELEVANT</thermalWarning>
								<thermalDanger>NOT_RELEVANT</thermalDanger>
								<Keying>NOT_RELEVANT</Keying>
								<Power>NOT_RELEVANT</Power>
								<Cooling>NOT_RELEVANT</Cooling>
								<Location>NOT_TESTED</Location>
								<Failure>NO_ERROR</Failure>
								<Degraded>NOT_TESTED</Degraded>
								<AC>NO_ERROR</AC>
								<i2c>NOT_RELEVANT</i2c>
								<oaRedundancy>NOT_RELEVANT</oaRedundancy>
							</DIAG>
							<ACINPUT>OK</ACINPUT>
							<ACTUALOUTPUT>263</ACTUALOUTPUT>
							<CAPACITY>2650</CAPACITY>
							<SN>5DRCA0AHL610QE</SN>
							<FWRI>0.00</FWRI>
							<PN>733459-B21</PN>
						</POWERSUPPLY>
						<POWERSUPPLY>
							<BAY>
								<CONNECTION>5</CONNECTION>
							</BAY>
							<STATUS>OK</STATUS>
							<DIAG>
								<FRU>NO_ERROR</FRU>
								<MgmtProc>NOT_RELEVANT</MgmtProc>
								<thermalWarning>NOT_RELEVANT</thermalWarning>
								<thermalDanger>NOT_RELEVANT</thermalDanger>
								<Keying>NOT_RELEVANT</Keying>
								<Power>NOT_RELEVANT</Power>
								<Cooling>NOT_RELEVANT</Cooling>
								<Location>NOT_TESTED</Location>
								<Failure>NO_ERROR</Failure>
								<Degraded>NOT_TESTED</Degraded>
								<AC>NO_ERROR</AC>
								<i2c>NOT_RELEVANT</i2c>
								<oaRedundancy>NOT_RELEVANT</oaRedundancy>
							</DIAG>
							<ACINPUT>OK</ACINPUT>
							<ACTUALOUTPUT>263</ACTUALOUTPUT>
							<CAPACITY>2650</CAPACITY>
							<SN>5DRCA0AHL610Q0</SN>
							<FWRI>0.00</FWRI>
							<PN>733459-B21</PN>
						</POWERSUPPLY>
						<POWERSUPPLY>
							<BAY>
								<CONNECTION>6</CONNECTION>
							</BAY>
							<STATUS>OK</STATUS>
							<DIAG>
								<FRU>NO_ERROR</FRU>
								<MgmtProc>NOT_RELEVANT</MgmtProc>
								<thermalWarning>NOT_RELEVANT</thermalWarning>
								<thermalDanger>NOT_RELEVANT</thermalDanger>
								<Keying>NOT_RELEVANT</Keying>
								<Power>NOT_RELEVANT</Power>
								<Cooling>NOT_RELEVANT</Cooling>
								<Location>NOT_TESTED</Location>
								<Failure>NO_ERROR</Failure>
								<Degraded>NOT_TESTED</Degraded>
								<AC>NO_ERROR</AC>
								<i2c>NOT_RELEVANT</i2c>
								<oaRedundancy>NOT_RELEVANT</oaRedundancy>
							</DIAG>
							<ACINPUT>OK</ACINPUT>
							<ACTUALOUTPUT>263</ACTUALOUTPUT>
							<CAPACITY>2650</CAPACITY>
							<SN>5DRCA0AHL610PW</SN>
							<FWRI>0.00</FWRI>
							<PN>733459-B21</PN>
						</POWERSUPPLY>
						<PDU>413374-B21</PDU>
					</POWER>
					<TEMPS>
						<TEMP>
							<LOCATION>9</LOCATION>
							<DESC>AMBIENT</DESC>
							<C>17</C>
							<THRESHOLD>
								<DESC>CAUTION</DESC>
								<C>42</C>
								<STATUS>Degraded</STATUS>
							</THRESHOLD>
							<THRESHOLD>
								<DESC>CRITICAL</DESC>
								<C>46</C>
								<STATUS>Non-Recoverable Error</STATUS>
							</THRESHOLD>
						</TEMP>
					</TEMPS>
					<VCM>
						<vcmMode>false</vcmMode>
						<vcmUrl>empty</vcmUrl>
						<vcmDomainName></vcmDomainName>
						<vcmDomainId></vcmDomainId>
					</VCM>
					<VM>
						<DVDDRIVE>ABSENT</DVDDRIVE>
					</VM>
				</INFRA2>
				<RK_TPLGY CNT="1">
					<RUID>09CZ372137H3</RUID>
					<ICMB ADDR="A9FE01F0" MFG="232" PROD_ID="0x0009" SER="CZ372137H3" UUID="09CZ372137H3">
						<LEFT />
						<RIGHT />
					</ICMB>
				</RK_TPLGY>
				<SPATIAL>
					<DISCOVERY_RACK>Not Supported</DISCOVERY_RACK>
					<DISCOVERY_DATA>Server does not detect Discovery Services</DISCOVERY_DATA>
					<TAG_VERSION></TAG_VERSION>
					<RACK_ID></RACK_ID>
					<RACK_ID_PN></RACK_ID_PN>
					<RACK_cUUID></RACK_cUUID>
					<RACK_DESCRIPTION></RACK_DESCRIPTION>
					<RACK_UHEIGHT></RACK_UHEIGHT>
					<UPOSITION></UPOSITION>
					<ULOCATION></ULOCATION>
					<cUUID>5A433930-3733-3132-3337-483320202020</cUUID>
					<UHEIGHT>1000</UHEIGHT>
					<UOFFSET>2</UOFFSET>
					<DEVICE_UPOSITION></DEVICE_UPOSITION>
				</SPATIAL>
			</RIMP>
			`),
	}
)

func setup() (r *C7000, err error) {
	viper.SetDefault("debug", true)
	mux = http.NewServeMux()
	server = httptest.NewTLSServer(mux)
	ip := strings.TrimPrefix(server.URL, "https://")
	username := "super"
	password := "test"

	for url := range answers {
		url := url
		mux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
			cookie := http.Cookie{Name: "sessionKey", Value: "sessionKey_test"}
			http.SetCookie(w, &cookie)
			w.Write(answers[url])
		})
	}

	r, err = New(ip, username, password)
	if err != nil {
		return r, err
	}

	return r, err
}

func tearDown() {
	server.Close()
}

func TestHpChassisFwVersion(t *testing.T) {
	expectedAnswer := "4.70"

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.FwVersion()
	if err != nil {
		t.Fatalf("Found errors calling chassis.FwVersion %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestHpChassisPassThru(t *testing.T) {
	expectedAnswer := "10G"

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.PassThru()
	if err != nil {
		t.Fatalf("Found errors calling chassis.PassThru %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestHpChassisSerial(t *testing.T) {
	expectedAnswer := "cz372137h3"

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.Serial()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Serial %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestHpChassisModel(t *testing.T) {
	expectedAnswer := "BladeSystem c7000 DDR2 Onboard Administrator with KVM"

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.Model()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Model %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestHpChassisName(t *testing.T) {
	expectedAnswer := "spare-cz372137h3"

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.Name()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Name %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestHpChassisStatus(t *testing.T) {
	expectedAnswer := "OK"

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.Status()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Status %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestHpChassisPowerKW(t *testing.T) {
	expectedAnswer := 2.406

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.PowerKw()
	if err != nil {
		t.Fatalf("Found errors calling chassis.PowerKW %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestHpChassisTempC(t *testing.T) {
	expectedAnswer := 17

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.TempC()
	if err != nil {
		t.Fatalf("Found errors calling chassis.TempC %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestHpChassisNics(t *testing.T) {
	expectedAnswer := []*devices.Nic{
		{
			MacAddress: "1c:98:ec:1f:82:73",
			Name:       "OA-1C98EC1F8273",
		},
		{
			MacAddress: "94:18:82:72:e9:f5",
			Name:       "OA-94188272E9F5",
		},
	}

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	nics, err := chassis.Nics()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Nics %v", err)
	}

	if len(nics) != len(expectedAnswer) {
		t.Fatalf("Expected %v nics: found %v nics", len(expectedAnswer), len(nics))
	}

	for pos, nic := range nics {
		if nic.MacAddress != expectedAnswer[pos].MacAddress || nic.Name != expectedAnswer[pos].Name {
			t.Errorf("Expected answer %v: found %v", expectedAnswer[pos], nic)
		}
	}

	tearDown()
}

func TestHpChassisFans(t *testing.T) {
	expectedAnswer := []*devices.Fan{
		{
			Status:     "OK",
			Position:   1,
			Model:      "Active Cool 200 Fan",
			CurrentRPM: 5502,
			PowerKw:    0.007,
		},
		{
			Status:     "OK",
			Position:   2,
			Model:      "Active Cool 200 Fan",
			CurrentRPM: 5500,
			PowerKw:    0.009,
		},
		{
			Status:     "OK",
			Position:   3,
			Model:      "Active Cool 200 Fan",
			CurrentRPM: 5500,
			PowerKw:    0.009,
		},
		{
			Status:     "OK",
			Position:   4,
			Model:      "Active Cool 200 Fan",
			CurrentRPM: 5499,
			PowerKw:    0.007,
		},
		{
			Status:     "OK",
			Position:   5,
			Model:      "Active Cool 200 Fan",
			CurrentRPM: 5499,
			PowerKw:    0.009,
		},
		{
			Status:     "OK",
			Position:   6,
			Model:      "Active Cool 200 Fan",
			CurrentRPM: 5499,
			PowerKw:    0.009,
		},
		{
			Status:     "OK",
			Position:   7,
			Model:      "Active Cool 200 Fan",
			CurrentRPM: 5500,
			PowerKw:    0.009,
		},
		{
			Status:     "OK",
			Position:   8,
			Model:      "Active Cool 200 Fan",
			CurrentRPM: 5500,
			PowerKw:    0.007,
		},
		{
			Status:     "OK",
			Position:   9,
			Model:      "Active Cool 200 Fan",
			CurrentRPM: 5498,
			PowerKw:    0.007,
		},
		{
			Status:     "OK",
			Position:   10,
			Model:      "Active Cool 200 Fan",
			CurrentRPM: 5500,
			PowerKw:    0.007,
		},
	}

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	fans, err := chassis.Fans()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Fans %v", err)
	}

	if len(fans) != len(expectedAnswer) {
		t.Fatalf("Expected %v fans: found %v fans", len(expectedAnswer), len(fans))
	}

	for pos, fan := range fans {
		if fan.Status != expectedAnswer[pos].Status ||
			fan.Position != expectedAnswer[pos].Position ||
			fan.Model != expectedAnswer[pos].Model ||
			fan.CurrentRPM != expectedAnswer[pos].CurrentRPM ||
			fan.PowerKw != expectedAnswer[pos].PowerKw {
			t.Errorf("Expected answer %v: found %v", expectedAnswer[pos], fan)
		}
	}

	tearDown()
}

func TestHpChassisPsu(t *testing.T) {
	expectedAnswer := []*devices.Psu{
		{
			Serial:     "5drca0ahl610qj",
			CapacityKw: 2.65,
			Status:     "OK",
			PowerKw:    0.263,
			PartNumber: "733459-B21",
		},
		{
			Serial:     "5drca0ahl610qe",
			CapacityKw: 2.65,
			Status:     "OK",
			PowerKw:    0.263,
			PartNumber: "733459-B21",
		},
		{
			Serial:     "5drca0ahl610q0",
			CapacityKw: 2.65,
			Status:     "OK",
			PowerKw:    0.263,
			PartNumber: "733459-B21",
		},
		{
			Serial:     "5drca0ahl610pw",
			CapacityKw: 2.65,
			Status:     "OK",
			PowerKw:    0.263,
			PartNumber: "733459-B21",
		},
	}

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	psus, err := chassis.Psus()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Psus %v", err)
	}

	if len(psus) != len(expectedAnswer) {
		t.Fatalf("Expected %v psus: found %v psus", len(expectedAnswer), len(psus))
	}

	for pos, psu := range psus {
		if psu.Serial != expectedAnswer[pos].Serial ||
			psu.CapacityKw != expectedAnswer[pos].CapacityKw ||
			psu.PowerKw != expectedAnswer[pos].PowerKw ||
			psu.Status != expectedAnswer[pos].Status ||
			psu.PartNumber != expectedAnswer[pos].PartNumber {
			t.Errorf("Expected answer %v: found %v", expectedAnswer[pos], psu)
		}
	}

	tearDown()
}

func TestHpChassisRole(t *testing.T) {
	expectedAnswer := true

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer := chassis.IsActive()

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestHpChassisInterface(t *testing.T) {
	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	_ = devices.Cmc(chassis)
	_ = devices.Configure(chassis)
	_ = devices.CmcSetup(chassis)

	tearDown()
}

func TestHpChassisBmcType(t *testing.T) {
	expectedAnswer := "c7000"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer := bmc.BmcType()
	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestUpdateCredentials(t *testing.T) {
	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	bmc.UpdateCredentials("newUsername", "newPassword")

	if bmc.username != "newUsername" {
		t.Fatalf("Expected username to be updated to 'newUsername' but is: %s", bmc.username)
	}

	if bmc.password != "newPassword" {
		t.Fatalf("Expected password to be updated to 'newPassword' but is: %s", bmc.password)
	}

	tearDown()
}

func TestChassisIsPsuRedundant(t *testing.T) {
	expectedAnswer := true

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.IsPsuRedundant()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Name %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestChassisRedundancyMode(t *testing.T) {
	expectedAnswer := "Grid"

	chassis, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := chassis.PsuRedundancyMode()
	if err != nil {
		t.Fatalf("Found errors calling chassis.Name %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}
