package c7000

import (
	"fmt"
	"reflect"

	"github.com/bmc-toolbox/bmclib/cfgresources"
	log "github.com/sirupsen/logrus"
)

func (c *C7000) ApplyCfg(config *cfgresources.ResourcesConfig) (err error) {
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
				for _, user := range userAccounts.([]*cfgresources.User) {
					err := c.applyUserParams(user)
					if err != nil {
						log.WithFields(log.Fields{
							"step":     "ApplyCfg",
							"Resource": cfg.Field(r).Kind(),
							"IP":       c.ip,
							"Model":    c.BmcType(),
							"Serial":   c.serial,
							"Error":    err,
						}).Warn("Unable to set user config.")
					}
				}

			case "Syslog":
				syslogCfg := cfg.Field(r).Interface().(*cfgresources.Syslog)
				err := c.applySyslogParams(syslogCfg)
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "ApplyCfg",
						"resource": cfg.Field(r).Kind(),
						"IP":       c.ip,
						"Model":    c.BmcType(),
						"Serial":   c.serial,
						"Error":    err,
					}).Warn("Unable to set Syslog config.")
					return err
				}
			case "Network":
				fmt.Printf("%s: %v : %s\n", resourceName, cfg.Field(r), cfg.Field(r).Kind())
			case "Ntp":
				ntpCfg := cfg.Field(r).Interface().(*cfgresources.Ntp)
				err := c.applyNtpParams(ntpCfg)
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "ApplyCfg",
						"resource": cfg.Field(r).Kind(),
						"IP":       c.ip,
						"Model":    c.BmcType(),
						"Serial":   c.serial,
						"Error":    err,
					}).Warn("Unable to set NTP config.")
					return err
				}
			case "LdapGroup":
				ldapGroups := cfg.Field(r).Interface()
				err := c.applyLdapGroupParams(ldapGroups.([]*cfgresources.LdapGroup))
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "applyLdapParams",
						"resource": "LdapGroup",
						"IP":       c.ip,
						"Model":    c.BmcType(),
						"Serial":   c.serial,
						"Error":    err,
					}).Warn("applyLdapGroupParams returned error.")
					return err
				}
			case "Ldap":
				ldapCfg := cfg.Field(r).Interface().(*cfgresources.Ldap)
				err := c.applyLdapParams(ldapCfg)
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "applyLdapParams",
						"resource": "Ldap",
						"IP":       c.ip,
						"Model":    c.BmcType(),
						"Serial":   c.serial,
						"Error":    err,
					}).Warn("applyLdapParams returned error.")
					return err
				}
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
func (c *C7000) isRoleValid(role string) bool {

	validRoles := []string{"admin", "user"}
	for _, v := range validRoles {
		if role == v {
			return true
		}
	}

	return false
}

//1. apply ldap group params
//2. enable ldap auth
//3. apply ldap server params
func (c *C7000) applyLdapParams(cfg *cfgresources.Ldap) error {

	err := c.applysetLdapInfo4(cfg)
	if err != nil {
		log.WithFields(log.Fields{
			"step":     "applyLdapParams",
			"resource": "Ldap",
			"IP":       c.ip,
			"Model":    c.BmcType(),
			"Serial":   c.serial,
			"Error":    err,
		}).Warn("applyLdapParams returned error.")
		return err
	}

	err = c.applyEnableLdapAuth(cfg.Enable)
	if err != nil {
		log.WithFields(log.Fields{
			"step":     "applyLdapParams",
			"resource": "Ldap",
			"IP":       c.ip,
			"Model":    c.BmcType(),
			"Serial":   c.serial,
			"Error":    err,
		}).Warn("applyLdapParams returned error.")
		return err
	}

	return err
}

