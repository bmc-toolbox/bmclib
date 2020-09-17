package discover

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/bmc-toolbox/bmclib/providers/dell/idrac8"
	"github.com/bmc-toolbox/bmclib/providers/dell/idrac9"
	"github.com/bmc-toolbox/bmclib/providers/hp/c7000"
	"github.com/bmc-toolbox/bmclib/providers/hp/ilo"
	"github.com/bmc-toolbox/bmclib/providers/supermicro/supermicrox"
	"github.com/bmc-toolbox/bmclib/providers/supermicro/supermicrox11"
	"github.com/bombsimon/logrusr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func init() {
	if viper.GetBool("debug") != true {
		viper.SetDefault("debug", true)
	}
}

var Answers = map[string][]byte{
	"op=FRU_INFO.XML&r=(0,0)": []byte(`<?xml version="1.0"?>
		<IPMI>
			<FRU_INFO RES="1">
				<DEVICE ID="0"/>
				<CHASSIS TYPE="1" PART_NUM="CSE-813MFTS-R407CBP" SERIAL_NUM="C813MLI52NF0380"/>
				<BOARD LAN="0" MFG_DATE="2020/05/05 03:51:00" PROD_NAME="X11SCM-F" MFC_NAME="Supermicro" SERIAL_NUM="WM205S000401" PART_NUM="X11SCM-F"/>
				<PRODUCT LAN="0" MFC_NAME="Supermicro" PROD_NAME="" PART_NUM="SYS-5019C-MR-PH004" VERSION="NONE" SERIAL_NUM="S402854X0700021" ASSET_TAG=""/>
			</FRU_INFO>
		</IPMI>`),
	"FRU_INFO.XML=(0,0)": []byte(`<?xml version="1.0"?>
		<IPMI>
		  <FRU_INFO RES="1">
			<DEVICE ID="0"/>
			<CHASSIS TYPE="1" PART_NUM="CSE-F414IS2-R2K04BP" SERIAL_NUM="CF414AF38N50003"/>
			<BOARD LAN="0" MFG_DATE="1996/01/01 00:00:00" PROD_NAME="X10DRFF-CTG" MFC_NAME="Supermicro" SERIAL_NUM="VM158S009467" PART_NUM="X10DRFF-CTG"/>
			<PRODUCT LAN="0" MFC_NAME="Supermicro" PROD_NAME="NONE" PART_NUM="SYS-F618H6-FTPTL+" VERSION="NONE" SERIAL_NUM="A19627226A05569" ASSET_TAG="NONE"/>
		  </FRU_INFO>
		</IPMI>`),
}

// setup creates a test server and returns a curried ScanAndConnect() function and a teardown func.
func setup(vendor string, answers map[string][]byte) (scanAndConnectCurry func(opts ...Option) (bmc interface{}, err error), cancel func()) {
	mux := http.NewServeMux()
	server := httptest.NewTLSServer(mux)
	ip := strings.TrimPrefix(server.URL, "https://")
	username := "super"
	password := "test"

	for url := range answers {
		url := url

		mux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
			_ = r.ParseForm()
			if url == "/cgi/login.cgi" && r.Method == http.MethodPost && r.Form.Get("name") != "" {
				_, _ = w.Write([]byte("../cgi/url_redirect.cgi?url_name=mainmenu"))
			} else {
				_, _ = w.Write(answers[url])
			}
		})
	}

	return func(opts ...Option) (bmc interface{}, err error) {
			l := logrus.New()
			//l.Level = logrus.TraceLevel
			opts = append(opts, WithLogger(logrusr.NewLogger(l)))
			return ScanAndConnect(ip, username, password, opts...)
		},
		server.Close
}

