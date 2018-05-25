package supermicrox10

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"github.com/ncode/bmclib/devices"
)

var (
	mux     *http.ServeMux
	server  *httptest.Server
	Answers = map[string][]byte{
		"FRU_INFO.XML=(0,0)": []byte(`<?xml version="1.0"?>
			<IPMI>
			  <FRU_INFO RES="1">
				<DEVICE ID="0"/>
				<CHASSIS TYPE="1" PART_NUM="CSE-F414IS2-R2K04BP" SERIAL_NUM="CF414AF38N50003"/>
				<BOARD LAN="0" MFG_DATE="1996/01/01 00:00:00" PROD_NAME="X10DRFF-CTG" MFC_NAME="Supermicro" SERIAL_NUM="VM158S009467" PART_NUM="X10DRFF-CTG"/>
				<PRODUCT LAN="0" MFC_NAME="Supermicro" PROD_NAME="NONE" PART_NUM="SYS-F618H6-FTPTL+" VERSION="NONE" SERIAL_NUM="A19627226A05569" ASSET_TAG="NONE"/>
			  </FRU_INFO>
			</IPMI>`),
		"Get_PlatformCap.XML=(0,0)": []byte(`<?xml version="1.0"?>
			<IPMI>
			  <Platform Cap="8004c039" FanModeSupport="1b" LanModeSupport="7" EnPowerSupplyPage="81" EnStorage="0" EnECExpand="0" EnMultiNode="1" EnX10TwinProMCUUpdate="1" EnPCIeSSD="0" EnAtomHDD="0" EnLANByPassMode="0" EnDP="0" EnSMBIOS="1" SmartCoolCap="0" SmartCooling="0" EnHDDPwrCtrl="0" TwinType="a5" TwinNodeNumber="00" EnBigTwinLCMCCPLDUpdate="0" EnSmartPower="0"/>
			</IPMI>`),
		"GENERIC_INFO.XML=(0,0)": []byte(`<?xml version="1.0"?>  <IPMI>  <GENERIC_INFO>  <GENERIC BMC_IP="010.193.171.016" BMC_MAC="0c:c4:7a:b8:22:64" WEB_VERSION="1.1" IPMIFW_TAG="BL_SUPERMICRO_X7SB3_2017-05-23_B" IPMIFW_VERSION="0325" IPMIFW_BLDTIME="05/23/2017" SESSION_TIMEOUT="00" SDR_VERSION="0000" FRU_VERSION="0000" BIOS_VERSION="        " />  <KERNAL VERSION="2.6.28.9 "/>  </GENERIC_INFO>  </IPMI>`),
		"Get_PlatformInfo.XML=(0,0)": []byte(`<?xml version="1.0"?>
			<IPMI>
			  <PLATFORM_INFO MB_MAC_NUM="2" MB_MAC_ADDR1="0c:c4:7a:bc:dc:1a" MB_MAC_ADDR2="0c:c4:7a:bc:dc:1b" BIOS_VERSION="2.0" BIOS_VERSION_EXIST="1" BIOS_BUILD_DATE="12/17/2015" BIOS_BUILD_DATE_EXIST="1" CPLD_VERSION_EXIST="1" CPLD_VERSION="01.a1.02" REDFISH_REV="1.0.1">
				<HOST_AND_USER HOSTNAME="" BMC_IP="010.193.171.016" SESS_USER_NAME="Administrator" USER_ACCESS="04" DHCP6C_DUID="0E 00 00 01 00 01 20 FA 0E 90 0C C4 7A B8 22 64 "/>
			  </PLATFORM_INFO>
			</IPMI>`),
		"CONFIG_INFO.XML=(0,0)": []byte(`<?xml version="1.0"?>
			<IPMI>
			  <CONFIG_INFO>
				<TOTAL_NUMBER LAN="1" USER="a"/>
				<LAN BMC_IP="010.193.171.016" BMC_MAC="0c-c4-7a-b8-22-64" BMC_NETMASK="255.255.255.000" GATEWAY_IP="010.193.171.254" GATEWAY_MAC="0c-c4-7a-b8-22-64" VLAN_ID="0000" DHCP_TOUT="0" DHCP_EN="1" RMCP_PORT="026f"/>
				<USER NAME="                " USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<USER NAME="Administrator" USER_ACCESS="04" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<USER NAME="" USER_ACCESS="00" IKVM_VIDEO_EN="1" IKVM_KM_EN="1" IKVM_KICK_EN="1" VUSB_EN="1"/>
				<SERVICE DNS_ADDR="000.000.000.000" ALERT_EN="0" SMTP_SERVER=" " SMTP_PORT="587" MAIL_ADDR="0;0;0;0;0;0;0;0;0;0;0;0;0;0;0;0;" MAIL_USR=" " MAIL_PWD=" " SMTP_SSL="0"/>
				<LDAP LDAP_SSL="0" LDAP_IP="000.000.000.000" LDAP_EN="0" Encryption_EN="1" TIMEOUT="00" LDAP_PORT="00000" BASE_DN=" " BINDDN=" "/>
				<DNS DNS_SERVER="10.252.13.2"/>
				<LAN_IF INTERFACE="2"/>
				<HOSTNAME NAME="testserver"/>
				<DHCP6C DUID="0E 00 00 01 00 01 20 FA 0E 90 0C C4 7A B8 22 64 "/>
				<LINK_INFO MII_LINK_CONF="0" MII_AUTO_NEGOTIATION="0" MII_DUPLEX="1" MII_SPEED="2" MII_OPERSTATE="1" NCSI_AUTO_NEGOTIATION="0" NCSI_SPEED_AND_DUPLEX="0" NCSI_OPERSTATE="0" DEV_IF_MODE="2" BOND0_PORT="0"/>
			  </CONFIG_INFO>
			</IPMI>`),
		"SMBIOS_INFO.XML=(0,0)": []byte(`<?xml version="1.0"?>
			<IPMI>
			  <BIOS VENDOR="American Megatrends Inc." VER="2.0" REL_DATE="12/17/2015"/>
			  <SYSTEM MANUFACTURER="Supermicro" PN="SYS-F618H6-FTPTL+" SN="A19627226A05569" SKUN="Default string"/>
			  <CPU TYPE="03h" SPEED="2200 MHz" PROC_UPGRADE="2bh" CORE="10" CORE_ENABLED="10" SOCKET="CPU2" MANUFACTURER="Intel" VER="Intel(R) Xeon(R) CPU E5-2630 0 @ 2.20GHz"/>
			  <CPU TYPE="03h" SPEED="2200 MHz" PROC_UPGRADE="2bh" CORE="10" CORE_ENABLED="10" SOCKET="CPU1" MANUFACTURER="Intel" VER="Intel(R) Xeon(R) CPU E5-2630 0 @ 2.20GHz"/>
			  <DIMM TYPE="1ah" SPEED="2133 MHz" CFG_SPEED="2133 MHz" SIZE="16384 MB" LOCATION="P2-DIMMH1" SN="10D12481" PN="HMA42GR7MFR4N-TF   " BANK_LOCATION="P1_Node1_Channel3_Dimm0" ASSET="P2-DIMMH1_AssetTag (date:15/40)" MANUFACTURER="Hynix Semiconductor"/>
			  <DIMM TYPE="1ah" SPEED="2133 MHz" CFG_SPEED="2133 MHz" SIZE="16384 MB" LOCATION="P2-DIMMG1" SN="10D12494" PN="HMA42GR7MFR4N-TF   " BANK_LOCATION="P1_Node1_Channel2_Dimm0" ASSET="P2-DIMMG1_AssetTag (date:15/40)" MANUFACTURER="Hynix Semiconductor"/>
			  <DIMM TYPE="1ah" SPEED="2133 MHz" CFG_SPEED="2133 MHz" SIZE="16384 MB" LOCATION="P2-DIMMF1" SN="10D1247D" PN="HMA42GR7MFR4N-TF   " BANK_LOCATION="P1_Node1_Channel1_Dimm0" ASSET="P2-DIMMF1_AssetTag (date:15/40)" MANUFACTURER="Hynix Semiconductor"/>
			  <DIMM TYPE="1ah" SPEED="2133 MHz" CFG_SPEED="2133 MHz" SIZE="16384 MB" LOCATION="P2-DIMME1" SN="10D12480" PN="HMA42GR7MFR4N-TF   " BANK_LOCATION="P1_Node1_Channel0_Dimm0" ASSET="P2-DIMME1_AssetTag (date:15/40)" MANUFACTURER="Hynix Semiconductor"/>
			  <DIMM TYPE="1ah" SPEED="2133 MHz" CFG_SPEED="2133 MHz" SIZE="16384 MB" LOCATION="P1-DIMMD1" SN="10D12482" PN="HMA42GR7MFR4N-TF   " BANK_LOCATION="P0_Node0_Channel3_Dimm0" ASSET="P1-DIMMD1_AssetTag (date:15/40)" MANUFACTURER="Hynix Semiconductor"/>
			  <DIMM TYPE="1ah" SPEED="2133 MHz" CFG_SPEED="2133 MHz" SIZE="16384 MB" LOCATION="P1-DIMMC1" SN="10D12520" PN="HMA42GR7MFR4N-TF   " BANK_LOCATION="P0_Node0_Channel2_Dimm0" ASSET="P1-DIMMC1_AssetTag (date:15/40)" MANUFACTURER="Hynix Semiconductor"/>
			  <DIMM TYPE="1ah" SPEED="2133 MHz" CFG_SPEED="2133 MHz" SIZE="16384 MB" LOCATION="P1-DIMMB1" SN="10D1247E" PN="HMA42GR7MFR4N-TF   " BANK_LOCATION="P0_Node0_Channel1_Dimm0" ASSET="P1-DIMMB1_AssetTag (date:15/40)" MANUFACTURER="Hynix Semiconductor"/>
			  <DIMM TYPE="1ah" SPEED="2133 MHz" CFG_SPEED="2133 MHz" SIZE="16384 MB" LOCATION="P1-DIMMA1" SN="10D12479" PN="HMA42GR7MFR4N-TF   " BANK_LOCATION="P0_Node0_Channel0_Dimm0" ASSET="P1-DIMMA1_AssetTag (date:15/40)" MANUFACTURER="Hynix Semiconductor"/>
			  <PowerSupply TYPE="Switching" STATUS="OK" IVRS="Auto-switch" UNPLUGGED="NO" PRESENT="YES" HOTREP="YES" MAXPOWER="2000 Watts" GROUP="2" LOCATION="SLOT 2" SN="P2K4ACG22QT0165" PN="PWS-2K04A-1R" ASSET="N/A" MANUFACTURER="SUPERMICRO" NAME="PWS-2K04A-1R" REV="1.1"/>
			  <PowerSupply TYPE="Switching" STATUS="OK" IVRS="Auto-switch" UNPLUGGED="NO" PRESENT="YES" HOTREP="YES" MAXPOWER="2000 Watts" GROUP="1" LOCATION="SLOT 1" SN="P2K4ACG22QT0168" PN="PWS-2K04A-1R" ASSET="N/A" MANUFACTURER="SUPERMICRO" NAME="PWS-2K04A-1R" REV="1.1"/>
			</IPMI>`),
		"Get_PSInfoReadings.XML=(0,0)": []byte(`<?xml version="1.0"?>
			<IPMI>
			  <PSInfo at_w_PSTimeoutValue="0" at_b_PSTimeoutEnable="0" BBP_TIMEOUT_VALUE="0">
				<PSItem a_b_PS_Status_I2C="1" psType="1" acInVoltage="e4" acInCurrent="6eb" dc12OutVoltage="7a" dc12OutCurrent="7762" temp1="21" temp2="28" fan1="1da0" fan2="247f" dcOutPower="177" acInPower="193" name=""/>
				<PSItem a_b_PS_Status_I2C="1" psType="1" acInVoltage="e4" acInCurrent="66c" dc12OutVoltage="7a" dc12OutCurrent="6f34" temp1="20" temp2="28" fan1="1955" fan2="1f58" dcOutPower="15f" acInPower="175" name=""/>
				<PSItem a_b_PS_Status_I2C="ff" psType="0" acInVoltage="0" acInCurrent="0" dc12OutVoltage="0" dc12OutCurrent="0" temp1="0" temp2="0" fan1="0" fan2="0" dcOutPower="0" acInPower="0" name=""/>
				<PSItem a_b_PS_Status_I2C="ff" psType="0" acInVoltage="0" acInCurrent="0" dc12OutVoltage="0" dc12OutCurrent="0" temp1="0" temp2="0" fan1="0" fan2="0" dcOutPower="0" acInPower="0" name=""/>
			  </PSInfo>
			</IPMI>`),
		"Get_NodeInfoReadings.XML=(0,0)": []byte(`<?xml version="1.0"?>
			<IPMI>
			  <NodeModule psPower="1229" psCurrent="5390" nNODE_Status="1" nNNODE="4" nMYID="3" nMCUFWVer="272" nFatTwin_bp_location="ff" nUsrDefSysName="" nSysName="SYS-F618H6-FTPTL+" nSysSerialNo="A19627226A05562" nBPID="255" nBPRevision="512" nChaName="CSE-F414IS2-R2K04BP" nChaSerialNo="CF414AF38N50022" nBPModelName="BPN-PDB-F418" nBPModelSerialNo="EB164S011414"/>
			  <NodeInfo>
				<Node ID="0" Present="1" PowerStatus="1" Power="270" Current="230" IP="10.193.171.12" NodePartNo="X10DRFF-CTG" NodeSerialNo="VM158S008970" CPU1Temp="57" CPU2Temp="64" SystemTemp="24"/>
				<Node ID="1" Present="1" PowerStatus="1" Power="284" Current="231" IP="10.193.171.15" NodePartNo="X10DRFF-CTG" NodeSerialNo="VM157S014256" CPU1Temp="59" CPU2Temp="63" SystemTemp="24"/>
				<Node ID="2" Present="1" PowerStatus="1" Power="270" Current="221" IP="10.193.171.13" NodePartNo="X10DRFF-CTG" NodeSerialNo="VM156S002490" CPU1Temp="56" CPU2Temp="60" SystemTemp="24"/>
				<Node ID="3" Present="1" PowerStatus="1" Power="252" Current="214" IP="127.0.0.1" NodePartNo="X10DRFF-CTG" NodeSerialNo="VM158S008739" CPU1Temp="55" CPU2Temp="57" SystemTemp="24"/>
			  </NodeInfo>
			</IPMI>`),
		"BIOS_LINCENSE_ACTIVATE.XML=(0,0)": []byte(`<?xml version="1.0"?>
			<IPMI>
			  <BIOS_LINCESNE CHECK="0"/>
			</IPMI>`),
		"POWER_INFO.XML=(0,0)": []byte(`<?xml version="1.0"?>  <IPMI>  <POWER_INFO>  <POWER STATUS="ON"/>  </POWER_INFO>  </IPMI>`),
	}
)

