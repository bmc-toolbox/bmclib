package idrac8

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bmc-toolbox/bmclib/cfgresources"
)

// Return bool value if the role is valid.
func isRoleValid(role string) bool {
	validRoles := []string{"admin", "user"}
	for _, v := range validRoles {
		if role == v {
			return true
		}
	}

	return false
}

// Return bool value if the role is valid.
func (i *IDrac8) validateUserCfg(cfgUsers []*cfgresources.User) error {
	for _, cfgUser := range cfgUsers {
		if cfgUser.Name == "" {
			return errors.New("user resource expects parameter: Name")
		}

		if cfgUser.Password == "" {
			return errors.New("user resource expects parameter: Password")
		}

		if !isRoleValid(cfgUser.Role) {
			return errors.New("user resource expects parameter Role to be one of 'admin', 'user'")
		}
	}

	return nil
}

// IDrac8 supports upto 16 users, user 0 is reserved
// this function returns an empty user slot that can be used for a new user account
func getEmptyUserSlot(idracUsers UserInfo) (userID int, user User, err error) {
	for userID, user := range idracUsers {
		if userID == 1 {
			continue
		}

		// from the web UI idrac8 doesn't allow removing users only disabling users.
		// There is a case where all user.UserName are NOT == "". This doens't mean that a new
		// user cannot be created. Disabled users regardless of whether user.UserName == "" can be
		// used for new user creation. FYI, ipmitool can remove the name: ipmitool user set name <id> ""
		if user.UserName == "" || user.Enable == "disabled" {
			return userID, user, err
		}
	}

	return 0, user, errors.New("all user account slots in use, remove an account before adding a new one")
}

// checks if a user is present in a given list
func userInIdrac(user string, usersInfo UserInfo) (userID int, userInfo User, exists bool) {
	for userID, userInfo := range usersInfo {
		if userInfo.UserName == user {
			return userID, userInfo, true
		}
	}

	return userID, userInfo, false
}

// PUTs user config
func (i *IDrac8) putUser(userID int, user User) error {
	idracPayload := make(map[string]User)
	idracPayload["iDRAC.Users"] = user

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		return fmt.Errorf("error unmarshalling User payload: %w", err)
	}

	endpoint := fmt.Sprintf("sysmgmt/2012/server/configgroup/iDRAC.Users.%d", userID)
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil {
		return fmt.Errorf("PUT request to set User config returned error: %w", err)
	}

	if statusCode != 200 {
		return fmt.Errorf("PUT request to set User config returned status code: %d", statusCode)
	}

	return nil
}
