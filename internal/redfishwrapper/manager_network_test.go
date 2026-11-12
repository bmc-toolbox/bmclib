package redfishwrapper

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

const (
	dellManagerPath            = "/redfish/v1/Managers/iDRAC.Embedded.1"
	dellEthernetInterfacesPath = dellManagerPath + "/EthernetInterfaces"
	dellEthernetInterfacePath  = dellEthernetInterfacesPath + "/NIC.1"
)

// newDellManagerNetworkClient serves the Dell service root and manager collection with the given
// manager handler; the EthernetInterfaces and NetworkProtocol handlers are left to the test.
func newDellManagerNetworkClient(t *testing.T, mux *http.ServeMux, manager http.HandlerFunc) *Client {
	t.Helper()

	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "dell/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Managers", endpointFunc(t, "dell/managers.json"))
	mux.HandleFunc(dellManagerPath, manager)

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))

	ctx := context.Background()
	require.NoError(t, client.Open(ctx))
	t.Cleanup(func() { _ = client.Close(ctx) })

	return client
}

func TestGetManagerEthernetInterfaces(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(dellEthernetInterfacesPath, endpointFunc(t, "dell/ethernetinterfaces.json"))
	mux.HandleFunc(dellEthernetInterfacePath, endpointFunc(t, "dell/ethernetinterface_nic_1.json"))
	client := newDellManagerNetworkClient(t, mux, endpointFunc(t, "dell/manager.idrac.embedded.1.json"))

	interfaces, err := client.GetManagerEthernetInterfaces(context.Background())
	require.NoError(t, err)
	require.Len(t, interfaces, 1)

	speed, mtu := 1000, 1500
	assert.Equal(t, bmc.ManagerEthernetInterface{
		ID:                  "NIC.1",
		Name:                "Manager Ethernet Interface",
		MACAddress:          "D0:8E:79:B3:70:1A",
		PermanentMACAddress: "D0:8E:79:B3:70:1A",
		HostName:            "idrac-7XK9Q33",
		FQDN:                "idrac-7XK9Q33.example.com",
		Enabled:             true,
		LinkStatus:          "LinkUp",
		SpeedMbps:           &speed,
		MTUSize:             &mtu,
		IPv4Addresses: []bmc.ManagerIPv4Address{
			{Address: "10.0.0.5", SubnetMask: "255.255.255.0", Gateway: "10.0.0.1", Origin: "Static"},
		},
		IPv6Addresses: []bmc.ManagerIPv6Address{
			{Address: "fe80::d28e:79ff:feb3:701a", PrefixLength: 64, Origin: "LinkLocal"},
		},
		DHCPv4Enabled: false,
		VLANEnabled:   false,
		VLANID:        1,
		NameServers:   []string{"10.0.0.2", "10.0.0.3"},
		Health:        "OK",
		State:         "Enabled",
	}, interfaces[0])
}

func TestGetManagerEthernetInterfacesMinimal(t *testing.T) {
	// An interface reporting only the mandatory properties maps to zero values and nil pointers.
	mux := http.NewServeMux()
	mux.HandleFunc(dellEthernetInterfacesPath, endpointFunc(t, "dell/ethernetinterfaces.json"))
	mux.HandleFunc(dellEthernetInterfacePath, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"@odata.id": "` + dellEthernetInterfacePath + `", "Id": "NIC.1", "MACAddress": "D0:8E:79:B3:70:1A"}`))
	})
	client := newDellManagerNetworkClient(t, mux, endpointFunc(t, "dell/manager.idrac.embedded.1.json"))

	interfaces, err := client.GetManagerEthernetInterfaces(context.Background())
	require.NoError(t, err)
	require.Len(t, interfaces, 1)
	assert.Equal(t, bmc.ManagerEthernetInterface{ID: "NIC.1", MACAddress: "D0:8E:79:B3:70:1A"}, interfaces[0])
}

func TestGetManagerEthernetInterfacesNotFound(t *testing.T) {
	// A manager without an EthernetInterfaces link.
	client := newDellManagerNetworkClient(t, http.NewServeMux(), func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"@odata.id": "` + dellManagerPath + `", "Id": "iDRAC.Embedded.1", "Name": "Manager"}`))
	})

	_, err := client.GetManagerEthernetInterfaces(context.Background())
	assert.ErrorIs(t, err, bmclibErrs.ErrManagerEthernetInterfacesNotFound)

	_, err = client.GetManagerNetworkProtocol(context.Background())
	assert.ErrorIs(t, err, bmclibErrs.ErrManagerNetworkProtocolNotFound)
}

