package bmc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/go-multierror"
)

type userTester struct {
	MakeNotOK    bool
	MakeErrorOut bool
}

func (p *userTester) UserCreate(ctx context.Context, user, pass, role string) (ok bool, err error) {
	if p.MakeErrorOut {
		return ok, errors.New("create user failed")
	}
	if p.MakeNotOK {
		return false, nil
	}
	return true, nil
}

func (p *userTester) UserUpdate(ctx context.Context, user, pass, role string) (ok bool, err error) {
	if p.MakeErrorOut {
		return ok, errors.New("update user failed")
	}
	if p.MakeNotOK {
		return false, nil
	}
	return true, nil
}

func (p *userTester) UserDelete(ctx context.Context, user string) (ok bool, err error) {
	if p.MakeErrorOut {
		return ok, errors.New("delete user failed")
	}
	if p.MakeNotOK {
		return false, nil
	}
	return true, nil
}

func (p *userTester) UserRead(ctx context.Context) (users []map[string]string, err error) {
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
		ctxTimeout   time.Duration
	}{
		{name: "success", want: true},
		{name: "not ok return", want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("failed to create user"), errors.New("failed to create user")}}},
		{name: "error", makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("create user failed"), errors.New("failed to create user")}}},
		{name: "error context timeout", makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to create user")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testImplementation := userTester{MakeErrorOut: tc.makeErrorOut, MakeNotOK: tc.makeNotOk}
			expectedResult := tc.want
			user := "ADMIN"
			pass := "ADMIN"
			role := "admin"
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			result, err := CreateUser(ctx, user, pass, role, []UserCreator{&testImplementation})
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
			result, err := CreateUserFromInterfaces(context.Background(), user, pass, role, generic)
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
		ctxTimeout   time.Duration
	}{
		{name: "success", want: true},
		{name: "not ok return", want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("failed to update user"), errors.New("failed to update user")}}},
		{name: "error", makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("update user failed"), errors.New("failed to update user")}}},
		{name: "error context timeout", makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to update user")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testImplementation := userTester{MakeErrorOut: tc.makeErrorOut, MakeNotOK: tc.makeNotOk}
			expectedResult := tc.want
			user := "ADMIN"
			pass := "ADMIN"
			role := "admin"
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			result, err := UpdateUser(ctx, user, pass, role, []UserUpdater{&testImplementation})
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
			result, err := UpdateUserFromInterfaces(context.Background(), user, pass, role, generic)
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
		ctxTimeout   time.Duration
	}{
		{name: "success", want: true},
		{name: "not ok return", want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("failed to delete user"), errors.New("failed to delete user")}}},
		{name: "error", makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("delete user failed"), errors.New("failed to delete user")}}},
		{name: "error context timeout", makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to delete user")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testImplementation := userTester{MakeErrorOut: tc.makeErrorOut, MakeNotOK: tc.makeNotOk}
			expectedResult := tc.want
			user := "ADMIN"
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			result, err := DeleteUser(ctx, user, []UserDeleter{&testImplementation})
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
			result, err := DeleteUserFromInterfaces(context.Background(), user, generic)
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
		ctxTimeout   time.Duration
	}{
		{name: "success", want: true},
		{name: "not ok return", want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("read users failed"), errors.New("failed to read users")}}},
		{name: "not ok return", want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to read users")}}, ctxTimeout: time.Nanosecond * 1},
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
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			testImplementation := userTester{MakeErrorOut: tc.makeErrorOut}
			expectedResult := users
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			result, err := ReadUsers(ctx, []UserReader{&testImplementation})
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
			result, err := ReadUsersFromInterfaces(context.Background(), generic)
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
