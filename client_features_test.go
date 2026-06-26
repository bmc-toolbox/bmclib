package bmclib

import (
	"context"
	"testing"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	"github.com/jacobweinstock/registrar"
)

// fakeExtendedProvider is a registry driver implementing a couple of the
// extended interfaces, used to verify the Client-level wiring (dispatch +
// metadata storage).
type fakeExtendedProvider struct {
	name string
	snmp bmc.SNMPConfig
}

func (f *fakeExtendedProvider) Name() string { return f.name }

func (f *fakeExtendedProvider) SNMP(_ context.Context) (bmc.SNMPConfig, error) {
	return f.snmp, nil
}

func (f *fakeExtendedProvider) SetSNMPAlertFilter(_ context.Context, _ map[string]any) error {
	return nil
}

func (f *fakeExtendedProvider) EnableSNMPv1Trap(_ context.Context, _ bool) error { return nil }

func (f *fakeExtendedProvider) EnableSNMPv3Trap(_ context.Context, _ bool) error { return nil }

func TestClientSNMPWiring(t *testing.T) {
	want := bmc.SNMPConfig{V1TrapEnabled: true, TrapPort: 162, CommunityNames: []string{"public"}}

	registry := registrar.NewRegistry()
	registry.Register("fake", "fake", nil, nil, &fakeExtendedProvider{name: "fake", snmp: want})

	cl := NewClient("", "", "", WithRegistry(registry))

	got, err := cl.SNMP(context.Background())
	if err != nil {
		t.Fatalf("SNMP: %v", err)
	}
	if got.TrapPort != want.TrapPort || !got.V1TrapEnabled {
		t.Errorf("SNMP = %+v, want %+v", got, want)
	}
	if md := cl.GetMetadata(); md.SuccessfulProvider != "fake" {
		t.Errorf("metadata SuccessfulProvider = %q, want %q", md.SuccessfulProvider, "fake")
	}
}

func TestClientSNMPNoProvider(t *testing.T) {
	registry := registrar.NewRegistry()
	registry.Register("bare", "bare", nil, nil, &testProvider{PName: "bare"})

	cl := NewClient("", "", "", WithRegistry(registry))

	if _, err := cl.SNMP(context.Background()); err == nil {
		t.Fatal("expected an error when no provider implements SNMPConfigurer")
	}
}
