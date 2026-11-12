package redfishwrapper

import (
	"context"
	"io"
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

// networkProtocolPatch is a PATCH request received by the NetworkProtocol handler.
type networkProtocolPatch struct {
	body    string
	ifMatch string
}

// dellNetworkProtocolHandler serves the Dell NetworkProtocol fixture and records PATCH requests.
func dellNetworkProtocolHandler(t *testing.T, patches *[]networkProtocolPatch) http.HandlerFunc {
	t.Helper()

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write(mustReadFile(t, "dell/networkprotocol.json"))
		case http.MethodPatch:
			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)
			*patches = append(*patches, networkProtocolPatch{body: string(body), ifMatch: r.Header.Get("If-Match")})
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}
}

func TestSetManagerNetworkService(t *testing.T) {
	port := 2222
	testCases := []struct {
		name    string
		service string
		enabled bool
		port    *int
		want    string
	}{
		{name: "disable", service: "IPMI", want: `{"IPMI":{"ProtocolEnabled":false}}`},
		{name: "enable a disabled service", service: "NTP", enabled: true, want: `{"NTP":{"ProtocolEnabled":true}}`},
		{name: "change the port", service: "SSH", enabled: true, port: &port, want: `{"SSH":{"Port":2222,"ProtocolEnabled":true}}`},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			var patches []networkProtocolPatch

			mux := http.NewServeMux()
			mux.HandleFunc(dellNetworkProtocolPath, dellNetworkProtocolHandler(t, &patches))
			client := newDellManagerNetworkClient(t, mux, endpointFunc(t, "dell/manager.idrac.embedded.1.json"))

			require.NoError(t, client.SetManagerNetworkService(context.Background(), tt.service, tt.enabled, tt.port))

			require.Len(t, patches, 1, "expected one PATCH of the NetworkProtocol resource")
			assert.JSONEq(t, tt.want, patches[0].body, "only the named service must be sent")
			assert.Empty(t, patches[0].ifMatch, "the fixture carries no ETag")
		})
	}
}

func TestSetManagerNetworkServiceRefused(t *testing.T) {
	zero, tooHigh := 0, 65536
	testCases := []struct {
		name    string
		service string
		port    *int
		errIs   error
		errMsg  string
	}{
		{name: "service absent from the resource", service: "SNMP", errIs: bmclibErrs.ErrManagerNetworkServiceNotFound},
		{name: "unknown service", service: "Status", errIs: bmclibErrs.ErrManagerNetworkServiceNotFound},
		{name: "service name is case sensitive", service: "ipmi", errIs: bmclibErrs.ErrManagerNetworkServiceNotFound},
		{name: "port zero", service: "SSH", port: &zero, errMsg: "port 0 is out of range"},
		{name: "port above 65535", service: "SSH", port: &tooHigh, errMsg: "port 65536 is out of range"},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			var patches []networkProtocolPatch

			mux := http.NewServeMux()
			mux.HandleFunc(dellNetworkProtocolPath, dellNetworkProtocolHandler(t, &patches))
			client := newDellManagerNetworkClient(t, mux, endpointFunc(t, "dell/manager.idrac.embedded.1.json"))

			err := client.SetManagerNetworkService(context.Background(), tt.service, true, tt.port)
			require.Error(t, err)
			if tt.errIs != nil {
				assert.ErrorIs(t, err, tt.errIs)
			}
			if tt.errMsg != "" {
				assert.ErrorContains(t, err, tt.errMsg)
			}
			assert.Empty(t, patches, "a refused change must not reach the BMC")
		})
	}
}

