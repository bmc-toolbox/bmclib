package idrac9

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
func (i *IDrac9) validateCfg(cfgUsers []*cfgresources.User) (err error) {

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

// iDrac9 supports upto 16 users, user 0 is reserved
// this function returns an empty user slot that can be used for a new user account
func getEmptyUserSlot(idracUsers userInfo) (userId int, user User, err error) {
	for userId, user := range idracUsers {

		if userId == 1 {
			continue
		}

		if user.UserName == "" {
			return userId, user, err
		}
	}

	return 0, user, errors.New("All user account slots in use, remove an account before adding a new one.")
}

// checks if a user is present in a given list
func userInIdrac(user string, usersInfo userInfo) (userId int, userInfo User, exists bool) {

	for userId, userInfo := range usersInfo {
		if userInfo.UserName == user {
			return userId, userInfo, true
		}
	}

	return userId, userInfo, false
}

// PUTs user config
func (i *IDrac9) putUser(userId int, user User) (err error) {

	idracPayload := make(map[string]User)
	idracPayload["iDRAC.Users"] = user

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error unmarshalling User payload: %s", err)
		return errors.New(msg)
	}

	endpoint := fmt.Sprintf("sysmgmt/2012/server/configgroup/iDRAC.Users.%d", userId)
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil || statusCode != 200 {
		msg := fmt.Sprintf("PUT request to set User config returned error, return code: %d", statusCode)
		return errors.New(msg)
	}

	return err
}

// PUTs ldap config
func (i *IDrac9) putLdap(ldap Ldap) (err error) {

	idracPayload := make(map[string]Ldap)
	idracPayload["iDRAC.LDAP"] = ldap

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error unmarshalling Ldap payload: %s", err)
		return errors.New(msg)
	}

	endpoint := fmt.Sprintf("sysmgmt/2012/server/configgroup/iDRAC.LDAP")
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil || statusCode != 200 {
		msg := fmt.Sprintf("PUT request to set User config returned error, return code: %d", statusCode)
		return errors.New(msg)
	}

	return err
}

// checks if a role group is in idrac
func ldapRoleGroupInIdrac(group string, roleGroups LdapRoleGroups) (roleId string, roleGroup LdapRoleGroup, exists bool) {

	for roleId, roleGroup := range roleGroups {
		if roleGroup.DN == group {
			return roleId, roleGroup, true
		}
	}

	return roleId, roleGroup, false
}

// iDrac9 supports upto 5 ldap role groups
// this function returns an empty user slot that can be used for a new ldap role group.
func getEmptyLdapRoleGroupSlot(roleGroups LdapRoleGroups) (roleId string, roleGroup LdapRoleGroup, err error) {

	for roleId, roleGroup := range roleGroups {
		if roleGroup.DN == "" {
			return roleId, roleGroup, err
		}
	}

	return roleId, roleGroup, errors.New("All Ldap Role Group slots in use, remove a Role group before adding a new one.")
}

// PUTs ldap role group config
func (i *IDrac9) putLdapRoleGroup(roleId string, ldapRoleGroup LdapRoleGroup) (err error) {

	idracPayload := make(map[string]LdapRoleGroup)
	idracPayload["iDRAC.LDAPRoleGroup"] = ldapRoleGroup

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error marshalling Ldap Role Group payload: %s", err)
		return errors.New(msg)
	}

	endpoint := fmt.Sprintf("sysmgmt/2012/server/configgroup/iDRAC.LDAPRoleGroup.%s", roleId)
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil || statusCode != 200 {
		msg := fmt.Sprintf("PUT request to set Ldap Role Group config returned error, return code: %d", statusCode)
		return errors.New(msg)
	}

	return err
}

// PUTs timezone config
func (i *IDrac9) putTimezone(timezone Timezone) (err error) {

	idracPayload := make(map[string]Timezone)
	idracPayload["iDRAC.Time"] = timezone

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error marshalling Timezone payload: %s", err)
		return errors.New(msg)
	}

	endpoint := fmt.Sprintf("sysmgmt/2012/server/configgroup/iDRAC.Time")
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil || statusCode != 200 {
		msg := fmt.Sprintf("PUT request to set Timezone config returned error, return code: %d", statusCode)
		return errors.New(msg)
	}

	return err
}

// PUTs NTP config
func (i *IDrac9) putNtpConfig(ntpConfig NtpConfig) (err error) {

	idracPayload := make(map[string]NtpConfig)
	idracPayload["iDRAC.NTPConfigGroup"] = ntpConfig

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error marshalling NTP payload: %s", err)
		return errors.New(msg)
	}

	endpoint := fmt.Sprintf("sysmgmt/2012/server/configgroup/iDRAC.NTPConfigGroup")
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil || statusCode != 200 {
		msg := fmt.Sprintf("PUT request to set Timezone config returned error, return code: %d", statusCode)
		return errors.New(msg)
	}

	return err
}

// PUTs NTP config
func (i *IDrac9) putSyslog(syslog Syslog) (err error) {

	idracPayload := make(map[string]Syslog)
	idracPayload["iDRAC.SysLog"] = syslog

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error marshalling Syslog payload: %s", err)
		return errors.New(msg)
	}

	endpoint := fmt.Sprintf("sysmgmt/2012/server/configgroup/iDRAC.Syslog")
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil || statusCode != 200 {
		msg := fmt.Sprintf("PUT request to set Syslog config returned error, return code: %d", statusCode)
		return errors.New(msg)
	}

	return err
}
