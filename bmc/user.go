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

// userProviders is an internal struct used to correlate an implementation/provider with its name
type userProviders struct {
	name        string
	userCreator UserCreator
	userUpdater UserUpdater
	userDeleter UserDeleter
	userReader  UserReader
}

// CreateUser creates a user using the passed in implementation
// if a metadata is passed in, it will be updated to be the name of the provider that successfully executed
func CreateUser(ctx context.Context, user, pass, role string, u []userProviders, metadata ...*Metadata) (ok bool, err error) {
	var metadataLocal Metadata
	defer func() {
		if len(metadata) > 0 && metadata[0] != nil {
			*metadata[0] = metadataLocal
		}
	}()
Loop:
	for _, elem := range u {
		if elem.userCreator == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
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
			return ok, nil
		}
	}
	return ok, multierror.Append(err, errors.New("failed to create user"))
}

// CreateUserFromInterfaces pass through to library function
// if a metadata is passed in, it will be updated to be the name of the provider that successfully executed
func CreateUserFromInterfaces(ctx context.Context, user, pass, role string, generic []interface{}, metadata ...*Metadata) (ok bool, err error) {
	userCreators := make([]userProviders, 0)
	for _, elem := range generic {
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
		return ok, multierror.Append(err, errors.New("no UserCreator implementations found"))
	}
	return CreateUser(ctx, user, pass, role, userCreators, metadata...)
}

// UpdateUser updates a user's settings
// if a metadata is passed in, it will be updated to be the name of the provider that successfully executed
func UpdateUser(ctx context.Context, user, pass, role string, u []userProviders, metadata ...*Metadata) (ok bool, err error) {
	var metadataLocal Metadata
	defer func() {
		if len(metadata) > 0 && metadata[0] != nil {
			*metadata[0] = metadataLocal
		}
	}()
Loop:
	for _, elem := range u {
		if elem.userUpdater == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
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
			return ok, nil
		}
	}
	return ok, multierror.Append(err, errors.New("failed to update user"))
}

// UpdateUserFromInterfaces pass through to library function
// if a metadata is passed in, it will be updated to be the name of the provider that successfully executed
func UpdateUserFromInterfaces(ctx context.Context, user, pass, role string, generic []interface{}, metadata ...*Metadata) (ok bool, err error) {
	userUpdaters := make([]userProviders, 0)
	for _, elem := range generic {
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
		return ok, multierror.Append(err, errors.New("no UserUpdater implementations found"))
	}
	return UpdateUser(ctx, user, pass, role, userUpdaters, metadata...)
}

// DeleteUser deletes a user from a BMC
// if a metadata is passed in, it will be updated to be the name of the provider that successfully executed
func DeleteUser(ctx context.Context, user string, u []userProviders, metadata ...*Metadata) (ok bool, err error) {
	var metadataLocal Metadata
	defer func() {
		if len(metadata) > 0 && metadata[0] != nil {
			*metadata[0] = metadataLocal
		}
	}()
Loop:
	for _, elem := range u {
		if elem.userDeleter == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
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
			return ok, nil
		}
	}
	return ok, multierror.Append(err, errors.New("failed to delete user"))
}

// DeleteUserFromInterfaces pass through to library function
// if a metadata is passed in, it will be updated to be the name of the provider that successfully executed
func DeleteUserFromInterfaces(ctx context.Context, user string, generic []interface{}, metadata ...*Metadata) (ok bool, err error) {
	userDeleters := make([]userProviders, 0)
	for _, elem := range generic {
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
		return ok, multierror.Append(err, errors.New("no UserDeleter implementations found"))
	}
	return DeleteUser(ctx, user, userDeleters, metadata...)
}

// ReadUsers returns all users from a BMC
// if a metadata is passed in, it will be updated to be the name of the provider that successfully executed
func ReadUsers(ctx context.Context, u []userProviders, metadata ...*Metadata) (users []map[string]string, err error) {
	var metadataLocal Metadata
	defer func() {
		if len(metadata) > 0 && metadata[0] != nil {
			*metadata[0] = metadataLocal
		}
	}()
Loop:
	for _, elem := range u {
		if elem.userReader == nil {
			continue
		}
		select {
		case <-ctx.Done():
			err = multierror.Append(err, ctx.Err())
			break Loop
		default:
			metadataLocal.ProvidersAttempted = append(metadataLocal.ProvidersAttempted, elem.name)
			users, readErr := elem.userReader.UserRead(ctx)
			if readErr != nil {
				err = multierror.Append(err, readErr)
				continue
			}
			metadataLocal.SuccessfulProvider = elem.name
			return users, nil
		}
	}
	return users, multierror.Append(err, errors.New("failed to read users"))
}

// ReadUsersFromInterfaces pass through to library function
// if a metadata is passed in, it will be updated to be the name of the provider that successfully executed
func ReadUsersFromInterfaces(ctx context.Context, generic []interface{}, metadata ...*Metadata) (users []map[string]string, err error) {
	userReaders := make([]userProviders, 0)
	for _, elem := range generic {
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
		return users, multierror.Append(err, errors.New("no UserReader implementations found"))
	}
	return ReadUsers(ctx, userReaders, metadata...)
}
