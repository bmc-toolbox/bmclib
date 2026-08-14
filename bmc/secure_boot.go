package bmc

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// SecureBootStateGetter provides retrieval of whether UEFI Secure Boot is enabled.
type SecureBootStateGetter interface {
	GetSecureBoot(ctx context.Context) (enabled bool, err error)
}

type secureBootStateGetterProvider struct {
	name string
	SecureBootStateGetter
}

// SecureBootSetter provides enabling/disabling UEFI Secure Boot.
type SecureBootSetter interface {
	SetSecureBoot(ctx context.Context, enable bool) (err error)
}

type secureBootSetterProvider struct {
	name string
	SecureBootSetter
}

// SecureBootKeysResetter provides resetting the UEFI Secure Boot key databases.
type SecureBootKeysResetter interface {
	ResetSecureBootKeys(ctx context.Context, resetType string) (err error)
}

type secureBootKeysResetterProvider struct {
	name string
	SecureBootKeysResetter
}

func secureBootState(ctx context.Context, generic []secureBootStateGetterProvider) (enabled bool, metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.SecureBootStateGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			enabled, vErr := elem.GetSecureBoot(ctx)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return enabled, metadata, nil
		}
	}

	return enabled, metadata, multierror.Append(err, errors.New("failure to get secure boot state"))
}

func setSecureBoot(ctx context.Context, generic []secureBootSetterProvider, enable bool) (metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.SecureBootSetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			vErr := elem.SetSecureBoot(ctx, enable)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return metadata, nil
		}
	}

	return metadata, multierror.Append(err, errors.New("failure to set secure boot state"))
}

func resetSecureBootKeys(ctx context.Context, generic []secureBootKeysResetterProvider, resetType string) (metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.SecureBootKeysResetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			vErr := elem.ResetSecureBootKeys(ctx, resetType)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return metadata, nil
		}
	}

	return metadata, multierror.Append(err, errors.New("failure to reset secure boot keys"))
}

// GetSecureBootStateFromInterfaces returns whether UEFI Secure Boot is enabled using
// the first successful SecureBootStateGetter implementation found in generic.
func GetSecureBootStateFromInterfaces(ctx context.Context, generic []interface{}) (enabled bool, metadata Metadata, err error) {
	implementations := make([]secureBootStateGetterProvider, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := secureBootStateGetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case SecureBootStateGetter:
			temp.SecureBootStateGetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a SecureBootStateGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return enabled, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no SecureBootStateGetter implementations found"),
			),
		)
	}

	return secureBootState(ctx, implementations)
}

// SetSecureBootFromInterfaces enables/disables UEFI Secure Boot using the first
// successful SecureBootSetter implementation found in generic.
func SetSecureBootFromInterfaces(ctx context.Context, generic []interface{}, enable bool) (metadata Metadata, err error) {
	implementations := make([]secureBootSetterProvider, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := secureBootSetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case SecureBootSetter:
			temp.SecureBootSetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a SecureBootSetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no SecureBootSetter implementations found"),
			),
		)
	}

	return setSecureBoot(ctx, implementations, enable)
}

// ResetSecureBootKeysFromInterfaces resets the UEFI Secure Boot key databases using
// the first successful SecureBootKeysResetter implementation found in generic.
func ResetSecureBootKeysFromInterfaces(ctx context.Context, generic []interface{}, resetType string) (metadata Metadata, err error) {
	implementations := make([]secureBootKeysResetterProvider, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := secureBootKeysResetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case SecureBootKeysResetter:
			temp.SecureBootKeysResetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a SecureBootKeysResetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no SecureBootKeysResetter implementations found"),
			),
		)
	}

	return resetSecureBootKeys(ctx, implementations, resetType)
}
