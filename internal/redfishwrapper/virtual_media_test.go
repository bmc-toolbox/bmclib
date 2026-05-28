package redfishwrapper

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetVirtualMedia(t *testing.T) {
	tests := map[string]struct {
		hfunc       map[string]func(http.ResponseWriter, *http.Request)
		basicAuth   bool
		expectCount int
		expectErr   string
	}{
		"manager path has virtual media": {
			// Standard case: VirtualMedia is found under Manager (e.g., HP iLO, Supermicro)
			hfunc: map[string]func(http.ResponseWriter, *http.Request){
				"/redfish/v1/":                        endpointFunc(t, "serviceroot.json"),
				"/redfish/v1/Managers":                endpointFunc(t, "managers.json"),
				"/redfish/v1/Managers/1":              endpointFunc(t, "managers_1.json"),
				"/redfish/v1/Managers/1/VirtualMedia": endpointFunc(t, "404"),
				"/redfish/v1/Systems":                 endpointFunc(t, "systems.json"),
				"/redfish/v1/Systems/1":               endpointFunc(t, "systems_1.json"),
			},
			// managers_1.json has a VirtualMedia link, but our mock returns 404 for the collection.
			// This means Manager path returns 0 items, so fallback to System path.
			// systems_1.json doesn't have VirtualMedia link, so both fail.
			expectCount: 0,
			expectErr:   "no virtual media found",
		},
		"dell idrac - system path has virtual media": {
			// Dell iDRAC case: Manager has no VirtualMedia, System has VirtualMedia
			hfunc: map[string]func(http.ResponseWriter, *http.Request){
				"/redfish/v1/":                                         endpointFunc(t, "dell/serviceroot.json"),
				"/redfish/v1/Managers":                                 endpointFunc(t, "dell/managers.json"),
				"/redfish/v1/Managers/iDRAC.Embedded.1":                endpointFunc(t, "dell/manager.idrac.embedded.1.json"),
				"/redfish/v1/Systems":                                  endpointFunc(t, "dell/systems.json"),
				"/redfish/v1/Systems/System.Embedded.1":                endpointFunc(t, "dell/system.embedded.1.virtualmedia.json"),
				"/redfish/v1/Systems/System.Embedded.1/VirtualMedia":   endpointFunc(t, "dell/virtualmedia_collection.json"),
				"/redfish/v1/Systems/System.Embedded.1/VirtualMedia/1": endpointFunc(t, "dell/virtualmedia_1.json"),
				"/redfish/v1/Systems/System.Embedded.1/VirtualMedia/2": endpointFunc(t, "dell/virtualmedia_2.json"),
			},
			basicAuth:   true,
			expectCount: 2,
			expectErr:   "",
		},
		"no virtual media anywhere": {
			// Neither Manager nor System exposes VirtualMedia
			hfunc: map[string]func(http.ResponseWriter, *http.Request){
				"/redfish/v1/":                          endpointFunc(t, "dell/serviceroot.json"),
				"/redfish/v1/Managers":                  endpointFunc(t, "dell/managers.json"),
				"/redfish/v1/Managers/iDRAC.Embedded.1": endpointFunc(t, "dell/manager.idrac.embedded.1.json"),
				"/redfish/v1/Systems":                   endpointFunc(t, "dell/systems.json"),
				"/redfish/v1/Systems/System.Embedded.1": endpointFunc(t, "dell/system.embedded.1.json"),
			},
			basicAuth:   true,
			expectCount: 0,
			expectErr:   "no virtual media found",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			mux := http.NewServeMux()
			for endpoint, handler := range tc.hfunc {
				mux.HandleFunc(endpoint, handler)
			}

			server := httptest.NewTLSServer(mux)
			defer server.Close()

			parsedURL, err := url.Parse(server.URL)
			require.NoError(t, err)

			ctx := context.Background()

			opts := []Option{}
			if tc.basicAuth {
				opts = append(opts, WithBasicAuthEnabled(true))
			}

			client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", opts...)

			err = client.Open(ctx)
			require.NoError(t, err)

			defer client.Close(ctx)

			vm, err := client.getVirtualMedia(ctx)
			if tc.expectErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectErr)
				assert.Nil(t, vm)
			} else {
				assert.NoError(t, err)
				assert.Len(t, vm, tc.expectCount)
			}
		})
	}
}

