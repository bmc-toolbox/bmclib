package c7000

import (
	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
	log "github.com/sirupsen/logrus"
)

// This ensures the compiler errors if this type is missing
// a method that should be implmented to satisfy the Configure interface.
var _ devices.Configure = (*C7000)(nil)

// Resources returns a slice of supported resources and
// the order they are to be applied in.
func (c *C7000) Resources() []string {
	return []string{
		"user",
		"syslog",
		"license",
		"ntp",
		"ldap_group",
		"ldap",
	}
}

// ResourcesSetup returns
// - slice of supported one time setup resources,
//   in the order they must be applied
// ResourcesSetup implements the BmcChassisSetup interface
// see cfgresources.SetupChassis for list of setup resources.
func (c *C7000) ResourcesSetup() []string {
	return []string{
		"add_blade_bmc_admins",
		"remove_blade_bmc_users",
		"dynamicpower",
		"bladespower",
	}
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

// ApplyCfg implements the BmcChassis interface
func (c *C7000) ApplyCfg(config *cfgresources.ResourcesConfig) (err error) {
	return nil
}

// Ldap applies LDAP configuration params.
// Ldap implements the Configure interface.
//1. apply ldap group params
//2. enable ldap auth
//3. apply ldap server params
func (c *C7000) Ldap(cfg *cfgresources.Ldap) (err error) {
	err = c.applysetLdapInfo4(cfg)
	if err != nil {
		log.WithFields(log.Fields{
			"step":     "applyLdapParams",
			"resource": "Ldap",
			"IP":       c.ip,
			"Model":    c.BmcType(),
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
			"statusCode": statusCode,
			"Error":      err,
		}).Warn("Ldap applyEnableLdapAuth apply request returned non 200.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":    c.ip,
		"Model": c.BmcType(),
	}).Debug("Ldap Enabled.")
	return err
}

// LdapGroup applies LDAP Group/Role related configuration
// LdapGroup implements the Configure interface.
// Actions carried out in order
// 1.  addLdapGroup
// 2.  setLdapGroupBayACL
// 3.  addLdapGroupBayAccess (done)
func (c *C7000) LdapGroup(cfg []*cfgresources.LdapGroup, cfgLdap *cfgresources.Ldap) (err error) {

	for _, group := range cfg {

		if group.Group == "" {
			log.WithFields(log.Fields{
				"step":      "applyLdapGroupParams",
				"Model":     c.BmcType(),
				"Ldap role": group.Role,
			}).Warn("Ldap resource parameter Group required but not declared.")
			return
		}

		//0. removeLdapGroup
		if !group.Enable {
			c.applyRemoveLdapGroup(group.Group)
			if err != nil {
				log.WithFields(log.Fields{
					"step":     "applyRemoveLdapGroup",
					"resource": "Ldap",
					"IP":       c.ip,
					"Model":    c.BmcType(),
					"Error":    err,
				}).Warn("Remove Ldap Group returned error.")
				return
			}

			continue
		}

		if !c.isRoleValid(group.Role) {
			log.WithFields(log.Fields{
				"step":  "applyLdapGroupParams",
				"role":  group.Role,
				"Model": c.BmcType(),
			}).Warn("Ldap resource Role must be a valid role: admin OR user.")
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
				"Error":    err,
			}).Warn("addLdapGroup returned error.")
			return
		}

		//2. setLdapGroupBayACL
		err = c.applyLdapGroupBayACL(group.Role, group.Group)
		if err != nil {
			log.WithFields(log.Fields{
				"step":     "setLdapGroupBayACL",
				"resource": "Ldap",
				"IP":       c.ip,
				"Model":    c.BmcType(),
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
				"Error":    err,
			}).Warn("addLdapGroup returned error.")
			return
		}

		if err != nil {
			log.WithFields(log.Fields{
				"step":  "applyLdapGroupParams",
				"IP":    c.ip,
				"Role":  group.Role,
				"Model": c.BmcType(),
				"Error": err,
			}).Warn("Unable to set LdapGroup config for role.")
		}
	}

	log.WithFields(log.Fields{
		"IP":    c.ip,
		"Model": c.BmcType(),
	}).Debug("Ldap config applied")
	return
}