// Apply Ldap server config params
// <hpoa:setLdapInfo4>
//   <hpoa:directoryServerAddress>machine.example.com</hpoa:directoryServerAddress>
//   <hpoa:directoryServerSslPort>636</hpoa:directoryServerSslPort>
//   <hpoa:directoryServerGCPort>0</hpoa:directoryServerGCPort>
//   <hpoa:userNtAccountNameMapping>false</hpoa:userNtAccountNameMapping>
//   <hpoa:enableServiceAccount>false</hpoa:enableServiceAccount>
//   <hpoa:serviceAccountName></hpoa:serviceAccountName>
//   <hpoa:serviceAccountPassword></hpoa:serviceAccountPassword>
//   <hpoa:searchContexts xmlns:hpoa="hpoa.xsd">
//    <hpoa:searchContext>ou=People,dc=activehotels,dc=com</hpoa:searchContext>
//    <hpoa:searchContext/>
//    <hpoa:searchContext/>
//    <hpoa:searchContext/>
//    <hpoa:searchContext/>
//    <hpoa:searchContext/>
//   </hpoa:searchContexts>
// </hpoa:setLdapInfo4>
func (c *C7000) applysetLdapInfo4(cfg *cfgresources.Ldap) (err error) {
	if cfg.Server == "" {
		log.WithFields(log.Fields{
			"step":  "applysetLdapInfo4",
			"Model": c.BmcType(),
		}).Warn("Ldap resource parameter Server required but not declared.")
		return err
	}

	if cfg.Port == 0 {
		log.WithFields(log.Fields{
			"step":  "applysetLdapInfo4",
			"Model": c.BmcType(),
		}).Warn("Ldap resource parameter Port required but not declared.")
		return err
	}

	if cfg.BaseDn == "" {
		log.WithFields(log.Fields{
			"step": "applysetLdapInfo4",
		}).Warn("Ldap resource parameter BaseDn required but not declared.")
		return err
	}

	searchcontexts := SearchContexts{Hpoa: "hpoa.xsd"}
	searchcontexts.SearchContext = append(searchcontexts.SearchContext, SearchContext{Text: cfg.BaseDn})
	for s := 1; s <= 6; s++ {
		searchcontexts.SearchContext = append(searchcontexts.SearchContext, SearchContext{Text: ""})
	}

	payload := setLdapInfo4{
		DirectoryServerAddress:   cfg.Server,
		DirectoryServerSslPort:   cfg.Port,
		DirectoryServerGCPort:    0,
		UserNtAccountNameMapping: false,
		EnableServiceAccount:     false,
		ServiceAccountName:       "",
		ServiceAccountPassword:   "",
		SearchContexts:           searchcontexts,
	}

	statusCode, _, err := c.postXML(payload)
	if statusCode != 200 || err != nil {
		log.WithFields(log.Fields{
			"step":       "applysetLdapInfo4",
			"IP":         c.ip,
			"Model":      c.BmcType(),
			"Serial":     c.serial,
			"statusCode": statusCode,
			"Error":      err,
		}).Warn("Ldap applysetLdapInfo4 apply request returned non 200.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":    c.ip,
		"Model": c.BmcType(),
	}).Debug("Ldap Server parameters applied.")
	return err
}

// <hpoa:enableLdapAuthentication>
//  <hpoa:enableLdap>true</hpoa:enableLdap>
//  <hpoa:enableLocalUsers>true</hpoa:enableLocalUsers>
// </hpoa:enableLdapAuthentication>
func (c *C7000) applyEnableLdapAuth(enable bool) (err error) {

	payload := enableLdapAuthentication{EnableLdap: enable, EnableLocalUsers: true}
	statusCode, _, err := c.postXML(payload)
	if statusCode != 200 || err != nil {
		log.WithFields(log.Fields{
			"step":       "applyEnableLdapAuth",
			"IP":         c.ip,
			"Model":      c.BmcType(),
			"Serial":     c.serial,
			"statusCode": statusCode,
			"Error":      err,
		}).Warn("Ldap applyEnableLdapAuth apply request returned non 200.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":     c.ip,
		"Model":  c.BmcType(),
		"Serial": c.serial,
	}).Debug("Ldap Enabled.")
	return err
}

