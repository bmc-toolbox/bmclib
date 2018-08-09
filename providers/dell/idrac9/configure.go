package idrac9

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/internal/helper"
	log "github.com/sirupsen/logrus"
)

func (i *IDrac9) ApplyCfg(config *cfgresources.ResourcesConfig) (err error) {
	cfg := reflect.ValueOf(config).Elem()

	//Each Field in ResourcesConfig struct is a ptr to a resource,
	//Here we figure the resources to be configured, i.e the ptr is not nil
	for r := 0; r < cfg.NumField(); r++ {
		resourceName := cfg.Type().Field(r).Name
		if cfg.Field(r).Pointer() != 0 {
			switch resourceName {
			case "User":
				//retrieve users resource values as an interface
				userAccounts := cfg.Field(r).Interface()

				//assert userAccounts interface to its actual type - A slice of ptrs to User
				err := i.applyUserParams(userAccounts.([]*cfgresources.User))
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "ApplyCfg",
						"resource": cfg.Field(r).Kind(),
						"IP":       i.ip,
						"Model":    i.BmcType(),
						"Serial":   i.Serial,
						"Error":    err,
					}).Warn("Unable to apply User config.")
				}
			case "Syslog":
			case "Network":
			case "Ntp":
			case "Ldap":
				ldapCfg := cfg.Field(r).Interface().(*cfgresources.Ldap)
				err := i.applyLdapServerParams(ldapCfg)
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "applyLdapParams",
						"resource": "Ldap",
						"IP":       i.ip,
						"Model":    i.BmcType(),
						"Serial":   i.Serial,
						"Error":    err,
					}).Warn("applyLdapServerParams returned error.")
				}
			case "LdapGroup":
				ldapGroupCfg := cfg.Field(r).Interface().([]*cfgresources.LdapGroup)
				err := i.applyLdapGroupParams(ldapGroupCfg)
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "applyLdapParams",
						"resource": "Ldap",
						"IP":       i.ip,
						"Model":    i.BmcType(),
						"Serial":   i.Serial,
						"Error":    err,
					}).Warn("applyLdapGroupParams returned error.")
				}
			case "Ssl":
			default:
				log.WithFields(log.Fields{
					"step": "ApplyCfg",
				}).Warn("Unknown resource.")
			}
		}
	}

	return err
}

// Iterates over iDrac users and adds/removes/modifies the user account
func (i *IDrac9) applyUserParams(cfgUsers []*cfgresources.User) (err error) {

	err = i.validateCfg(cfgUsers)
	if err != nil {
		msg := "Config validation failed."
		log.WithFields(log.Fields{
			"step":   "applyUserParams",
			"IP":     i.ip,
			"Model":  i.BmcType(),
			"Serial": i.Serial,
			"Error":  err,
		}).Warn(msg)
		return errors.New(msg)
	}

	idracUsers, err := i.queryUsers()
	if err != nil {
		msg := "Unable to query existing users"
		log.WithFields(log.Fields{
			"step":   "applyUserParams",
			"IP":     i.ip,
			"Model":  i.BmcType(),
			"Serial": i.Serial,
			"Error":  err,
		}).Warn(msg)
		return errors.New(msg)
	}

	//for each configuration user
	for _, cfgUser := range cfgUsers {

		userId, userInfo, uExists := userInIdrac(cfgUser.Name, idracUsers)

		//user to be added/updated
		if cfgUser.Enable {

			//new user to be added
			if uExists == false {
				userId, userInfo, err = getEmptyUserSlot(idracUsers)
				if err != nil {
					log.WithFields(log.Fields{
						"IP":     i.ip,
						"Model":  i.BmcType(),
						"Serial": i.Serial,
						"step":   helper.WhosCalling(),
						"User":   cfgUser.Name,
						"Error":  err,
					}).Warn("Unable to add new User.")
					continue
				}
			}

			userInfo.Enable = "Enabled"
			userInfo.SolEnable = "Enabled"
			userInfo.UserName = cfgUser.Name
			userInfo.Password = cfgUser.Password

			//set appropriate privileges
			if cfgUser.Role == "admin" {
				userInfo.Privilege = "511"
				userInfo.IpmiLanPrivilege = "Administrator"
			} else {
				userInfo.Privilege = "499"
				userInfo.IpmiLanPrivilege = "Operator"
			}

			err = i.putUser(userId, userInfo)
			if err != nil {
				log.WithFields(log.Fields{
					"IP":     i.ip,
					"Model":  i.BmcType(),
					"Serial": i.Serial,
					"step":   helper.WhosCalling(),
					"User":   cfgUser.Name,
					"Error":  err,
				}).Warn("Add/Update user request failed.")
				continue
			}

		} // end if cfgUser.Enable

		//if the user exists but is disabled in our config, remove the user
		if cfgUser.Enable == false && uExists == true {
			endpoint := fmt.Sprintf("sysmgmt/2017/server/user?userid=%d", userId)
			statusCode, response, err := i.delete_(endpoint)
			if err != nil {
				log.WithFields(log.Fields{
					"IP":         i.ip,
					"Model":      i.BmcType(),
					"Serial":     i.Serial,
					"step":       helper.WhosCalling(),
					"User":       cfgUser.Name,
					"Error":      err,
					"StatusCode": statusCode,
					"Response":   response,
				}).Warn("Delete user request failed.")
				continue
			}
		}

		log.WithFields(log.Fields{
			"IP":     i.ip,
			"Model":  i.BmcType(),
			"Serial": i.Serial,
			"User":   cfgUser.Name,
		}).Info("User parameters applied.")

	}

	return err
}

