package idrac8

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/internal/helper"
	log "github.com/sirupsen/logrus"
)

func (i *IDrac8) ApplyCfg(config *cfgresources.ResourcesConfig) (err error) {
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
				for id, user := range userAccounts.([]*cfgresources.User) {

					//the dells have user id 1 set to a anon user, so we start with 2.
					userId := id + 2
					//setup params to post
					err := i.applyUserParams(user, userId)
					if err != nil {
						log.WithFields(log.Fields{
							"step":     "ApplyCfg",
							"Resource": cfg.Field(r).Kind(),
							"IP":       i.ip,
							"Model":    i.BmcType(),
							"Serial":   i.serial,
							"Error":    err,
						}).Warn("Unable to set user config.")
					}
				}

			case "Syslog":
				syslogCfg := cfg.Field(r).Interface().(*cfgresources.Syslog)
				err := i.applySyslogParams(syslogCfg)
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "ApplyCfg",
						"resource": cfg.Field(r).Kind(),
						"IP":       i.ip,
						"Model":    i.BmcType(),
						"Serial":   i.serial,
						"Error":    err,
					}).Warn("Unable to set Syslog config.")
				}
			case "Network":
				networkCfg := cfg.Field(r).Interface().(*cfgresources.Network)
				err := i.applyNetworkParams(networkCfg)
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "ApplyCfg",
						"resource": cfg.Field(r).Kind(),
						"IP":       i.ip,
						"Model":    i.BmcType(),
						"Serial":   i.serial,
						"Error":    err,
					}).Warn("Unable to set Network config.")
				}
			case "Ntp":
				ntpCfg := cfg.Field(r).Interface().(*cfgresources.Ntp)
				err := i.applyNtpParams(ntpCfg)
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "ApplyCfg",
						"resource": cfg.Field(r).Kind(),
						"IP":       i.ip,
						"Model":    i.BmcType(),
						"Serial":   i.serial,
					}).Warn("Unable to set NTP config.")
				}
			case "Ldap":
				ldapCfg := cfg.Field(r).Interface().(*cfgresources.Ldap)
				i.applyLdapParams(ldapCfg)
			case "LdapGroup":
				ldapGroups := cfg.Field(r).Interface()
				err := i.applyLdapGroupParams(ldapGroups.([]*cfgresources.LdapGroup), config.Ldap)
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "applyLdapParams",
						"resource": "Ldap",
						"IP":       i.ip,
						"Model":    i.BmcType(),
						"Serial":   i.serial,
						"Error":    err,
					}).Warn("applyLdapGroupParams returned error.")
				}
			case "License":
			case "Ssl":
			default:
				log.WithFields(log.Fields{
					"step":     "ApplyCfg",
					"resource": resourceName,
				}).Warn("Unknown resource.")
			}
		}
	}

	return err
}

// Encodes the string the way idrac expects credentials to be sent
// foobar == @066@06f@06f@062@061@072
// convert ever character to its hex equiv, and prepend @0
func encodeCred(s string) string {
	r := ""
	for _, c := range s {
		r += fmt.Sprintf("@0%x", c)
	}

	return r
}

// escapeLdapString escapes ldap parameters strings
func escapeLdapString(s string) string {
	r := ""
	for _, c := range s {
		if c == '=' || c == ',' {
			r += fmt.Sprintf("\\%c", c)
		} else {
			r += string(c)
		}
	}

	return r
}

// Return bool value if the role is valid.
func (i *IDrac8) isRoleValid(role string) bool {

	validRoles := []string{"admin", "user"}
	for _, v := range validRoles {
		if role == v {
			return true
		}
	}

	return false
}

