package bmc

import (
	"context"
	"errors"
	"testing"

	"github.com/bmc-toolbox/bmclib/logging"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
)

type userTester struct {
	MakeNotOK    bool
	MakeErrorOut bool
}

func (p *userTester) UserCreate(ctx context.Context, log logr.Logger, user, pass, role string) (ok bool, err error) {
	if p.MakeErrorOut {
		return ok, errors.New("create user failed")
	}
	if p.MakeNotOK {
		return false, nil
	}
	return true, nil
}

func (p *userTester) UserUpdate(ctx context.Context, log logr.Logger, user, pass, role string) (ok bool, err error) {
	if p.MakeErrorOut {
		return ok, errors.New("update user failed")
	}
	if p.MakeNotOK {
		return false, nil
	}
	return true, nil
}

func (p *userTester) UserDelete(ctx context.Context, log logr.Logger, user string) (ok bool, err error) {
	if p.MakeErrorOut {
		return ok, errors.New("delete user failed")
	}
	if p.MakeNotOK {
		return false, nil
	}
	return true, nil
}

func (p *userTester) UserRead(ctx context.Context, log logr.Logger) (users []map[string]string, err error) {
	if p.MakeErrorOut {
		return users, errors.New("read users failed")
	}

	users = []map[string]string{
		{
			"Auth":   "true",
			"Callin": "true",
			"ID":     "2",
			"Link":   "false",
			"Name":   "ADMIN",
		},
	}
	return users, nil
}

func TestUserCreate(t *testing.T) {
	testCases := []struct {
		name         string
		makeErrorOut bool
		makeNotOk    bool
		want         bool
		err          error
	}{
		{name: "success", want: true},
		{name: "not ok return", want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("failed to create user"), errors.New("failed to create user")}}},
		{name: "error", makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("create user failed"), errors.New("failed to create user")}}},
	}
	log := logging.DefaultLogger()
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testImplementation := userTester{MakeErrorOut: tc.makeErrorOut, MakeNotOK: tc.makeNotOk}
			expectedResult := tc.want
			user := "ADMIN"
			pass := "ADMIN"
			role := "admin"
			result, err := CreateUser(context.Background(), log, user, pass, role, []UserCreator{&testImplementation})
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}

			} else {
				diff := cmp.Diff(expectedResult, result)
				if diff != "" {
					t.Fatal(diff)
				}
			}

		})
	}
}

func TestCreateUserFromInterfaces(t *testing.T) {
	testCases := []struct {
		name              string
		err               error
		badImplementation bool
		want              bool
	}{
		{name: "success", want: true},
		{name: "no implementations found", badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a UserCreator implementation: *struct {}"), errors.New("no UserCreator implementations found")}}},
	}
	log := logging.DefaultLogger()
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := userTester{}
				generic = []interface{}{&testImplementation}
			}
			expectedResult := tc.want
			user := "ADMIN"
			pass := "ADMIN"
			role := "admin"
			result, err := CreateUserFromInterfaces(context.Background(), log, user, pass, role, generic)
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}

			} else {
				diff := cmp.Diff(expectedResult, result)
				if diff != "" {
					t.Fatal(diff)
				}
			}

		})
	}
}

func TestUpdateUser(t *testing.T) {
	testCases := []struct {
		name         string
		makeErrorOut bool
		makeNotOk    bool
		want         bool
		err          error
	}{
		{name: "success", want: true},
		{name: "not ok return", want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("failed to update user"), errors.New("failed to update user")}}},
		{name: "error", makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("update user failed"), errors.New("failed to update user")}}},
	}
	log := logging.DefaultLogger()
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testImplementation := userTester{MakeErrorOut: tc.makeErrorOut, MakeNotOK: tc.makeNotOk}
			expectedResult := tc.want
			user := "ADMIN"
			pass := "ADMIN"
			role := "admin"
			result, err := UpdateUser(context.Background(), log, user, pass, role, []UserUpdater{&testImplementation})
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}

			} else {
				diff := cmp.Diff(expectedResult, result)
				if diff != "" {
					t.Fatal(diff)
				}
			}

		})
	}
}