// Applies LDAP server params
func (i *IDrac9) applyLdapServerParams(cfg *cfgresources.Ldap) (err error) {
	params := map[string]string{
		"Enable":               "Disabled",
		"Port":                 "636",
		"UserAttribute":        "uid",
		"GroupAttribute":       "memberUid",
		"GroupAttributeIsDN":   "Enabled",
		"CertValidationEnable": "Disabled",
		"SearchFilter":         "objectClass=posixAccount",
	}

	if cfg.Server == "" {
		msg := "Ldap resource parameter Server required but not declared."
		log.WithFields(log.Fields{
			"IP":     i.ip,
			"Model":  i.BmcType(),
			"Serial": i.Serial,
			"step":   helper.WhosCalling,
		}).Warn(msg)
		return errors.New(msg)
	}

	if cfg.BaseDn == "" {
		msg := "Ldap resource parameter BaseDn required but not declared."
		log.WithFields(log.Fields{
			"Model": i.BmcType(),
			"step":  helper.WhosCalling,
		}).Warn(msg)
		return errors.New(msg)
	}

	if cfg.Enable {
		params["Enable"] = "Enabled"
	}

	if cfg.Port == 0 {
		params["Port"] = string(cfg.Port)
	}

	if cfg.UserAttribute != "" {
		params["UserAttribute"] = cfg.UserAttribute
	}

	if cfg.GroupAttribute != "" {
		params["GroupAttribute"] = cfg.GroupAttribute
	}

	if cfg.SearchFilter != "" {
		params["SearchFilter"] = cfg.SearchFilter
	}

	payload := Ldap{
		BaseDN:               cfg.BaseDn,
		BindDN:               cfg.BindDn,
		CertValidationEnable: params["CertValidationEnable"],
		Enable:               params["Enable"],
		GroupAttribute:       params["GroupAttribute"],
		GroupAttributeIsDN:   params["GroupAttributeIsDN"],
		Port:                 params["Port"],
		SearchFilter:         params["SearchFilter"],
		Server:               cfg.Server,
		UserAttribute:        params["UserAttribute"],
	}

	err = i.putLdap(payload)
	if err != nil {
		msg := "Ldap params PUT request failed."
		log.WithFields(log.Fields{
			"IP":     i.ip,
			"Model":  i.BmcType(),
			"Serial": i.Serial,
			"step":   helper.WhosCalling(),
			"Error":  err,
		}).Warn(msg)
		return errors.New("Ldap params put request failed.")
	}

	return err
}

// Iterates over iDrac Ldap role groups and adds/removes/modifies ldap role groups
func (i *IDrac9) applyLdapGroupParams(cfg []*cfgresources.LdapGroup) (err error) {

	idracLdapRoleGroups, err := i.queryLdapRoleGroups()
	if err != nil {
		msg := "Unable to query existing users"
		log.WithFields(log.Fields{
			"step":   "applyUserParams",
			"IP":     i.ip,
			"Model":  i.BmcType(),
			"Serial": i.Serial,
			"Error":  err,
		}).Warn(msg)
		return errors.New(msg)
	}

	//for each configuration ldap role group
	for _, cfgRole := range cfg {
		roleId, role, rExists := ldapRoleGroupInIdrac(cfgRole.Group, idracLdapRoleGroups)
		fmt.Printf("%+v <-> %+v <-> %+v", roleId, role, rExists)

		//role to be added/updated
		if cfgRole.Enable {

			//new role to be added
			if rExists == false {
				roleId, role, err = getEmptyLdapRoleGroupSlot(idracLdapRoleGroups)
				if err != nil {
					log.WithFields(log.Fields{
						"IP":              i.ip,
						"Model":           i.BmcType(),
						"Serial":          i.Serial,
						"step":            helper.WhosCalling(),
						"Ldap Role Group": cfgRole.Group,
						"Role Group DN":   cfgRole.Role,
						"Error":           err,
					}).Warn("Unable to add new Ldap Role Group.")
					continue
				}
			}

			role.DN = cfgRole.Group

			//set appropriate privileges
			if cfgRole.Role == "admin" {
				role.Privilege = "511"
			} else {
				role.Privilege = "499"
			}

			err = i.putLdapRoleGroup(roleId, role)
			if err != nil {
				log.WithFields(log.Fields{
					"IP":              i.ip,
					"Model":           i.BmcType(),
					"Serial":          i.Serial,
					"step":            helper.WhosCalling(),
					"Ldap Role Group": cfgRole.Group,
					"Role Group DN":   cfgRole.Role,
					"Error":           err,
				}).Warn("Add/Update LDAP Role Group request failed.")
				continue
			}

		} // end if cfgUser.Enable

		//if the role exists but is disabled in our config, remove the role
		if cfgRole.Enable == false && rExists == true {

			role.DN = ""
			role.Privilege = "0"
			err = i.putLdapRoleGroup(roleId, role)
			if err != nil {
				log.WithFields(log.Fields{
					"IP":              i.ip,
					"Model":           i.BmcType(),
					"Serial":          i.Serial,
					"step":            helper.WhosCalling(),
					"Ldap Role Group": cfgRole.Group,
					"Role Group DN":   cfgRole.Role,
					"Error":           err,
				}).Warn("Remove LDAP Role Group request failed.")
				continue
			}
		}

		log.WithFields(log.Fields{
			"IP":              i.ip,
			"Model":           i.BmcType(),
			"Serial":          i.Serial,
			"Ldap Role Group": cfgRole.Role,
		}).Info("Ldap Role Group parameters applied.")

	}

	return err
}
