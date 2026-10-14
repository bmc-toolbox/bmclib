package redfishwrapper

import (
	"context"

	"github.com/stmcginnis/gofish/schemas"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

// GetSecureBoot returns whether UEFI Secure Boot is currently enabled for the system.
func (c *Client) GetSecureBoot(ctx context.Context) (enabled bool, err error) {
	sys, err := c.System()
	if err != nil {
		return false, err
	}

	if !c.compatibleOdataID(sys.ODataID, knownSystemsOdataIDs) {
		return false, bmclibErrs.ErrRedfishSystemOdataID
	}

	secureBoot, err := sys.SecureBoot()
	if err != nil {
		return false, err
	}
	if secureBoot == nil {
		return false, bmclibErrs.ErrRedfishVersionIncompatible
	}

	return secureBoot.SecureBootEnable, nil
}

// SetSecureBoot enables or disables UEFI Secure Boot for the system. The system
// must be in UEFI boot mode for this to take effect.
func (c *Client) SetSecureBoot(ctx context.Context, enable bool) (err error) {
	sys, err := c.System()
	if err != nil {
		return err
	}

	if !c.compatibleOdataID(sys.ODataID, knownSystemsOdataIDs) {
		return bmclibErrs.ErrRedfishSystemOdataID
	}

	secureBoot, err := sys.SecureBoot()
	if err != nil {
		return err
	}
	if secureBoot == nil {
		return bmclibErrs.ErrRedfishVersionIncompatible
	}

	secureBoot.SecureBootEnable = enable

	return secureBoot.Update()
}

// ResetSecureBootKeys resets the UEFI Secure Boot key databases.
func (c *Client) ResetSecureBootKeys(ctx context.Context, resetType bmc.ResetSecureBootKeysType) (err error) {
	sys, err := c.System()
	if err != nil {
		return err
	}

	if !c.compatibleOdataID(sys.ODataID, knownSystemsOdataIDs) {
		return bmclibErrs.ErrRedfishSystemOdataID
	}

	secureBoot, err := sys.SecureBoot()
	if err != nil {
		return err
	}
	if secureBoot == nil {
		return bmclibErrs.ErrRedfishVersionIncompatible
	}

	_, err = secureBoot.ResetKeys(schemas.ResetKeysType(resetType))

	return err
}

// secureBootDatabase returns the single UEFI Secure Boot key database matching
// database from the SecureBootDatabases collection reported by the BMC.
func (c *Client) secureBootDatabase(database bmc.SecureBootDatabase) (*schemas.SecureBootDatabase, error) {
	sys, err := c.System()
	if err != nil {
		return nil, err
	}

	if !c.compatibleOdataID(sys.ODataID, knownSystemsOdataIDs) {
		return nil, bmclibErrs.ErrRedfishSystemOdataID
	}

	secureBoot, err := sys.SecureBoot()
	if err != nil {
		return nil, err
	}
	if secureBoot == nil {
		return nil, bmclibErrs.ErrRedfishVersionIncompatible
	}

	databases, err := secureBoot.SecureBootDatabases()
	if err != nil {
		return nil, err
	}

	for _, db := range databases {
		if db.DatabaseID == string(database) {
			return db, nil
		}
	}

	return nil, bmclibErrs.ErrSecureBootDatabaseNotFound
}

// ResetSecureBootDatabaseKeys resets a single UEFI Secure Boot key database
// rather than the whole SecureBoot subsystem, leaving every other database -
// including PK - untouched. DeletePK is not a valid per-database reset type.
func (c *Client) ResetSecureBootDatabaseKeys(ctx context.Context, database bmc.SecureBootDatabase, resetType bmc.ResetSecureBootDatabaseKeysType) (err error) {
	db, err := c.secureBootDatabase(database)
	if err != nil {
		return err
	}

	_, err = db.ResetKeys(schemas.SecureBootDatabaseResetKeysType(resetType))

	return err
}

// ImportSecureBootCertificate enrolls a PEM-encoded certificate into a single
// UEFI Secure Boot key database without disturbing any certificate already
// present - it is additive, unlike ResetSecureBootDatabaseKeys.
func (c *Client) ImportSecureBootCertificate(ctx context.Context, database bmc.SecureBootDatabase, certificatePEM string) (err error) {
	db, err := c.secureBootDatabase(database)
	if err != nil {
		return err
	}

	_, err = db.AddCertificate(certificatePEM, schemas.PEMCertificateType, "")

	return err
}
