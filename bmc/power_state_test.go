package bmc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
)

func TestParsePowerState(t *testing.T) {
	testCases := map[string]struct {
		raw  string
		want PowerState
	}{
		// Redfish ComputerSystem.PowerState
		"redfish On":               {raw: "On", want: PowerStateOn},
		"redfish Off":              {raw: "Off", want: PowerStateOff},
		"redfish PoweringOn":       {raw: "PoweringOn", want: PowerStatePoweringOn},
		"redfish PoweringOff":      {raw: "PoweringOff", want: PowerStatePoweringOff},
		"redfish Paused":           {raw: "Paused", want: PowerStatePaused},
		"redfish empty":            {raw: "", want: PowerStateUnknown},
		"redfish trailing newline": {raw: "On\n", want: PowerStateOn},
		// intelamt, asrockrack, rpc, homeassistant
		"lower case on":             {raw: "on", want: PowerStateOn},
		"lower case off":            {raw: "off", want: PowerStateOff},
		"homeassistant unavailable": {raw: "unavailable", want: PowerStateUnknown},
		"homeassistant unknown":     {raw: "unknown", want: PowerStateUnknown},
		// case and whitespace
		"surrounding whitespace":  {raw: "  On \t", want: PowerStateOn},
		"upper case":              {raw: "ON", want: PowerStateOn},
		"mixed case":              {raw: "oFf", want: PowerStateOff},
		"lower case powering off": {raw: "poweringoff", want: PowerStatePoweringOff},
		// raw provider values only
		"normalized powering_off": {raw: "powering_off", want: PowerStateUnknown},
		"normalized unknown":      {raw: string(PowerStateUnknown), want: PowerStateUnknown},
		// ipmitool, goipmi
		"ipmi on":                         {raw: "Chassis Power is on", want: PowerStateOn},
		"ipmi off":                        {raw: "Chassis Power is off", want: PowerStateOff},
		"ipmi mixed case":                 {raw: "Chassis Power Is ON", want: PowerStateOn},
		"ipmitool on newline":             {raw: "Chassis Power is on\n", want: PowerStateOn},
		"ipmitool off crlf":               {raw: "Chassis Power is off\r\n", want: PowerStateOff},
		"ipmitool warning before result":  {raw: "Unable to Get Channel Cipher Suites\nChassis Power is off\n", want: PowerStateOff},
		"ipmitool warning after result":   {raw: "Chassis Power is on\nUnable to Get Channel Cipher Suites\n", want: PowerStateOn},
		"ipmitool only warning":           {raw: "Unable to Get Channel Cipher Suites\n", want: PowerStateUnknown},
		"ipmitool conflicting results":    {raw: "Chassis Power is on\nChassis Power is off", want: PowerStateUnknown},
		"ipmitool duplicated result":      {raw: "Chassis Power is on\nChassis Power is on", want: PowerStateUnknown},
		"ipmi result mixed with redfish":  {raw: "On\nChassis Power is on", want: PowerStateUnknown},
		"ipmi unexpected value":           {raw: "Chassis Power is unknown", want: PowerStateUnknown},
		"ipmi trailing text":              {raw: "Chassis Power is on (soft)", want: PowerStateUnknown},
		"multi line without ipmi result":  {raw: "on\nfault: PSU failure, shutting down", want: PowerStateUnknown},
		"multi line conflicting values":   {raw: "On\nOff", want: PowerStateUnknown},
		"substring of on":                 {raw: "Online", want: PowerStateUnknown},
		"substring of off":                {raw: "Power off pending", want: PowerStateUnknown},
		"substring of powering off value": {raw: "PoweringOff in progress", want: PowerStateUnknown},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if diff := cmp.Diff(tc.want, ParsePowerState(tc.raw)); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

type rawPowerStateTester struct {
	name string
	raw  string
	err  error
}

func (r *rawPowerStateTester) PowerStateGet(ctx context.Context) (string, error) {
	return r.raw, r.err
}

func (r *rawPowerStateTester) Name() string {
	return r.name
}

func TestGetNormalizedPowerState(t *testing.T) {
	testCases := map[string]struct {
		providers      []*rawPowerStateTester
		ctxTimeout     time.Duration
		want           PowerState
		wantSuccessful string
		wantFailed     map[string]string
		err            error
	}{
		"success": {
			providers:      []*rawPowerStateTester{{name: "redfish", raw: "PoweringOff"}},
			want:           PowerStatePoweringOff,
			wantSuccessful: "redfish",
			wantFailed:     map[string]string{},
		},
		"unknown falls back to next provider": {
			providers:      []*rawPowerStateTester{{name: "redfish", raw: ""}, {name: "ipmitool", raw: "Chassis Power is on\n"}},
			want:           PowerStateOn,
			wantSuccessful: "ipmitool",
			wantFailed:     map[string]string{"redfish": `"": unrecognized power status`},
		},
		"error falls back to next provider": {
			providers:      []*rawPowerStateTester{{name: "redfish", err: errors.New("boom")}, {name: "ipmitool", raw: "Chassis Power is off"}},
			want:           PowerStateOff,
			wantSuccessful: "ipmitool",
			wantFailed:     map[string]string{"redfish": "boom"},
		},
		"all unknown": {
			providers:  []*rawPowerStateTester{{name: "homeassistant", raw: "unavailable"}},
			want:       PowerStateUnknown,
			wantFailed: map[string]string{"homeassistant": `"unavailable": unrecognized power status`},
			err:        &multierror.Error{Errors: []error{errors.New(`provider: homeassistant: "unavailable": unrecognized power status`), errors.New("failed to get power state")}},
		},
		"context timeout": {
			providers:  []*rawPowerStateTester{{name: "redfish", raw: "On"}},
			ctxTimeout: time.Nanosecond,
			want:       PowerStateUnknown,
			wantFailed: map[string]string{},
			err:        &multierror.Error{Errors: []error{errors.New("context deadline exceeded")}},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			if tc.ctxTimeout == time.Nanosecond {
				<-ctx.Done()
			}

			providers := make([]powerProviders, 0, len(tc.providers))
			for _, p := range tc.providers {
				providers = append(providers, powerProviders{name: p.name, powerStateGetter: p})
			}
			state, metadata, err := getNormalizedPowerState(ctx, time.Second, providers)
			if tc.err != nil {
				if err == nil {
					t.Fatal("expected error")
				}
				if diff := cmp.Diff(tc.err.Error(), err.Error()); diff != "" {
					t.Fatal(diff)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.want, state); diff != "" {
				t.Fatal(diff)
			}
			if diff := cmp.Diff(tc.wantSuccessful, metadata.SuccessfulProvider); diff != "" {
				t.Fatal(diff)
			}
			if diff := cmp.Diff(tc.wantFailed, metadata.FailedProviderDetail); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}

func TestGetNormalizedPowerStateFromInterfaces(t *testing.T) {
	testCases := map[string]struct {
		generic []interface{}
		want    PowerState
		err     error
	}{
		"success": {
			generic: []interface{}{&rawPowerStateTester{name: "redfish", raw: "Off"}},
			want:    PowerStateOff,
		},
		"no implementations found": {
			generic: []interface{}{&struct{}{}},
			want:    PowerStateUnknown,
			err:     &multierror.Error{Errors: []error{errors.New("not a PowerStateGetter implementation: *struct {}"), errors.New("no PowerStateGetter implementations found")}},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			state, _, err := GetNormalizedPowerStateFromInterfaces(context.Background(), time.Second, tc.generic)
			if tc.err != nil {
				if err == nil {
					t.Fatal("expected error")
				}
				if diff := cmp.Diff(tc.err.Error(), err.Error()); diff != "" {
					t.Fatal(diff)
				}
			} else if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tc.want, state); diff != "" {
				t.Fatal(diff)
			}
		})
	}
}