// attempts to add the user
// if the user exists, update the users password.
func (i *IDrac8) applyUserParams(cfg *cfgresources.User, Id int) (err error) {

	if cfg.Name == "" {
		log.WithFields(log.Fields{
			"step": "apply-user-cfg",
		}).Fatal("User resource expects parameter: Name.")
	}

	if cfg.Password == "" {
		log.WithFields(log.Fields{
			"step": "apply-user-cfg",
		}).Fatal("User resource expects parameter: Password.")
	}

	if !i.isRoleValid(cfg.Role) {
		log.WithFields(log.Fields{
			"step":     "apply-user-cfg",
			"Username": cfg.Name,
		}).Warn("User resource Role must be declared and a must be a valid role: 'admin' OR 'user'.")
		return
	}

	var enable string
	if cfg.Enable == false {
		enable = "Disabled"
	} else {
		enable = "Enabled"
	}

	user := User{UserName: encodeCred(cfg.Name), Password: encodeCred(cfg.Password), Enable: enable, SolEnable: "Enabled"}

	switch cfg.Role {
	case "admin":
		user.Privilege = "511"
		user.IpmiLanPrivilege = "Administrator"
	case "user":
		user.Privilege = "497"
		user.IpmiLanPrivilege = "Operator"

	}

	data := make(map[string]User)
	data["iDRAC.Users"] = user

	payload, err := json.Marshal(data)
	if err != nil {
		log.WithFields(log.Fields{
			"step": helper.WhosCalling(),
		}).Warn("Unable to marshal syslog payload.")
		return err
	}

	endpoint := fmt.Sprintf("sysmgmt/2012/server/configgroup/iDRAC.Users.%d", Id)
	response, err := i.put(endpoint, payload)
	if err != nil {
		log.WithFields(log.Fields{
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"response": string(response),
		}).Warn("PUT request failed.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":     i.ip,
		"Model":  i.BmcType(),
		"Serial": i.serial,
		"User":   cfg.Name,
	}).Debug("User parameters applied.")

	return err
}

func (i *IDrac8) applySyslogParams(cfg *cfgresources.Syslog) (err error) {

	var port int
	enable := "Enabled"

	if cfg.Server == "" {
		log.WithFields(log.Fields{
			"step": helper.WhosCalling(),
		}).Warn("Syslog resource expects parameter: Server.")
		return
	}

	if cfg.Port == 0 {
		log.WithFields(log.Fields{
			"step": helper.WhosCalling(),
		}).Debug("Syslog resource port set to default: 514.")
		port = 514
	} else {
		port = cfg.Port
	}

	if cfg.Enable != true {
		enable = "Disabled"
		log.WithFields(log.Fields{
			"step": helper.WhosCalling(),
		}).Debug("Syslog resource declared with enable: false.")
	}

	data := make(map[string]Syslog)

	data["iDRAC.SysLog"] = Syslog{
		Port:    strconv.Itoa(port),
		Server1: cfg.Server,
		Server2: "",
		Server3: "",
		Enable:  enable,
	}

	payload, err := json.Marshal(data)
	if err != nil {
		log.WithFields(log.Fields{
			"step": helper.WhosCalling(),
		}).Warn("Unable to marshal syslog payload.")
		return err
	}

	endpoint := "sysmgmt/2012/server/configgroup/iDRAC.SysLog"
	response, err := i.put(endpoint, payload)
	if err != nil {
		log.WithFields(log.Fields{
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"response": string(response),
		}).Warn("PUT request failed.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":     i.ip,
		"Model":  i.BmcType(),
		"Serial": i.serial,
	}).Debug("Syslog parameters applied.")

	return err
}

func (i *IDrac8) applyNtpParams(cfg *cfgresources.Ntp) (err error) {

	if cfg.Server1 == "" {
		log.WithFields(log.Fields{
			"step": "apply-ntp-cfg",
		}).Warn("NTP resource expects parameter: server1.")
		return
	}

	if cfg.Timezone == "" {
		log.WithFields(log.Fields{
			"step": "apply-ntp-cfg",
		}).Warn("NTP resource expects parameter: timezone.")
		return
	}

	i.applyTimezoneParam(cfg.Timezone)
	i.applyNtpServerParam(cfg)

	return err
}

func (i *IDrac8) applyNtpServerParam(cfg *cfgresources.Ntp) {

	var enable int
	if cfg.Enable != true {
		log.WithFields(log.Fields{
			"step": helper.WhosCalling(),
		}).Debug("Ntp resource declared with enable: false.")
		enable = 0
	} else {
		enable = 1
	}

	//https://10.193.251.10/data?set=tm_ntp_int_opmode:1, \\
	//                               tm_ntp_str_server1:ntp0.lhr4.example.com, \\
	//                               tm_ntp_str_server2:ntp0.ams4.example.com, \\
	//                               tm_ntp_str_server3:ntp0.fra4.example.com
	queryStr := fmt.Sprintf("set=tm_ntp_int_opmode:%d,", enable)
	queryStr += fmt.Sprintf("tm_ntp_str_server1:%s,", cfg.Server1)
	queryStr += fmt.Sprintf("tm_ntp_str_server2:%s,", cfg.Server2)
	queryStr += fmt.Sprintf("tm_ntp_str_server3:%s,", cfg.Server3)

	//GET - params as query string
	//ntp servers

	endpoint := fmt.Sprintf("data?%s", queryStr)
	response, err := i.get(endpoint, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":       i.ip,
			"Model":    i.BmcType(),
			"Serial":   i.serial,
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"response": string(response),
		}).Warn("GET request failed.")
	}

	log.WithFields(log.Fields{
		"IP":     i.ip,
		"Model":  i.BmcType(),
		"Serial": i.serial,
	}).Debug("NTP servers param applied.")

}

