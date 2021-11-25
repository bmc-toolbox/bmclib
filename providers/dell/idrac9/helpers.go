package idrac9

import (
	"encoding/json"
	"errors"
	"fmt"
)

// IDRAC9 supports upto 16 users, user 0 is reserved.
// This function returns an empty user slot that can be used for a new user account.
func getEmptyUserSlot(idracUsers userInfo) (userID int, user User, err error) {
	for userID, user := range idracUsers {
		if userID == 1 {
			continue
		}

		if user.UserName == "" {
			return userID, user, err
		}
	}

	return 0, user, errors.New("All user account slots are in use, remove an account before adding a new one.")
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
func (i *IDrac9) putUser(userID int, user UserInfo) (err error) {
	idracPayload := make(map[string]UserInfo)
	idracPayload["iDRAC.Users"] = user

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error unmarshalling user payload: %s", err)
		return errors.New(msg)
	}

	endpoint := fmt.Sprintf("sysmgmt/2012/server/configgroup/iDRAC.Users.%d", userID)
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil {
		return fmt.Errorf("PUT request to set User config failed with error %s!", err.Error())
	} else if statusCode != 200 {
		return fmt.Errorf("PUT request to set User config failed with status code %d!", statusCode)
	}

	return nil
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

	endpoint := "sysmgmt/2012/server/configgroup/iDRAC.LDAP"
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil {
		return fmt.Errorf("PUT request to set LDAP config failed with error %s!", err.Error())
	} else if statusCode != 200 {
		return fmt.Errorf("PUT request to set LDAP config failed with status code %d!", statusCode)
	}

	return nil
}

func (i *IDrac9) putLdapRoleGroup(roleID string, ldapRoleGroup LdapRoleGroup) (err error) {
	idracPayload := make(map[string]LdapRoleGroup)
	idracPayload["iDRAC.LDAPRoleGroup"] = ldapRoleGroup

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error marshaling LDAP Role Group payload: %s", err)
		return errors.New(msg)
	}

	endpoint := fmt.Sprintf("sysmgmt/2012/server/configgroup/iDRAC.LDAPRoleGroup.%s", roleID)
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil {
		return fmt.Errorf("PUT request to set LDAPRoleGroup config failed with error %s!", err.Error())
	} else if statusCode != 200 {
		return fmt.Errorf("PUT request to set LDAPRoleGroup config failed with status code %d!", statusCode)
	}

	return nil
}

// PUTs timezone config
func (i *IDrac9) putTimezone(timezone Timezone) (err error) {
	idracPayload := make(map[string]Timezone)
	idracPayload["iDRAC.Time"] = timezone

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error marshaling Timezone payload: %s", err)
		return errors.New(msg)
	}

	endpoint := "sysmgmt/2012/server/configgroup/iDRAC.Time"
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil {
		return fmt.Errorf("PUT request to set Timezone config failed with error %s!", err.Error())
	} else if statusCode != 200 {
		return fmt.Errorf("PUT request to set Timezone config failed with status code %d!", statusCode)
	}

	return nil
}

// PUTs NTP config
func (i *IDrac9) putNtpConfig(ntpConfig NtpConfig) (err error) {
	idracPayload := make(map[string]NtpConfig)
	idracPayload["iDRAC.NTPConfigGroup"] = ntpConfig

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error marshaling NTP payload: %s", err)
		return errors.New(msg)
	}

	endpoint := "sysmgmt/2012/server/configgroup/iDRAC.NTPConfigGroup"
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil {
		return fmt.Errorf("PUT request to set NTP config failed with error %s!", err.Error())
	} else if statusCode != 200 {
		return fmt.Errorf("PUT request to set NTP config failed with status code %d!", statusCode)
	}

	return nil
}

// PUTs NTP config
func (i *IDrac9) putSyslog(syslog Syslog) (err error) {
	idracPayload := make(map[string]Syslog)
	idracPayload["iDRAC.SysLog"] = syslog

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error marshaling Syslog payload: %s", err)
		return errors.New(msg)
	}

	endpoint := "sysmgmt/2012/server/configgroup/iDRAC.Syslog"
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil {
		return fmt.Errorf("PUT request to set Syslog config failed with error %s!", err.Error())
	} else if statusCode != 200 {
		return fmt.Errorf("PUT request to set Syslog config failed with status code %d!", statusCode)
	}

	return nil
}

// PUTs IPv4 config
func (i *IDrac9) putIPv4(ipv4 Ipv4) (err error) {
	idracPayload := make(map[string]Ipv4)
	idracPayload["iDRAC.IPv4"] = ipv4

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error marshaling Ipv4 payload: %s", err)
		return errors.New(msg)
	}

	endpoint := "sysmgmt/2012/server/configgroup/iDRAC.IPv4"
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil {
		return fmt.Errorf("PUT request to set IPv4 config failed with error %s!", err.Error())
	} else if statusCode != 200 {
		return fmt.Errorf("PUT request to set IPv4 config failed with status code %d!", statusCode)
	}

	return nil
}