func TestGetManagerNetworkLinkNotFound(t *testing.T) {
	// A manager that advertises NetworkProtocol but answers 404 is skipped; a 404 on the
	// EthernetInterfaces collection is a collection failure and is returned.
	mux := http.NewServeMux()
	mux.HandleFunc(dellEthernetInterfacesPath, notFound)
	mux.HandleFunc(dellNetworkProtocolPath, notFound)
	client := newDellManagerNetworkClient(t, mux, endpointFunc(t, "dell/manager.idrac.embedded.1.json"))

	_, err := client.GetManagerNetworkProtocol(context.Background())
	assert.ErrorIs(t, err, bmclibErrs.ErrManagerNetworkProtocolNotFound)

	_, err = client.GetManagerEthernetInterfaces(context.Background())
	require.Error(t, err)
	assert.ErrorContains(t, err, "manager iDRAC.Embedded.1 EthernetInterfaces")
}

// A failure other than 404 while collecting is returned, not skipped.
func TestGetManagerNetworkErrorSurfaced(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(dellEthernetInterfacesPath, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc(dellNetworkProtocolPath, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	client := newDellManagerNetworkClient(t, mux, endpointFunc(t, "dell/manager.idrac.embedded.1.json"))

	_, err := client.GetManagerEthernetInterfaces(context.Background())
	require.Error(t, err)
	assert.ErrorContains(t, err, "manager iDRAC.Embedded.1 EthernetInterfaces")

	_, err = client.GetManagerNetworkProtocol(context.Background())
	require.Error(t, err)
	assert.ErrorContains(t, err, "manager iDRAC.Embedded.1 NetworkProtocol")
}

func TestGetManagerNetworkProtocol(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(dellNetworkProtocolPath, endpointFunc(t, "dell/networkprotocol.json"))
	client := newDellManagerNetworkClient(t, mux, endpointFunc(t, "dell/manager.idrac.embedded.1.json"))

	protocol, err := client.GetManagerNetworkProtocol(context.Background())
	require.NoError(t, err)

	https, ssh, ipmi, ntp := 443, 22, 623, 123
	assert.Equal(t, &bmc.ManagerNetworkProtocol{
		HostName: "idrac-7XK9Q33",
		Services: []bmc.ManagerNetworkService{
			{Name: "HTTPS", Enabled: true, Port: &https},
			{Name: "SSH", Enabled: true, Port: &ssh},
			{Name: "IPMI", Enabled: true, Port: &ipmi},
			{Name: "NTP", Enabled: false, Port: &ntp},
		},
	}, protocol, "services absent from the resource (HTTP, SNMP, Telnet, ...) must be omitted")
}

func TestGetManagerNetworkProtocolServiceWithoutPort(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(dellNetworkProtocolPath, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{
			"@odata.id": "` + dellNetworkProtocolPath + `",
			"Id": "NetworkProtocol",
			"HostName": "bmc",
			"FQDN": "bmc.example.com",
			"Telnet": {"ProtocolEnabled": false},
			"NTP": {"ProtocolEnabled": true, "NTPServers": ["10.0.0.9"]},
			"SNMP": null
		}`))
	})
	client := newDellManagerNetworkClient(t, mux, endpointFunc(t, "dell/manager.idrac.embedded.1.json"))

	protocol, err := client.GetManagerNetworkProtocol(context.Background())
	require.NoError(t, err)
	assert.Equal(t, &bmc.ManagerNetworkProtocol{
		HostName:   "bmc",
		FQDN:       "bmc.example.com",
		NTPServers: []string{"10.0.0.9"},
		Services: []bmc.ManagerNetworkService{
			{Name: "Telnet", Enabled: false},
			{Name: "NTP", Enabled: true},
		},
	}, protocol, "a present service without Port has a nil Port; a null service is absent")
}