func setup() (r *SupermicroX10, err error) {
	viper.SetDefault("debug", false)
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
		w.Write(Answers[string(query)])
	})

	mux.HandleFunc("/cgi/login.cgi", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("../cgi/url_redirect.cgi?url_name=mainmenu"))
	})

	r, err = New(ip, username, password)
	if err != nil {
		return r, err
	}

	return r, err
}

func tearDown() {
	server.Close()
}

func TestLogin(t *testing.T) {
	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	err = bmc.Login()
	if err != nil {
		t.Errorf("Unable to login: %v", err)
	}

	tearDown()
}

func TestSerial(t *testing.T) {
	expectedAnswer := "a19627226a05569_vm158s009467"

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

func TestModel(t *testing.T) {
	expectedAnswer := "X10DRFF-CTG"

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
	expectedAnswer := "Supermicro"

	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	answer, err := bmc.BmcType()
	if err != nil {
		t.Fatalf("Found errors calling bmc.BmcType %v", err)
	}

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

	answer, err := bmc.BmcVersion()
	if err != nil {
		t.Fatalf("Found errors calling bmc.BmcVersion %v", err)
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
	expectedAnswer := "NotSupported"

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
	expectedAnswer := 128

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
	expectedAnswerCPUType := "intel(r) xeon(r) cpu e5-2630"
	expectedAnswerCPUCount := 2
	expectedAnswerCore := 10
	expectedAnswerHyperthread := 10

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
	expectedAnswer := "2.0"

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
	expectedAnswer := 0.252

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
	expectedAnswer := 24

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
		&devices.Nic{
			MacAddress: "0c:c4:7a:b8:22:64",
			Name:       "bmc",
		},
		&devices.Nic{
			MacAddress: "0c:c4:7a:bc:dc:1a",
			Name:       "eth0",
		},
		&devices.Nic{
			MacAddress: "0c:c4:7a:bc:dc:1b",
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

func TestPoweState(t *testing.T) {
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

func TestIDracInterface(t *testing.T) {
	bmc, err := setup()
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}
	_ = devices.Bmc(bmc)
	tearDown()
}
