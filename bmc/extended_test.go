package bmc

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeSecureBoot is a test provider implementing SecureBootManager (and Provider
// for a stable name).
type fakeSecureBoot struct {
	name    string
	state   SecureBootState
	getErr  error
	setErr  error
	setSeen *bool
}

func (f *fakeSecureBoot) Name() string { return f.name }

func (f *fakeSecureBoot) GetSecureBoot(_ context.Context) (SecureBootState, error) {
	return f.state, f.getErr
}

func (f *fakeSecureBoot) SetSecureBoot(_ context.Context, _ bool) error {
	if f.setSeen != nil {
		*f.setSeen = true
	}
	return f.setErr
}

func (f *fakeSecureBoot) ResetSecureBootKeys(_ context.Context, _ string) error { return nil }

// bareProvider implements Provider but none of the extended interfaces.
type bareProvider struct{ name string }

func (b *bareProvider) Name() string { return b.name }

const testTimeout = 5 * time.Second

func TestRunProviderRead_Success(t *testing.T) {
	want := SecureBootState{Enabled: true, Mode: "SetupMode"}
	providers := []interface{}{
		&bareProvider{name: "bare"},
		&fakeSecureBoot{name: "lenovo", state: want},
	}

	got, metadata, err := GetSecureBootFromInterfaces(context.Background(), testTimeout, providers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("state = %+v, want %+v", got, want)
	}
	if metadata.SuccessfulProvider != "lenovo" {
		t.Errorf("SuccessfulProvider = %q, want %q", metadata.SuccessfulProvider, "lenovo")
	}
}

func TestRunProviderRead_NoImplementations(t *testing.T) {
	providers := []interface{}{&bareProvider{name: "bare"}, nil}

	_, _, err := GetSecureBootFromInterfaces(context.Background(), testTimeout, providers)
	if err == nil {
		t.Fatal("expected an error when no provider implements the interface")
	}
	if !containsStr(err.Error(), "no SecureBootManager implementations found") {
		t.Errorf("error = %q, want it to mention no implementations", err.Error())
	}
}

func TestRunProviderRead_FailureRecorded(t *testing.T) {
	providers := []interface{}{
		&fakeSecureBoot{name: "lenovo", getErr: errors.New("boom")},
	}

	_, metadata, err := GetSecureBootFromInterfaces(context.Background(), testTimeout, providers)
	if err == nil {
		t.Fatal("expected an error when the provider call fails")
	}
	if detail := metadata.FailedProviderDetail["lenovo"]; detail == "" {
		t.Error("expected FailedProviderDetail to record the failing provider")
	}
}

func TestRunProviderAction_Success(t *testing.T) {
	var seen bool
	providers := []interface{}{&fakeSecureBoot{name: "lenovo", setSeen: &seen}}

	metadata, err := SetSecureBootFromInterfaces(context.Background(), testTimeout, true, providers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !seen {
		t.Error("expected SetSecureBoot to be invoked")
	}
	if metadata.SuccessfulProvider != "lenovo" {
		t.Errorf("SuccessfulProvider = %q, want %q", metadata.SuccessfulProvider, "lenovo")
	}
}

func TestRunProviderAction_Error(t *testing.T) {
	providers := []interface{}{&fakeSecureBoot{name: "lenovo", setErr: errors.New("denied")}}

	_, err := SetSecureBootFromInterfaces(context.Background(), testTimeout, true, providers)
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