// Actions carried out in order
// 1.  addLdapGroup
// 2.  setLdapGroupBayAcl
// 3.  addLdapGroupBayAccess (done)
func (c *C7000) applyLdapGroupParams(cfg []*cfgresources.LdapGroup) (err error) {

	for _, group := range cfg {

		if !c.isRoleValid(group.Role) {
			log.WithFields(log.Fields{
				"step":   "applyLdapGroupParams",
				"role":   group.Role,
				"Model":  c.BmcType(),
				"Serial": c.serial,
			}).Warn("Ldap resource Role must be a valid role: admin OR user.")
			return
		}

		if group.Group == "" {
			log.WithFields(log.Fields{
				"step":      "applyLdapGroupParams",
				"Model":     c.BmcType(),
				"Ldap role": group.Role,
				"Serial":    c.serial,
			}).Warn("Ldap resource parameter Group required but not declared.")
			return
		}

		//1. addLdapGroup
		err = c.applyAddLdapGroup(group.Group)
		if err != nil {
			log.WithFields(log.Fields{
				"step":     "applyAddLdapGroup",
				"resource": "Ldap",
				"IP":       c.ip,
				"Model":    c.BmcType(),
				"Serial":   c.serial,
				"Error":    err,
			}).Warn("addLdapGroup returned error.")
			return
		}

		//2. setLdapGroupBayAcl
		err = c.applyLdapGroupBayAcl(group.Role, group.Group)
		if err != nil {
			log.WithFields(log.Fields{
				"step":     "setLdapGroupBayAcl",
				"resource": "Ldap",
				"IP":       c.ip,
				"Model":    c.BmcType(),
				"Serial":   c.serial,
				"Error":    err,
			}).Warn("addLdapGroup returned error.")
			return
		}

		//3. applyAddLdapGroupBayAccess
		err = c.applyAddLdapGroupBayAccess(group.Group)
		if err != nil {
			log.WithFields(log.Fields{
				"step":     "applyAddLdapGroupBayAccess",
				"resource": "Ldap",
				"IP":       c.ip,
				"Model":    c.BmcType(),
				"Serial":   c.serial,
				"Error":    err,
			}).Warn("addLdapGroup returned error.")
			return
		}

		if err != nil {
			log.WithFields(log.Fields{
				"step":   "applyLdapGroupParams",
				"IP":     c.ip,
				"Role":   group.Role,
				"Model":  c.BmcType(),
				"Serial": c.serial,
				"Error":  err,
			}).Warn("Unable to set LdapGroup config for role.")
		}
	}

	log.WithFields(log.Fields{
		"IP":     c.ip,
		"Model":  c.BmcType(),
		"Serial": c.serial,
	}).Debug("Ldap config applied")
	return
}

// LDAP setup group, soap actions in order.
// <hpoa:addLdapGroup>
//  <hpoa:ldapGroup>bmcAdmins</hpoa:ldapGroup>
// </hpoa:addLdapGroup>
func (c *C7000) applyAddLdapGroup(group string) (err error) {

	payload := addLdapGroup{LdapGroup: ldapGroup{Text: group}}
	statusCode, _, err := c.postXML(payload)
	if statusCode == 200 || statusCode == 500 { // 500 indicates the group exists.
		log.WithFields(log.Fields{
			"step":       "applyAddLdapGroup",
			"IP":         c.ip,
			"Model":      c.BmcType(),
			"Serial":     c.serial,
			"statusCode": statusCode,
		}).Debug("Ldap applyAddLdapGroup applied.")
		return nil
	}

	if statusCode >= 300 || err != nil {
		log.WithFields(log.Fields{
			"step":       "applyAddLdapGroup",
			"IP":         c.ip,
			"Model":      c.BmcType(),
			"Serial":     c.serial,
			"statusCode": statusCode,
		}).Warn("Ldap applyAddLdapGroup request returned non 200.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":    c.ip,
		"Model": c.BmcType(),
	}).Debug("Ldap group added.")
	return nil
}

// Applies ldap group ACL
// <hpoa:setLdapGroupBayAcl>
//  <hpoa:ldapGroup>bmcAdmins</hpoa:ldapGroup>
//  <hpoa:acl>ADMINISTRATOR</hpoa:acl>
// </hpoa:setLdapGroupBayAcl>
func (c *C7000) applyLdapGroupBayAcl(role string, group string) (err error) {

	var userAcl string

	if role == "admin" {
		userAcl = "ADMINISTRATOR"
	} else {
		userAcl = "USER"
	}

	payload := setLdapGroupBayAcl{LdapGroup: ldapGroup{Text: group}, Acl: Acl{Text: userAcl}}
	statusCode, _, err := c.postXML(payload)
	if statusCode != 200 || err != nil {
		log.WithFields(log.Fields{
			"step":       "applyLdapGroupBayAcl",
			"IP":         c.ip,
			"Model":      c.BmcType(),
			"Serial":     c.serial,
			"statusCode": statusCode,
			"Error":      err,
		}).Warn("LDAP applyLdapGroupBayAcl request returned non 200.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":     c.ip,
		"Model":  c.BmcType(),
		"Serial": c.serial,
		"Role":   role,
		"Group":  group,
	}).Debug("Ldap group ACL added.")
	return err
}

