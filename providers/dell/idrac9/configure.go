package idrac9

import (
	"encoding/json"
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
			case "LdapGroup":
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