//applies ldap config parameters
func (i *IDrac8) applyLdapParams(cfg *cfgresources.Ldap) {
	// LDAP settings
	// Notes: - all non-numeric,alphabetic characters are escaped
	//        - the idrac posts each payload twice?
	//        - requests can be either POST or GET except for the final one - postset?ldapconf

	//Set ldap groups

	r := i.applyLdapServerParam(cfg)
	if r != 0 {
		return
	}

	r = i.applyLdapSearchFilterParam(cfg)
	if r != 0 {
		return
	}

	log.WithFields(log.Fields{
		"IP":     i.ip,
		"Model":  i.BmcType(),
		"Serial": i.serial,
	}).Debug("Ldap config applied.")

}

// Applies ldap server param
// https://10.193.251.10/data?set=xGLServer:ldaps.prod.blah.com
func (i *IDrac8) applyLdapServerParam(cfg *cfgresources.Ldap) int {

	if cfg.Server == "" {
		log.WithFields(log.Fields{
			"step": "applyLdapServerParam",
		}).Warn("Ldap resource parameter Server required but not declared.")
		return 1
	}

	endpoint := fmt.Sprintf("data?set=xGLServer:%s", cfg.Server)
	response, err := i.get(endpoint, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":       i.ip,
			"Model":    i.BmcType(),
			"Serial":   i.serial,
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"response": string(response),
		}).Warn("GET request failed.")
		return 1
	}

	log.WithFields(log.Fields{
		"IP":     i.ip,
		"Model":  i.BmcType(),
		"Serial": i.serial,
	}).Debug("Ldap server param set.")

	return 0
}

// Applies ldap search filter param.
// set=xGLSearchFilter:objectClass\=posixAccount
func (i *IDrac8) applyLdapSearchFilterParam(cfg *cfgresources.Ldap) int {
	if cfg.SearchFilter == "" {
		log.WithFields(log.Fields{
			"step": "applyLdapSearchFilterParam",
		}).Warn("Ldap resource parameter SearchFilter required but not declared.")
		return 1
	}

	endpoint := fmt.Sprintf("data?set=xGLSearchFilter:%s", escapeLdapString(cfg.SearchFilter))
	response, err := i.get(endpoint, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":       i.ip,
			"Model":    i.BmcType(),
			"Serial":   i.serial,
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"response": string(response),
		}).Warn("GET request failed.")
		return 1
	}

	log.WithFields(log.Fields{
		"IP":     i.ip,
		"Model":  i.BmcType(),
		"Serial": i.serial,
	}).Debug("Ldap search filter param applied.")

	return 0
}

