package bmc

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// ResetSecureBootKeysType identifies the type of Secure Boot key reset to perform.
type ResetSecureBootKeysType string

// ResetSecureBootKeysType constants enumerate the supported reset types for the whole Secure Boot subsystem.
const (
	ResetSecureBootKeysTypeResetAllKeysToDefault ResetSecureBootKeysType = "ResetAllKeysToDefault"
	ResetSecureBootKeysTypeDeleteAllKeys         ResetSecureBootKeysType = "DeleteAllKeys"
	ResetSecureBootKeysTypeDeletePK              ResetSecureBootKeysType = "DeletePK"
)

// ResetSecureBootDatabaseKeysType identifies the type of per-database Secure Boot key reset to perform.
type ResetSecureBootDatabaseKeysType string

// ResetSecureBootDatabaseKeysType constants enumerate the supported reset types for a single Secure Boot database.
const (
	ResetSecureBootDatabaseKeysTypeResetAllKeysToDefault ResetSecureBootDatabaseKeysType = "ResetAllKeysToDefault"
	ResetSecureBootDatabaseKeysTypeDeleteAllKeys         ResetSecureBootDatabaseKeysType = "DeleteAllKeys"
)

// SecureBootDatabase identifies a UEFI Secure Boot key database.
type SecureBootDatabase string

// SecureBootDatabase constants enumerate the supported Secure Boot key databases.
const (
	SecureBootDatabaseDB  SecureBootDatabase = "db"
	SecureBootDatabaseKEK SecureBootDatabase = "KEK"
	SecureBootDatabasePK  SecureBootDatabase = "PK"
	SecureBootDatabaseDBX SecureBootDatabase = "dbx"
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
	ResetSecureBootKeys(ctx context.Context, resetType ResetSecureBootKeysType) (err error)
}

type secureBootKeysResetterProvider struct {
	name string
	SecureBootKeysResetter
}

// SecureBootDatabaseKeysResetter provides resetting a single UEFI Secure Boot key
// database (e.g. "db", "KEK"), leaving every other database - including PK -
// untouched.
type SecureBootDatabaseKeysResetter interface {
	ResetSecureBootDatabaseKeys(ctx context.Context, database SecureBootDatabase, resetType ResetSecureBootDatabaseKeysType) (err error)
}

type secureBootDatabaseKeysResetterProvider struct {
	name string
	SecureBootDatabaseKeysResetter
}

// SecureBootCertificateImporter provides enrolling a certificate into a single
// UEFI Secure Boot key database (e.g. "db", "KEK") without disturbing any
// certificate already present.
type SecureBootCertificateImporter interface {
	ImportSecureBootCertificate(ctx context.Context, database SecureBootDatabase, certificatePEM string) (err error)
}

type secureBootCertificateImporterProvider struct {
	name string
	SecureBootCertificateImporter
}

// SecureBootKeyManagementSetter controls whether the platform will accept
// custom UEFI Secure Boot keys, as opposed to using only the key set the
// firmware shipped with. Vendors expose this differently (a BIOS attribute on
// some platforms, a setup-menu-only setting on others, and not at all on
// platforms that never gate enrollment).
//
// Implementations MUST NOT alter the contents of any Secure Boot key database
// as a side effect. Callers rely on this to sequence a key-store reset and
// this call independently. A platform whose equivalent setting is coupled to
// key-store initialization cannot satisfy this contract and MUST NOT
// implement this interface.
//
// Implementations MUST be idempotent.
//
// rebootRequired reports that the change is staged and will not be in effect
// until the host has been power cycled.
type SecureBootKeyManagementSetter interface {
	SetSecureBootKeyManagement(ctx context.Context, enable bool) (rebootRequired bool, err error)
}

type secureBootKeyManagementSetterProvider struct {
	name string
	SecureBootKeyManagementSetter
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

func resetSecureBootKeys(ctx context.Context, generic []secureBootKeysResetterProvider, resetType ResetSecureBootKeysType) (metadata Metadata, err error) {
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

func resetSecureBootDatabaseKeys(ctx context.Context, generic []secureBootDatabaseKeysResetterProvider, database SecureBootDatabase, resetType ResetSecureBootDatabaseKeysType) (metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.SecureBootDatabaseKeysResetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			vErr := elem.ResetSecureBootDatabaseKeys(ctx, database, resetType)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return metadata, nil
		}
	}

	return metadata, multierror.Append(err, errors.New("failure to reset secure boot database keys"))
}