// Set blade, interconnect access
//<hpoa:addLdapGroupBayAccess>
// <hpoa:ldapGroup>bmcAdmins</hpoa:ldapGroup>
// <hpoa:bays xmlns:hpoa="hpoa.xsd">
//  <hpoa:oaAccess>true</hpoa:oaAccess>
//  <hpoa:bladeBays>
//   <hpoa:blade xmlns:hpoa="hpoa.xsd">
//    <hpoa:bayNumber>1</hpoa:bayNumber>
//    <hpoa:access>true</hpoa:access>
//   </hpoa:blade>
//  <hpoa:blade xmlns:hpoa="hpoa.xsd">
//    <hpoa:bayNumber>2</hpoa:bayNumber>
//    <hpoa:access>true</hpoa:access>
//    </hpoa:blade>
//    .... repeat for number of blades in a c7000 chassis ~ 16 max
// </hpoa:bladeBays>
// <hpoa:interconnectTrayBays>
//  <hpoa:interconnectTray xmlns:hpoa="hpoa.xsd">
//  <hpoa:bayNumber>1</hpoa:bayNumber>
//  <hpoa:access>true</hpoa:access>
// </hpoa:interconnectTray>
// <hpoa:interconnectTray xmlns:hpoa="hpoa.xsd">
//  <hpoa:bayNumber>2</hpoa:bayNumber>
//  <hpoa:access>true</hpoa:access>
//  </hpoa:interconnectTray>
// ...  repeat for number of interconnect bays in chassis ~ 8
//  <hpoa:interconnectTrayBays>
// </hpoa:bays>
//</hpoa:addLdapGroupBayAccess>

func (c *C7000) applyAddLdapGroupBayAccess(group string) (err error) {
	//group = "bmcAdmins"

	//setup blade bays payload
	bladebays := bladeBays{}
	for b := 1; b <= 16; b++ {
		baynumber := bayNumber{Text: b}
		access := access{Text: true}
		blade := blade{Hpoa: "hpoa.xsd", BayNumber: baynumber, Access: access}
		bladebays.Blade = append(bladebays.Blade, blade)
	}

	//setup interconnect tray bays payload
	interconnecttraybays := interconnectTrayBays{}
	for t := 1; t <= 8; t++ {
		access := access{Text: true}
		baynumber := bayNumber{Text: t}
		interconnecttray := interconnectTray{Hpoa: "hpoa.xsd", Access: access, BayNumber: baynumber}
		interconnecttraybays.InterconnectTray = append(interconnecttraybays.InterconnectTray, interconnecttray)
	}

	//setup the bays payload
	bayz := bays{
		Hpoa:                 "hpoa.xsd",
		OaAccess:             oaAccess{Text: true},
		BladeBays:            bladebays,
		InterconnectTrayBays: interconnecttraybays,
	}

	payload := addLdapGroupBayAccess{
		LdapGroup: ldapGroup{Text: group},
		Bays:      bayz,
	}

	statusCode, _, err := c.postXML(payload)
	if statusCode != 200 || err != nil {
		log.WithFields(log.Fields{
			"step":       "applyAddLdapGroupBayAccess",
			"IP":         c.ip,
			"Model":      c.BmcType(),
			"Serial":     c.serial,
			"statusCode": statusCode,
			"Error":      err,
		}).Warn("LDAP applyAddLdapGroupBayAccess apply request returned non 200.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":     c.ip,
		"Model":  c.BmcType(),
		"Serial": c.serial,
		"Group":  group,
	}).Debug("Ldap interconnect and bay ACLs added.")
	return err
}