// Sets up ldap role groups
//data?set=xGLGroup1Name:cn\=bmcAdmins\,ou\=Group\,dc\=activehotels\,dc\=com
func (i *IDrac8) applyLdapGroupParams(cfgGroup []*cfgresources.LdapGroup, cfgLdap *cfgresources.Ldap) (err error) {

	groupId := 1

	//set to decide what privileges the group should have
	//497 == operator
	//511 == administrator (full privileges)
	privId := "0"

	//groupPrivilegeParam is populated per group and is passed to i.applyLdapRoleGroupPrivParam
	groupPrivilegeParam := ""

	//first some preliminary checks
	if cfgLdap.Port == 0 {
		msg := "Ldap resource parameter Port required but not declared"
		log.WithFields(log.Fields{
			"step": "applyLdapRoleGroupPrivParam",
		}).Warn(msg)
		errors.New(msg)
	}

	if cfgLdap.BaseDn == "" {
		msg := "Ldap resource parameter BaseDn required but not declared."
		log.WithFields(log.Fields{
			"step": "applyLdapRoleGroupPrivParam",
		}).Warn(msg)
		errors.New(msg)
	}

	if cfgLdap.UserAttribute == "" {
		msg := "Ldap resource parameter userAttribute required but not declared."
		log.WithFields(log.Fields{
			"step": "applyLdapRoleGroupPrivParam",
		}).Warn(msg)
		errors.New(msg)
	}

	if cfgLdap.GroupAttribute == "" {
		msg := "Ldap resource parameter groupAttribute required but not declared."
		log.WithFields(log.Fields{
			"step": "applyLdapRoleGroupPrivParam",
		}).Warn(msg)
		errors.New(msg)
	}

	//for each ldap group
	for _, group := range cfgGroup {

		//if a group has been set to disable in the config,
		//its configuration is skipped and removed.
		if !group.Enable {
			continue
		}

		if group.Role == "" {
			msg := "Ldap resource parameter Role required but not declared."
			log.WithFields(log.Fields{
				"Role": group.Role,
				"step": "applyLdapGroupParams",
			}).Warn(msg)
			continue
		}

		if group.Group == "" {
			msg := "Ldap resource parameter Group required but not declared."
			log.WithFields(log.Fields{
				"Role": group.Role,
				"step": "applyLdapGroupParams",
			}).Warn(msg)
			return errors.New(msg)
		}

		if group.GroupBaseDn == "" {
			msg := "Ldap resource parameter GroupBaseDn required but not declared."
			log.WithFields(log.Fields{
				"Role": group.Role,
				"step": "applyLdapGroupParams",
			}).Warn(msg)
			return errors.New(msg)
		}

		if !i.isRoleValid(group.Role) {
			msg := "Ldap resource Role must be a valid role: admin OR user."
			log.WithFields(log.Fields{
				"Role": group.Role,
				"step": "applyLdapGroupParams",
			}).Warn(msg)
			return errors.New(msg)
		}

		groupDn := fmt.Sprintf("%s,%s", group.Group, group.GroupBaseDn)
		groupDn = escapeLdapString(groupDn)

		endpoint := fmt.Sprintf("data?set=xGLGroup%dName:%s", groupId, groupDn)
		response, err := i.get(endpoint, nil)
		if err != nil {
			log.WithFields(log.Fields{
				"IP":       i.ip,
				"Model":    i.BmcType(),
				"Serial":   i.serial,
				"endpoint": endpoint,
				"step":     "applyLdapGroupParams",
				"response": string(response),
			}).Warn("GET request failed.")
			return err
		}

		log.WithFields(log.Fields{
			"IP":     i.ip,
			"Model":  i.BmcType(),
			"Serial": i.serial,
			"Role":   group.Role,
		}).Debug("Ldap GroupDN config applied.")

		switch group.Role {
		case "user":
			privId = "497"
		case "admin":
			privId = "511"
		}

		groupPrivilegeParam += fmt.Sprintf("xGLGroup%dPriv:%s,", groupId, privId)
		groupId += 1

	}

	//set the rest of the group privileges to 0
	for i := groupId + 1; i <= 5; i++ {
		groupPrivilegeParam += fmt.Sprintf("xGLGroup%dPriv:0,", i)
	}

	err = i.applyLdapRoleGroupPrivParam(cfgLdap, groupPrivilegeParam)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":     i.ip,
			"Model":  i.BmcType(),
			"Serial": i.serial,
			"step":   "applyLdapGroupParams",
		}).Warn("Unable to set Ldap Role Group Privileges.")
		return err
	}
	return err
}

