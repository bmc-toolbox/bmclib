package bmc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// PowerState is a normalized, provider independent power state of a machine.
//
// PowerStateGetter implementations return the provider's raw power state
// string (e.g. Redfish "PoweringOff", ipmitool "Chassis Power is on"). Use
// ParsePowerState to convert such a string into a PowerState.
type PowerState string

// PowerState values. Transition states are kept distinct from
// PowerStateUnknown so callers waiting for a final state can keep polling.
const (
	PowerStateOn          PowerState = "on"
	PowerStateOff         PowerState = "off"
	PowerStatePoweringOn  PowerState = "powering_on"
	PowerStatePoweringOff PowerState = "powering_off"
	PowerStatePaused      PowerState = "paused"
	PowerStateUnknown     PowerState = "unknown"
)

// rawPowerStates maps the single line raw power states returned by the
// bmclib providers, lower cased, to their normalized PowerState: the Redfish
// ComputerSystem.PowerState values (redfish, dell, lenovo, supermicro,
// openbmc), and "on"/"off" from intelamt, asrockrack, rpc and homeassistant.
var rawPowerStates = map[string]PowerState{
	"on":          PowerStateOn,
	"off":         PowerStateOff,
	"poweringon":  PowerStatePoweringOn,
	"poweringoff": PowerStatePoweringOff,
	"paused":      PowerStatePaused,
}

// ipmiPowerStates maps the "chassis power status" output lines of the
// ipmitool and goipmi providers, lower cased, to their normalized PowerState.
var ipmiPowerStates = map[string]PowerState{
	"chassis power is on":  PowerStateOn,
	"chassis power is off": PowerStateOff,
}

// ParsePowerState converts a raw power state string, as returned by
// PowerStateGetter.PowerStateGet or Client.GetPowerState, into a PowerState.
// It accepts raw provider values only, not PowerState values.
//
// Matching is exact, never by substring, so "PoweringOff" is never mistaken
// for "off". It ignores case and surrounding whitespace, as providers differ
// in capitalization ("On" from Redfish, "on" from intelamt).
//
// A single line must be one of the known provider values. Multiple lines are
// only accepted for IPMI output, as ipmitool may print warnings (e.g. "Unable
// to Get Channel Cipher Suites") around its "Chassis Power is on/off" result:
// exactly one IPMI result must be present, other lines are ignored unless
// they are themselves a power state value.
//
// PowerStateUnknown is returned for anything else.
func ParsePowerState(raw string) PowerState {
	raw = strings.ToLower(strings.TrimSpace(raw))
	if !strings.Contains(raw, "\n") {
		if s, ok := rawPowerStates[raw]; ok {
			return s
		}
		if s, ok := ipmiPowerStates[raw]; ok {
			return s
		}

		return PowerStateUnknown
	}

	state := PowerStateUnknown
	for _, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if _, ok := rawPowerStates[line]; ok {
			return PowerStateUnknown
		}
		s, ok := ipmiPowerStates[line]
		if !ok {
			continue
		}
		if state != PowerStateUnknown {
			return PowerStateUnknown
		}
		state = s
	}

	return state
}

// getNormalizedPowerState gets the normalized power state for a BMC, trying
// all interface implementations passed in. A provider whose raw power state
// is not recognized by ParsePowerState counts as failed, so the next provider
// is tried.
func getNormalizedPowerState(ctx context.Context, timeout time.Duration, p []powerProviders) (state PowerState, m Metadata, err error) {
	metadataLocal := Metadata{
		FailedProviderDetail: make(map[string]string),
	}
	for _, elem := range p {
		if elem.powerStateGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return PowerStateUnknown, metadataLocal, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			ctx, cancel := context.WithTimeout(ctx, timeout)
			raw, stateErr := elem.powerStateGetter.PowerStateGet(ctx)
			cancel()
			if stateErr == nil {
				if state = ParsePowerState(raw); state == PowerStateUnknown {
					stateErr = errors.WithMessagef(bmclibErrs.ErrPowerStatusUnknown, "%q", raw)
				}
			}
			if stateErr != nil {
				err = multierror.Append(err, errors.WithMessagef(stateErr, "provider: %v", elem.name))
				metadataLocal.FailedProviderDetail[elem.name] = stateErr.Error()
				continue
			}
			metadataLocal.SuccessfulProvider = elem.name
			return state, metadataLocal, nil
		}
	}
	return PowerStateUnknown, metadataLocal, multierror.Append(err, errors.New("failed to get power state"))
}

// GetNormalizedPowerStateFromInterfaces identifies implementations of the
// PowerStateGetter interface and passes the found implementations to the
// getNormalizedPowerState() wrapper.
func GetNormalizedPowerStateFromInterfaces(ctx context.Context, timeout time.Duration, generic []interface{}) (state PowerState, metadata Metadata, err error) {
	metadata = newMetadata()

	powerStateGetter := make([]powerProviders, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := powerProviders{name: getProviderName(elem)}
		switch p := elem.(type) {
		case PowerStateGetter:
			temp.powerStateGetter = p
			powerStateGetter = append(powerStateGetter, temp)
		default:
			e := fmt.Sprintf("not a PowerStateGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(powerStateGetter) == 0 {
		return PowerStateUnknown, metadata, multierror.Append(err, errors.New("no PowerStateGetter implementations found"))
	}
	return getNormalizedPowerState(ctx, timeout, powerStateGetter)
}