func TestSetManagerNetworkServiceIfMatch(t *testing.T) {
	newMux := func(patches *[]networkProtocolPatch) *http.ServeMux {
		handler := dellNetworkProtocolHandler(t, patches)
		mux := http.NewServeMux()
		mux.HandleFunc(dellNetworkProtocolPath, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("ETag", `W/"7d1a"`)
			handler(w, r)
		})

		return mux
	}

	t.Run("resource ETag is sent", func(t *testing.T) {
		var patches []networkProtocolPatch
		client := newDellManagerNetworkClient(t, newMux(&patches), endpointFunc(t, "dell/manager.idrac.embedded.1.json"))

		require.NoError(t, client.SetManagerNetworkService(context.Background(), "IPMI", false, nil))
		require.Len(t, patches, 1)
		assert.Equal(t, `W/"7d1a"`, patches[0].ifMatch)
	})

	t.Run("ETag matching disabled", func(t *testing.T) {
		var patches []networkProtocolPatch
		client := newDellManagerNetworkClient(t, newMux(&patches), endpointFunc(t, "dell/manager.idrac.embedded.1.json"))
		client.disableEtagMatch = true

		require.NoError(t, client.SetManagerNetworkService(context.Background(), "IPMI", false, nil))
		require.Len(t, patches, 1)
		assert.Empty(t, patches[0].ifMatch)
	})
}

func TestSetManagerNetworkServiceRejected(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(dellNetworkProtocolPath, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write(mustReadFile(t, "dell/networkprotocol.json"))
		default:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"Port is in use"}}`))
		}
	})
	client := newDellManagerNetworkClient(t, mux, endpointFunc(t, "dell/manager.idrac.embedded.1.json"))

	port := 443
	err := client.SetManagerNetworkService(context.Background(), "SSH", true, &port)
	assert.ErrorIs(t, err, bmclibErrs.ErrManagerNetworkProtocolSet)
	assert.ErrorContains(t, err, "Port is in use")
}

func TestSetManagerNetworkServiceNoResource(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(dellNetworkProtocolPath, notFound)
	client := newDellManagerNetworkClient(t, mux, endpointFunc(t, "dell/manager.idrac.embedded.1.json"))

	err := client.SetManagerNetworkService(context.Background(), "IPMI", false, nil)
	assert.ErrorIs(t, err, bmclibErrs.ErrManagerNetworkProtocolNotFound)
}

func TestSetManagerNTPServers(t *testing.T) {
	testCases := []struct {
		name    string
		servers []string
		want    string
	}{
		{name: "set", servers: []string{"10.0.0.1", "ntp.example.org"}, want: `{"NTP":{"NTPServers":["10.0.0.1","ntp.example.org"]}}`},
		{name: "clear with nil", want: `{"NTP":{"NTPServers":[]}}`},
		{name: "clear with an empty list", servers: []string{}, want: `{"NTP":{"NTPServers":[]}}`},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			var patches []networkProtocolPatch

			mux := http.NewServeMux()
			mux.HandleFunc(dellNetworkProtocolPath, dellNetworkProtocolHandler(t, &patches))
			client := newDellManagerNetworkClient(t, mux, endpointFunc(t, "dell/manager.idrac.embedded.1.json"))

			require.NoError(t, client.SetManagerNTPServers(context.Background(), tt.servers))

			require.Len(t, patches, 1, "expected one PATCH of the NetworkProtocol resource")
			assert.JSONEq(t, tt.want, patches[0].body, "only NTP.NTPServers must be sent")
		})
	}
}

func TestSetManagerNTPServersRejected(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(dellNetworkProtocolPath, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write(mustReadFile(t, "dell/networkprotocol.json"))
		default:
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"error":{"message":"too many NTP servers"}}`))
		}
	})
	client := newDellManagerNetworkClient(t, mux, endpointFunc(t, "dell/manager.idrac.embedded.1.json"))

	err := client.SetManagerNTPServers(context.Background(), []string{"a", "b", "c", "d"})
	assert.ErrorIs(t, err, bmclibErrs.ErrManagerNetworkProtocolSet)
	assert.ErrorContains(t, err, "too many NTP servers")
}

func TestSetManagerNTPServersNoResource(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc(dellNetworkProtocolPath, notFound)
	client := newDellManagerNetworkClient(t, mux, endpointFunc(t, "dell/manager.idrac.embedded.1.json"))

	err := client.SetManagerNTPServers(context.Background(), []string{"10.0.0.1"})
	assert.ErrorIs(t, err, bmclibErrs.ErrManagerNetworkProtocolNotFound)
}
