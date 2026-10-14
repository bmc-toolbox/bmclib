package bmc

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakePower is a test provider implementing PowerReader and PowerCapSetter
// (and Provider for a stable name).
type fakePower struct {
	name    string
	info    PowerInfo
	getErr  error
	setErr  error
	setSeen *bool
}

func (f *fakePower) Name() string { return f.name }

func (f *fakePower) ReadPower(_ context.Context) (PowerInfo, error) {
	return f.info, f.getErr
}

func (f *fakePower) SetPowerCap(_ context.Context, _ *float64) error {
	if f.setSeen != nil {
		*f.setSeen = true
	}
	return f.setErr
}

// bareProvider implements Provider but none of the extended interfaces.
type bareProvider struct{ name string }

func (b *bareProvider) Name() string { return b.name }

const testTimeout = 5 * time.Second

func TestRunProviderRead_Success(t *testing.T) {
	want := PowerInfo{ConsumedWatts: 240, CapacityWatts: 1100}
	providers := []interface{}{
		&bareProvider{name: "bare"},
		&fakePower{name: "lenovo", info: want},
	}

	got, metadata, err := ReadPowerFromInterfaces(context.Background(), testTimeout, providers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ConsumedWatts != want.ConsumedWatts || got.CapacityWatts != want.CapacityWatts {
		t.Errorf("info = %+v, want %+v", got, want)
	}
	if metadata.SuccessfulProvider != "lenovo" {
		t.Errorf("SuccessfulProvider = %q, want %q", metadata.SuccessfulProvider, "lenovo")
	}
}

func TestRunProviderRead_NoImplementations(t *testing.T) {
	providers := []interface{}{&bareProvider{name: "bare"}, nil}

	_, _, err := ReadPowerFromInterfaces(context.Background(), testTimeout, providers)
	if err == nil {
		t.Fatal("expected an error when no provider implements the interface")
	}
	if !containsStr(err.Error(), "no PowerReader implementations found") {
		t.Errorf("error = %q, want it to mention no implementations", err.Error())
	}
}

func TestRunProviderRead_FailureRecorded(t *testing.T) {
	providers := []interface{}{
		&fakePower{name: "lenovo", getErr: errors.New("boom")},
	}

	_, metadata, err := ReadPowerFromInterfaces(context.Background(), testTimeout, providers)
	if err == nil {
		t.Fatal("expected an error when the provider call fails")
	}
	if detail := metadata.FailedProviderDetail["lenovo"]; detail == "" {
		t.Error("expected FailedProviderDetail to record the failing provider")
	}
}

func TestRunProviderAction_Success(t *testing.T) {
	var seen bool
	providers := []interface{}{&fakePower{name: "lenovo", setSeen: &seen}}

	limit := 500.0
	metadata, err := SetPowerCapFromInterfaces(context.Background(), testTimeout, &limit, providers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !seen {
		t.Error("expected SetPowerCap to be invoked")
	}
	if metadata.SuccessfulProvider != "lenovo" {
		t.Errorf("SuccessfulProvider = %q, want %q", metadata.SuccessfulProvider, "lenovo")
	}
}

func TestRunProviderAction_Error(t *testing.T) {
	providers := []interface{}{&fakePower{name: "lenovo", setErr: errors.New("denied")}}

	_, err := SetPowerCapFromInterfaces(context.Background(), testTimeout, nil, providers)
	if err == nil {
		t.Fatal("expected an error when the action fails")
	}
}

func containsStr(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