// Apply ldap group privileges
//https://10.193.251.10/postset?ldapconf
// data=LDAPEnableMode:3,xGLNameSearchEnabled:0,xGLBaseDN:ou%5C%3DPeople%5C%2Cdc%5C%3Dactivehotels%5C%2Cdc%5C%3Dcom,xGLUserLogin:uid,xGLGroupMem:memberUid,xGLBindDN:,xGLCertValidationEnabled:1,xGLGroup1Priv:511,xGLGroup2Priv:97,xGLGroup3Priv:0,xGLGroup4Priv:0,xGLGroup5Priv:0,xGLServerPort:636
func (i *IDrac8) applyLdapRoleGroupPrivParam(cfg *cfgresources.Ldap, groupPrivilegeParam string) (err error) {

	baseDn := escapeLdapString(cfg.BaseDn)
	payload := "data=LDAPEnableMode:3,"  //setup generic ldap
	payload += "xGLNameSearchEnabled:0," //lookup ldap server from dns
	payload += fmt.Sprintf("xGLBaseDN:%s,", baseDn)
	payload += fmt.Sprintf("xGLUserLogin:%s,", cfg.UserAttribute)
	payload += fmt.Sprintf("xGLGroupMem:%s,", cfg.GroupAttribute)

	//if bindDn was declared, we set it.
	if cfg.BindDn != "" {
		bindDn := escapeLdapString(cfg.BindDn)
		payload += fmt.Sprintf("xGLBindDN:%s,", bindDn)
	} else {
		payload += "xGLBindDN:,"
	}

	payload += "xGLCertValidationEnabled:0," //we may want to be able to set this from config
	payload += groupPrivilegeParam
	payload += fmt.Sprintf("xGLServerPort:%d", cfg.Port)

	//fmt.Println(payload)
	endpoint := "postset?ldapconf"
	responseCode, responseBody, err := i.post(endpoint, []byte(payload))
	if err != nil || responseCode != 200 {
		log.WithFields(log.Fields{
			"IP":           i.ip,
			"Model":        i.BmcType(),
			"Serial":       i.serial,
			"endpoint":     endpoint,
			"step":         helper.WhosCalling(),
			"responseCode": responseCode,
			"response":     string(responseBody),
		}).Warn("POST request failed.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":     i.ip,
		"Model":  i.BmcType(),
		"Serial": i.serial,
	}).Debug("Ldap Group role privileges applied.")

	return err
}

func (i *IDrac8) applyTimezoneParam(timezone string) {
	//POST - params as query string
	//timezone
	//https://10.193.251.10/data?set=tm_tz_str_zone:CET

	endpoint := fmt.Sprintf("data?set=tm_tz_str_zone:%s", timezone)
	response, err := i.get(endpoint, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":       i.ip,
			"Model":    i.BmcType(),
			"Serial":   i.serial,
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"response": string(response),
		}).Warn("GET request failed.")
	}

	log.WithFields(log.Fields{
		"IP":     i.ip,
		"Model":  i.BmcType(),
		"Serial": i.serial,
	}).Debug("Timezone param applied.")

}

func (i *IDrac8) applyNetworkParams(cfg *cfgresources.Network) (err error) {

	params := map[string]int{
		"EnableIPv4":              1,
		"DHCPEnable":              1,
		"DNSFromDHCP":             1,
		"EnableSerialOverLan":     1,
		"EnableSerialRedirection": 1,
		"EnableIpmiOverLan":       1,
	}

	if !cfg.DNSFromDHCP {
		params["DNSFromDHCP"] = 0
	}

	if !cfg.IpmiEnable {
		params["EnableIpmiOverLan"] = 0
	}

	if !cfg.SolEnable {
		params["EnableSerialOverLan"] = 0
		params["EnableSerialRedirection"] = 0
	}

	endpoint := "data?set"
	payload := fmt.Sprintf("dhcpForDNSDomain:%d,", params["DNSFromDHCP"])
	payload += fmt.Sprintf("ipmiLAN:%d,", params["EnableIpmiOverLan"])
	payload += fmt.Sprintf("serialOverLanEnabled:%d,", params["EnableSerialOverLan"])
	payload += fmt.Sprintf("serialOverLanBaud:3,") //115.2 kbps
	payload += fmt.Sprintf("serialOverLanPriv:0,") //Administrator
	payload += fmt.Sprintf("racRedirectEna:%d,", params["EnableSerialRedirection"])
	payload += fmt.Sprintf("racEscKey:^\\\\")

	responseCode, responseBody, err := i.post(endpoint, []byte(payload))
	if err != nil || responseCode != 200 {
		log.WithFields(log.Fields{
			"IP":           i.ip,
			"Model":        i.BmcType(),
			"Serial":       i.serial,
			"endpoint":     endpoint,
			"step":         helper.WhosCalling(),
			"responseCode": responseCode,
			"response":     string(responseBody),
		}).Warn("POST request to set Network params failed.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":     i.ip,
		"Model":  i.BmcType(),
		"Serial": i.Serial,
	}).Debug("Network config parameters applied.")
	return err
}