func importSecureBootCertificate(ctx context.Context, generic []secureBootCertificateImporterProvider, database SecureBootDatabase, certificatePEM string) (metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.SecureBootCertificateImporter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			vErr := elem.ImportSecureBootCertificate(ctx, database, certificatePEM)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return metadata, nil
		}
	}

	return metadata, multierror.Append(err, errors.New("failure to import secure boot certificate"))
}

func setSecureBootKeyManagement(ctx context.Context, generic []secureBootKeyManagementSetterProvider, enable bool) (rebootRequired bool, metadata Metadata, err error) {
	metadata = newMetadata()
Loop:
	for _, elem := range generic {
		if elem.SecureBootKeyManagementSetter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadata.ProvidersAttempted = append(metadata.ProvidersAttempted, elem.name)
			rebootRequired, vErr := elem.SetSecureBootKeyManagement(ctx, enable)
			if vErr != nil {
				err = multierror.Append(err, errors.WithMessagef(vErr, "provider: %v", elem.name))
				continue
			}
			metadata.SuccessfulProvider = elem.name
			return rebootRequired, metadata, nil
		}
	}

	return rebootRequired, metadata, multierror.Append(err, errors.New("failure to set secure boot key management"))
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
func ResetSecureBootKeysFromInterfaces(ctx context.Context, generic []interface{}, resetType ResetSecureBootKeysType) (metadata Metadata, err error) {
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

// ResetSecureBootDatabaseKeysFromInterfaces resets a single UEFI Secure Boot key
// database using the first successful SecureBootDatabaseKeysResetter
// implementation found in generic.
func ResetSecureBootDatabaseKeysFromInterfaces(ctx context.Context, generic []interface{}, database SecureBootDatabase, resetType ResetSecureBootDatabaseKeysType) (metadata Metadata, err error) {
	implementations := make([]secureBootDatabaseKeysResetterProvider, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := secureBootDatabaseKeysResetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case SecureBootDatabaseKeysResetter:
			temp.SecureBootDatabaseKeysResetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a SecureBootDatabaseKeysResetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no SecureBootDatabaseKeysResetter implementations found"),
			),
		)
	}

	return resetSecureBootDatabaseKeys(ctx, implementations, database, resetType)
}

// ImportSecureBootCertificateFromInterfaces enrolls a certificate into a single
// UEFI Secure Boot key database using the first successful
// SecureBootCertificateImporter implementation found in generic.
func ImportSecureBootCertificateFromInterfaces(ctx context.Context, generic []interface{}, database SecureBootDatabase, certificatePEM string) (metadata Metadata, err error) {
	implementations := make([]secureBootCertificateImporterProvider, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := secureBootCertificateImporterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case SecureBootCertificateImporter:
			temp.SecureBootCertificateImporter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a SecureBootCertificateImporter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no SecureBootCertificateImporter implementations found"),
			),
		)
	}

	return importSecureBootCertificate(ctx, implementations, database, certificatePEM)
}

// SetSecureBootKeyManagementFromInterfaces enables/disables acceptance of custom UEFI
// Secure Boot keys using the first successful SecureBootKeyManagementSetter
// implementation found in generic.
func SetSecureBootKeyManagementFromInterfaces(ctx context.Context, generic []interface{}, enable bool) (rebootRequired bool, metadata Metadata, err error) {
	implementations := make([]secureBootKeyManagementSetterProvider, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := secureBootKeyManagementSetterProvider{name: getProviderName(elem)}
		switch p := elem.(type) {
		case SecureBootKeyManagementSetter:
			temp.SecureBootKeyManagementSetter = p
			implementations = append(implementations, temp)
		default:
			e := fmt.Sprintf("not a SecureBootKeyManagementSetter implementation: %T", p)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(implementations) == 0 {
		return rebootRequired, metadata, multierror.Append(
			err,
			errors.Wrap(
				bmclibErrs.ErrProviderImplementation,
				("no SecureBootKeyManagementSetter implementations found"),
			),
		)
	}

	return setSecureBootKeyManagement(ctx, implementations, enable)
}