// PUTs SerialOverLan config
func (i *IDrac9) putSerialOverLan(serialOverLan SerialOverLan) (err error) {
	idracPayload := make(map[string]SerialOverLan)
	idracPayload["iDRAC.IPMISOL"] = serialOverLan

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error marshaling SerialOverLan payload: %s", err)
		return errors.New(msg)
	}

	endpoint := "sysmgmt/2012/server/configgroup/iDRAC.IPMISOL"
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil {
		return fmt.Errorf("PUT request to set SerialOverLAN config failed with error %s!", err.Error())
	} else if statusCode != 200 {
		return fmt.Errorf("PUT request to set SerialOverLAN config failed with status code %d!", statusCode)
	}

	return nil
}

// PUTs SerialRedirection config
func (i *IDrac9) putSerialRedirection(serialRedirection SerialRedirection) (err error) {
	idracPayload := make(map[string]SerialRedirection)
	idracPayload["iDRAC.SerialRedirection"] = serialRedirection

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error marshaling serialRedirection payload: %s", err)
		return errors.New(msg)
	}

	endpoint := "sysmgmt/2012/server/configgroup/iDRAC.SerialRedirection"
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil {
		return fmt.Errorf("PUT request to set SerialRedirection config failed with error %s!", err.Error())
	} else if statusCode != 200 {
		return fmt.Errorf("PUT request to set SerialRedirection config failed with status code %d!", statusCode)
	}

	return nil
}

// PUTs IpmiOverLan config
func (i *IDrac9) putIpmiOverLan(ipmiOverLan IpmiOverLan) (err error) {
	idracPayload := make(map[string]IpmiOverLan)
	idracPayload["iDRAC.IPMILan"] = ipmiOverLan

	payload, err := json.Marshal(idracPayload)
	if err != nil {
		msg := fmt.Sprintf("Error marshaling ipmiOverLan payload: %s", err)
		return errors.New(msg)
	}

	endpoint := "sysmgmt/2012/server/configgroup/iDRAC.IPMILAN"
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil {
		return fmt.Errorf("PUT request to set IPMIOverLAN config failed with error %s!", err.Error())
	} else if statusCode != 200 {
		return fmt.Errorf("PUT request to set IPMIOverLAN config failed with status code %d!", statusCode)
	}

	return nil
}

// PUTs CSR request - request for a CSR based on given attributes.
func (i *IDrac9) putCSR(csrInfo CSRInfo) (err error) {
	m := map[string]CSRInfo{"iDRAC.Security": csrInfo}
	payload, err := json.Marshal(m)
	if err != nil {
		msg := fmt.Sprintf("Error marshaling CSRInfo payload: %s", err)
		return errors.New(msg)
	}

	endpoint := "sysmgmt/2012/server/configgroup/iDRAC.Security"
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil {
		return fmt.Errorf("PUT request to set CSR config failed with error %s!", err.Error())
	} else if statusCode != 200 {
		return fmt.Errorf("PUT request to set CSR attributes failed with status code %d!", statusCode)
	}

	return nil
}

// putAlertConfig sets up alert filtering/sysloging
// see alertConfigPayload listed in model.go for payload.
func (i *IDrac9) putAlertConfig() (err error) {
	endpoint := "sysmgmt/2012/server/eventpolicy"
	statusCode, _, err := i.put(endpoint, alertConfigPayload)
	if err != nil {
		return fmt.Errorf("PUT request to set AlertConfig config failed with error %s!", err.Error())
	} else if statusCode != 200 {
		return fmt.Errorf("PUT request to set AlertConfig config failed with status code %d!", statusCode)
	}

	return nil
}

// putAlertEnable - enables/disables alerts
func (i *IDrac9) putAlertEnable(alertEnable AlertEnable) (err error) {
	m := map[string]AlertEnable{"iDRAC.IPMILan": alertEnable}
	payload, err := json.Marshal(m)
	if err != nil {
		msg := fmt.Sprintf("Error marshaling payload: %s", err)
		return errors.New(msg)
	}

	endpoint := "sysmgmt/2012/server/configgroup/iDRAC.IPMILAN"
	statusCode, _, err := i.put(endpoint, payload)
	if err != nil {
		return fmt.Errorf("PUT request to set AlertEnable config failed with error %s!", err.Error())
	} else if statusCode != 200 {
		return fmt.Errorf("PUT request to set AlertEnable config failed with status code %d!", statusCode)
	}

	return nil
}
