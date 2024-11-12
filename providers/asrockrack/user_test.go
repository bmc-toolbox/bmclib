package asrockrack

import (
	"context"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
)

// NOTE: user accounts are defined in mock_test.go as JSON payload in the userPayload var

type testCase struct {
	user  string
	pass  string
	role  string
	ok    bool
	err   error
	tName string
}

var (
	// common set of test cases
	testCases = []testCase{

		{
			"foo",
			"baz",
			"",
			false,
			bmclibErrs.ErrInvalidUserRole,
			"role not defined",
		},
		{
			"foo",
			"",
			"Administrator",
			false,
			bmclibErrs.ErrUserParamsRequired,
			"param not defined",
		},
	}
)

func Test_UserRead(t *testing.T) {
	expected := []map[string]string{
		{
			"RoleID": "administrator",
			"ID":     "2",
			"Name":   "admin",
		},
		{
			"ID":     "3",
			"Name":   "foo",
			"RoleID": "administrator",
		},
	}

	err := aClient.httpsLogin(context.TODO())
	if err != nil {
		t.Errorf("login: %s", err.Error())
	}

	users, err := aClient.UserRead(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, expected, users)

	for _, tt := range testCases {
		ok, err := aClient.UserCreate(context.TODO(), tt.user, tt.pass, tt.role)
		assert.Equal(t, errors.Is(err, tt.err), true, tt.tName)
		assert.Equal(t, tt.ok, ok, tt.tName)
	}

	// test account retrieval failure error
	os.Setenv("TEST_FAIL_QUERY", "womp womp")
	defer os.Unsetenv("TEST_FAIL_QUERY")

	_, err = aClient.UserRead(context.TODO())
	assert.Equal(t, errors.Is(err, bmclibErrs.ErrRetrievingUserAccounts), true)
}

func Test_UserCreate(t *testing.T) {

	tests := testCases
	tests = append(tests,
		[]testCase{{
			"root",
			"calvin",
			"Administrator",
			true,
			nil,
			"user account is created",
		},
			{
				"admin",
				"foo",
				"Administrator",
				false,
				bmclibErrs.ErrUserAccountExists,
				"account already exists",
			},
		}...,
	)

	err := aClient.httpsLogin(context.TODO())
	if err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		ok, err := aClient.UserCreate(context.TODO(), tt.user, tt.pass, tt.role)
		assert.Equal(t, errors.Is(err, tt.err), true, tt.tName)
		assert.Equal(t, tt.ok, ok, tt.tName)
	}
}

func Test_UserUpdate(t *testing.T) {
	tests := testCases
	tests = append(tests,
		[]testCase{
			{
				"admin",
				"calvin",
				"Administrator",
				true,
				nil,
				"user account is updated",
			},
			{
				"badmin",
				"calvin",
				"Administrator",
				false,
				bmclibErrs.ErrUserAccountNotFound,
				"user account not present",
			},
		}...,
	)

	err := aClient.httpsLogin(context.TODO())
	if err != nil {
		t.Error(err)
	}

	for _, tt := range tests {
		ok, err := aClient.UserUpdate(context.TODO(), tt.user, tt.pass, tt.role)
		assert.Equal(t, errors.Is(err, tt.err), true, tt.tName)
		assert.Equal(t, tt.ok, ok, tt.tName)
	}
}

func Test_createUser(t *testing.T) {
	err := aClient.httpsLogin(context.TODO())
	if err != nil {
		t.Errorf("login: %s", err.Error())
	}

	account := &UserAccount{
		ID:                           3,
		Name:                         "foobar",
		Access:                       1,
		Kvm:                          1,
		Vmedia:                       1,
		NetworkPrivilege:             "administrator",
		FixedUserCount:               2,
		OEMProprietaryLevelPrivilege: 1,
		PrivilegeLimitSerial:         "none",
		SSHKey:                       "Not Available",
		CreationTime:                 4802,
		Changepassword:               0,
		UserOperation:                0,
		Password:                     "",
		ConfirmPassword:              "",
		PasswordSize:                 "",
	}

	err = aClient.createUpdateUser(context.TODO(), account)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "/api/settings/users/3", httpRequestTestVar.URL.String())
	assert.Equal(t, http.MethodPut, httpRequestTestVar.Method)
	var contentType string
	for k, v := range httpRequestTestVar.Header {
		if k == "Content-Type" {
			contentType = v[0]
		}
	}

	assert.Equal(t, "application/json", contentType)

}

func Test_userAccounts(t *testing.T) {
	err := aClient.httpsLogin(context.TODO())
	if err != nil {
		t.Errorf("login: %s", err.Error())
	}

	account0 := &UserAccount{
		ID:                           1,
		Name:                         "anonymous",
		Access:                       0,
		Kvm:                          1,
		Vmedia:                       1,
		NetworkPrivilege:             "administrator",
		FixedUserCount:               2,
		OEMProprietaryLevelPrivilege: 1,
		PrivilegeLimitSerial:         "none",
		SSHKey:                       "Not Available",
		CreationTime:                 4802,
		Changepassword:               0,
		UserOperation:                0,
		Password:                     "",
		ConfirmPassword:              "",
		PasswordSize:                 "",
		EmailFormat:                  "ami_format",
	}

	accounts, err := aClient.listUsers(context.TODO())
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, 10, len(accounts))
	assert.Equal(t, account0, accounts[0])
}