func TestProbes(t *testing.T) {
	testt := []struct {
		name     string
		wantHint string
		wantType interface{}
	}{
		{
			name:     "SupermicroX",
			wantHint: ProbeSupermicrox,
			wantType: (*supermicrox.SupermicroX)(nil),
		},
		{
			name:     "SupermicroX11",
			wantHint: ProbeSupermicrox11,
			wantType: (*supermicrox11.SupermicroX)(nil),
		},
		{
			name:     "IDrac9",
			wantHint: ProbeIdrac9,
			wantType: (*idrac9.IDrac9)(nil),
		},
		{
			name:     "IDrac8",
			wantHint: ProbeIdrac8,
			wantType: (*idrac8.IDrac8)(nil),
		},
		{
			name:     "C7000",
			wantHint: ProbeHpC7000,
			wantType: (*c7000.C7000)(nil),
		},
		{
			name:     "Ilo",
			wantHint: ProbeHpIlo,
			wantType: (*ilo.Ilo)(nil),
		},
	}

	for _, tt := range testt {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			scanAndConnect, cancel := setup(tt.name, _answers[tt.name])
			defer cancel()

			hintCallBack := checkHint(t, tt.wantHint)

			for _, hint := range _hints {
				bmc, err := scanAndConnect(WithProbeHint(hint), WithHintCallBack(hintCallBack))
				if err != nil {
					t.Fatalf("error calling ScanAndConnect(): %v", err)
				}

				if reflect.TypeOf(tt.wantType) != reflect.TypeOf(bmc) {
					t.Errorf("Want %T, got %T", tt.wantType, bmc)

				}
			}

			/*
				// just for completeness and backward compatibility, do it again w/o optional params
				bmc, err := scanAndConnect()
				if err != nil {
					t.Fatalf("error calling ScanAndConnect(): %v", err)
				}

				if reflect.TypeOf(tt.wantType) != reflect.TypeOf(bmc) {
					t.Errorf("Want %T, got %T", tt.wantType, bmc)

				}
			*/
		})
	}
}

func checkHint(t *testing.T, want string) func(string) error {
	return func(got string) error {
		t.Helper()

		if got != want {
			t.Errorf("hint call back returned wrong hint: want: %q, got: %q", want, got)
		}

		return nil
	}
}

