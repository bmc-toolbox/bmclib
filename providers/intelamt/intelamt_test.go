package intelamt

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/jacobweinstock/iamt"
)

type mock struct {
	errSetPXE      error
	errIsPoweredOn error
	poweredON      bool
	errPowerOn     error
	errPowerOff    error
	errPowerCycle  error
	errOpen        error
}

func (m *mock) Open(ctx context.Context) error {
	return m.errOpen
}

func (m *mock) Close(ctx context.Context) error {
	return nil
}

func (m *mock) IsPoweredOn(ctx context.Context) (bool, error) {
	if m.errIsPoweredOn != nil {
		return false, m.errIsPoweredOn
	}
	return m.poweredON, nil
}

func (m *mock) PowerOn(ctx context.Context) error {
	return m.errPowerOn
}

func (m *mock) PowerOff(ctx context.Context) error {
	return m.errPowerOff
}

func (m *mock) PowerCycle(ctx context.Context) error {
	return m.errPowerCycle
}

func (m *mock) SetPXE(ctx context.Context) error {
	return m.errSetPXE
}

func TestClose(t *testing.T) {
	conn := &Conn{client: &mock{}}
	if err := conn.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestName(t *testing.T) {
	conn := &Conn{client: &mock{}}
	if diff := cmp.Diff(conn.Name(), ProviderName); diff != "" {
		t.Fatal(diff)
	}
}

func TestBootDeviceSet(t *testing.T) {
	tests := map[string]struct {
		want     bool
		err      error
		failCall bool
		device   string
	}{
		"success":                   {want: true, device: "pxe"},
		"invalid boot device":       {want: false, err: errors.New("only pxe boot device is supported for AMT provider"), device: "invalid"},
		"failed to set boot device": {want: false, failCall: true, err: errors.New(""), device: "pxe"},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := &mock{}
			if tt.failCall {
				m = &mock{errSetPXE: tt.err}
			}
			conn := &Conn{client: m}
			ctx := context.Background()
			if err := conn.Open(ctx); err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			got, err := conn.BootDeviceSet(ctx, tt.device, false, false)
			if err != nil && tt.err == nil {
				t.Fatalf("expected nil error, got: %v", err)
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestPowerStateGet(t *testing.T) {
	tests := map[string]struct {
		want string
		err  error
	}{
		"power on":                  {want: "on"},
		"power off":                 {want: "off"},
		"invalid power state":       {want: "", err: errors.New("invalid power state: invalid")},
		"failed to set power state": {want: "", err: errors.New("")},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var state bool
			switch tt.want {
			case "on":
				state = true
			case "off":
				state = false
			default:
			}
			m := &mock{poweredON: state, errIsPoweredOn: tt.err}
			conn := &Conn{client: m}
			ctx := context.Background()
			if err := conn.Open(ctx); err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			got, err := conn.PowerStateGet(ctx)
			if err != nil && tt.err == nil {
				t.Fatalf("expected nil error, got: %v", err)
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestPowerSet(t *testing.T) {
	tests := map[string]struct {
		want      bool
		err       error
		poweredOn bool
		wantState string
	}{
		"power on success":     {want: true, wantState: "on"},
		"power on success 2":   {want: true, wantState: "on", poweredOn: true},
		"power on failed":      {want: false, wantState: "on", err: errors.New("failed to power on")},
		"power off success":    {want: true, wantState: "off"},
		"power off success 2":  {want: true, wantState: "off", poweredOn: true},
		"power off failed":     {want: false, poweredOn: true, wantState: "off", err: errors.New("failed to power off")},
		"power cycle success":  {want: true, wantState: "cycle"},
		"power cycle failed":   {want: false, wantState: "cycle", err: errors.New("failed to power cycle")},
		"power cycle failed 2": {want: false, wantState: "cycle", poweredOn: false, err: errors.New("failed to power cycle")},
		"invalid power state":  {want: false, wantState: "unknown", err: errors.New("requested state type unknown")},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := &mock{}
			switch name {
			case "power on failed":
				m.errPowerOn = tt.err
			case "power off failed":
				m.errPowerOff = tt.err
			case "power cycle failed":
				m.errPowerCycle = tt.err
			case "power cycle failed 2":
				m.errPowerCycle = tt.err
				m.errPowerOn = tt.err
			default:
			}
			m.poweredON = tt.poweredOn
			conn := &Conn{client: m}
			ctx := context.Background()
			if err := conn.Open(ctx); err != nil {
				t.Fatal(err)
			}
			defer conn.Close()
			got, err := conn.PowerSet(ctx, tt.wantState)
			if err != nil && tt.err == nil {
				t.Fatalf("expected nil error, got: %v", err)
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestCompatible(t *testing.T) {
	tests := map[string]struct {
		want       bool
		failOnOpen bool
	}{
		"success":         {want: true},
		"failed on open":  {want: false, failOnOpen: true},
		"failed on power": {want: false},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			m := &mock{}
			if !tt.want {
				if tt.failOnOpen {
					m.errOpen = errors.New("failed to open")
				} else {
					m.errIsPoweredOn = errors.New("failed to power on")
				}
			}
			conn := &Conn{client: m}
			ctx := context.Background()
			defer conn.Close()
			got := conn.Compatible(ctx)
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestNew(t *testing.T) {
	wantClient := iamt.NewClient(logr.Discard(), "localhost", "", "admin", "pass")
	want := &Conn{client: wantClient}
	got := New(logr.Discard(), "localhost", "", "admin", "pass")
	t.Log(got == nil)
	c := Conn{}
	a := iamt.Client{}
	l := logr.Logger{}
	if diff := cmp.Diff(got, want, cmpopts.IgnoreUnexported(c, a, l)); diff != "" { //nolint:govet
		t.Fatal(diff)
	}
}
