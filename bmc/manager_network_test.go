package bmc

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type mockManagerEthernetInterfacesGetter struct {
	interfaces []ManagerEthernetInterface
	err        error
}

func (m *mockManagerEthernetInterfacesGetter) GetManagerEthernetInterfaces(ctx context.Context) ([]ManagerEthernetInterface, error) {
	return m.interfaces, m.err
}

func (m *mockManagerEthernetInterfacesGetter) Name() string {
	return "mock"
}

type mockManagerNetworkProtocolGetter struct {
	protocol *ManagerNetworkProtocol
	err      error
}

func (m *mockManagerNetworkProtocolGetter) GetManagerNetworkProtocol(ctx context.Context) (*ManagerNetworkProtocol, error) {
	return m.protocol, m.err
}

func (m *mockManagerNetworkProtocolGetter) Name() string {
	return "mock"
}

func TestGetManagerEthernetInterfacesFromInterfaces(t *testing.T) {
	interfaces := []ManagerEthernetInterface{{ID: "NIC.1", MACAddress: "d0:8e:79:b3:70:1a", Enabled: true}}

	testCases := []struct {
		name               string
		generic            []interface{}
		errMsg             string
		expectedInterfaces []ManagerEthernetInterface
	}{
		{
			name:               "success",
			generic:            []interface{}{&mockManagerEthernetInterfacesGetter{interfaces: interfaces}},
			expectedInterfaces: interfaces,
		},
		{
			name:    "not an implementation",
			generic: []interface{}{"foo"},
			errMsg:  "no ManagerEthernetInterfacesGetter implementations found",
		},
		{
			name:    "no implementations",
			generic: []interface{}{},
			errMsg:  "no ManagerEthernetInterfacesGetter implementations found",
		},
		{
			name:    "error from getter",
			generic: []interface{}{&mockManagerEthernetInterfacesGetter{err: errors.New("foobar")}},
			errMsg:  "foobar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got, metadata, err := GetManagerEthernetInterfacesFromInterfaces(context.Background(), tt.generic)

			if tt.errMsg == "" {
				assert.NoError(t, err)
				assert.Equal(t, "mock", metadata.SuccessfulProvider)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}

			assert.Equal(t, tt.expectedInterfaces, got)
		})
	}
}

func TestGetManagerNetworkProtocolFromInterfaces(t *testing.T) {
	port := 443
	protocol := &ManagerNetworkProtocol{
		HostName: "bmc",
		Services: []ManagerNetworkService{{Name: "HTTPS", Enabled: true, Port: &port}},
	}

	testCases := []struct {
		name             string
		generic          []interface{}
		errMsg           string
		expectedProtocol *ManagerNetworkProtocol
	}{
		{
			name:             "success",
			generic:          []interface{}{&mockManagerNetworkProtocolGetter{protocol: protocol}},
			expectedProtocol: protocol,
		},
		{
			name:    "not an implementation",
			generic: []interface{}{"foo"},
			errMsg:  "no ManagerNetworkProtocolGetter implementations found",
		},
		{
			name:    "no implementations",
			generic: []interface{}{},
			errMsg:  "no ManagerNetworkProtocolGetter implementations found",
		},
		{
			name:    "error from getter",
			generic: []interface{}{&mockManagerNetworkProtocolGetter{err: errors.New("foobar")}},
			errMsg:  "foobar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			got, metadata, err := GetManagerNetworkProtocolFromInterfaces(context.Background(), tt.generic)

			if tt.errMsg == "" {
				assert.NoError(t, err)
				assert.Equal(t, "mock", metadata.SuccessfulProvider)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}

			assert.Equal(t, tt.expectedProtocol, got)
		})
	}
}

type mockManagerNetworkServiceSetter struct {
	err error
}

func (m *mockManagerNetworkServiceSetter) SetManagerNetworkService(ctx context.Context, _ string, _ bool, _ *int) error {
	return m.err
}

func (m *mockManagerNetworkServiceSetter) Name() string {
	return "mock"
}

func TestSetManagerNetworkServiceFromInterfaces(t *testing.T) {
	port := 2222
	testCases := []struct {
		name    string
		generic []interface{}
		port    *int
		errMsg  string
	}{
		{
			name:    "success",
			generic: []interface{}{&mockManagerNetworkServiceSetter{}},
		},
		{
			name:    "success, with port",
			generic: []interface{}{&mockManagerNetworkServiceSetter{}},
			port:    &port,
		},
		{
			name:    "not an implementation",
			generic: []interface{}{"foo"},
			errMsg:  "no ManagerNetworkServiceSetter implementations found",
		},
		{
			name:    "no implementations",
			generic: []interface{}{},
			errMsg:  "no ManagerNetworkServiceSetter implementations found",
		},
		{
			name:    "error from setter",
			generic: []interface{}{&mockManagerNetworkServiceSetter{err: errors.New("foobar")}},
			errMsg:  "foobar",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			metadata, err := SetManagerNetworkServiceFromInterfaces(context.Background(), tt.generic, "SSH", true, tt.port)

			if tt.errMsg == "" {
				assert.NoError(t, err)
				assert.Equal(t, "mock", metadata.SuccessfulProvider)
			} else {
				assert.ErrorContains(t, err, tt.errMsg)
			}
		})
	}
}