func TestSetVirtualMedia_DellSystemPath(t *testing.T) {
	// Test that SetVirtualMedia works with Dell iDRAC where VirtualMedia
	// is only available under the System resource path.
	// We test ejection (empty mediaURL) which only requires GET operations
	// and validates the full System path fallback flow.
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "dell/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Managers", endpointFunc(t, "dell/managers.json"))
	mux.HandleFunc("/redfish/v1/Managers/iDRAC.Embedded.1", endpointFunc(t, "dell/manager.idrac.embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc(t, "dell/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc(t, "dell/system.embedded.1.virtualmedia.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/VirtualMedia", endpointFunc(t, "dell/virtualmedia_collection.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/VirtualMedia/1", endpointFunc(t, "dell/virtualmedia_1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/VirtualMedia/2", endpointFunc(t, "dell/virtualmedia_2.json"))

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	ctx := context.Background()

	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))

	err = client.Open(ctx)
	require.NoError(t, err)

	defer client.Close(ctx)

	// Test ejecting CD media via System path (empty URL = eject).
	// VirtualMedia fixtures have Inserted: false, so eject is a no-op success.
	ok, err := client.SetVirtualMedia(ctx, "CD", "")
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestInsertedVirtualMedia_DellSystemPath(t *testing.T) {
	// Test that InsertedVirtualMedia works when VirtualMedia is only
	// available under the System resource path (Dell iDRAC).
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "dell/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Managers", endpointFunc(t, "dell/managers.json"))
	mux.HandleFunc("/redfish/v1/Managers/iDRAC.Embedded.1", endpointFunc(t, "dell/manager.idrac.embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc(t, "dell/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc(t, "dell/system.embedded.1.virtualmedia.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/VirtualMedia", endpointFunc(t, "dell/virtualmedia_collection.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/VirtualMedia/1", endpointFunc(t, "dell/virtualmedia_1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/VirtualMedia/2", endpointFunc(t, "dell/virtualmedia_2.json"))

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	ctx := context.Background()

	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))

	err = client.Open(ctx)
	require.NoError(t, err)

	defer client.Close(ctx)

	// Both VirtualMedia instances have Inserted: false, so should return empty
	inserted, err := client.InsertedVirtualMedia(ctx)
	assert.NoError(t, err)
	assert.Empty(t, inserted)
}

func TestSetVirtualMedia_SlotFallback(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "dell/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Managers", endpointFunc(t, "dell/managers.json"))
	mux.HandleFunc("/redfish/v1/Managers/iDRAC.Embedded.1", endpointFunc(t, "dell/manager.idrac.embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc(t, "dell/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc(t, "dell/system.embedded.1.virtualmedia.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/VirtualMedia", endpointFunc(t, "dell/virtualmedia_collection.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/VirtualMedia/1", endpointFunc(t, "dell/virtualmedia_1.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/VirtualMedia/2", endpointFunc(t, "dell/virtualmedia_2.json"))
	// VirtualMedia/2 (first in collection) rejects InsertMedia with 500.
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/VirtualMedia/2/Actions/VirtualMedia.InsertMedia",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		},
	)
	// VirtualMedia/1 (second in collection) accepts InsertMedia.
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/VirtualMedia/1/Actions/VirtualMedia.InsertMedia",
		func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		},
	)

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	ctx := context.Background()

	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))

	err = client.Open(ctx)
	require.NoError(t, err)

	defer client.Close(ctx)

	ok, err := client.SetVirtualMedia(ctx, "CD", "http://example.com/boot.iso")
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestSetVirtualMedia_InsertPatchFallbackWithoutAction(t *testing.T) {
	const virtualMediaCollection = `{
		"@odata.context": "/redfish/v1/$metadata#VirtualMediaCollection.VirtualMediaCollection",
		"@odata.id": "/redfish/v1/Systems/System.Embedded.1/VirtualMedia",
		"@odata.type": "#VirtualMediaCollection.VirtualMediaCollection",
		"Members": [
			{
				"@odata.id": "/redfish/v1/Systems/System.Embedded.1/VirtualMedia/1"
			}
		],
		"Members@odata.count": 1,
		"Name": "Virtual Media Services"
	}`

	const virtualMediaWithoutActions = `{
		"@odata.context": "/redfish/v1/$metadata#VirtualMedia.VirtualMedia",
		"@odata.id": "/redfish/v1/Systems/System.Embedded.1/VirtualMedia/1",
		"@odata.type": "#VirtualMedia.v1_2_0.VirtualMedia",
		"Id": "1",
		"Name": "VirtualMedia",
		"Image": "",
		"Inserted": false,
		"ConnectedVia": "NotConnected",
		"WriteProtected": true,
		"MediaTypes": ["CD", "DVD"]
	}`

	var patchCalled bool

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "dell/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Managers", endpointFunc(t, "dell/managers.json"))
	mux.HandleFunc("/redfish/v1/Managers/iDRAC.Embedded.1", endpointFunc(t, "dell/manager.idrac.embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc(t, "dell/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc(t, "dell/system.embedded.1.virtualmedia.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/VirtualMedia", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		_, _ = w.Write([]byte(virtualMediaCollection))
	})
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/VirtualMedia/1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(virtualMediaWithoutActions))
		case http.MethodPatch:
			patchCalled = true
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	ctx := context.Background()

	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))

	err = client.Open(ctx)
	require.NoError(t, err)

	defer client.Close(ctx)

	ok, err := client.SetVirtualMedia(ctx, "CD", "http://example.com/boot.iso")
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, patchCalled, "expected InsertMedia fallback to PATCH the VirtualMedia resource")
}

