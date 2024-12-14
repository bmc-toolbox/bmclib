package bmc

import (
	"context"
	"errors"
	"fmt"
	"time"

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

// userProviders is an internal struct used to correlate an implementation/provider with its name
type userProviders struct {
	name        string
	userCreator UserCreator
	userUpdater UserUpdater
	userDeleter UserDeleter
	userReader  UserReader
}

// createUser creates a user using the passed in implementation
func createUser(ctx context.Context, timeout time.Duration, user, pass, role string, u []userProviders) (ok bool, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range u {
		if elem.userCreator == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return false, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			ok, createErr := elem.userCreator.UserCreate(ctx, user, pass, role)
			if createErr != nil {
				err = multierror.Append(err, createErr)
				continue
			}
			if !ok {
				err = multierror.Append(err, errors.New("failed to create user"))
				continue
			}
			metadataLocal.SuccessfulProvider = elem.name
			return ok, metadataLocal, nil
		}
	}
	return ok, metadataLocal, multierror.Append(err, errors.New("failed to create user"))
}

// CreateUsersFromInterfaces identifies implementations of the UserCreator interface and passes them to the createUser() wrapper method.
func CreateUserFromInterfaces(ctx context.Context, timeout time.Duration, user, pass, role string, generic []interface{}) (ok bool, metadata Metadata, err error) {
	userCreators := make([]userProviders, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := userProviders{name: getProviderName(elem)}
		switch u := elem.(type) {
		case UserCreator:
			temp.userCreator = u
			userCreators = append(userCreators, temp)
		default:
			e := fmt.Sprintf("not a UserCreator implementation: %T", u)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(userCreators) == 0 {
		return ok, metadata, multierror.Append(err, errors.New("no UserCreator implementations found"))
	}
	return createUser(ctx, timeout, user, pass, role, userCreators)
}

// updateUser updates a user's settings
func updateUser(ctx context.Context, timeout time.Duration, user, pass, role string, u []userProviders) (ok bool, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range u {
		if elem.userUpdater == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return false, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			ok, UpdateErr := elem.userUpdater.UserUpdate(ctx, user, pass, role)
			if UpdateErr != nil {
				err = multierror.Append(err, UpdateErr)
				continue
			}
			if !ok {
				err = multierror.Append(err, errors.New("failed to update user"))
				continue
			}
			metadataLocal.SuccessfulProvider = elem.name
			return ok, metadataLocal, nil
		}
	}
	return ok, metadataLocal, multierror.Append(err, errors.New("failed to update user"))
}

// UpdateUsersFromInterfaces identifies implementations of the UserUpdater interface and passes them to the updateUser() wrapper method.
func UpdateUserFromInterfaces(ctx context.Context, timeout time.Duration, user, pass, role string, generic []interface{}) (ok bool, metadata Metadata, err error) {
	userUpdaters := make([]userProviders, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := userProviders{name: getProviderName(elem)}
		switch u := elem.(type) {
		case UserUpdater:
			temp.userUpdater = u
			userUpdaters = append(userUpdaters, temp)
		default:
			e := fmt.Sprintf("not a UserUpdater implementation: %T", u)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(userUpdaters) == 0 {
		return ok, metadata, multierror.Append(err, errors.New("no UserUpdater implementations found"))
	}
	return updateUser(ctx, timeout, user, pass, role, userUpdaters)
}

// deleteUser deletes a user from a BMC
func deleteUser(ctx context.Context, timeout time.Duration, user string, u []userProviders) (ok bool, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range u {
		if elem.userDeleter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return false, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			ok, deleteErr := elem.userDeleter.UserDelete(ctx, user)
			if deleteErr != nil {
				err = multierror.Append(err, deleteErr)
				continue
			}
			if !ok {
				err = multierror.Append(err, errors.New("failed to delete user"))
				continue
			}
			metadataLocal.SuccessfulProvider = elem.name
			return ok, metadataLocal, nil
		}
	}
	return ok, metadataLocal, multierror.Append(err, errors.New("failed to delete user"))
}

// DeleteUsersFromInterfaces identifies implementations of the UserDeleter interface and passes them to the deleteUser() wrapper method.
func DeleteUserFromInterfaces(ctx context.Context, timeout time.Duration, user string, generic []interface{}) (ok bool, metadata Metadata, err error) {
	userDeleters := make([]userProviders, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := userProviders{name: getProviderName(elem)}
		switch u := elem.(type) {
		case UserDeleter:
			temp.userDeleter = u
			userDeleters = append(userDeleters, temp)
		default:
			e := fmt.Sprintf("not a UserDeleter implementation: %T", u)
			err = multierror.Append(err, errors.New(e))
		}
	}
	if len(userDeleters) == 0 {
		return ok, metadata, multierror.Append(err, errors.New("no UserDeleter implementations found"))
	}
	return deleteUser(ctx, timeout, user, userDeleters)
}

// readUsers returns all users from a BMC
func readUsers(ctx context.Context, timeout time.Duration, u []userProviders) (users []map[string]string, metadata Metadata, err error) {
	var metadataLocal Metadata

	for _, elem := range u {
		if elem.userReader == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())

			return users, metadata, err
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()
			users, readErr := elem.userReader.UserRead(ctx)
			if readErr != nil {
				err = multierror.Append(err, readErr)
				continue
			}
			metadataLocal.SuccessfulProvider = elem.name
			return users, metadataLocal, nil
		}
	}
	return users, metadataLocal, multierror.Append(err, errors.New("failed to read users"))
}

// ReadUsersFromInterfaces identifies implementations of the UserReader interface and passes them to the readUsers() wrapper method.
func ReadUsersFromInterfaces(ctx context.Context, timeout time.Duration, generic []interface{}) (users []map[string]string, metadata Metadata, err error) {
	userReaders := make([]userProviders, 0)
	for _, elem := range generic {
		if elem == nil {
			continue
		}
		temp := userProviders{name: getProviderName(elem)}
		switch u := elem.(type) {
		case UserReader:
			temp.userReader = u
			userReaders = append(userReaders, temp)
		default:
			e := fmt.Sprintf("not a UserReader implementation: %T", u)
			err = multierror.Append(errors.New(e))
		}
	}
	if len(userReaders) == 0 {
		return users, metadata, multierror.Append(err, errors.New("no UserReader implementations found"))
	}
	return readUsers(ctx, timeout, userReaders)
}
