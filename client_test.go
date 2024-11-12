package bmclib

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/jacobweinstock/registrar"
	"github.com/metal-toolbox/bmclib/logging"
	"gopkg.in/go-playground/assert.v1"
)

func TestBMC(t *testing.T) {
	t.Skip("needs ipmitool and real ipmi server")
	host := "127.0.0.1"
	user := "admin"
	pass := "admin"

	log := logging.DefaultLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cl := NewClient(host, user, pass, WithLogger(log), WithPerProviderTimeout(5*time.Second))
	if err := cl.Open(ctx); err != nil {
		t.Logf("%+v", cl.GetMetadata())
		t.Fatal(err)
	}
	defer cl.Close(ctx)
	t.Logf("metadata for Open: %+v", cl.GetMetadata())

	cl.Registry.Drivers = cl.Registry.PreferDriver("non-existent")
	state, err := cl.GetPowerState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)
	t.Logf("metadata for GetPowerState: %+v", cl.GetMetadata())

	cl.Registry.Drivers = cl.Registry.PreferDriver("ipmitool")
	state, err = cl.PreferProvider("gofish").GetPowerState(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(state)
	t.Logf("metadata for GetPowerState: %+v", cl.GetMetadata())

	users, err := cl.ReadUsers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(users)
	t.Logf("metadata for ReadUsers: %+v", cl.GetMetadata())

	t.Fatal()
}

func TestWithRedfishVersionsNotCompatible(t *testing.T) {
	host := "127.0.0.1"
	user := "ADMIN"
	pass := "ADMIN"

	tests := []struct {
		name     string
		versions []string
	}{
		{
			"no versions",
			[]string{},
		},
		{
			"with versions",
			[]string{"1.2.3", "4.5.6"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := NewClient(host, user, pass, WithRedfishVersionsNotCompatible(tt.versions))
			assert.Equal(t, tt.versions, cl.providerConfig.gofish.VersionsNotCompatible)
		})
	}
}

func TestWithRedfishBasicAuth(t *testing.T) {
	host := "127.0.0.1"
	user := "ADMIN"
	pass := "ADMIN"

	tests := []struct {
		name    string
		enabled bool
	}{
		{
			"disabled",
			false,
		},
		{
			"enabled",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var opts []Option
			if tt.enabled {
				opts = append(opts, WithRedfishUseBasicAuth(true))
			}

			cl := NewClient(host, user, pass, opts...)
			assert.Equal(t, tt.enabled, cl.providerConfig.gofish.UseBasicAuth)
		})
	}
}

func TestWithConnectionTimeout(t *testing.T) {
	host := "127.0.0.1"
	user := "ADMIN"
	pass := "ADMIN"

	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			"no connection timeout",
			0,
		},
		{
			"with timeout",
			5 * time.Second,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := NewClient(host, user, pass, WithPerProviderTimeout(tt.timeout))
			assert.Equal(t, tt.timeout, cl.perProviderTimeout(nil))
		})
	}
}

func TestDefaultTimeout(t *testing.T) {
	tests := map[string]struct {
		ctx  context.Context
		want func(n int) time.Duration
	}{
		"no per provider timeout": {
			ctx:  context.Background(),
			want: func(n int) time.Duration { return 30 * time.Second },
		},
		"with per provider timeout": {
			ctx: func() context.Context {
				c, d := context.WithTimeout(context.Background(), 5*time.Second)
				defer d()
				return c
			}(),
			want: func(n int) time.Duration {
				v := (4999 * time.Millisecond / time.Duration(n))
				return v.Round(time.Millisecond)
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := NewClient("", "", "")
			got := c.defaultTimeout(tt.ctx)
			if !equalWithinErrorMargin(got.Round(time.Millisecond), tt.want(len(c.Registry.Drivers))) {
				diff := cmp.Diff(got.Round(time.Millisecond), tt.want(len(c.Registry.Drivers)))
				t.Errorf("unexpected timeout (-want +got):\n%s", diff)
			}
		})
	}
}

func equalWithinErrorMargin(a, b time.Duration) bool {
	return (a - b) < 10*time.Millisecond
}

type testProvider struct {
	PName        string
	Powerstate   string
	BootdeviceOK bool
	Err          error
}

func (t *testProvider) Name() string {
	if t.PName != "" {
		return t.PName
	}
	return "tester"
}

func (t *testProvider) Open(ctx context.Context) error {
	return t.Err
}

func (t *testProvider) Close(ctx context.Context) error {
	return t.Err
}

func (t *testProvider) PowerStateGet(ctx context.Context) (string, error) {
	return t.Powerstate, t.Err
}

func (t *testProvider) PowerSet(ctx context.Context, state string) error {
	return t.Err
}

func (t *testProvider) BootDeviceSet(ctx context.Context, bootDevice string, setPersistent, efiBoot bool) (ok bool, err error) {
	return t.BootdeviceOK, t.Err
}

func registryNames(r []*registrar.Driver) []string {
	var names []string
	for _, d := range r {
		names = append(names, d.Name)
	}
	return names
}

func TestOpenFiltered(t *testing.T) {
	registry := registrar.NewRegistry()
	registry.Register("tester1", "tester1", nil, nil, &testProvider{PName: "tester1"})
	registry.Register("tester2", "tester2", nil, nil, &testProvider{PName: "tester2"})
	registry.Register("tester3", "tester3", nil, nil, &testProvider{PName: "tester3"})
	cl := NewClient("", "", "", WithRegistry(registry))
	if err := cl.PreferProvider("tester3").Open(context.Background()); err != nil {
		t.Fatal(err)
	}
	defer cl.Close(context.Background())
	want := []string{"tester3", "tester1", "tester2"}
	if diff := cmp.Diff(cl.GetMetadata().ProvidersAttempted, want); diff != "" {
		t.Errorf("diff: %s", diff)
	}
	want = []string{"tester1", "tester2", "tester3"}
	if diff := cmp.Diff(registryNames(cl.Registry.Drivers), want); diff != "" {
		t.Errorf("diff: %s", diff)
	}
}