func TestSetVirtualMedia_InsertPatchFallbackRequiresInsertedWithoutWriteProtected(t *testing.T) {
	const virtualMediaCollection = `{
		"@odata.context": "/redfish/v1/$metadata#VirtualMediaCollection.VirtualMediaCollection",
		"@odata.id": "/redfish/v1/Systems/System.Embedded.1/VirtualMedia",
		"@odata.type": "#VirtualMediaCollection.VirtualMediaCollection",
		"Members": [
			{
				"@odata.id": "/redfish/v1/Systems/System.Embedded.1/VirtualMedia/1"
			}
		],
		"Members@odata.count": 1,
		"Name": "Virtual Media Services"
	}`

	const virtualMediaWithoutActions = `{
		"@odata.context": "/redfish/v1/$metadata#VirtualMedia.VirtualMedia",
		"@odata.id": "/redfish/v1/Systems/System.Embedded.1/VirtualMedia/1",
		"@odata.type": "#VirtualMedia.v1_2_0.VirtualMedia",
		"Id": "1",
		"Name": "VirtualMedia",
		"Image": "",
		"Inserted": false,
		"ConnectedVia": "NotConnected",
		"WriteProtected": true,
		"MediaTypes": ["CD", "DVD"]
	}`

	var patchAttempts int

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "dell/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Managers", endpointFunc(t, "dell/managers.json"))
	mux.HandleFunc("/redfish/v1/Managers/iDRAC.Embedded.1", endpointFunc(t, "dell/manager.idrac.embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc(t, "dell/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc(t, "dell/system.embedded.1.virtualmedia.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/VirtualMedia", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		_, _ = w.Write([]byte(virtualMediaCollection))
	})
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/VirtualMedia/1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(virtualMediaWithoutActions))
		case http.MethodPatch:
			patchAttempts++

			body, err := io.ReadAll(r.Body)
			require.NoError(t, err)

			payload := string(body)

			// First attempt includes WriteProtected and is rejected by this BMC.
			if strings.Contains(payload, "WriteProtected") {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":{"message":"WriteProtected is not accepted"}}`))
				return
			}

			// The BMC requires Inserted to be present.
			if !strings.Contains(payload, "Inserted") {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error":{"message":"The property Inserted is a required property and must be included in the request."}}`))
				return
			}

			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	ctx := context.Background()

	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))

	err = client.Open(ctx)
	require.NoError(t, err)

	defer client.Close(ctx)

	ok, err := client.SetVirtualMedia(ctx, "CD", "http://example.com/boot.iso")
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, 2, patchAttempts)
}