// LDAP remove group, soap actions in order.
// <hpoa:removeLdapGroup>
//  <hpoa:ldapGroup>bmcAdmins</hpoa:ldapGroup>
// </hpoa:removeLdapGroup>
func (c *C7000) applyRemoveLdapGroup(group string) (err error) {

	payload := removeLdapGroup{LdapGroup: ldapGroup{Text: group}}
	statusCode, _, err := c.postXML(payload)
	if statusCode == 200 || statusCode == 500 { // 500 indicates the group exists.
		log.WithFields(log.Fields{
			"step":       "applyRemoveLdapGroup",
			"IP":         c.ip,
			"Model":      c.BmcType(),
			"statusCode": statusCode,
		}).Debug("Ldap applyRemoveLdapGroup applied.")
		return nil
	}

	if statusCode >= 300 || err != nil {
		log.WithFields(log.Fields{
			"step":       "applyRemoveLdapGroup",
			"IP":         c.ip,
			"Model":      c.BmcType(),
			"statusCode": statusCode,
		}).Warn("Ldap applyRemoveLdapGroup request returned non 200.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":    c.ip,
		"Model": c.BmcType(),
		"Group": group,
	}).Debug("Ldap group removed.")
	return nil
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
			"statusCode": statusCode,
		}).Debug("Ldap applyAddLdapGroup applied.")
		return nil
	}

	if statusCode >= 300 || err != nil {
		log.WithFields(log.Fields{
			"step":       "applyAddLdapGroup",
			"IP":         c.ip,
			"Model":      c.BmcType(),
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
func (c *C7000) applyLdapGroupBayACL(role string, group string) (err error) {

	var userACL string

	if role == "admin" {
		userACL = "ADMINISTRATOR"
	} else {
		userACL = "USER"
	}

	payload := setLdapGroupBayACL{LdapGroup: ldapGroup{Text: group}, ACL: ACL{Text: userACL}}
	statusCode, _, err := c.postXML(payload)
	if statusCode != 200 || err != nil {
		log.WithFields(log.Fields{
			"step":       "applyLdapGroupBayACL",
			"IP":         c.ip,
			"Model":      c.BmcType(),
			"statusCode": statusCode,
			"Error":      err,
		}).Warn("LDAP applyLdapGroupBayACL request returned non 200.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":    c.ip,
		"Model": c.BmcType(),
		"Role":  role,
		"Group": group,
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
			"statusCode": statusCode,
			"Error":      err,
		}).Warn("LDAP applyAddLdapGroupBayAccess apply request returned non 200.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":    c.ip,
		"Model": c.BmcType(),
		"Group": group,
	}).Debug("Ldap interconnect and bay ACLs added.")
	return err
}

// User applies the User configuration resource,
// if the user exists, it updates the users password,
// User implements the Configure interface.
func (c *C7000) User(users []*cfgresources.User) (err error) {

	for _, cfg := range users {
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
				"IP":    c.ip,
				"Model": c.BmcType(),
				"User":  cfg.Name,
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
				"Return code": statusCode,
			}).Debug("User already exists, setting password.")

			//update user password
			err := c.setUserPassword(cfg.Name, cfg.Password)
			if err != nil {
				return err
			}

			//update user acl
			err = c.setUserACL(cfg.Name, cfg.Role)
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
			"IP":    c.ip,
			"Model": c.BmcType(),
			"user":  cfg.Name,
		}).Debug("User cfg applied.")

	}
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
			"return code": statusCode,
			"Error":       err,
		}).Warn("Unable to set user password.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":    c.ip,
		"Model": c.BmcType(),
		"user":  user,
	}).Debug("User password set.")
	return err
}

func (c *C7000) setUserACL(user string, role string) (err error) {

	var aclRole string
	if role == "admin" {
		aclRole = "ADMINISTRATOR"
	} else {
		aclRole = "OPERATOR"
	}

	u := Username{Text: user}
	a := ACL{Text: aclRole}

	payload := SetUserBayACL{Username: u, ACL: a}

	statusCode, _, err := c.postXML(payload)
	if err != nil {
		log.WithFields(log.Fields{
			"step":        "setUserACL",
			"user":        user,
			"ACL":         role,
			"IP":          c.ip,
			"Model":       c.BmcType(),
			"return code": statusCode,
			"Error":       err,
		}).Warn("Unable to set user ACL.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":    c.ip,
		"Model": c.BmcType(),
		"User":  user,
		"ACL":   role,
	}).Debug("User ACL set.")
	return err
}

