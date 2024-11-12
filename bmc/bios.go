package bmc

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
	"github.com/pkg/errors"
)

type BiosConfigurationGetter interface {
	GetBiosConfiguration(ctx context.Context) (biosConfig map[string]string, err error)
}

type biosConfigurationGetterProvider struct {
	name string
	BiosConfigurationGetter
}

type BiosConfigurationSetter interface {
	SetBiosConfiguration(ctx context.Context, biosConfig map[string]string) (err error)
	SetBiosConfigurationFromFile(ctx context.Context, cfg string) (err error)
}

type biosConfigurationSetterProvider struct {
	name string
	BiosConfigurationSetter
}

type BiosConfigurationResetter interface {
	ResetBiosConfiguration(ctx context.Context) (err error)
}

type biosConfigurationResetterProvider struct {
	name string
	BiosConfigurationResetter
}

func biosConfiguration(ctx context.Context, generic []biosConfigurationGetterProvider) (biosConfig map[string]string, metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.BiosConfigurationGetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			biosConfig, vErr := elem.GetBiosConfiguration(ctx)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				err = multierror.Append(err, vErr)
				continue

			}
			metadata.SuccessfulProvider = elem.name
			return biosConfig, metadata, nil
		}
	}

	return biosConfig, metadata, multierror.Append(err, errors.New("failure to get bios configuration"))
}

func setBiosConfiguration(ctx context.Context, generic []biosConfigurationSetterProvider, biosConfig map[string]string) (metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.BiosConfigurationSetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			vErr := elem.SetBiosConfiguration(ctx, biosConfig)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				err = multierror.Append(err, vErr)
				continue

			}
			metadata.SuccessfulProvider = elem.name
			return metadata, nil
		}
	}

	return metadata, multierror.Append(err, errors.New("failure to set bios configuration"))
}

func setBiosConfigurationFromFile(ctx context.Context, generic []biosConfigurationSetterProvider, cfg string) (metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.BiosConfigurationSetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			vErr := elem.SetBiosConfigurationFromFile(ctx, cfg)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				err = multierror.Append(err, vErr)
				continue

			}
			metadata.SuccessfulProvider = elem.name
			return metadata, nil
		}
	}

	return metadata, multierror.Append(err, errors.New("failure to set bios configuration from file"))
}

func resetBiosConfiguration(ctx context.Context, generic []biosConfigurationResetterProvider) (metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.BiosConfigurationResetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			vErr := elem.ResetBiosConfiguration(ctx)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				err = multierror.Append(err, vErr)
				continue

			}
			metadata.SuccessfulProvider = elem.name
			return metadata, nil
		}
	}

	return metadata, multierror.Append(err, errors.New("failure to reset bios configuration"))
}

func GetBiosConfigurationInterfaces(ctx context.Context, generic []interface{}) (biosConfig map[string]string, metadata Metadata, err error) {
	implementations := make([]biosConfigurationGetterProvider, 0)
	for _, elem := range generic {
		temp := biosConfigurationGetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case BiosConfigurationGetter:
			temp.BiosConfigurationGetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a BiosConfigurationGetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return biosConfig, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no BiosConfigurationGetter implementations found"),
			),
		)
	}

	return biosConfiguration(ctx, implementations)
}

func SetBiosConfigurationInterfaces(ctx context.Context, generic []interface{}, biosConfig map[string]string) (metadata Metadata, err error) {
	implementations := make([]biosConfigurationSetterProvider, 0)
	for _, elem := range generic {
		temp := biosConfigurationSetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case BiosConfigurationSetter:
			temp.BiosConfigurationSetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a BiosConfigurationSetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no BiosConfigurationSetter implementations found"),
			),
		)
	}

	return setBiosConfiguration(ctx, implementations, biosConfig)
}

func SetBiosConfigurationFromFileInterfaces(ctx context.Context, generic []interface{}, cfg string) (metadata Metadata, err error) {
	implementations := make([]biosConfigurationSetterProvider, 0)
	for _, elem := range generic {
		temp := biosConfigurationSetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case BiosConfigurationSetter:
			temp.BiosConfigurationSetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a BiosConfigurationSetterFromFile implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no BiosConfigurationSetterFromFile implementations found"),
			),
		)
	}

	return setBiosConfigurationFromFile(ctx, implementations, cfg)
}

func ResetBiosConfigurationInterfaces(ctx context.Context, generic []interface{}) (metadata Metadata, err error) {
	implementations := make([]biosConfigurationResetterProvider, 0)
	for _, elem := range generic {
		temp := biosConfigurationResetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case BiosConfigurationResetter:
			temp.BiosConfigurationResetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a BiosConfigurationResetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no BiosConfigurationResetter implementations found"),
			),
		)
	}

	return resetBiosConfiguration(ctx, implementations)
}
