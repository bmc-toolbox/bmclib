package supermicrox11

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bombsimon/logrusr"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var (
	mux     *http.ServeMux
	server  *httptest.Server
	Answers = map[string][]byte{
		"op=FRU_INFO.XML&r=(0,0)": []byte(`<?xml version="1.0"?>
			<IPMI>
				<FRU_INFO RES="1">
					<DEVICE ID="0"/>
					<CHASSIS TYPE="1" PART_NUM="CSE-813MFTS-R407CBP" SERIAL_NUM="C813MLI52NF0380"/>
					<BOARD LAN="0" MFG_DATE="2020/05/05 03:51:00" PROD_NAME="X11SCM-F" MFC_NAME="Supermicro" SERIAL_NUM="WM205S000401" PART_NUM="X11SCM-F"/>
					<PRODUCT LAN="0" MFC_NAME="Supermicro" PROD_NAME="" PART_NUM="SYS-5019C-MR-PH004" VERSION="NONE" SERIAL_NUM="S402854X0700021" ASSET_TAG=""/>
				</FRU_INFO>
			</IPMI>`),
		"op=Get_PlatformCap.XML&r=(0,0)": []byte(`<?xml version="1.0"?>
			<IPMI>
				<Platform Cap="4d009" FanModeSupport="17" LanModeSupport="7" EnPowerSupplyPage="2" EnStorage="0" EnAOMPwrCtl="0" EnMCUMultiNode="0" EnCPLDNode="0" EnPCIeSSD="0" EnAtomHDD="0" EnLANByPassMode="0" EnTroubleShoot="1" EnDP="0" EnSMBIOS="1" SmartCoolCap="0" SmartCooling="0" TwinProMyID="0" HasSticker="1"/>
			</IPMI>`),
		"op=GENERIC_INFO.XML&r=(0,0)": []byte(`<?xml version="1.0"?>  <IPMI>  <GENERIC_INFO>  <GENERIC BMC_IP="010.193.171.016" BMC_MAC="0c:c4:7a:b8:22:64" WEB_VERSION="1.1" IPMIFW_TAG="BL_SUPERMICRO_X7SB3_2017-05-23_B" IPMIFW_VERSION="0325" IPMIFW_BLDTIME="05/23/2017" SESSION_TIMEOUT="00" SDR_VERSION="0000" FRU_VERSION="0000" BIOS_VERSION="        " />  <KERNAL VERSION="2.6.28.9 "/>  </GENERIC_INFO>  </IPMI>`),
		"op=Get_PlatformInfo.XML&r=(0,0)": []byte(`<?xml version="1.0"?>
		    <IPMI>
		      <PLATFORM_INFO MB_MAC_NUM="2" MB_MAC_ADDR1="3c:ec:ef:6b:0b:bc" MB_MAC_ADDR2="3c:ec:ef:6b:0b:bd" CPLD_REV="03.b3.05" BIOS_VERSION="1.4" BIOS_BUILD_DATE="05/26/2020" REDFISH_REV="1.0.1">
			    <HOST_AND_USER HOSTNAME="" BMC_IP="010.236.131.035" SESS_USER_NAME="test" USER_ACCESS="04" DHCP6C_DUID="0E 00 00 01 00 01 26 B4 AB FA 3C EC EF 6B 0C 20 "/>
		      </PLATFORM_INFO>
		    </IPMI>`),
		"op=CONFIG_INFO.XML&r=(0,0)": []byte(`<?xml version="1.0"?>
			<IPMI>
			<CONFIG_INFO>
				<TOTAL_NUMBER LAN="1" USER="a"/>
				<LAN BMC_IP="010.236.131.035" BMC_MAC="3c:ec:ef:6b:0c:20" BMC_NETMASK="255.255.255.128" GATEWAY_IP="010.236.131.001" GATEWAY_MAC="3c:ec:ef:6b:0c:20" VLAN_ID="0000" DHCP_TOUT="0" DHCP_EN="1" RMCP_PORT="026f"/>
				<USER NAME="                " USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1" U_STATUS="0" U_LOCKED="0"/>
				<USER NAME="ADMIN" USER_ACCESS="04" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1" U_STATUS="1" U_LOCKED="0"/>
				<USER NAME="test" USER_ACCESS="04" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1" U_STATUS="1" U_LOCKED="0"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1" U_STATUS="0" U_LOCKED="0"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1" U_STATUS="0" U_LOCKED="0"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1" U_STATUS="0" U_LOCKED="0"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1" U_STATUS="0" U_LOCKED="0"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1" U_STATUS="0" U_LOCKED="0"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1" U_STATUS="0" U_LOCKED="0"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1" U_STATUS="0" U_LOCKED="0"/>
				<DNS DNS_SERVER="147.75.207.208"/>
				<LAN_IF INTERFACE="2"/>
				<HOSTNAME NAME="testserver"/>
				<DHCP6C DUID="0E 00 00 01 00 01 26 B4 AB FA 3C EC EF 6B 0C 20 "/>
				<LINK_INFO MII_LINK_CONF="0" MII_AUTO_NEGOTIATION="0" MII_DUPLEX="1" MII_SPEED="2" MII_OPERSTATE="1" NCSI_AUTO_NEGOTIATION="0" NCSI_SPEED_AND_DUPLEX="0" NCSI_OPERSTATE="0" DEV_IF_MODE="2" BOND0_PORT="0"/>
				<IP_PROTOCOL_STATUS IP4_STATUS="1" IP6_STATUS="1"/>
			</CONFIG_INFO>
			</IPMI>`),
		"op=SMBIOS_INFO.XML&r=(0,0)": []byte(`<?xml version="1.0"?>
			<IPMI>
			  <BIOS VENDOR="American Megatrends Inc." VER="1.4" REL_DATE="05/26/2020"/>
			  <SYSTEM MANUFACTURER="Supermicro" PN="SYS-5019C-MR-PH004" SN="S402854X0700021" SKUN="To be filled by O.E.M."/>
			  <CPU TYPE="03h" SPEED="3400 MHz" PROC_UPGRADE="32h" CORE="8" CORE_ENABLED="8" SOCKET="CPU" MANUFACTURER="Intel(R) Corporation" VER="Intel(R) Xeon(R) E-2278G CPU @ 3.40GHz" SN="To Be Filled By O.E.M." ASSET="To Be Filled By O.E.M." PN="To Be Filled By O.E.M."/>
			  <DIMM TYPE="1ah" SPEED="2667 MHz" CFG_SPEED="2667 MHz" SIZE="16384 MiB" LOCATION="DIMMB2" SN="F0FECF32" PN="18ADF2G72AZ-2G6E1   " BANK_LOCATION="P0_Node0_Channel1_Dimm1" ASSET="TestAsset0" MANUFACTURER="Micron"/>
			  <DIMM TYPE="1ah" SPEED="2667 MHz" CFG_SPEED="2667 MHz" SIZE="16384 MiB" LOCATION="DIMMA2" SN="F0FECF2E" PN="18ADF2G72AZ-2G6E1   " BANK_LOCATION="P0_Node0_Channel0_Dimm1" ASSET="TestAsset0" MANUFACTURER="Micron"/>
			  <PowerSupply TYPE="Switching" STATUS="OK" IVRS="Auto-switch" UNPLUGGED="NO" PRESENT="YES" HOTREP="YES" MAXPOWER="400 Watts" GROUP="2" LOCATION="PSU2" SN="P407PCJ50WT1746" PN="PWS-407P-1R" ASSET="N/A" MANUFACTURER="SUPERMICRO" NAME="PWS-407P-1R" REV="1.2"/>
			  <PowerSupply TYPE="Switching" STATUS="OK" IVRS="Auto-switch" UNPLUGGED="NO" PRESENT="YES" HOTREP="YES" MAXPOWER="400 Watts" GROUP="1" LOCATION="PSU1" SN="P407PCJ50WT1747" PN="PWS-407P-1R" ASSET="N/A" MANUFACTURER="SUPERMICRO" NAME="PWS-407P-1R" REV="1.2"/>
			  <LAN ETH_INTERFACE="EthernetInterface 1" ID="1" ETH_NAME="OnBoard LAN 1" DESC="" STATE="Disabled" HEALTH="OK" MAC="3c:ec:ef:6b:0b:bc" SPEED="0" FQDN=""/>
			  <LAN ETH_INTERFACE="EthernetInterface 2" ID="2" ETH_NAME="OnBoard LAN 2" DESC="" STATE="Disabled" HEALTH="OK" MAC="3c:ec:ef:6b:0b:bd" SPEED="0" FQDN=""/>
			</IPMI>`),
		"op=Get_PSInfoReadings.XML=(0,0)": []byte(`<?xml version="1.0"?>
			<IPMI>
			  <PSInfo at_w_PSTimeoutValue="0" at_b_PSTimeoutEnable="0" BBP_TIMEOUT_VALUE="0">
			    <PSItem psStatus="fe" psUnitType="Power Supply" psPMBusState="PS OK" psBBPState="Not present" acInVoltage="d1" acInCurrent="94" dc12OutVoltage="7a" dc12OutCurrent="4a3" temp1="29" temp2="30" fan1="1180" fan2="0" dcOutPower="e" acInPower="15" isDCPower="0" PSname="P407PCJ50WT1747"/>
			    <PSItem psStatus="fe" psUnitType="Power Supply" psPMBusState="PS OK" psBBPState="Not present" acInVoltage="d0" acInCurrent="9c" dc12OutVoltage="7a" dc12OutCurrent="4e2" temp1="25" temp2="28" fan1="1460" fan2="0" dcOutPower="f" acInPower="16" isDCPower="0" PSname="P407PCJ50WT1746"/>
			    <PSItem psStatus="0" psUnitType="Power Supply" psPMBusState="Not present" psBBPState="Not present" acInVoltage="0" acInCurrent="0" dc12OutVoltage="0" dc12OutCurrent="0" temp1="0" temp2="0" fan1="0" fan2="0" dcOutPower="0" acInPower="0" isDCPower="0" PSname=""/>
			  <PSItem psStatus="0" psUnitType="Power Supply" psPMBusState="Not present" psBBPState="Not present" acInVoltage="0" acInCurrent="0" dc12OutVoltage="0" dc12OutCurrent="0" temp1="0" temp2="0" fan1="0" fan2="0" dcOutPower="0" acInPower="0" isDCPower="0" PSname=""/>
		  </PSInfo>
			</IPMI>`),
		"op=POWER_CONSUMPTION.XML&r=(0,0)": []byte(`<?xml version="1.0"?>
		<IPMI>
		  <POWER HAVERAGE="43" DAVERAGE="88" WAVERAGE="88" HMINIMUM="23" DMINIMUM="41" WMINIMUM="41" HMINTIME="2020/07/10 09:12:13" DMINTIME="2020/07/10 09:12:13" WMINTIME="2020/07/09 16:44:46" HMAXIMUM="62" DMAXIMUM="144" WMAXIMUM="144" HMAXTIME="2020/07/10 09:13:12" DMAXTIME="2020/07/09 21:56:33" WMAXTIME="2020/07/09 21:56:33"/>
		  <HOUR>
			<FMINS0 MAX="58" AVR="57" MIN="56"/>
			<FMINS1 MAX="58" AVR="56" MIN="55"/>
			<FMINS2 MAX="62" AVR="44" MIN="23"/>
			<FMINS3 MAX="41" AVR="41" MIN="41"/>
			<FMINS4 MAX="44" AVR="42" MIN="41"/>
			<FMINS5 MAX="41" AVR="41" MIN="41"/>
			<FMINS6 MAX="44" AVR="41" MIN="41"/>
			<FMINS7 MAX="41" AVR="41" MIN="41"/>
			<FMINS8 MAX="41" AVR="41" MIN="41"/>
			<FMINS9 MAX="42" AVR="41" MIN="41"/>
			<FMINS10 MAX="41" AVR="41" MIN="41"/>
			<FMINS11 MAX="42" AVR="41" MIN="41"/>
		  </HOUR>
		  <DAY>
			<HOUR0 MAX="62" AVR="43" MIN="23"/>
			<HOUR1 MAX="56" AVR="41" MIN="40"/>
			<HOUR2 MAX="47" AVR="41" MIN="40"/>
			<HOUR3 MAX="44" AVR="41" MIN="40"/>
			<HOUR4 MAX="51" AVR="41" MIN="41"/>
			<HOUR5 MAX="44" AVR="41" MIN="41"/>
			<HOUR6 MAX="63" AVR="41" MIN="41"/>
			<HOUR7 MAX="44" AVR="41" MIN="41"/>
			<HOUR8 MAX="62" AVR="41" MIN="41"/>
			<HOUR9 MAX="143" AVR="78" MIN="41"/>
			<HOUR10 MAX="143" AVR="142" MIN="142"/>
			<HOUR11 MAX="144" AVR="135" MIN="41"/>
			<HOUR12 MAX="136" AVR="123" MIN="42"/>
			<HOUR13 MAX="136" AVR="122" MIN="42"/>
			<HOUR14 MAX="135" AVR="123" MIN="42"/>
			<HOUR15 MAX="136" AVR="110" MIN="42"/>
			<HOUR16 MAX="140" AVR="68" MIN="41"/>
			<HOUR17 MAX="0" AVR="0" MIN="0"/>
			<HOUR18 MAX="0" AVR="0" MIN="0"/>
			<HOUR19 MAX="0" AVR="0" MIN="0"/>
			<HOUR20 MAX="0" AVR="0" MIN="0"/>
			<HOUR21 MAX="0" AVR="0" MIN="0"/>
			<HOUR22 MAX="0" AVR="0" MIN="0"/>
			<HOUR23 MAX="0" AVR="0" MIN="0"/>
		  </DAY>
		  <WEEK>
			<DAY0 MAX="144" AVR="88" MIN="41"/>
			<DAY1 MAX="0" AVR="0" MIN="0"/>
			<DAY2 MAX="0" AVR="0" MIN="0"/>
			<DAY3 MAX="0" AVR="0" MIN="0"/>
			<DAY4 MAX="0" AVR="0" MIN="0"/>
			<DAY5 MAX="0" AVR="0" MIN="0"/>
			<DAY6 MAX="0" AVR="0" MIN="0"/>
			<DAY7 MAX="0" AVR="0" MIN="0"/>
			<DAY8 MAX="0" AVR="0" MIN="0"/>
			<DAY9 MAX="0" AVR="0" MIN="0"/>
			<DAY10 MAX="0" AVR="0" MIN="0"/>
			<DAY11 MAX="0" AVR="0" MIN="0"/>
			<DAY12 MAX="0" AVR="0" MIN="0"/>
			<DAY13 MAX="0" AVR="0" MIN="0"/>
		  </WEEK>
		  <NOW MAX="58" AVR="14" MIN="56"/>
		  <PEAK MAX="144" MIN="23" Current="43" PMAXTIME="2020/07/09 21:56:33" PMINTIME="2020/07/10 09:12:13"/>
		  <BBP TIMEOUT="0" BBPSUPPORT="0"/>
		</IPMI>`),
		"op=POWER_INFO.XML&r=(0,0)":             []byte(`<?xml version="1.0"?>  <IPMI>  <POWER_INFO>  <POWER STATUS="ON"/>  </POWER_INFO>  </IPMI>`),
		"op=SYS_HEALTH.XML&r=(1,ff)":            []byte(`<?xml version="1.0"?>  <IPMI>  <SYS_HEALTH STATUS="3"/> </IPMI>`),
		"op=BIOS_LINCENSE_ACTIVATE.XML&r=(0,0)": []byte(`<?xml version="1.0"?> <IPMI> <BIOS_LINCESNE CHECK="0"/> </IPMI>`),
		"op=SENSOR_INFO.XML&r=(1,ff)": []byte(`<?xml version="1.0"?>
		<IPMI>
		  <SENSOR_INFO>
			<SENSOR ID="001" NUMBER="01" NAME="CPU Temp" READING="1ec000" OPTION="c0" UNR="64" UC="64" UNC="5f" LNC="0a" LC="05" LNR="05" STYPE="01" RTYPE="01" ERTYPE="01" UNIT1="00" UNIT="01" L="00" M="0100" B="0000" RB="00"/>
			<SENSOR ID="002" NUMBER="0a" NAME="PCH Temp" READING="22c000" OPTION="c0" UNR="69" UC="5a" UNC="55" LNC="0a" LC="05" LNR="05" STYPE="01" RTYPE="01" ERTYPE="01" UNIT1="00" UNIT="01" L="00" M="0100" B="0000" RB="00"/>
			<SENSOR ID="003" NUMBER="0b" NAME="System Temp" READING="19c000" OPTION="c0" UNR="5a" UC="55" UNC="50" LNC="0a" LC="05" LNR="05" STYPE="01" RTYPE="01" ERTYPE="01" UNIT1="00" UNIT="01" L="00" M="0100" B="0000" RB="00"/>
			<SENSOR ID="004" NUMBER="0c" NAME="Peripheral Temp" READING="1dc000" OPTION="c0" UNR="5a" UC="55" UNC="50" LNC="0a" LC="05" LNR="05" STYPE="01" RTYPE="01" ERTYPE="01" UNIT1="00" UNIT="01" L="00" M="0100" B="0000" RB="00"/>
			<SENSOR ID="005" NUMBER="10" NAME="VcpuVRM Temp" READING="1fc000" OPTION="c0" UNR="69" UC="64" UNC="5f" LNC="0a" LC="05" LNR="05" STYPE="01" RTYPE="01" ERTYPE="01" UNIT1="00" UNIT="01" L="00" M="0100" B="0000" RB="00"/>
			<SENSOR ID="006" NUMBER="4b" NAME="M2NVMeSSD Temp1" READING="000000" OPTION="00" UNR="4b" UC="46" UNC="41" LNC="0a" LC="05" LNR="05" STYPE="01" RTYPE="01" ERTYPE="01" UNIT1="00" UNIT="01" L="00" M="0100" B="0000" RB="00"/>
			<SENSOR ID="007" NUMBER="4c" NAME="M2NVMeSSD Temp2" READING="000000" OPTION="00" UNR="4b" UC="46" UNC="41" LNC="0a" LC="05" LNR="05" STYPE="01" RTYPE="01" ERTYPE="01" UNIT1="00" UNIT="01" L="00" M="0100" B="0000" RB="00"/>
			<SENSOR ID="008" NUMBER="b0" NAME="DIMMA1 Temp" READING="000000" OPTION="00" UNR="5a" UC="55" UNC="50" LNC="0a" LC="05" LNR="05" STYPE="01" RTYPE="01" ERTYPE="01" UNIT1="00" UNIT="01" L="00" M="0100" B="0000" RB="00"/>
			<SENSOR ID="009" NUMBER="b1" NAME="DIMMA2 Temp" READING="1dc000" OPTION="c0" UNR="5a" UC="55" UNC="50" LNC="0a" LC="05" LNR="05" STYPE="01" RTYPE="01" ERTYPE="01" UNIT1="00" UNIT="01" L="00" M="0100" B="0000" RB="00"/>
			<SENSOR ID="00a" NUMBER="b2" NAME="DIMMB1 Temp" READING="000000" OPTION="00" UNR="5a" UC="55" UNC="50" LNC="0a" LC="05" LNR="05" STYPE="01" RTYPE="01" ERTYPE="01" UNIT1="00" UNIT="01" L="00" M="0100" B="0000" RB="00"/>  
		  </SENSOR_INFO>
		</IPMI>`),
		"op=Get_NodeInfoReadings.XML&r=(0,0)": []byte(`<?xml version="1.0"?>
		<IPMI>
		  <NodeModule nCONFIG_ID="0" nMYID="0" nMCUFWVer="0.00" nFatTwin_bp_location="0" nUsrDefSysName="" nSysName="" nSysSerialNo="" nBPID="0" nBPRevision="0.00" nChaName="" nChaSerialNo="" nBPModelName="" nBPModelSerialNo="" nFanAlert="0" TwinType="0"/>
		  <NodeInfo>
			<Node ID="1" Present="0" PowerStatus="0" Power="0" Current="0" IP="0.0.0.0" NodePartNo="" NodeSerialNo="" CPU1Temp="0" CPU2Temp="0" SystemTemp="0" FanSpeed="0"/>
			<Node ID="2" Present="0" PowerStatus="0" Power="0" Current="0" IP="0.0.0.0" NodePartNo="" NodeSerialNo="" CPU1Temp="0" CPU2Temp="0" SystemTemp="0" FanSpeed="0"/>
			<Node ID="3" Present="0" PowerStatus="0" Power="0" Current="0" IP="0.0.0.0" NodePartNo="" NodeSerialNo="" CPU1Temp="0" CPU2Temp="0" SystemTemp="0" FanSpeed="0"/>
			<Node ID="4" Present="0" PowerStatus="0" Power="0" Current="0" IP="0.0.0.0" NodePartNo="" NodeSerialNo="" CPU1Temp="0" CPU2Temp="0" SystemTemp="0" FanSpeed="0"/>
		  </NodeInfo>
		</IPMI>`),
	}
)