// Applies user bay access to each blade, interconnect,
// see applyAddLdapGroupBayAccess() for details.
func (c *C7000) applyAddUserBayAccess(user string) (err error) {

	//The c7000 wont allow changes to the bay acls for the reserved Administrator user.
	if user == "Administrator" {
		return nil
	}

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
			"statusCode": statusCode,
			"Error":      err,
		}).Warn("LDAP applyAddUserBayAccess apply request returned non 200.")
		return err
	}

	log.WithFields(log.Fields{
		"step":  "applyAddUserBayAccess",
		"IP":    c.ip,
		"Model": c.BmcType(),
		"user":  user,
	}).Debug("User account related interconnect and bay ACLs added.")
	return err
}

// Ntp applies NTP configuration params
// Ntp implements the Configure interface.
//
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
func (c *C7000) Ntp(cfg *cfgresources.Ntp) (err error) {

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
			"StatusCode": statusCode,
			"Error":      err,
		}).Warn("NTP apply request returned non 200.")
		return err
	}

	err = c.applyNtpTimezoneParam(cfg.Timezone)
	if err != nil {
		log.WithFields(log.Fields{
			"step":  "applyNtpParams",
			"IP":    c.ip,
			"Model": c.BmcType(),
			"Error": err,
		}).Warn("Unable to apply NTP timezone config.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":    c.ip,
		"Model": c.BmcType(),
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
			"Error":      err,
			"StatusCode": statusCode,
		}).Warn("NTP applyNtpTimezoneParam request returned non 200.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":    c.ip,
		"Model": c.BmcType(),
	}).Debug("Timezone config applied.")
	return err
}

// Syslog applies the Syslog configuration resource
// Syslog implements the Configure interface
// Applies syslog parameters
// 1. set syslog server
// 2. set syslog port
// 3. enable syslog
// theres no option to set the port
func (c *C7000) Syslog(cfg *cfgresources.Syslog) (err error) {

	var port int
	if cfg.Server == "" {
		log.WithFields(log.Fields{
			"step":  "applySyslogParams",
			"IP":    c.ip,
			"Model": c.BmcType(),
		}).Warn("Syslog resource expects parameter: Server.")
		return
	}

	if cfg.Port == 0 {
		log.WithFields(log.Fields{
			"step":  "applySyslogParams",
			"IP":    c.ip,
			"Model": c.BmcType(),
		}).Debug("Syslog resource port set to default: 514.")
		port = 514
	} else {
		port = cfg.Port
	}

	if cfg.Enable != true {
		log.WithFields(log.Fields{
			"step":  "applySyslogParams",
			"IP":    c.ip,
			"Model": c.BmcType(),
		}).Debug("Syslog resource declared with enable: false.")
	}

	c.applySyslogServer(cfg.Server)
	c.applySyslogPort(port)
	c.applySyslogEnabled(cfg.Enable)

	log.WithFields(log.Fields{
		"IP":    c.ip,
		"Model": c.BmcType(),
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
			"Error":      err,
			"StatusCode": statusCode,
		}).Warn("Syslog set server request returned non 200.")
		return
	}

	log.WithFields(log.Fields{
		"IP":    c.ip,
		"Model": c.BmcType(),
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
			"Error":      err,
			"StatusCode": statusCode,
		}).Warn("Syslog set port request returned non 200.")
		return
	}

	log.WithFields(log.Fields{
		"IP":    c.ip,
		"Model": c.BmcType(),
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
			"Error":      err,
			"StatusCode": statusCode,
		}).Warn("Syslog enable request returned non 200.")
		return
	}

	log.WithFields(log.Fields{
		"IP":    c.ip,
		"Model": c.BmcType(),
	}).Debug("Syslog enabled.")
	return

}

// Network method implements the Configure interface
func (c *C7000) Network(cfg *cfgresources.Network) error {
	return nil
}

// SetLicense implements the Configure interface
func (c *C7000) SetLicense(*cfgresources.License) error {
	return nil
}

// Bios method implements the Configure interface
func (c *C7000) Bios(cfg *cfgresources.Bios) error {
	return nil
}
