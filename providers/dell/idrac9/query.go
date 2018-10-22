package idrac9

import (
	"encoding/json"

	"github.com/bmc-toolbox/bmclib/internal/helper"
	log "github.com/sirupsen/logrus"
)

func (i *IDrac9) Screenshot() (response []byte, extension string, err error) {

	endpoint := "capconsole/scapture0.png"
	extension = "png"

	response, err = i.get(endpoint, &map[string]string{})
	if err != nil {
		return []byte{}, extension, err
	}

	return response, extension, err
}

func (i *IDrac9) queryUsers() (users map[int]User, err error) {

	endpoint := "sysmgmt/2012/server/configgroup/iDRAC.Users"

	data, err := i.get(endpoint, &map[string]string{})
	if err != nil {
		log.WithFields(log.Fields{
			"IP":       i.ip,
			"Model":    i.BmcType(),
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"Error":    err,
		}).Warn("GET request failed.")
		return users, err
	}

	userData := make(idracUsers)
	err = json.Unmarshal(data, &userData)
	if err != nil {
		log.WithFields(log.Fields{
			"step":     "queryUserInfo",
			"resource": "User",
			"IP":       i.ip,
			"Model":    i.BmcType(),
			"Error":    err,
		}).Warn("Unable to unmarshal payload.")
		return users, err
	}

	return userData["iDRAC.Users"], err
}

func (i *IDrac9) queryLdapRoleGroups() (ldapRoleGroups LdapRoleGroups, err error) {

	endpoint := "sysmgmt/2012/server/configgroup/iDRAC.LDAPRoleGroup"

	data, err := i.get(endpoint, &map[string]string{})
	if err != nil {
		log.WithFields(log.Fields{
			"IP":       i.ip,
			"Model":    i.BmcType(),
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"Error":    err,
		}).Warn("GET request failed.")
		return ldapRoleGroups, err
	}

	idracLdapRoleGroups := make(idracLdapRoleGroups)
	err = json.Unmarshal(data, &idracLdapRoleGroups)
	if err != nil {
		log.WithFields(log.Fields{
			"step":     "queryUserInfo",
			"resource": "User",
			"IP":       i.ip,
			"Model":    i.BmcType(),
			"Error":    err,
		}).Warn("Unable to unmarshal payload.")
		return ldapRoleGroups, err
	}

	return idracLdapRoleGroups["iDRAC.LDAPRoleGroup"], err
}