func TestUpdateUserFromInterfaces(t *testing.T) {
	testCases := []struct {
		name              string
		err               error
		badImplementation bool
		want              bool
	}{
		{name: "success", want: true},
		{name: "no implementations found", badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a UserUpdater implementation: *struct {}"), errors.New("no UserUpdater implementations found")}}},
	}
	log := logging.DefaultLogger()
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := userTester{}
				generic = []interface{}{&testImplementation}
			}
			expectedResult := tc.want
			user := "ADMIN"
			pass := "ADMIN"
			role := "admin"
			result, err := UpdateUserFromInterfaces(context.Background(), log, user, pass, role, generic)
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}

			} else {
				diff := cmp.Diff(expectedResult, result)
				if diff != "" {
					t.Fatal(diff)
				}
			}

		})
	}
}

func TestDeleteUser(t *testing.T) {
	testCases := []struct {
		name         string
		makeErrorOut bool
		makeNotOk    bool
		want         bool
		err          error
	}{
		{name: "success", want: true},
		{name: "not ok return", want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("failed to delete user"), errors.New("failed to delete user")}}},
		{name: "error", makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("delete user failed"), errors.New("failed to delete user")}}},
	}
	log := logging.DefaultLogger()
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testImplementation := userTester{MakeErrorOut: tc.makeErrorOut, MakeNotOK: tc.makeNotOk}
			expectedResult := tc.want
			user := "ADMIN"
			result, err := DeleteUser(context.Background(), log, user, []UserDeleter{&testImplementation})
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}

			} else {
				diff := cmp.Diff(expectedResult, result)
				if diff != "" {
					t.Fatal(diff)
				}
			}

		})
	}
}

func TestDeleteUserFromInterfaces(t *testing.T) {
	testCases := []struct {
		name              string
		err               error
		badImplementation bool
		want              bool
	}{
		{name: "success", want: true},
		{name: "no implementations found", badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a UserDeleter implementation: *struct {}"), errors.New("no UserDeleter implementations found")}}},
	}
	log := logging.DefaultLogger()
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := userTester{}
				generic = []interface{}{&testImplementation}
			}
			expectedResult := tc.want
			user := "ADMIN"
			result, err := DeleteUserFromInterfaces(context.Background(), log, user, generic)
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}

			} else {
				diff := cmp.Diff(expectedResult, result)
				if diff != "" {
					t.Fatal(diff)
				}
			}

		})
	}
}

func TestReadUsers(t *testing.T) {
	testCases := []struct {
		name         string
		makeErrorOut bool
		want         bool
		err          error
	}{
		{name: "success", want: true},
		{name: "not ok return", want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("read users failed"), errors.New("failed to read users")}}},
	}
	users := []map[string]string{
		{
			"Auth":   "true",
			"Callin": "true",
			"ID":     "2",
			"Link":   "false",
			"Name":   "ADMIN",
		},
	}
	log := logging.DefaultLogger()
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testImplementation := userTester{MakeErrorOut: tc.makeErrorOut}
			expectedResult := users
			result, err := ReadUsers(context.Background(), log, []UserReader{&testImplementation})
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}

			} else {
				diff := cmp.Diff(expectedResult, result)
				if diff != "" {
					t.Fatal(diff)
				}
			}

		})
	}
}

func TestReadUsersFromInterfaces(t *testing.T) {
	testCases := []struct {
		name              string
		err               error
		badImplementation bool
		want              bool
	}{
		{name: "success", want: true},
		{name: "no implementations found", badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a UserReader implementation: *struct {}"), errors.New("no UserReader implementations found")}}},
	}
	users := []map[string]string{
		{
			"Auth":   "true",
			"Callin": "true",
			"ID":     "2",
			"Link":   "false",
			"Name":   "ADMIN",
		},
	}
	log := logging.DefaultLogger()
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := userTester{}
				generic = []interface{}{&testImplementation}
			}
			expectedResult := users
			result, err := ReadUsersFromInterfaces(context.Background(), log, generic)
			if err != nil {
				diff := cmp.Diff(tc.err.Error(), err.Error())
				if diff != "" {
					t.Fatal(diff)
				}

			} else {
				diff := cmp.Diff(expectedResult, result)
				if diff != "" {
					t.Fatal(diff)
				}
			}

		})
	}
}