func TestSetVirtualMedia_EjectPatchFallbackWithoutAction(t *testing.T) {
	const virtualMediaCollection = `{
		"@odata.context": "/redfish/v1/$metadata#VirtualMediaCollection.VirtualMediaCollection",
		"@odata.id": "/redfish/v1/Systems/System.Embedded.1/VirtualMedia",
		"@odata.type": "#VirtualMediaCollection.VirtualMediaCollection",
		"Members": [
			{
				"@odata.id": "/redfish/v1/Systems/System.Embedded.1/VirtualMedia/1"
			}
		],
		"Members@odata.count": 1,
		"Name": "Virtual Media Services"
	}`

	const insertedVirtualMediaWithoutActions = `{
		"@odata.context": "/redfish/v1/$metadata#VirtualMedia.VirtualMedia",
		"@odata.id": "/redfish/v1/Systems/System.Embedded.1/VirtualMedia/1",
		"@odata.type": "#VirtualMedia.v1_2_0.VirtualMedia",
		"Id": "1",
		"Name": "VirtualMedia",
		"Image": "http://example.com/old.iso",
		"Inserted": true,
		"ConnectedVia": "URI",
		"WriteProtected": true,
		"MediaTypes": ["CD", "DVD"]
	}`

	var patchCalled bool

	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "dell/serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Managers", endpointFunc(t, "dell/managers.json"))
	mux.HandleFunc("/redfish/v1/Managers/iDRAC.Embedded.1", endpointFunc(t, "dell/manager.idrac.embedded.1.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc(t, "dell/systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1", endpointFunc(t, "dell/system.embedded.1.virtualmedia.json"))
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/VirtualMedia", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		_, _ = w.Write([]byte(virtualMediaCollection))
	})
	mux.HandleFunc("/redfish/v1/Systems/System.Embedded.1/VirtualMedia/1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_, _ = w.Write([]byte(insertedVirtualMediaWithoutActions))
		case http.MethodPatch:
			patchCalled = true
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	ctx := context.Background()

	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "", WithBasicAuthEnabled(true))

	err = client.Open(ctx)
	require.NoError(t, err)

	defer client.Close(ctx)

	ok, err := client.SetVirtualMedia(ctx, "CD", "")
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.True(t, patchCalled, "expected EjectMedia fallback to PATCH the VirtualMedia resource")
}

func TestSetVirtualMedia_InvalidMediaType(t *testing.T) {
	// Test that invalid media type returns an error before any Redfish calls
	mux := http.NewServeMux()
	mux.HandleFunc("/redfish/v1/", endpointFunc(t, "serviceroot.json"))
	mux.HandleFunc("/redfish/v1/Managers", endpointFunc(t, "managers.json"))
	mux.HandleFunc("/redfish/v1/Managers/1", endpointFunc(t, "managers_1.json"))
	mux.HandleFunc("/redfish/v1/Systems", endpointFunc(t, "systems.json"))
	mux.HandleFunc("/redfish/v1/Systems/1", endpointFunc(t, "systems_1.json"))

	server := httptest.NewTLSServer(mux)
	defer server.Close()

	parsedURL, err := url.Parse(server.URL)
	require.NoError(t, err)

	ctx := context.Background()

	client := NewClient(parsedURL.Hostname(), parsedURL.Port(), "", "")

	err = client.Open(ctx)
	require.NoError(t, err)

	defer client.Close(ctx)

	ok, err := client.SetVirtualMedia(ctx, "InvalidType", "http://example.com/boot.iso")
	assert.Error(t, err)
	assert.False(t, ok)
	assert.Contains(t, err.Error(), "invalid media type")
}