var (
	_hints = []string{
		"",
		"garbage",
		ProbeHpIlo,
		ProbeIdrac8,
		ProbeIdrac9,
		ProbeSupermicrox,
		ProbeSupermicrox11,
		ProbeHpC7000,
		ProbeM1000e,
		ProbeQuanta,
		ProbeHpCl100,
	}

	_answers = map[string]map[string][]byte{
		"IDrac8": {"/session": []byte(`{"aimGetProp" : {"hostname" :"machine","gui_str_title_bar" :"","OEMHostName" :"machine.example.com","fwVersion" :"2.50.33","sysDesc" :"PowerEdge M630","status" : "OK"}}`)},
		"IDrac9": {
			"/sysmgmt/2015/bmc/info":    []byte(`{"Attributes":{"ADEnabled":"Disabled","BuildVersion":"37","FwVer":"3.15.15.15","GUITitleBar":"spare-H16Z4M2","IsOEMBranded":"0","License":"Enterprise","SSOEnabled":"Disabled","SecurityPolicyMessage":"By accessing this computer, you confirm that such access complies with your organization's security policy.","ServerGen":"14G","SrvPrcName":"NULL","SystemLockdown":"Disabled","SystemModelName":"PowerEdge M640","TFAEnabled":"Disabled","iDRACName":"spare-H16Z4M2"}}`),
			"/sysmgmt/2015/bmc/session": []byte(`{"status": "good", "authResult": 7, "forwardUrl": "something", "errorMsg": "none"}`),
		},
		"SupermicroX": {"/cgi/login.cgi": []byte("ATEN International"),
			"/cgi/ipmi.cgi": []byte(`<?xml version="1.0"?>
		<IPMI>
		  <FRU_INFO RES="1">
			<DEVICE ID="0"/>
			<CHASSIS TYPE="1" PART_NUM="CSE-F414IS2-R2K04BP" SERIAL_NUM="CF414AF38N50003"/>
			<BOARD LAN="0" MFG_DATE="1996/01/01 00:00:00" PROD_NAME="X10DRFF-CTG" MFC_NAME="Supermicro" SERIAL_NUM="VM158S009467" PART_NUM="X10DRFF-CTG"/>
			<PRODUCT LAN="0" MFC_NAME="Supermicro" PROD_NAME="NONE" PART_NUM="SYS-F618H6-FTPTL+" VERSION="NONE" SERIAL_NUM="A19627226A05569" ASSET_TAG="NONE"/>
		  </FRU_INFO>
		</IPMI>`)},
		"SupermicroX11": {"/cgi/login.cgi": []byte("ATEN International"), "/cgi/ipmi.cgi": []byte(`<?xml version="1.0"?>
		<IPMI>
			<FRU_INFO RES="1">
				<DEVICE ID="0"/>
				<CHASSIS TYPE="1" PART_NUM="CSE-813MFTS-R407CBP" SERIAL_NUM="C813MLI52NF0380"/>
				<BOARD LAN="0" MFG_DATE="2020/05/05 03:51:00" PROD_NAME="X11SCM-F" MFC_NAME="Supermicro" SERIAL_NUM="WM205S000401" PART_NUM="X11SCM-F"/>
				<PRODUCT LAN="0" MFC_NAME="Supermicro" PROD_NAME="" PART_NUM="SYS-5019C-MR-PH004" VERSION="NONE" SERIAL_NUM="S402854X0700021" ASSET_TAG=""/>
			</FRU_INFO>
		</IPMI>`)},

		"Quanta": {"/page/login.html": []byte("Quanta")},
		"C7000": {
			"/xmldata": []byte(`<RIMP>
			<MP>
					<ST>1</ST>
					<PRIM>true</PRIM>
					<PN>BladeSystem c7000 DDR2 Onboard Administrator with KVM</PN>
					<FWRI>4.90</FWRI>
					<HWRI>65.49</HWRI>
					<SN>OB51CP6651    </SN>
					<UUID>09OB51CP6651    </UUID>
					<STE>false</STE>
					<USESTE>false</USESTE>
					<SSO>false</SSO>
					<CIMOM>false</CIMOM>
					<ERS>0</ERS>
			</MP>
			<INFRA2>
					<RACK>UnnamedRack</RACK>
					<ENCL>prodch-R01B13B</ENCL>
					<DATETIME>2020-02-11T14:31:25+01:00</DATETIME>
					<TIMEZONE>CET</TIMEZONE>
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
					<ENCL_SN>CZ35230K30</ENCL_SN>
					<PART>681844-B21</PART>
					<UUID>09CZ35230K30</UUID>
					<UIDSTATUS>OFF</UIDSTATUS>
					<ADDR>A9FE019C</ADDR>
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
							</BAYS>
							<BLADE>
									<BAY>
											<CONNECTION>1</CONNECTION>
									</BAY>
									<MGMTIPADDR>10.213.34.213</MGMTIPADDR>
									<MGMTIPV6ADDR_LL>fe80::7210:6fff:feb0:ec02/64</MGMTIPV6ADDR_LL>
									<MGMTIPV6ADDR_SLAAC>2a01:5041:4000:18:7210:6fff:feb0:ec02/64</MGMTIPV6ADDR_SLAAC>
									<MGMTDNSNAME>example.lom.com</MGMTDNSNAME>
									<MGMTPN>iLO4</MGMTPN>
									<MGMTFWVERSION>2.70 May 07 2019</MGMTFWVERSION>
									<PN>813198-B21</PN>
									<BLADEROMVER>I36 09/12/2016</BLADEROMVER>
									<NAME>example.lom.com</NAME>
									<PWRM>1.0.9</PWRM>
									<VLAN>1</VLAN>
									<SPN>ProLiant BL460c Gen9</SPN>
									<BSN>CZ3632K2SR</BSN>
									<UUID>813198CZ3632K2SR</UUID>
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
																	<WWPN>5C:B9:01:C9:DE:20</WWPN>
																	<TYPE>INTERCONNECT_TYPE_ETH</TYPE>
																	<STATUS>OK</STATUS>
																	<GUIDS>
																			<GUID>
																					<TYPE>C</TYPE>
																					<FUNCTION>a</FUNCTION>
																					<GUID_STRING>5C:B9:01:C9:DE:20</GUID_STRING>
																					</GUID>
																			<GUID>
																					<TYPE>H</TYPE>
																					<FUNCTION>b</FUNCTION>
																					<GUID_STRING>5C:B9:01:C9:DE:21</GUID_STRING>
																					</GUID>
																			<GUID>
																					<TYPE>G</TYPE>
																					<FUNCTION>b</FUNCTION>
																					<GUID_STRING>20:00:5C:B9:01:C9:DE:21</GUID_STRING>
																					</GUID>
																			</GUIDS>
															</PORT>
															<PORT>
																	<NUMBER>2</NUMBER>
																	<WWPN>5C:B9:01:C9:DE:28</WWPN>
																	<TYPE>INTERCONNECT_TYPE_ETH</TYPE>
																	<STATUS>OK</STATUS>
																	<GUIDS>
																			<GUID>
																					<TYPE>C</TYPE>
																					<FUNCTION>a</FUNCTION>
																					<GUID_STRING>5C:B9:01:C9:DE:28</GUID_STRING>
																					</GUID>
																			<GUID>
																					<TYPE>H</TYPE>
																					<FUNCTION>b</FUNCTION>
																					<GUID_STRING>5C:B9:01:C9:DE:29</GUID_STRING>
																					</GUID>
																			<GUID>
																					<TYPE>G</TYPE>
																					<FUNCTION>b</FUNCTION>
																					<GUID_STRING>20:00:5C:B9:01:C9:DE:29</GUID_STRING>
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
													<C>25</C>
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
											<POWER_CONSUMED>168</POWER_CONSUMED>
									</POWER>
									<VMSTAT>
											<SUPPORT>VM_SUPPORTED</SUPPORT>
											<CDROMSTAT>VM_DEV_STATUS_DISCONNECTED</CDROMSTAT>
											<CDROMURL></CDROMURL>
											<FLOPPYSTAT>VM_DEV_STATUS_DISCONNECTED</FLOPPYSTAT>
											<FLOPPYURL></FLOPPYURL>
									</VMSTAT>
									<cUUID>31333138-3839-5A43-3336-33324B325352</cUUID>
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
							</BAYS>
							<SWITCH>
									<BAY>
											<CONNECTION>1</CONNECTION>
									</BAY>
									<MGMTIPADDR>0.0.0.0</MGMTIPADDR>
									<MGMTURL></MGMTURL>
									<BSN>7C992900L3</BSN>
									<PN>854194-B21</PN>
									<FWRI>1.10</FWRI>
									<FABRICTYPE>INTERCONNECT_TYPE_ETH</FABRICTYPE>
									<SPN>HPE 10GbE Pass-Thru Module II</SPN>
									<MANUFACTURER>HPE</MANUFACTURER>
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
											</SLOT>
									</PORTMAP>
									<TEMPS>
											<TEMP>
													<LOCATION>13</LOCATION>
													<DESC>AMBIENT</DESC>
													<C>37</C>
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
											<POWER_ON_WATTAGE>57</POWER_ON_WATTAGE>
											<POWER_OFF_WATTAGE>3</POWER_OFF_WATTAGE>
									</POWER>
							</SWITCH>
							<SWITCH>
									<BAY>
											<CONNECTION>2</CONNECTION>
									</BAY>
									<MGMTIPADDR>0.0.0.0</MGMTIPADDR>
									<MGMTURL></MGMTURL>
									<BSN>TWT505V06B</BSN>
									<PN>406740-B21</PN>
									<FWRI>[Unknown]</FWRI>
									<FABRICTYPE>INTERCONNECT_TYPE_ETH</FABRICTYPE>
									<SPN>HP 1Gb Ethernet Pass-Thru Module for c-Class BladeSystem</SPN>
									<MANUFACTURER>HP</MANUFACTURER>
									<STATUS>OK</STATUS>
									<DIAG>
											<FRU>NO_ERROR</FRU>
											<MgmtProc>NO_ERROR</MgmtProc>
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
															<BLADEMEZZPORTNUMBER>2</BLADEMEZZPORTNUMBER>
															<STATUS>OK</STATUS>
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
													<C>30</C>
													<THRESHOLD>
															<DESC>CAUTION</DESC>
															<C>72</C>
															<STATUS>Degraded</STATUS>
													</THRESHOLD>
													<THRESHOLD>
															<DESC>CRITICAL</DESC>
															<C>80</C>
															<STATUS>Non-Recoverable Error</STATUS>
													</THRESHOLD>
											</TEMP>
									</TEMPS>
									<THERMAL>OK</THERMAL>
									<POWER>
											<POWERSTATE>ON</POWERSTATE>
											<POWER_ON_WATTAGE>32</POWER_ON_WATTAGE>
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
							<MGMTIPADDR>10.213.34.4</MGMTIPADDR>
							<NAME>OA-FC15B41BD3B1</NAME>
							<ROLE>ACTIVE</ROLE>
							<STATUS>OK</STATUS>
							<FWRI>4.90</FWRI>
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
							<WIZARDSTATUS>LCD_WIZARD_COMPLETE</WIZARDSTATUS>
							<YOUAREHERE>true</YOUAREHERE>
							<BSN>OB51CP6651    </BSN>
							<UUID>09OB51CP6651    </UUID>
							<SPN>BladeSystem c7000 DDR2 Onboard Administrator with KVM</SPN>
							<MANUFACTURER>HP</MANUFACTURER>
							<TEMPS>
									<TEMP>
											<LOCATION>17</LOCATION>
											<DESC>AMBIENT</DESC>
											<C>47</C>
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
							<MACADDR>FC:15:B4:1B:D3:B1</MACADDR>
							<IPV6STATUS>ENABLED</IPV6STATUS>
							<MGMTIPv6ADDR1>2a01:5041:4000:18:fe15:b4ff:fe1b:d3b1/64</MGMTIPv6ADDR1>
							<MGMTIPv6ADDR2>fe80::fe15:b4ff:fe1b:d3b1/64</MGMTIPv6ADDR2>
					</MANAGER>
					<MANAGER>
							<BAY>
									<CONNECTION>2</CONNECTION>
							</BAY>
							<MGMTIPADDR>10.213.34.2</MGMTIPADDR>
							<NAME>OA-3863BB307D1F</NAME>
							<ROLE>STANDBY</ROLE>
							<STATUS>OK</STATUS>
							<FWRI>4.90</FWRI>
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
							<WIZARDSTATUS>LCD_WIZARD_COMPLETE</WIZARDSTATUS>
							<YOUAREHERE>false</YOUAREHERE>
							<BSN>OB54CP5578    </BSN>
							<UUID>09OB54CP5578    </UUID>
							<SPN>BladeSystem c7000 DDR2 Onboard Administrator with KVM</SPN>
							<MANUFACTURER>HP</MANUFACTURER>
							<TEMPS>
									<TEMP>
											<LOCATION>17</LOCATION>
											<DESC>AMBIENT</DESC>
											<C>47</C>
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
							<MACADDR>38:63:BB:30:7D:1F</MACADDR>
							<IPV6STATUS>ENABLED</IPV6STATUS>
							<MGMTIPv6ADDR3>2a01:5041:4000:18:3a63:bbff:fe30:7d1f/64</MGMTIPv6ADDR3>
							<MGMTIPv6ADDR4>fe80::3a63:bbff:fe30:7d1f/64</MGMTIPv6ADDR4>
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
									<FWRI>2.8.3</FWRI>
									<IMAGE_URL>/cgi-bin/getLCDImage?oaSessionKey=</IMAGE_URL>
									<PIN_ENABLED>false</PIN_ENABLED>
									<BUTTON_LOCK_ENABLED>false</BUTTON_LOCK_ENABLED>
									<USERNOTES>Upload up to^six lines of^text information and your^320x240 bitmap using the^Onboard Administrator^web user interface</USERNOTES>
									<PN>441203-001</PN>
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
													</BAYS>
							<FAN>
									<BAY>
											<CONNECTION>1</CONNECTION>
									</BAY>
									<STATUS>OK</STATUS>
									<PN>412140-B21</PN>
									<PRODUCTNAME>Active Cool 200 Fan</PRODUCTNAME>
									<PWR_USED>7</PWR_USED>
									<RPM_CUR>6135</RPM_CUR>
									<RPM_MAX>18000</RPM_MAX>
									<RPM_MIN>600</RPM_MIN>
							</FAN>
					</FANS>
					<POWER>
							<TYPE>INTERNAL_DC</TYPE>
							<STATUS>OK</STATUS>
							<CAPACITY>4900</CAPACITY>
							<OUTPUT_POWER>8184</OUTPUT_POWER>
							<POWER_CONSUMED>3556</POWER_CONSUMED>
							<REDUNDANT_CAPACITY>1344</REDUNDANT_CAPACITY>
							<REDUNDANCY>REDUNDANT</REDUNDANCY>
							<REDUNDANCYMODE>AC_REDUNDANT</REDUNDANCYMODE>
							<WANTED_PS>4</WANTED_PS>
							<NEEDED_PS>2</NEEDED_PS>
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
									<ACTUALOUTPUT>404</ACTUALOUTPUT>
									<CAPACITY>2450</CAPACITY>
									<SN>5BGXF0AHL8B0TJ</SN>
									<FWRI>0.00</FWRI>
									<PN>588603-B21</PN>
							</POWERSUPPLY>
								<PDU>413374-B21</PDU>
					</POWER>
					<TEMPS>
							<TEMP>
									<LOCATION>9</LOCATION>
									<DESC>AMBIENT</DESC>
									<C>26</C>
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
					<RUID>09CZ35230K30</RUID>
					<ICMB ADDR="A9FE019C" MFG="232" PROD_ID="0x0009" SER="CZ35230K30" UUID="09CZ35230K30">
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
					<cUUID>5A433930-3533-3332-304B-333020202020</cUUID>
					<UHEIGHT>1000</UHEIGHT>
					<UOFFSET>2</UOFFSET>
					<DEVICE_UPOSITION></DEVICE_UPOSITION>
			</SPATIAL>
	</RIMP>`),
		},
		"Ilo": {
			"/xmldata": []byte(`<RIMP>
			<HSI>
			<SBSN>CZ3629FY8B</SBSN>
			<SPN>ProLiant BL460c Gen9</SPN>
			<UUID>813198CZ3629FY8B</UUID>
			<SP>1</SP>
			<cUUID>31333138-3839-5A43-3336-323946593842</cUUID>
			<VIRTUAL>
			<STATE>Inactive</STATE>
			<VID>
			<BSN></BSN>
			<cUUID></cUUID>
			</VID>
			</VIRTUAL>
			<PRODUCTID> 813198-B21</PRODUCTID>
			<NICS>
			<NIC>
			<PORT>1</PORT>
			<DESCRIPTION>iLO 4</DESCRIPTION>
			<LOCATION>Embedded</LOCATION>
			<MACADDR>70:10:6f:af:80:0a</MACADDR>
			<IPADDR>10.183.202.144</IPADDR>
			<STATUS>OK</STATUS>
			</NIC>
			<NIC>
			<PORT>1</PORT>
			<DESCRIPTION>HP FlexFabric 10Gb 2-port 536FLB Adapter</DESCRIPTION>
			<LOCATION>Embedded</LOCATION>
			<MACADDR>ec:b1:d7:b8:ac:c0</MACADDR>
			<IPADDR></IPADDR>
			<STATUS>Unknown</STATUS>
			</NIC>
			<NIC>
			<PORT>2</PORT>
			<DESCRIPTION>HP FlexFabric 10Gb 2-port 536FLB Adapter</DESCRIPTION>
			<LOCATION>Embedded</LOCATION>
			<MACADDR>ec:b1:d7:b8:ac:c8</MACADDR>
			<IPADDR></IPADDR>
			<STATUS>Unknown</STATUS>
			</NIC>
			</NICS>
			</HSI>
			<MP>
			<ST>1</ST>
			<PN>Integrated Lights-Out 4 (iLO 4)</PN>
			<FWRI>2.54</FWRI>
			<BBLK></BBLK>
			<HWRI>ASIC: 17</HWRI>
			<SN>ILOCZ3629FY8B</SN>
			<UUID>ILO813198CZ3629FY8B</UUID>
			<IPM>1</IPM>
			<SSO>0</SSO>
			<PWRM>1.0.9</PWRM>
			<ERS>0</ERS>
			<EALERT>1</EALERT>
			</MP>
			<BLADESYSTEM>
			<BAY>9</BAY>
			<MANAGER>
			<TYPE>Onboard Administrator</TYPE>
			<MGMTIPADDR>10.183.202.135</MGMTIPADDR>
			<MGMTIPv6ADDR1>FE80::9657:A5FF:FE5F:AC1</MGMTIPv6ADDR1>
			<MGMTIPv6ADDR2>2A01:5041:0:16:9657:A5FF:FE5F:AC1</MGMTIPv6ADDR2>
			<RACK>UnnamedRack</RACK>
			<ENCL>prodch-r17</ENCL>
			<ST>2</ST></MANAGER>
			</BLADESYSTEM>
			<SPATIAL>
			<DISCOVERY_RACK>Not Supported</DISCOVERY_RACK>
			<DISCOVERY_DATA>Server does not detect Location Discovery Services</DISCOVERY_DATA>
			<TAG_VERSION>0</TAG_VERSION>
			<RACK_ID>0</RACK_ID>
			<RACK_ID_PN>0</RACK_ID_PN>
			<RACK_DESCRIPTION>0</RACK_DESCRIPTION>
			<RACK_UHEIGHT>0</RACK_UHEIGHT>
			<UPOSITION>0</UPOSITION>
			<ULOCATION>0</ULOCATION>
			<cUUID>31333138-3839-5A43-3336-323946593842</cUUID>
			<UHEIGHT>10.00</UHEIGHT>
			<UOFFSET>2</UOFFSET>
			<BAY>9</BAY>
			<ENCLOSURE_cUUID>5A433930-3633-3932-4659-3139 0 0 0 0 1 0 0 0</ENCLOSURE_cUUID>
			</SPATIAL>
			<HEALTH>
			<STATUS>2</STATUS>
			</HEALTH>
			</RIMP>`),
		},
	}
)
