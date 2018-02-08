package discover

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ncode/bmc/idrac8"
	"github.com/ncode/bmc/idrac9"
	"github.com/ncode/bmc/ilo"

	"github.com/spf13/viper"
)

var (
	mux     *http.ServeMux
	server  *httptest.Server
	answers = map[string]map[string][]byte{
		"BladeIlo": map[string][]byte{
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
		"BladeIDrac8": map[string][]byte{"/session": []byte(`{"aimGetProp" : {"hostname" :"incubatordb-2011","gui_str_title_bar" :"","OEMHostName" :"machine.example.com","fwVersion" :"2.50.33","sysDesc" :"PowerEdge M630","status" : "OK"}}`)},
		"BladeIDrac9": map[string][]byte{"/sysmgmt/2015/bmc/info": []byte(`{"Attributes":{"ADEnabled":"Disabled","BuildVersion":"37","FwVer":"3.15.15.15","GUITitleBar":"spare-H16Z4M2","IsOEMBranded":"0","License":"Enterprise","SSOEnabled":"Disabled","SecurityPolicyMessage":"By accessing this computer, you confirm that such access complies with your organization's security policy.","ServerGen":"14G","SrvPrcName":"NULL","SystemLockdown":"Disabled","SystemModelName":"PowerEdge M640","TFAEnabled":"Disabled","iDRACName":"spare-H16Z4M2"}}`)},
	}
)

func setup(answers map[string][]byte) (bmc interface{}, err error) {
	viper.SetDefault("debug", true)
	mux = http.NewServeMux()
	server = httptest.NewTLSServer(mux)
	ip := strings.TrimPrefix(server.URL, "https://")
	username := "super"
	password := "test"

	for url := range answers {
		url := url
		mux.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
			w.Write(answers[url])
		})
	}

	bmc, err = ScanAndConnect(ip, username, password)
	if err != nil {
		return bmc, err
	}

	return bmc, err
}

func tearDown() {
	server.Close()
}

func TestFindBladedIlo(t *testing.T) {
	bmc, err := setup(answers["BladeIlo"])
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	if answer, ok := bmc.(*ilo.Ilo); !ok {
		fmt.Println(ok)
		t.Errorf("Expected answer %T: found %T", &ilo.Ilo{}, answer)
	}

	tearDown()
}

func TestFindBladeIDrac8(t *testing.T) {
	bmc, err := setup(answers["BladeIDrac8"])
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	if answer, ok := bmc.(*idrac8.IDrac8); !ok {
		t.Errorf("Expected answer %T: found %T", &idrac8.IDrac8{}, answer)
	}

	tearDown()
}

func TestFindBladeIDrac9(t *testing.T) {
	bmc, err := setup(answers["BladeIDrac9"])
	if err != nil {
		t.Fatalf("Found errors during the test setup %v", err)
	}

	if answer, ok := bmc.(*idrac9.IDrac9); !ok {
		t.Errorf("Expected answer %T: found %T", &idrac9.IDrac9{}, answer)
	}

	tearDown()
}