// attempts to add the user
// if the user exists, update the users password.
func (c *C7000) applyUserParams(cfg *cfgresources.User) (err error) {

	if cfg.Name == "" {
		log.WithFields(log.Fields{
			"step":  "applyUserParams",
			"Model": c.BmcType(),
		}).Fatal("User resource expects parameter: Name.")
	}

	if cfg.Password == "" {
		log.WithFields(log.Fields{
			"step":  "applyUserParams",
			"Model": c.BmcType(),
		}).Fatal("User resource expects parameter: Password.")
	}

	if c.isRoleValid(cfg.Role) == false {
		log.WithFields(log.Fields{
			"step":  "applyUserParams",
			"Model": c.BmcType(),
			"Role":  cfg.Role,
		}).Fatal("User resource Role must be declared and a valid role: admin.")
	}

	username := Username{Text: cfg.Name}
	password := Password{Text: cfg.Password}

	//if user account is disabled, remove the user
	if cfg.Enable == false {
		payload := RemoveUser{Username: username}
		statusCode, _, _ := c.postXML(payload)

		//user doesn't exist
		if statusCode != 400 {
			return err
		}

		log.WithFields(log.Fields{
			"IP":     c.ip,
			"Model":  c.BmcType(),
			"Serial": c.serial,
			"User":   cfg.Name,
		}).Debug("User removed.")

		//user exists and was removed.
		return err
	}

	payload := AddUser{Username: username, Password: password}
	statusCode, _, err := c.postXML(payload)
	if err != nil {
		return err
	}

	//user exists
	if statusCode == 400 {
		log.WithFields(log.Fields{
			"step":        "applyUserParams",
			"user":        cfg.Name,
			"IP":          c.ip,
			"Model":       c.BmcType(),
			"Serial":      c.serial,
			"Return code": statusCode,
		}).Debug("User already exists, setting password.")

		//update user password
		err := c.setUserPassword(cfg.Name, cfg.Password)
		if err != nil {
			return err
		}

		//update user acl
		err = c.setUserAcl(cfg.Name, cfg.Role)
		if err != nil {
			return err
		}

		//updates user blade bay access acls
		err = c.applyAddUserBayAccess(cfg.Name)
		if err != nil {
			return err
		}

	}

	log.WithFields(log.Fields{
		"IP":     c.ip,
		"Model":  c.BmcType(),
		"Serial": c.serial,
		"user":   cfg.Name,
	}).Debug("User cfg applied.")
	return err
}

func (c *C7000) setUserPassword(user string, password string) (err error) {

	u := Username{Text: user}
	p := Password{Text: password}
	payload := SetUserPassword{Username: u, Password: p}

	statusCode, _, err := c.postXML(payload)
	if err != nil {
		log.WithFields(log.Fields{
			"step":        "setUserPassword",
			"user":        user,
			"IP":          c.ip,
			"Model":       c.BmcType(),
			"Serial":      c.serial,
			"return code": statusCode,
			"Error":       err,
		}).Warn("Unable to set user password.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":     c.ip,
		"Model":  c.BmcType(),
		"Serial": c.serial,
		"user":   user,
	}).Debug("User password set.")
	return err
}

func (c *C7000) setUserAcl(user string, role string) (err error) {

	var aclRole string
	if role == "admin" {
		aclRole = "ADMINISTRATOR"
	} else {
		aclRole = "OPERATOR"
	}

	u := Username{Text: user}
	a := Acl{Text: aclRole}

	payload := SetUserBayAcl{Username: u, Acl: a}

	statusCode, _, err := c.postXML(payload)
	if err != nil {
		log.WithFields(log.Fields{
			"step":        "setUserAcl",
			"user":        user,
			"Acl":         role,
			"IP":          c.ip,
			"Model":       c.BmcType(),
			"Serial":      c.serial,
			"return code": statusCode,
			"Error":       err,
		}).Warn("Unable to set user Acl.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":     c.ip,
		"Model":  c.BmcType(),
		"Serial": c.serial,
		"User":   user,
		"Acl":    role,
	}).Debug("User ACL set.")
	return err
}

