package bmc

import (
	"context"
	"errors"
	"fmt"

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
		if elem != nil {
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
	}
	return ok, multierror.Append(err, errors.New("failed to create user"))
}

// CreateUserFromInterfaces pass through to library function
func CreateUserFromInterfaces(ctx context.Context, user, pass, role string, generic []interface{}) (ok bool, err error) {
	var userCreators []UserCreator
	for _, elem := range generic {
		switch u := elem.(type) {
		case UserCreator:
			userCreators = append(userCreators, u)
		default:
		}
	}
	if len(userCreators) == 0 {
		return ok, errors.New("no UserCreator implementations found")
	}
	return CreateUser(ctx, user, pass, role, userCreators)
}

// UpdateUser updates a user's settings
func UpdateUser(ctx context.Context, user, pass, role string, u []UserUpdater) (ok bool, err error) {
	for _, elem := range u {
		if elem != nil {
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
	}
	return ok, multierror.Append(err, errors.New("failed to update user"))
}

// UpdateUserFromInterfaces pass through to library function
func UpdateUserFromInterfaces(ctx context.Context, user, pass, role string, generic []interface{}) (ok bool, err error) {
	var userUpdaters []UserUpdater
	for _, elem := range generic {
		switch u := elem.(type) {
		case UserUpdater:
			userUpdaters = append(userUpdaters, u)
		default:
		}
	}
	if len(userUpdaters) == 0 {
		return ok, errors.New("no UserUpdater implementations found")
	}
	return UpdateUser(ctx, user, pass, role, userUpdaters)
}

// DeleteUser deletes a user from a BMC
func DeleteUser(ctx context.Context, user string, u []UserDeleter) (ok bool, err error) {
	for _, elem := range u {
		if elem != nil {
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
	}
	return ok, multierror.Append(err, errors.New("failed to delete user"))
}

// DeleteUserFromInterfaces pass through to library function
func DeleteUserFromInterfaces(ctx context.Context, user string, generic []interface{}) (ok bool, err error) {
	var userDeleters []UserDeleter
	for _, elem := range generic {
		switch u := elem.(type) {
		case UserDeleter:
			userDeleters = append(userDeleters, u)
		default:
		}
	}
	if len(userDeleters) == 0 {
		return ok, errors.New("no UserDeleter implementations found")
	}
	return DeleteUser(ctx, user, userDeleters)
}

// ReadUsers returns all users from a BMC
func ReadUsers(ctx context.Context, u []UserReader) (users []map[string]string, err error) {
	for _, elem := range u {
		if elem != nil {
			users, err = elem.UserRead(ctx)
			if err != nil {
				err = multierror.Append(err, err)
				continue
			}
			return users, err
		}
	}
	return users, multierror.Append(err, errors.New("failed to read users"))
}

// ReadUsersFromInterfaces pass through to library function
func ReadUsersFromInterfaces(ctx context.Context, generic []interface{}) (users []map[string]string, err error) {
	var userReaders []UserReader
	for _, elem := range generic {
		switch u := elem.(type) {
		case UserReader:
			userReaders = append(userReaders, u)
		default:
			e := fmt.Sprintf("not a UserReader implementation: %T", u)
			err = multierror.Append(errors.New(e))
		}
	}
	if len(userReaders) == 0 {
		return users, multierror.Append(err, errors.New("no UserReader implementations found"))
	}
	return ReadUsers(ctx, userReaders)
}
