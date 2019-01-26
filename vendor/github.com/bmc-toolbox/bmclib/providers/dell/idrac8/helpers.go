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
func (i *IDrac8) validateUserCfg(cfgUsers []*cfgresources.User) (err error) {

	for _, cfgUser := range cfgUsers {
		if cfgUser.Name == "" {
			msg := "User resource expects parameter: Name."
			return errors.New(msg)
		}

		if cfgUser.Password == "" {
			msg := "User resource expects parameter: Password."
			return errors.New(msg)
		}

		if !isRoleValid(cfgUser.Role) {
			msg := "User resource expects parameter Role to be one of 'admin', 'user'"
			return errors.New(msg)
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

		if user.UserName == "" {
			return userID, user, err
		}
	}

	return 0, user, errors.New("All user account slots in use, remove an account before adding a new one")
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
func (i *IDrac8) putUser(userID int, user User) (err error) {

	idracPayload := make(map[string]User)
	idracPayload["iDRAC.Users"] = user

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error unmarshalling User payload: %s", err)
		return errors.New(msg)
	}

	endpoint := fmt.Sprintf("sysmgmt/2012/server/configgroup/iDRAC.Users.%d", userID)
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil || statusCode != 200 {
		msg := fmt.Sprintf("PUT request to set User config returned error, return code: %d", statusCode)
		return errors.New(msg)
	}

	return err
}
