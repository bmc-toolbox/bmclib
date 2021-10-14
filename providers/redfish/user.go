package redfish

import (
	"context"

	"github.com/pkg/errors"
	"github.com/stmcginnis/gofish/redfish"
)

var (
	ErrNoUserSlotsAvailable = errors.New("no user account slots available")
	ErrUserNotPresent       = errors.New("given user not present")
	ErrUserPassParams       = errors.New("user and pass parameters required")
	ErrUserExists           = errors.New("user exists")
	ErrInvalidUserRole      = errors.New("invalid user role")
	ValidRoles              = []string{"Administrator", "Operator", "ReadOnly", "None"}
)

// UserRead returns a list of enabled user accounts
func (c *Conn) UserRead(ctx context.Context) (users []map[string]string, err error) {
	service, err := c.conn.Service.AccountService()
	if err != nil {
		return nil, err
	}

	accounts, err := service.Accounts()
	if err != nil {
		return nil, err
	}

	users = make([]map[string]string, 0)

	for _, account := range accounts {
		if account.Enabled {
			user := map[string]string{
				"ID":       account.ID,
				"Name":     account.Name,
				"Username": account.UserName,
				"RoleID":   account.RoleID,
			}
			users = append(users, user)
		}
	}

	return users, nil
}

// UserUpdate updates a user password and role
func (c *Conn) UserUpdate(ctx context.Context, user, pass, role string) (ok bool, err error) {
	service, err := c.conn.Service.AccountService()
	if err != nil {
		return false, err
	}

	accounts, err := service.Accounts()
	if err != nil {
		return false, err
	}

	for _, account := range accounts {
		if account.UserName == user {
			var change bool
			if pass != "" {
				account.Password = pass
				change = true
			}
			if role != "" {
				account.RoleID = role
				change = true
			}

			if change {
				err := account.Update()
				if err != nil {
					return false, err
				}
				return true, nil
			}
		}
	}

	return ok, ErrUserNotPresent
}

// UserCreate adds a new user account
func (c *Conn) UserCreate(ctx context.Context, user, pass, role string) (ok bool, err error) {
	if !StringInSlice(role, ValidRoles) {
		return false, ErrInvalidUserRole
	}

	if user == "" || pass == "" {
		return false, ErrUserPassParams
	}

	service, err := c.conn.Service.AccountService()
	if err != nil {
		return false, err
	}

	// fetch current list of accounts
	accounts, err := service.Accounts()
	if err != nil {
		return false, err
	}

	// identify account slot not in use
	for _, account := range accounts {
		// Dell iDracs don't want us to create accounts in these slots
		if StringInSlice(account.ID, []string{"1"}) {
			continue
		}

		account := account
		if account.UserName == user {
			return false, errors.Wrap(ErrUserExists, user)
		}

		if !account.Enabled && account.UserName == "" {
			account.Enabled = true
			account.UserName = user
			account.Password = pass
			account.RoleID = role
			account.AccountTypes = []redfish.AccountTypes{"Redfish", "OEM"}

			err := account.Update()
			if err != nil {
				return false, err
			}
			return true, nil
		}
	}

	return false, ErrNoUserSlotsAvailable
}

func StringInSlice(str string, sl []string) bool {
	for _, s := range sl {
		if str == s {
			return true
		}
	}
	return false
}
