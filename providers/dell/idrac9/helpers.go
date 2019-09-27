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
func getEmptyUserSlot(idracUsers userInfo) (userID int, user User, err error) {
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
func userInIdrac(user string, usersInfo userInfo) (userID int, userInfo User, exists bool) {

	for userID, userInfo := range usersInfo {
		if userInfo.UserName == user {
			return userID, userInfo, true
		}
	}

	return userID, userInfo, false
}

// PUTs user config
func (i *IDrac9) putUser(userID int, user User) (err error) {

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
func ldapRoleGroupInIdrac(group string, roleGroups LdapRoleGroups) (roleID string, roleGroup LdapRoleGroup, exists bool) {

	for roleID, roleGroup := range roleGroups {
		if roleGroup.DN == group {
			return roleID, roleGroup, true
		}
	}

	return roleID, roleGroup, false
}

// iDrac9 supports upto 5 ldap role groups
// this function returns an empty user slot that can be used for a new ldap role group.
func getEmptyLdapRoleGroupSlot(roleGroups LdapRoleGroups) (roleID string, roleGroup LdapRoleGroup, err error) {

	for roleID, roleGroup := range roleGroups {
		if roleGroup.DN == "" {
			return roleID, roleGroup, err
		}
	}

	return roleID, roleGroup, errors.New("All Ldap Role Group slots in use, remove a Role group before adding a new one")
}

// PUTs ldap role group config
func (i *IDrac9) putLdapRoleGroup(roleID string, ldapRoleGroup LdapRoleGroup) (err error) {

	idracPayload := make(map[string]LdapRoleGroup)
	idracPayload["iDRAC.LDAPRoleGroup"] = ldapRoleGroup

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error marshalling Ldap Role Group payload: %s", err)
		return errors.New(msg)
	}

	endpoint := fmt.Sprintf("sysmgmt/2012/server/configgroup/iDRAC.LDAPRoleGroup.%s", roleID)
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

// PUTs IPv4 config
func (i *IDrac9) putIPv4(ipv4 Ipv4) (err error) {

	idracPayload := make(map[string]Ipv4)
	idracPayload["iDRAC.IPv4"] = ipv4

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error marshalling Ipv4 payload: %s", err)
		return errors.New(msg)
	}

	endpoint := fmt.Sprintf("sysmgmt/2012/server/configgroup/iDRAC.IPv4")
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil || statusCode != 200 {
		msg := fmt.Sprintf("PUT request to set IPv4 config returned error, return code: %d", statusCode)
		return errors.New(msg)
	}

	return err
}

// PUTs SerialOverLan config
func (i *IDrac9) putSerialOverLan(serialOverLan SerialOverLan) (err error) {

	idracPayload := make(map[string]SerialOverLan)
	idracPayload["iDRAC.IPMISOL"] = serialOverLan

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error marshalling SerialOverLan payload: %s", err)
		return errors.New(msg)
	}

	endpoint := fmt.Sprintf("sysmgmt/2012/server/configgroup/iDRAC.IPMISOL")
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil || statusCode != 200 {
		msg := fmt.Sprintf("PUT request to set SerialOverLan config returned error, return code: %d", statusCode)
		return errors.New(msg)
	}

	return err
}

// PUTs SerialRedirection config
func (i *IDrac9) putSerialRedirection(serialRedirection SerialRedirection) (err error) {

	idracPayload := make(map[string]SerialRedirection)
	idracPayload["iDRAC.SerialRedirection"] = serialRedirection

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error marshalling serialRedirection payload: %s", err)
		return errors.New(msg)
	}

	endpoint := fmt.Sprintf("sysmgmt/2012/server/configgroup/iDRAC.SerialRedirection")
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil || statusCode != 200 {
		msg := fmt.Sprintf("PUT request to set serialRedirection config returned error, return code: %d", statusCode)
		return errors.New(msg)
	}

	return err
}

// PUTs IpmiOverLan config
func (i *IDrac9) putIpmiOverLan(ipmiOverLan IpmiOverLan) (err error) {

	idracPayload := make(map[string]IpmiOverLan)
	idracPayload["iDRAC.IPMILan"] = ipmiOverLan

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error marshalling ipmiOverLan payload: %s", err)
		return errors.New(msg)
	}

	endpoint := fmt.Sprintf("sysmgmt/2012/server/configgroup/iDRAC.IPMILAN")
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil || statusCode != 200 {
		msg := fmt.Sprintf("PUT request to set IpmiOverLan config returned error, return code: %d", statusCode)
		return errors.New(msg)
	}

	return err
}

// PUTs CSR request - request for a CSR based on given attributes.
func (i *IDrac9) putCSR(csrInfo CSRInfo) (err error) {

	m := map[string]CSRInfo{"iDRAC.Security": csrInfo}
	payload, err := json.Marshal(m)
	if err != nil {
		msg := fmt.Sprintf("Error marshalling CSRInfo payload: %s", err)
		return errors.New(msg)
	}

	endpoint := fmt.Sprintf("sysmgmt/2012/server/configgroup/iDRAC.Security")
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil || statusCode != 200 {
		msg := fmt.Sprintf("PUT request to set CSR attributes returned: %d", statusCode)
		return errors.New(msg)
	}

	return err
}

// putAlertConfig sets up alert filtering/sysloging
// see alertConfigPayload listed in model.go for payload.
func (i *IDrac9) putAlertConfig() (err error) {

	endpoint := fmt.Sprintf("sysmgmt/2012/server/eventpolicy")
	statusCode, _, err := i.put(endpoint, alertConfigPayload)
	if err != nil || statusCode != 200 {
		msg := fmt.Sprintf("PUT request to set attributes returned: %d", statusCode)
		return errors.New(msg)
	}

	return err
}

// putAlertEnable - enables/disables alerts
func (i *IDrac9) putAlertEnable(alertEnable AlertEnable) (err error) {

	m := map[string]AlertEnable{"iDRAC.IPMILan": alertEnable}
	payload, err := json.Marshal(m)
	if err != nil {
		msg := fmt.Sprintf("Error marshalling payload: %s", err)
		return errors.New(msg)
	}

	endpoint := fmt.Sprintf("sysmgmt/2012/server/configgroup/iDRAC.IPMILAN")
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil || statusCode != 200 {
		msg := fmt.Sprintf("PUT request to set attributes returned: %d", statusCode)
		return errors.New(msg)
	}

	return err
}