// Applies user bay access to each blade, interconnect,
// see applyAddLdapGroupBayAccess() for details.
func (c *C7000) applyAddUserBayAccess(user string) (err error) {

	//setup blade bays payload
	bladebays := bladeBays{}
	for b := 1; b <= 16; b++ {
		baynumber := bayNumber{Text: b}
		access := access{Text: true}
		blade := blade{Hpoa: "hpoa.xsd", BayNumber: baynumber, Access: access}
		bladebays.Blade = append(bladebays.Blade, blade)
	}

	//setup interconnect tray bays payload
	interconnecttraybays := interconnectTrayBays{}
	for t := 1; t <= 8; t++ {
		access := access{Text: true}
		baynumber := bayNumber{Text: t}
		interconnecttray := interconnectTray{Hpoa: "hpoa.xsd", Access: access, BayNumber: baynumber}
		interconnecttraybays.InterconnectTray = append(interconnecttraybays.InterconnectTray, interconnecttray)
	}

	//setup the bays payload
	bayz := bays{
		Hpoa:                 "hpoa.xsd",
		OaAccess:             oaAccess{Text: true},
		BladeBays:            bladebays,
		InterconnectTrayBays: interconnecttraybays,
	}

	payload := addUserBayAccess{
		Username: Username{Text: user},
		Bays:     bayz,
	}

	statusCode, _, err := c.postXML(payload)
	if statusCode != 200 || err != nil {
		log.WithFields(log.Fields{
			"step":       "applyAddUserBayAccess",
			"IP":         c.ip,
			"Model":      c.BmcType(),
			"Serial":     c.serial,
			"statusCode": statusCode,
			"Error":      err,
		}).Warn("LDAP applyAddUserBayAccess apply request returned non 200.")
		return err
	}

	log.WithFields(log.Fields{
		"step":   "applyAddUserBayAccess",
		"IP":     c.ip,
		"Model":  c.BmcType(),
		"Serial": c.serial,
		"user":   user,
	}).Debug("User account related interconnect and bay ACLs added.")
	return err
}

// Applies ntp parameters
// 1. SOAP call to set the NTP server params
// 2. SOAP call to set TZ
// 1.
// <hpoa:configureNtp>
//   <hpoa:ntpPrimary>ntp0.example.com</hpoa:ntpPrimary>
//   <hpoa:ntpSecondary>ntp1.example.com</hpoa:ntpSecondary>
//   <hpoa:ntpPoll>720</hpoa:ntpPoll>
//  </hpoa:configureNtp>
// 2.
// <hpoa:setEnclosureTimeZone>
//  <hpoa:timeZone>CET</hpoa:timeZone>
// </hpoa:setEnclosureTimeZone>
func (c *C7000) applyNtpParams(cfg *cfgresources.Ntp) (err error) {

	if cfg.Server1 == "" {
		log.WithFields(log.Fields{
			"step":  "applyNtpParams",
			"Model": c.BmcType(),
		}).Warn("NTP resource expects parameter: server1.")
		return
	}

	if cfg.Timezone == "" {
		log.WithFields(log.Fields{
			"step":  "applyNtpParams",
			"Model": c.BmcType(),
		}).Warn("NTP resource expects parameter: timezone.")
		return
	}

	if cfg.Enable != true {
		log.WithFields(log.Fields{
			"step":  "applyNtpParams",
			"Model": c.BmcType(),
		}).Debug("Ntp resource declared with enable: false.")
		return
	}

	//setup ntp XML payload
	ntppoll := NtpPoll{Text: "720"} //default period to poll the NTP server
	primaryServer := NtpPrimary{Text: cfg.Server1}
	secondaryServer := NtpSecondary{Text: cfg.Server2}
	payload := configureNtp{NtpPrimary: primaryServer, NtpSecondary: secondaryServer, NtpPoll: ntppoll}

	//fmt.Printf("%s\n", output)
	statusCode, _, err := c.postXML(payload)
	if err != nil || statusCode != 200 {
		log.WithFields(log.Fields{
			"step":       "applyNtpParams",
			"IP":         c.ip,
			"Model":      c.BmcType(),
			"Serial":     c.serial,
			"StatusCode": statusCode,
			"Error":      err,
		}).Warn("NTP apply request returned non 200.")
		return err
	}

	err = c.applyNtpTimezoneParam(cfg.Timezone)
	if err != nil {
		log.WithFields(log.Fields{
			"step":   "applyNtpParams",
			"IP":     c.ip,
			"Model":  c.BmcType(),
			"Serial": c.serial,
			"Error":  err,
		}).Warn("Unable to apply NTP timezone config.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":     c.ip,
		"Model":  c.BmcType(),
		"Serial": c.serial,
	}).Debug("Date and time config applied.")
	return err
}

//applies timezone
// TODO: validate timezone string.
func (c *C7000) applyNtpTimezoneParam(timezone string) (err error) {

	//setup timezone XML payload
	payload := setEnclosureTimeZone{Timezone: timeZone{Text: timezone}}

	statusCode, _, err := c.postXML(payload)
	if err != nil || statusCode != 200 {
		log.WithFields(log.Fields{
			"step":       "applyNtpTimezoneParam",
			"IP":         c.ip,
			"Model":      c.BmcType(),
			"Serial":     c.serial,
			"Error":      err,
			"StatusCode": statusCode,
		}).Warn("NTP applyNtpTimezoneParam request returned non 200.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":     c.ip,
		"Model":  c.BmcType(),
		"Serial": c.serial,
	}).Debug("Timezone config applied.")
	return err
}

