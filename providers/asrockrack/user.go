package asrockrack

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
	"github.com/metal-toolbox/bmclib/internal"
)

var (
	// TODO: standardize these across Redfish, IPMI, Vendor GUI
	validRoles = []string{"Administrator", "Operator", "User"}
)

// UserAccount is a ASRR BMC user account struct
type UserAccount struct {
	ID                           int    `json:"id"`
	Name                         string `json:"name"`
	Access                       int    `json:"access"`
	AccessByChannel              string `json:"accessByChannel,omitempty"`
	Kvm                          int    `json:"kvm"`
	Vmedia                       int    `json:"vmedia"`
	NetworkPrivilege             string `json:"network_privilege"`
	FixedUserCount               int    `json:"fixed_user_count"`
	OEMProprietaryLevelPrivilege int    `json:"OEMProprietary_level_Privilege"`
	Privilege                    string `json:"privilege,omitempty"`
	PrivilegeByChannel           string `json:"privilegeByChannel,omitempty"`
	PrivilegeLimitSerial         string `json:"privilege_limit_serial"`
	SSHKey                       string `json:"ssh_key"`
	CreationTime                 int    `json:"creation_time"`
	Changepassword               int    `json:"changepassword"`
	UserOperation                int    `json:"UserOperation"`
	Password                     string `json:"password"`
	ConfirmPassword              string `json:"confirm_password"`
	PasswordSize                 string `json:"password_size"`
	PrevSNMP                     int    `json:"prev_snmp"`
	SNMP                         int    `json:"snmp"`
	SNMPAccess                   string `json:"snmp_access"`
	SNMPAuthenticationProtocol   string `json:"snmp_authentication_protocol"`
	EmailFormat                  string `json:"email_format"`
	EmailID                      string `json:"email_id"`
}

// UserRead returns a list of enabled user accounts
func (a *ASRockRack) UserRead(ctx context.Context) (users []map[string]string, err error) {
	err = a.Open(ctx)
	if err != nil {
		return nil, err
	}

	accounts, err := a.listUsers(ctx)
	if err != nil {
		return nil, errors.Wrap(bmclibErrs.ErrRetrievingUserAccounts, err.Error())
	}

	users = make([]map[string]string, 0)
	for _, account := range accounts {
		if account.Access == 1 {
			user := map[string]string{
				"ID":     fmt.Sprintf("%d", account.ID),
				"Name":   account.Name,
				"RoleID": account.NetworkPrivilege,
			}
			users = append(users, user)
		}
	}

	return users, nil
}

// UserCreate adds a new user account
func (a *ASRockRack) UserCreate(ctx context.Context, user, pass, role string) (ok bool, err error) {
	if !internal.StringInSlice(role, validRoles) {
		return false, bmclibErrs.ErrInvalidUserRole
	}

	if user == "" || pass == "" || role == "" {
		return false, bmclibErrs.ErrUserParamsRequired
	}

	// fetch current list of accounts
	accounts, err := a.listUsers(ctx)
	if err != nil {
		return false, errors.Wrap(bmclibErrs.ErrRetrievingUserAccounts, err.Error())
	}

	// identify account slot not in use
	for _, account := range accounts {
		// ASRR BMCs have a reserved slot 1 for a disabled Anonymous, no idea why.
		if account.ID == 1 {
			continue
		}

		account := account
		if account.Name == user {
			return false, errors.Wrap(bmclibErrs.ErrUserAccountExists, user)
		}

		if account.Access == 0 && account.Name == "" {
			newAccount := newUserAccount(account.ID, user, pass, strings.ToLower(role))
			err := a.createUpdateUser(ctx, newAccount)
			if err != nil {
				return false, err
			}
			return true, nil
		}
	}

	return false, bmclibErrs.ErrNoUserSlotsAvailable
}

//

// UserUpdate updates a user password and role
func (a *ASRockRack) UserUpdate(ctx context.Context, user, pass, role string) (ok bool, err error) {
	if !internal.StringInSlice(role, validRoles) {
		return false, bmclibErrs.ErrInvalidUserRole
	}

	if user == "" || pass == "" || role == "" {
		return false, bmclibErrs.ErrUserParamsRequired
	}

	accounts, err := a.listUsers(ctx)
	if err != nil {
		return false, errors.Wrap(bmclibErrs.ErrRetrievingUserAccounts, err.Error())
	}

	role = strings.ToLower(role)

	// identify account slot not in use
	for _, account := range accounts {
		account := account
		if account.Name == user {
			user := newUserAccount(account.ID, user, pass, role)

			user.AccessByChannel = account.AccessByChannel
			user.PrivilegeByChannel = account.PrivilegeByChannel
			user.Privilege = role

			if role == "administrator" {
				user.PrivilegeLimitSerial = "none"
				user.UserOperation = 1
				user.CreationTime = 6000 // doesn't mean anything.
			}

			err := a.createUpdateUser(ctx, user)
			if err != nil {
				return false, errors.Wrap(bmclibErrs.ErrUserAccountUpdate, err.Error())
			}

			return true, nil
		}
	}

	return ok, errors.Wrap(bmclibErrs.ErrUserAccountNotFound, user)
}

// newUserAccount returns a user account object populated with the given attributes and certain defaults
//
// note: the role parameter must be validated before being passed to this constructor
func newUserAccount(id int, user, pass, role string) *UserAccount {
	return &UserAccount{
		ID:                           id,
		Name:                         user,
		Access:                       1, // Access enabled
		Kvm:                          1,
		Vmedia:                       1,
		NetworkPrivilege:             role,
		FixedUserCount:               2, // No idea what this is about
		OEMProprietaryLevelPrivilege: 1,
		PrivilegeLimitSerial:         role,
		SSHKey:                       "Not Available",
		CreationTime:                 0,
		Changepassword:               1,
		UserOperation:                0,
		Password:                     pass,
		ConfirmPassword:              pass,
		PasswordSize:                 "bytes_16", // bytes_20 for larger passwords
		EmailFormat:                  "AMI-Format",
	}
}