func init() {
	if viper.GetBool("debug") != true {
		viper.SetDefault("debug", true)
	}
}

func setup() (r *SupermicroX, err error) {
	mux = http.NewServeMux()
	server = httptest.NewTLSServer(mux)
	ip := strings.TrimPrefix(server.URL, "https://")
	username := "super"
	password := "test"

	mux.HandleFunc("/cgi/ipmi.cgi", func(w http.ResponseWriter, r *http.Request) {
		query, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(Answers[string(query)])
	})

	mux.HandleFunc("/cgi/login.cgi", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("../cgi/url_redirect.cgi?url_name=mainmenu"))
	})

	testLog := logrus.New()
	r, err = New(context.TODO(), ip, username, password, logrusr.NewLogger(testLog))
	if err != nil {
		return r, err
	}

	return r, err
}

func tearDown() {
	server.Close()
}

func TestSerial(t *testing.T) {
	expectedAnswer := "wm205s000401"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.Serial()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Serial %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestChassisSerial(t *testing.T) {
	expectedAnswer := "c813mli52nf0380"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.ChassisSerial()
	if err != nil {
		t.Fatalf("Found errors calling bmc.ChassisSerial %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestModel(t *testing.T) {
	expectedAnswer := "X11SCM-F"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.Model()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Model %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestBmcType(t *testing.T) {
	expectedAnswer := "x11"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer := bmc.HardwareType()
	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestBmcVersion(t *testing.T) {
	expectedAnswer := "0325"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.Version()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Version %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestName(t *testing.T) {
	expectedAnswer := "testserver"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.Name()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Name %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestStatus(t *testing.T) {
	expectedAnswer := "Unhealthy"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.Status()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Status %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestMemory(t *testing.T) {
	expectedAnswer := 32

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.Memory()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Memory %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestCPU(t *testing.T) {
	expectedAnswerCPUType := "intel(r) xeon(r) e-2278g cpu"
	expectedAnswerCPUCount := 1
	expectedAnswerCore := 8
	expectedAnswerHyperthread := 8

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	cpuType, cpuCount, core, ht, err := bmc.CPU()
	if err != nil {
		t.Fatalf("Found errors calling bmc.CPU %v", err)
	}

	if cpuType != expectedAnswerCPUType {
		t.Errorf("Expected cpuType answer %v: found %v", expectedAnswerCPUType, cpuType)
	}

	if cpuCount != expectedAnswerCPUCount {
		t.Errorf("Expected cpuCount answer %v: found %v", expectedAnswerCPUCount, cpuCount)
	}

	if core != expectedAnswerCore {
		t.Errorf("Expected core answer %v: found %v", expectedAnswerCore, core)
	}

	if ht != expectedAnswerHyperthread {
		t.Errorf("Expected ht answer %v: found %v", expectedAnswerHyperthread, ht)
	}

	tearDown()
}

func TestBiosVersion(t *testing.T) {
	expectedAnswer := "1.4"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.BiosVersion()
	if err != nil {
		t.Fatalf("Found errors calling bmc.BiosVersion %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestPowerKW(t *testing.T) {
	expectedAnswer := 0.043

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.PowerKw()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerKW %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestTempC(t *testing.T) {
	expectedAnswer := 19

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.TempC()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Temp %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestNics(t *testing.T) {
	expectedAnswer := []*devices.Nic{
		{
			MacAddress: "0c:c4:7a:b8:22:64",
			Name:       "bmc",
		},
		{
			MacAddress: "3c:ec:ef:6b:0b:bc",
			Name:       "eth0",
		},
		{
			MacAddress: "3c:ec:ef:6b:0b:bd",
			Name:       "eth1",
		},
	}

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	nics, err := bmc.Nics()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Nics %v", err)
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

func TestLicense(t *testing.T) {
	expectedName := "oob"
	expectedLicType := "Activated"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	name, licType, err := bmc.License()
	if err != nil {
		t.Fatalf("Found errors calling bmc.License %v", err)
	}

	if name != expectedName {
		t.Errorf("Expected name %v: found %v", expectedName, name)
	}

	if licType != expectedLicType {
		t.Errorf("Expected name %v: found %v", expectedLicType, licType)
	}

	tearDown()
}

func TestIsBlade(t *testing.T) {
	expectedAnswer := false

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.IsBlade()
	if err != nil {
		t.Fatalf("Found errors calling bmc.IsBlade %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestSlot(t *testing.T) {
	expectedAnswer := 1

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.Slot()
	if err != nil {
		t.Fatalf("Found errors calling bmc.Position %v", err)
	}

	if answer != expectedAnswer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestPowerState(t *testing.T) {
	expectedAnswer := "on"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.PowerState()
	if err != nil {
		t.Fatalf("Found errors calling bmc.PowerState %v", err)
	}

	if expectedAnswer != answer {
		t.Errorf("Expected answer %v: found %v", expectedAnswer, answer)
	}

	tearDown()
}

func TestIBmcInterface(t *testing.T) {
	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	_ = devices.Bmc(bmc)
	_ = devices.Configure(bmc)
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
