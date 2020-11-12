package bmc

import (
	"context"
	"errors"

	"github.com/hashicorp/go-multierror"
)

// UserCreator creates a user on a BMC
type UserCreator interface {
	UserCreate(ctx context.Context, user, pass, role string) (ok bool, err error)
}

// UserUpdater updates a user on a BMC
type UserUpdater interface {
	UserUpdate(ctx context.Context, user, pass, role string) (ok bool, err error)
}

// UserDeleter deletes a user on a BMC
type UserDeleter interface {
	UserDelete(ctx context.Context, user string) (ok bool, err error)
}

// UserReader lists all users on a BMC
type UserReader interface {
	UserRead(ctx context.Context) (users []map[string]string, err error)
}

// CreateUser creates a user using the passed in implementation
func CreateUser(ctx context.Context, user, pass, role string, u []UserCreator) (ok bool, err error) {
	for _, elem := range u {
		ok, err = elem.UserCreate(ctx, user, pass, role)
		if err != nil {
			err = multierror.Append(err, err)
			continue
		}
		if !ok {
			err = multierror.Append(err, err)
			continue
		}
		return ok, err
	}
	return ok, multierror.Append(err, errors.New("failed to create user"))
}

// UpdateUser updates a user's settings
func UpdateUser(ctx context.Context, user, pass, role string, u []UserUpdater) (ok bool, err error) {
	for _, elem := range u {
		ok, err = elem.UserUpdate(ctx, user, pass, role)
		if err != nil {
			err = multierror.Append(err, err)
			continue
		}
		if !ok {
			err = multierror.Append(err, err)
			continue
		}
		return ok, err
	}
	return ok, multierror.Append(err, errors.New("failed to update user"))
}

// DeleteUser deletes a user from a BMC
func DeleteUser(ctx context.Context, user string, u []UserDeleter) (ok bool, err error) {
	for _, elem := range u {
		ok, err = elem.UserDelete(ctx, user)
		if err != nil {
			err = multierror.Append(err, err)
			continue
		}
		if !ok {
			err = multierror.Append(err, err)
			continue
		}
		return ok, err
	}
	return ok, multierror.Append(err, errors.New("failed to delete user"))
}

// ReadUsers returns all users from a BMC
func ReadUsers(ctx context.Context, u []UserReader) (users []map[string]string, err error) {
	for _, elem := range u {
		users, err = elem.UserRead(ctx)
		if err != nil {
			err = multierror.Append(err, err)
			continue
		}
		return users, err
	}
	return users, multierror.Append(err, errors.New("failed to delete user"))
}
