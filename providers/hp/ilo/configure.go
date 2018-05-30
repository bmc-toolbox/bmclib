package ilo

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ncode/bmclib/cfgresources"
	log "github.com/sirupsen/logrus"
	"reflect"
	"runtime"
)

// returns the calling function.
func funcName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
}

func (i *Ilo) ApplyCfg(config *cfgresources.ResourcesConfig) (err error) {

	//check sessionKey is available
	if i.sessionKey == "" {
		msg := "Expected sessionKey not found, unable to configure BMC."
		log.WithFields(log.Fields{
			"step":  "Login()",
			"IP":    i.ip,
			"Model": i.ModelId(),
		}).Warn(msg)
		return errors.New(msg)
	}

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
						"Model":    i.ModelId(),
						"Error":    err,
					}).Warn("Unable to set User config.")
				}

			case "Syslog":
				syslogCfg := cfg.Field(r).Interface().(*cfgresources.Syslog)
				err := i.applySyslogParams(syslogCfg)
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "ApplyCfg",
						"resource": cfg.Field(r).Kind(),
						"IP":       i.ip,
						"Model":    i.ModelId(),
						"Error":    err,
					}).Warn("Unable to set Syslog config.")
				}
			case "Network":
				fmt.Printf("%s: %v : %s\n", resourceName, cfg.Field(r), cfg.Field(r).Kind())
			case "Ntp":
				ntpCfg := cfg.Field(r).Interface().(*cfgresources.Ntp)
				err := i.applyNtpParams(ntpCfg)
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "ApplyCfg",
						"resource": cfg.Field(r).Kind(),
						"IP":       i.ip,
						"Model":    i.ModelId(),
						"Error":    err,
					}).Warn("Unable to set NTP config.")
				}
			case "LdapGroup":
				ldapGroups := cfg.Field(r).Interface()
				fmt.Println(ldapGroups)
			case "Ldap":
				ldapCfg := cfg.Field(r).Interface().(*cfgresources.Ldap)
				fmt.Println(ldapCfg)
			case "Ssl":
				fmt.Printf("%s: %v : %s\n", resourceName, cfg.Field(r), cfg.Field(r).Kind())
			default:
				log.WithFields(log.Fields{
					"step": "ApplyCfg",
				}).Warn("Unknown resource.")
				//fmt.Printf("%v\n", cfg.Field(r))

			}
		}
	}

	return err
}

// Return bool value if the role is valid.
func (i *Ilo) isRoleValid(role string) bool {

	validRoles := []string{"admin", "user"}
	for _, v := range validRoles {
		if role == v {
			return true
		}
	}

	return false
}

// checks if a user is present in a given list
func userExists(user string, usersInfo []UserInfo) (userInfo UserInfo, exists bool) {

	for _, userInfo := range usersInfo {
		if userInfo.UserName == user {
			return userInfo, true
		}
	}

	return userInfo, false
}

