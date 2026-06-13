package lenovo

import (
	"context"

	"github.com/bmc-toolbox/bmclib/v2/bmc"
	"github.com/pkg/errors"
)

// compile-time assertions that the provider implements the user interfaces.
var (
	_ bmc.UserReader  = (*Conn)(nil)
	_ bmc.UserCreator = (*Conn)(nil)
	_ bmc.UserUpdater = (*Conn)(nil)
	_ bmc.UserDeleter = (*Conn)(nil)
)

var (
	// ErrUserParams is returned when required user parameters are missing.
	ErrUserParams = errors.New("user, pass and role are required parameters")
	// ErrUserExists is returned when creating an account whose username is taken.
	ErrUserExists = errors.New("user account already exists")
	// ErrUserNotFound is returned when the named account does not exist.
	ErrUserNotFound = errors.New("user account not found")
	// ErrNoUserSlots is returned when no empty account slot is available.
	ErrNoUserSlots = errors.New("no user account slots available")
)

// UserRead returns the XCC accounts that have a username assigned.
//
// Each entry carries the account id, name, username, role and enabled state.
// Implements bmc.UserReader.
func (c *Conn) UserRead(ctx context.Context) (users []map[string]string, err error) {
	service, err := c.redfishwrapper.AccountService()
	if err != nil {
		return nil, err
	}

	accounts, err := service.Accounts()
	if err != nil {
		return nil, err
	}

	users = make([]map[string]string, 0)
	for _, account := range accounts {
		if account.UserName == "" {
			continue
		}
		users = append(users, map[string]string{
			"ID":       account.ID,
			"Name":     account.Name,
			"Username": account.UserName,
			"RoleID":   account.RoleID,
			"Enabled":  boolToString(account.Enabled),
		})
	}

	return users, nil
}

// UserCreate creates an XCC account.
//
// It first attempts the standard Redfish POST to the accounts collection. When
// the XCC rejects POST (Intel Purley-based systems), it falls back to PATCHing
// an empty account slot. Implements bmc.UserCreator.
func (c *Conn) UserCreate(ctx context.Context, user, pass, role string) (ok bool, err error) {
	if user == "" || pass == "" || role == "" {
		return false, ErrUserParams
	}

	service, err := c.redfishwrapper.AccountService()
	if err != nil {
		return false, err
	}

	accounts, err := service.Accounts()
	if err != nil {
		return false, err
	}

	for _, account := range accounts {
		if account.UserName == user {
			return false, errors.Wrap(ErrUserExists, user)
		}
	}

	// Primary path: Redfish POST to the accounts collection.
	if _, postErr := service.CreateAccount(user, pass, role); postErr == nil {
		return true, nil
	}

	// Fallback path (Intel Purley-based systems): PATCH an empty account slot.
	for _, account := range accounts {
		if account.Enabled || account.UserName != "" {
			continue
		}

		account.Enabled = true
		account.UserName = user
		account.Password = pass
		account.RoleID = role

		if err := account.Update(); err != nil {
			return false, err
		}

		return true, nil
	}

	return false, ErrNoUserSlots
}

// UserUpdate updates an account's password and/or role, matched by username.
//
// Implements bmc.UserUpdater.
func (c *Conn) UserUpdate(ctx context.Context, user, pass, role string) (ok bool, err error) {
	service, err := c.redfishwrapper.AccountService()
	if err != nil {
		return false, err
	}

	accounts, err := service.Accounts()
	if err != nil {
		return false, err
	}

	for _, account := range accounts {
		if account.UserName != user {
			continue
		}

		var changed bool
		if pass != "" {
			account.Password = pass
			changed = true
		}
		if role != "" {
			account.RoleID = role
			changed = true
		}

		if !changed {
			return true, nil
		}

		if err := account.Update(); err != nil {
			return false, err
		}

		return true, nil
	}

	return false, ErrUserNotFound
}

// UserDelete removes an account, matched by username, by clearing the slot
// (PATCH) — the XCC-compatible delete path. Implements bmc.UserDeleter.
func (c *Conn) UserDelete(ctx context.Context, user string) (ok bool, err error) {
	service, err := c.redfishwrapper.AccountService()
	if err != nil {
		return false, err
	}

	accounts, err := service.Accounts()
	if err != nil {
		return false, err
	}

	for _, account := range accounts {
		if account.UserName != user {
			continue
		}

		account.Enabled = false
		account.UserName = ""
		account.Password = ""

		if err := account.Update(); err != nil {
			return false, err
		}

		return true, nil
	}

	return false, ErrUserNotFound
}

// boolToString renders a bool as "true"/"false" for the UserRead string map.
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