// Applies syslog parameters
// 1. set syslog server
// 2. set syslog port
// 3. enable syslog
// theres no option to set the port
func (c *C7000) applySyslogParams(cfg *cfgresources.Syslog) (err error) {

	var port int
	if cfg.Server == "" {
		log.WithFields(log.Fields{
			"step":   "applySyslogParams",
			"IP":     c.ip,
			"Model":  c.BmcType(),
			"Serial": c.serial,
		}).Warn("Syslog resource expects parameter: Server.")
		return
	}

	if cfg.Port == 0 {
		log.WithFields(log.Fields{
			"step":   "applySyslogParams",
			"IP":     c.ip,
			"Model":  c.BmcType(),
			"Serial": c.serial,
		}).Debug("Syslog resource port set to default: 514.")
		port = 514
	} else {
		port = cfg.Port
	}

	if cfg.Enable != true {
		log.WithFields(log.Fields{
			"step":   "applySyslogParams",
			"IP":     c.ip,
			"Model":  c.BmcType(),
			"Serial": c.serial,
		}).Debug("Syslog resource declared with enable: false.")
	}

	c.applySyslogServer(cfg.Server)
	c.applySyslogPort(port)
	c.applySyslogEnabled(cfg.Enable)

	log.WithFields(log.Fields{
		"IP":     c.ip,
		"Model":  c.BmcType(),
		"Serial": c.serial,
	}).Debug("Syslog config applied.")
	return err
}

// Sets syslog server
// <hpoa:setRemoteSyslogServer>
//  <hpoa:server>foobar</hpoa:server>
// </hpoa:setRemoteSyslogServer>
func (c *C7000) applySyslogServer(server string) {

	payload := SetRemoteSyslogServer{Server: server}
	statusCode, _, err := c.postXML(payload)
	if err != nil || statusCode != 200 {
		log.WithFields(log.Fields{
			"step":       "applySyslogServer",
			"IP":         c.ip,
			"Model":      c.BmcType(),
			"Serial":     c.serial,
			"Error":      err,
			"StatusCode": statusCode,
		}).Warn("Syslog set server request returned non 200.")
		return
	}

	log.WithFields(log.Fields{
		"IP":     c.ip,
		"Model":  c.BmcType(),
		"Serial": c.serial,
	}).Debug("Syslog server set.")
	return
}

// Sets syslog port
// <hpoa:setRemoteSyslogPort>
//  <hpoa:port>514</hpoa:port>
// </hpoa:setRemoteSyslogPort>
func (c *C7000) applySyslogPort(port int) {
	payload := SetRemoteSyslogPort{Port: port}
	statusCode, _, err := c.postXML(payload)
	if err != nil || statusCode != 200 {
		log.WithFields(log.Fields{
			"step":       "applySyslogPort",
			"IP":         c.ip,
			"Model":      c.BmcType(),
			"Serial":     c.serial,
			"Error":      err,
			"StatusCode": statusCode,
		}).Warn("Syslog set port request returned non 200.")
		return
	}

	log.WithFields(log.Fields{
		"IP":     c.ip,
		"Model":  c.BmcType(),
		"Serial": c.serial,
	}).Debug("Syslog port set.")
	return
}

// Enables syslogging
// <hpoa:setRemoteSyslogEnabled>
//  <hpoa:enabled>true</hpoa:enabled>
// </hpoa:setRemoteSyslogEnabled>
func (c *C7000) applySyslogEnabled(enabled bool) {

	payload := SetRemoteSyslogEnabled{Enabled: enabled}
	statusCode, _, err := c.postXML(payload)
	if err != nil || statusCode != 200 {
		log.WithFields(log.Fields{
			"step":       "SetRemoteSyslogEnabled",
			"IP":         c.ip,
			"Model":      c.BmcType(),
			"Serial":     c.serial,
			"Error":      err,
			"StatusCode": statusCode,
		}).Warn("Syslog enable request returned non 200.")
		return
	}

	log.WithFields(log.Fields{
		"IP":     c.ip,
		"Model":  c.BmcType(),
		"Serial": c.serial,
	}).Debug("Syslog enabled.")
	return

}