// attempts to add the user
// if the user exists, update the users password.
func (i *Ilo) applyUserParams(users []*cfgresources.User) (err error) {

	existingUsers, err := i.queryUsers()
	if err != nil {
		msg := "Unable to query existing users"
		log.WithFields(log.Fields{
			"step":  "applyUserParams",
			"IP":    i.ip,
			"Model": i.ModelId(),
			"Error": err,
		}).Warn(msg)
		return errors.New(msg)
	}

	for _, user := range users {

		var postPayload bool

		if user.Name == "" {
			msg := "User resource expects parameter: Name."
			log.WithFields(log.Fields{
				"step": "applyUserParams",
			}).Warn(msg)
			return errors.New(msg)
		}

		if user.Password == "" {
			msg := "User resource expects parameter: Password."
			log.WithFields(log.Fields{
				"step":     "applyUserParams",
				"Username": user.Name,
			}).Warn(msg)
			return errors.New(msg)
		}

		if !i.isRoleValid(user.Role) {
			msg := "User resource Role must be declared and a must be a valid role: 'admin' OR 'user'."
			log.WithFields(log.Fields{
				"step":     "applyUserParams",
				"Username": user.Name,
			}).Warn(msg)
			return errors.New(msg)
		}

		//retrive userInfo
		userinfo, uexists := userExists(user.Name, existingUsers)
		//set session key
		userinfo.SessionKey = i.sessionKey

		//if the user is enabled setup parameters
		if user.Enable {
			userinfo.RemoteConsPriv = 1
			userinfo.VirtualMediaPriv = 1
			userinfo.ResetPriv = 1
			userinfo.UserPriv = 1
			userinfo.Password = user.Password

			if user.Role == "admin" {
				userinfo.ConfigPriv = 1
				userinfo.LoginPriv = 1
			} else if user.Role == "user" {
				userinfo.ConfigPriv = 0
				userinfo.LoginPriv = 0
			}

			//if the user exists, modify it
			if uexists {
				userinfo.Method = "mod_user"
				userinfo.UserId = userinfo.Id
				userinfo.UserName = user.Name
				userinfo.LoginName = user.Name
				userinfo.Password = user.Password
			} else {
				userinfo.Method = "add_user"
				userinfo.UserName = user.Name
				userinfo.LoginName = user.Name
				userinfo.Password = user.Password
			}

			postPayload = true
		}

		//if the user is disabled remove it
		if user.Enable == false && uexists {
			userinfo.Method = "del_user"
			userinfo.UserId = userinfo.Id
			log.WithFields(log.Fields{
				"IP":    i.ip,
				"Model": i.ModelId(),
				"User":  user.Name,
			}).Info("User disabled in config, will be removed.")
			postPayload = true
		}

		if postPayload {
			payload, err := json.Marshal(userinfo)
			if err != nil {
				log.WithFields(log.Fields{
					"IP":    i.ip,
					"Model": i.ModelId(),
					"step":  funcName(),
					"User":  user.Name,
					"Error": err,
				}).Warn("Unable to marshal userInfo payload to set User config.")
				continue
			}

			endpoint := "json/user_info"
			statusCode, response, err := i.post(endpoint, payload, false)
			if err != nil || statusCode != 200 {
				log.WithFields(log.Fields{
					"IP":         i.ip,
					"Model":      i.ModelId(),
					"endpoint":   endpoint,
					"step":       funcName(),
					"User":       user.Name,
					"StatusCode": statusCode,
					"response":   string(response),
					"Error":      err,
				}).Warn("POST request to set User config returned error.")
				continue
			}

			log.WithFields(log.Fields{
				"IP":    i.ip,
				"Model": i.ModelId(),
				"User":  user.Name,
			}).Info("User parameters applied.")

		}
	}

	return err
}

func (i *Ilo) applySyslogParams(cfg *cfgresources.Syslog) (err error) {

	//var port int
	//enable := "Enabled"

	//if cfg.Server == "" {
	//	log.WithFields(log.Fields{
	//		"step": funcName(),
	//	}).Warn("Syslog resource expects parameter: Server.")
	//	return
	//}

	//if cfg.Port == 0 {
	//	log.WithFields(log.Fields{
	//		"step": funcName(),
	//	}).Debug("Syslog resource port set to default: 514.")
	//	port = 514
	//} else {
	//	port = cfg.Port
	//}

	//if cfg.Enable != true {
	//	enable = "Disabled"
	//	log.WithFields(log.Fields{
	//		"step": funcName(),
	//	}).Debug("Syslog resource declared with enable: false.")
	//}

	return err
}

func (i *Ilo) applyNtpParams(cfg *cfgresources.Ntp) (err error) {

	if cfg.Server1 == "" {
		log.WithFields(log.Fields{
			"step": funcName(),
		}).Warn("NTP resource expects parameter: server1.")
		return
	}

	if cfg.Timezone == "" {
		log.WithFields(log.Fields{
			"step": "apply-ntp-cfg",
		}).Warn("NTP resource expects parameter: timezone.")
		return
	}

	return err
}
