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

func (p *userTester) Name() string {
	return "test provider"
}

func TestUserCreate(t *testing.T) {
	testCases := map[string]struct {
		makeErrorOut bool
		makeNotOk    bool
		want         bool
		err          error
		ctxTimeout   time.Duration
	}{
		"success":               {want: true},
		"not ok return":         {want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("failed to create user"), errors.New("failed to create user")}}},
		"error":                 {makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("create user failed"), errors.New("failed to create user")}}},
		"error context timeout": {makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to create user")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
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
			result, err := CreateUser(ctx, user, pass, role, []userProviders{{"", &testImplementation, nil, nil, nil}})
			if err != nil {
				diff := cmp.Diff(err.Error(), tc.err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			} else {
				diff := cmp.Diff(result, expectedResult)
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestCreateUserFromInterfaces(t *testing.T) {
	testCases := map[string]struct {
		err               error
		badImplementation bool
		want              bool
		withMetadata      bool
	}{
		"success":                  {want: true},
		"success with metadata":    {want: true, withMetadata: true},
		"no implementations found": {badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a UserCreator implementation: *struct {}"), errors.New("no UserCreator implementations found")}}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
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
			var result bool
			var err error
			var metadata Metadata
			if tc.withMetadata {
				result, err = CreateUserFromInterfaces(context.Background(), user, pass, role, generic, &metadata)
			} else {
				result, err = CreateUserFromInterfaces(context.Background(), user, pass, role, generic)
			}
			if err != nil {
				if tc.err != nil {
					diff := cmp.Diff(err.Error(), tc.err.Error())
					if diff != "" {
						t.Fatal(diff)
					}
				} else {
					t.Fatal(err)
				}
			} else {
				diff := cmp.Diff(result, expectedResult)
				if diff != "" {
					t.Fatal(diff)
				}
			}
			if tc.withMetadata {
				if diff := cmp.Diff(metadata.SuccessfulProvider, "test provider"); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestUpdateUser(t *testing.T) {
	testCases := map[string]struct {
		makeErrorOut bool
		makeNotOk    bool
		want         bool
		err          error
		ctxTimeout   time.Duration
	}{
		"success":               {want: true},
		"not ok return":         {want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("failed to update user"), errors.New("failed to update user")}}},
		"error":                 {makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("update user failed"), errors.New("failed to update user")}}},
		"error context timeout": {makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to update user")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
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
			result, err := UpdateUser(ctx, user, pass, role, []userProviders{{"", nil, &testImplementation, nil, nil}})
			if err != nil {
				diff := cmp.Diff(err.Error(), tc.err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			} else {
				diff := cmp.Diff(result, expectedResult)
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestUpdateUserFromInterfaces(t *testing.T) {
	testCases := map[string]struct {
		err               error
		badImplementation bool
		want              bool
		withMetadata      bool
	}{
		"success":                  {want: true},
		"success with metadata":    {want: true, withMetadata: true},
		"no implementations found": {badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a UserUpdater implementation: *struct {}"), errors.New("no UserUpdater implementations found")}}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
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
			var result bool
			var err error
			var metadata Metadata
			if tc.withMetadata {
				result, err = UpdateUserFromInterfaces(context.Background(), user, pass, role, generic, &metadata)
			} else {
				result, err = UpdateUserFromInterfaces(context.Background(), user, pass, role, generic)
			}
			if err != nil {
				if tc.err != nil {
					diff := cmp.Diff(err.Error(), tc.err.Error())
					if diff != "" {
						t.Fatal(diff)
					}
				} else {
					t.Fatal(err)
				}
			} else {
				diff := cmp.Diff(result, expectedResult)
				if diff != "" {
					t.Fatal(diff)
				}
			}
			if tc.withMetadata {
				if diff := cmp.Diff(metadata.SuccessfulProvider, "test provider"); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	testCases := map[string]struct {
		makeErrorOut bool
		makeNotOk    bool
		want         bool
		err          error
		ctxTimeout   time.Duration
	}{
		"success":               {want: true},
		"not ok return":         {want: false, makeNotOk: true, err: &multierror.Error{Errors: []error{errors.New("failed to delete user"), errors.New("failed to delete user")}}},
		"error":                 {makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("delete user failed"), errors.New("failed to delete user")}}},
		"error context timeout": {makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to delete user")}}, ctxTimeout: time.Nanosecond * 1},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testImplementation := userTester{MakeErrorOut: tc.makeErrorOut, MakeNotOK: tc.makeNotOk}
			expectedResult := tc.want
			user := "ADMIN"
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			result, err := DeleteUser(ctx, user, []userProviders{{"", nil, nil, &testImplementation, nil}})
			if err != nil {
				diff := cmp.Diff(err.Error(), tc.err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			} else {
				diff := cmp.Diff(result, expectedResult)
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestDeleteUserFromInterfaces(t *testing.T) {
	testCases := map[string]struct {
		err               error
		badImplementation bool
		want              bool
		withMetadata      bool
	}{
		"success":                  {want: true},
		"success with metadata":    {want: true, withMetadata: true},
		"no implementations found": {badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a UserDeleter implementation: *struct {}"), errors.New("no UserDeleter implementations found")}}},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
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
			var result bool
			var err error
			var metadata Metadata
			if tc.withMetadata {
				result, err = DeleteUserFromInterfaces(context.Background(), user, generic, &metadata)
			} else {
				result, err = DeleteUserFromInterfaces(context.Background(), user, generic)
			}
			if err != nil {
				if tc.err != nil {
					diff := cmp.Diff(err.Error(), tc.err.Error())
					if diff != "" {
						t.Fatal(diff)
					}
				} else {
					t.Fatal(err)
				}
			} else {
				diff := cmp.Diff(result, expectedResult)
				if diff != "" {
					t.Fatal(diff)
				}
			}
			if tc.withMetadata {
				if diff := cmp.Diff(metadata.SuccessfulProvider, "test provider"); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestReadUsers(t *testing.T) {
	testCases := map[string]struct {
		makeErrorOut bool
		want         bool
		err          error
		ctxTimeout   time.Duration
	}{
		"success":               {want: true},
		"not ok return":         {want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("read users failed"), errors.New("failed to read users")}}},
		"error context timeout": {want: false, makeErrorOut: true, err: &multierror.Error{Errors: []error{errors.New("context deadline exceeded"), errors.New("failed to read users")}}, ctxTimeout: time.Nanosecond * 1},
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
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			testImplementation := userTester{MakeErrorOut: tc.makeErrorOut}
			expectedResult := users
			if tc.ctxTimeout == 0 {
				tc.ctxTimeout = time.Second * 3
			}
			ctx, cancel := context.WithTimeout(context.Background(), tc.ctxTimeout)
			defer cancel()
			result, err := ReadUsers(ctx, []userProviders{{"", nil, nil, nil, &testImplementation}})
			if err != nil {
				diff := cmp.Diff(err.Error(), tc.err.Error())
				if diff != "" {
					t.Fatal(diff)
				}
			} else {
				diff := cmp.Diff(result, expectedResult)
				if diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}

func TestReadUsersFromInterfaces(t *testing.T) {
	testCases := map[string]struct {
		err               error
		badImplementation bool
		want              bool
		withMetadata      bool
	}{
		"success":                  {want: true},
		"success with metadata":    {want: true, withMetadata: true},
		"no implementations found": {badImplementation: true, err: &multierror.Error{Errors: []error{errors.New("not a UserReader implementation: *struct {}"), errors.New("no UserReader implementations found")}}},
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
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			var generic []interface{}
			if tc.badImplementation {
				badImplementation := struct{}{}
				generic = []interface{}{&badImplementation}
			} else {
				testImplementation := userTester{}
				generic = []interface{}{&testImplementation}
			}
			expectedResult := users
			var result []map[string]string
			var err error
			var metadata Metadata
			if tc.withMetadata {
				result, err = ReadUsersFromInterfaces(context.Background(), generic, &metadata)
			} else {
				result, err = ReadUsersFromInterfaces(context.Background(), generic)
			}
			if err != nil {
				if tc.err != nil {
					diff := cmp.Diff(err.Error(), tc.err.Error())
					if diff != "" {
						t.Fatal(diff)
					}
				} else {
					t.Fatal(err)
				}
			} else {
				diff := cmp.Diff(result, expectedResult)
				if diff != "" {
					t.Fatal(diff)
				}
			}
			if tc.withMetadata {
				if diff := cmp.Diff(metadata.SuccessfulProvider, "test provider"); diff != "" {
					t.Fatal(diff)
				}
			}
		})
	}
}
