package c7000

import (
	"crypto/x509"
	"errors"
	"fmt"

	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
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
// ResourcesSetup implements the CmcSetup interface
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

// ApplyCfg implements the Cmc interface
func (c *C7000) ApplyCfg(config *cfgresources.ResourcesConfig) (err error) {
	return nil
}

// Power implemented the Configure interface
func (c *C7000) Power(cfg *cfgresources.Power) (err error) {
	return nil
}

// Applies LDAP configuration params.
// Implements the Configure interface.
// 1. Apply LDAP group params
// 2. Enable LDAP auth
// 3. Apply LDAP server params
func (c *C7000) Ldap(cfg *cfgresources.Ldap) (err error) {
	err = c.applysetLdapInfo4(cfg)
	if err != nil {
		c.log.V(1).Error(err, "applyLdapParams returned error.",
			"step", "applyLdapParams",
			"resource", "Ldap",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
		)
		return err
	}

	err = c.applyEnableLdapAuth(cfg.Enable)
	if err != nil {
		c.log.V(1).Error(err, "applyLdapParams returned error.",
			"step", "applyLdapParams",
			"resource", "Ldap",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
		)
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
		c.log.V(1).Info("Ldap resource parameter Server required but not declared.",
			"step", "applysetLdapInfo4",
			"HardwareType", c.HardwareType(),
		)
		return err
	}

	if cfg.Port == 0 {
		c.log.V(1).Info("Ldap resource parameter Port required but not declared.",
			"step", "applysetLdapInfo4",
			"HardwareType", c.HardwareType(),
		)
		return err
	}

	if cfg.BaseDn == "" {
		c.log.V(1).Info("Ldap resource parameter BaseDn required but not declared.",
			"step", "applysetLdapInfo4",
		)
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
		c.log.V(1).Info("Ldap applysetLdapInfo4 apply request returned non 200.",
			"step", "applysetLdapInfo4",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
			"statusCode", statusCode,
			"Error", internal.ErrStringOrEmpty(err),
		)
		return err
	}

	c.log.V(1).Info("Ldap Server parameters applied.",
		"IP", c.ip,
		"HardwareType", c.HardwareType(),
	)
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
		c.log.V(1).Info("Ldap applyEnableLdapAuth apply request returned non 200.",
			"step", "applyEnableLdapAuth",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
			"statusCode", statusCode,
			"Error", internal.ErrStringOrEmpty(err),
		)
		return err
	}

	c.log.V(1).Info("Ldap Enabled.",
		"IP", c.ip,
		"HardwareType", c.HardwareType(),
	)
	return err
}

// LdapGroups applies LDAP Group/Role related configuration
// LdapGroups implements the Configure interface.
// Actions carried out in order
// 1.  addLdapGroup
// 2.  setLdapGroupBayACL
// 3.  addLdapGroupBayAccess (done)
func (c *C7000) LdapGroups(cfgGroups []*cfgresources.LdapGroup, cfgLdap *cfgresources.Ldap) (err error) {
	for _, group := range cfgGroups {
		if group.Group == "" {
			c.log.V(1).Info("Ldap resource parameter Group required but not declared.",
				"step", "applyLdapGroupParams",
				"HardwareType", c.HardwareType(),
				"Ldap role", group.Role,
			)
			return
		}

		// 0. removeLdapGroup
		if !group.Enable {
			err = c.applyRemoveLdapGroup(group.Group)
			if err != nil {
				c.log.V(1).Error(err, "Remove Ldap Group returned error.",
					"step", "applyRemoveLdapGroup",
					"resource", "Ldap",
					"IP", c.ip,
					"HardwareType", c.HardwareType(),
				)
				return
			}

			continue
		}

		if !c.isRoleValid(group.Role) {
			c.log.V(1).Info("Ldap resource Role must be a valid role: admin OR user.",
				"step", "applyLdapGroupParams",
				"role", group.Role,
				"HardwareType", c.HardwareType(),
			)
			return
		}

		// 1. addLdapGroup
		err = c.applyAddLdapGroup(group.Group)
		if err != nil {
			c.log.V(1).Error(err, "addLdapGroup returned error.",
				"step", "applyAddLdapGroup",
				"resource", "Ldap",
				"IP", c.ip,
				"HardwareType", c.HardwareType(),
			)
			return
		}

		// 2. setLdapGroupBayACL
		err = c.applyLdapGroupBayACL(group.Role, group.Group)
		if err != nil {
			c.log.V(1).Error(err, "addLdapGroup returned error.",
				"step", "setLdapGroupBayACL",
				"resource", "Ldap",
				"IP", c.ip,
				"HardwareType", c.HardwareType(),
			)
			return
		}

		// 3. applyAddLdapGroupBayAccess
		err = c.applyAddLdapGroupBayAccess(group.Group)
		if err != nil {
			c.log.V(1).Error(err, "addLdapGroup returned error.",
				"step", "applyAddLdapGroupBayAccess",
				"resource", "Ldap",
				"IP", c.ip,
				"HardwareType", c.HardwareType(),
			)
			return
		}
	}

	c.log.V(1).Info("Ldap config applied",
		"IP", c.ip,
		"HardwareType", c.HardwareType(),
	)
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
		c.log.V(1).Info("Ldap applyRemoveLdapGroup applied.",
			"step", "applyRemoveLdapGroup",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
			"statusCode", statusCode,
		)
		return nil
	}

	if statusCode >= 300 || err != nil {
		c.log.V(1).Info("Ldap applyRemoveLdapGroup request returned non 200.",
			"step", "applyRemoveLdapGroup",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
			"statusCode", statusCode,
		)
		return err
	}

	c.log.V(1).Info("Ldap group removed.",
		"IP", c.ip,
		"HardwareType", c.HardwareType(),
		"Group", group,
	)
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
		c.log.V(1).Info("Ldap applyAddLdapGroup applied.",
			"step", "applyAddLdapGroup",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
			"statusCode", statusCode,
		)
		return nil
	}

	if statusCode >= 300 || err != nil {
		c.log.V(1).Info("Ldap applyAddLdapGroup request returned non 200.",
			"step", "applyAddLdapGroup",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
			"statusCode", statusCode,
		)
		return err
	}

	c.log.V(1).Info("Ldap group added.",
		"IP", c.ip,
		"HardwareType", c.HardwareType(),
	)
	return nil
}

// Applies ldap group ACL
// <hpoa:setLdapGroupBayAcl>
//  <hpoa:ldapGroup>bmcAdmins</hpoa:ldapGroup>
//  <hpoa:acl>ADMINISTRATOR</hpoa:acl>
// </hpoa:setLdapGroupBayAcl>
func (c *C7000) applyLdapGroupBayACL(role string, group string) (err error) {
	userACL := "USER"
	if role == "admin" {
		userACL = "ADMINISTRATOR"
	}

	payload := setLdapGroupBayACL{LdapGroup: ldapGroup{Text: group}, ACL: ACL{Text: userACL}}
	statusCode, _, err := c.postXML(payload)
	if statusCode != 200 || err != nil {
		if err == nil {
			err = fmt.Errorf("XML POST failed with status code %d.", statusCode)
		}

		c.log.V(1).Error(err, "applyLdapGroupBayACL(): POST request failed.",
			"step", "applyLdapGroupBayACL",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
			"StatusCode", statusCode,
		)
		return err
	}

	c.log.V(1).Info("Ldap group ACL added.",
		"IP", c.ip,
		"HardwareType", c.HardwareType(),
		"Role", role,
		"Group", group,
	)
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
	// setup blade bays payload
	bladebays := bladeBays{}
	for b := 1; b <= 16; b++ {
		baynumber := bayNumber{Text: b}
		access := access{Text: true}
		blade := blade{Hpoa: "hpoa.xsd", BayNumber: baynumber, Access: access}
		bladebays.Blade = append(bladebays.Blade, blade)
	}

	// setup interconnect tray bays payload
	interconnecttraybays := interconnectTrayBays{}
	for t := 1; t <= 8; t++ {
		access := access{Text: true}
		baynumber := bayNumber{Text: t}
		interconnecttray := interconnectTray{Hpoa: "hpoa.xsd", Access: access, BayNumber: baynumber}
		interconnecttraybays.InterconnectTray = append(interconnecttraybays.InterconnectTray, interconnecttray)
	}

	// setup the bays payload
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
		if err == nil {
			err = fmt.Errorf("POST XML request failed with status code %d.", statusCode)
		}

		c.log.V(1).Error(err, "applyAddLdapGroupBayAccess(): POST request failed.",
			"step", "applyAddLdapGroupBayAccess",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
			"StatusCode", statusCode,
		)
		return err
	}

	c.log.V(1).Info("Ldap interconnect and bay ACLs added.",
		"IP", c.ip,
		"HardwareType", c.HardwareType(),
		"Group", group,
	)
	return err
}

// Applies the User configuration resource.
// Implements the Configure interface.
// If the user exists, updates their password.
func (c *C7000) User(users []*cfgresources.User) (err error) {
	for _, cfg := range users {
		if cfg.Name == "" {
			err = errors.New("user resource expects parameter: Name")
			c.log.V(1).Error(err, "user resource expects parameter: Name",
				"step", "applyUserParams",
				"HardwareType", c.HardwareType(),
			)
			return err
		}

		if cfg.Password == "" {
			err = errors.New("user resource expects parameter: Password")
			c.log.V(1).Error(err, "user resource expects parameter: Password",
				"step", "applyUserParams",
				"HardwareType", c.HardwareType(),
			)
			return err
		}

		if !c.isRoleValid(cfg.Role) {
			err = errors.New("user resource Role must be declared and a valid role: admin")
			c.log.V(1).Error(err, "user resource Role must be declared and a valid role: admin",
				"step", "applyUserParams",
				"HardwareType", c.HardwareType(),
				"Role", cfg.Role,
			)
			return err
		}

		username := Username{Text: cfg.Name}
		password := Password{Text: cfg.Password}

		// User account is disabled? Remove them.
		if !cfg.Enable {
			payload := RemoveUser{Username: username}
			statusCode, _, _ := c.postXML(payload)

			// User doesn't exist? Nothing to do, success claimed!
			if statusCode != 400 {
				return err
			}

			c.log.V(1).Info("User removed.",
				"IP", c.ip,
				"HardwareType", c.HardwareType(),
				"User", cfg.Name,
			)

			return nil
		}

		payload := AddUser{Username: username, Password: password}
		statusCode, _, err := c.postXML(payload)
		if err != nil {
			return err
		}

		if statusCode == 400 {
			c.log.V(1).Info("User already exists, setting password.",
				"step", "applyUserParams",
				"user", cfg.Name,
				"IP", c.ip,
				"HardwareType", c.HardwareType(),
				"Return code", statusCode,
			)

			err := c.setUserPassword(cfg.Name, cfg.Password)
			if err != nil {
				return err
			}

			err = c.setUserACL(cfg.Name, cfg.Role)
			if err != nil {
				return err
			}

			err = c.applyAddUserBayAccess(cfg.Name)
			if err != nil {
				return err
			}
		}

		c.log.V(1).Info("User config applied.",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
			"user", cfg.Name,
		)
	}
	return err
}

func (c *C7000) setUserPassword(user string, password string) (err error) {
	u := Username{Text: user}
	p := Password{Text: password}
	payload := SetUserPassword{Username: u, Password: p}

	statusCode, _, err := c.postXML(payload)
	if err != nil {
		c.log.V(1).Error(err, "unable to set user password.",
			"step", "setUserPassword",
			"user", user,
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
			"StatusCode", statusCode,
		)
		return err
	}

	c.log.V(1).Info("User password set.",
		"IP", c.ip,
		"HardwareType", c.HardwareType(),
		"user", user,
	)
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
		c.log.V(1).Error(err, "unable to set user ACL.",
			"step", "setUserACL",
			"user", user,
			"ACL", role,
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
			"StatusCode", statusCode,
		)
		return err
	}

	c.log.V(1).Info("User ACL set.",
		"IP", c.ip,
		"HardwareType", c.HardwareType(),
		"User", user,
		"ACL", role,
	)
	return err
}

// Applies user bay access to each blade, interconnect,
// see applyAddLdapGroupBayAccess() for details.
func (c *C7000) applyAddUserBayAccess(user string) (err error) {
	// The c7000 won't allow changes to the bay ACLs for the reserved Administrator user.
	if user == "Administrator" {
		return nil
	}

	// setup blade bays payload
	bladebays := bladeBays{}
	for b := 1; b <= 16; b++ {
		baynumber := bayNumber{Text: b}
		access := access{Text: true}
		blade := blade{Hpoa: "hpoa.xsd", BayNumber: baynumber, Access: access}
		bladebays.Blade = append(bladebays.Blade, blade)
	}

	// setup interconnect tray bays payload
	interconnecttraybays := interconnectTrayBays{}
	for t := 1; t <= 8; t++ {
		access := access{Text: true}
		baynumber := bayNumber{Text: t}
		interconnecttray := interconnectTray{Hpoa: "hpoa.xsd", Access: access, BayNumber: baynumber}
		interconnecttraybays.InterconnectTray = append(interconnecttraybays.InterconnectTray, interconnecttray)
	}

	// setup the bays payload
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
		c.log.V(1).Error(err, "LDAP applyAddUserBayAccess apply request returned non 200.",
			"step", "applyAddUserBayAccess",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
			"statusCode", statusCode,
			"Error", internal.ErrStringOrEmpty(err),
		)
		return err
	}

	c.log.V(1).Info("User account related interconnect and bay ACLs added.",
		"step", "applyAddUserBayAccess",
		"IP", c.ip,
		"HardwareType", c.HardwareType(),
		"user", user,
	)
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
		c.log.V(1).Info("NTP resource expects parameter: server1.",
			"step", "applyNtpParams",
			"HardwareType", c.HardwareType(),
		)
		return
	}

	if cfg.Timezone == "" {
		c.log.V(1).Info("NTP resource expects parameter: timezone.",
			"step", "applyNtpParams",
			"HardwareType", c.HardwareType(),
		)
		return
	}

	if !cfg.Enable {
		c.log.V(1).Info("Ntp resource declared with enable: false.",
			"step", "applyNtpParams",
			"HardwareType", c.HardwareType(),
		)
		return
	}

	// setup ntp XML payload
	ntppoll := NtpPoll{Text: "720"} // default period to poll the NTP server
	primaryServer := NtpPrimary{Text: cfg.Server1}
	secondaryServer := NtpSecondary{Text: cfg.Server2}
	payload := configureNtp{NtpPrimary: primaryServer, NtpSecondary: secondaryServer, NtpPoll: ntppoll}

	statusCode, _, err := c.postXML(payload)
	if err != nil || statusCode != 200 {
		c.log.V(1).Info("NTP apply request returned non 200.",
			"step", "applyNtpParams",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
			"StatusCode", statusCode,
		)
		return err
	}

	err = c.applyNtpTimezoneParam(cfg.Timezone)
	if err != nil {
		c.log.V(1).Error(err, "Unable to apply NTP timezone config.",
			"step", "applyNtpParams",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
		)
		return err
	}

	c.log.V(1).Info("Date and time config applied.",
		"IP", c.ip,
		"HardwareType", c.HardwareType(),
	)
	return err
}

// TODO: validate timezone string.
func (c *C7000) applyNtpTimezoneParam(timezone string) (err error) {
	// setup timezone XML payload
	payload := setEnclosureTimeZone{Timezone: timeZone{Text: timezone}}

	statusCode, _, err := c.postXML(payload)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("POST XML request failed with status code %d.", statusCode)
		}

		c.log.V(1).Error(err, "applyNtpTimezoneParam(): POST request failed.",
			"step", "applyNtpTimezoneParam",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
			"StatusCode", statusCode,
		)
		return err
	}

	c.log.V(1).Info("Timezone config applied.",
		"IP", c.ip,
		"HardwareType", c.HardwareType(),
	)
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
		c.log.V(1).Info("Syslog resource expects parameter: Server.",
			"step", "applySyslogParams",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
		)
		return
	}

	if cfg.Port == 0 {
		c.log.V(1).Info("Syslog resource port set to default: 514.",
			"step", "applySyslogParams",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
		)
		port = 514
	} else {
		port = cfg.Port
	}

	if !cfg.Enable {
		c.log.V(1).Info("Syslog resource declared with enable: false.",
			"step", "applySyslogParams",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
		)
	}

	c.applySyslogServer(cfg.Server)
	c.applySyslogPort(port)
	c.applySyslogEnabled(cfg.Enable)

	c.log.V(1).Info("Syslog config applied.",
		"IP", c.ip,
		"HardwareType", c.HardwareType(),
	)
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
		c.log.V(1).Error(err, "Syslog set server request returned non 200.",
			"step", "applySyslogServer",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
			"Error", internal.ErrStringOrEmpty(err),
			"StatusCode", statusCode,
		)
		return
	}

	c.log.V(1).Info("Syslog server set.",
		"IP", c.ip,
		"HardwareType", c.HardwareType(),
	)
}

// Sets syslog port
// <hpoa:setRemoteSyslogPort>
//  <hpoa:port>514</hpoa:port>
// </hpoa:setRemoteSyslogPort>
func (c *C7000) applySyslogPort(port int) {
	payload := SetRemoteSyslogPort{Port: port}
	statusCode, _, err := c.postXML(payload)
	if err != nil || statusCode != 200 {
		c.log.V(1).Error(err, "Syslog set port request returned non 200.",
			"step", "applySyslogPort",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
			"Error", internal.ErrStringOrEmpty(err),
			"StatusCode", statusCode,
		)
		return
	}

	c.log.V(1).Info("Syslog port set.",
		"IP", c.ip,
		"HardwareType", c.HardwareType(),
	)
}

// Enables syslogging
// <hpoa:setRemoteSyslogEnabled>
//  <hpoa:enabled>true</hpoa:enabled>
// </hpoa:setRemoteSyslogEnabled>
func (c *C7000) applySyslogEnabled(enabled bool) {
	payload := SetRemoteSyslogEnabled{Enabled: enabled}
	statusCode, _, err := c.postXML(payload)
	if err != nil || statusCode != 200 {
		c.log.V(1).Error(err, "Syslog enable request returned non 200.",
			"step", "SetRemoteSyslogEnabled",
			"IP", c.ip,
			"HardwareType", c.HardwareType(),
			"Error", internal.ErrStringOrEmpty(err),
			"StatusCode", statusCode,
		)
		return
	}

	c.log.V(1).Info("Syslog enabled.",
		"IP", c.ip,
		"HardwareType", c.HardwareType(),
	)
}

// Network method implements the Configure interface
func (c *C7000) Network(cfg *cfgresources.Network) (bool, error) {
	return false, nil
}

// SetLicense implements the Configure interface
func (c *C7000) SetLicense(*cfgresources.License) error {
	return nil
}

// Bios method implements the Configure interface
func (c *C7000) Bios(cfg *cfgresources.Bios) error {
	return nil
}

// GenerateCSR generates a CSR request on the BMC.
// GenerateCSR implements the Configure interface.
func (c *C7000) GenerateCSR(cert *cfgresources.HTTPSCertAttributes) ([]byte, error) {
	return []byte{}, nil
}

// UploadHTTPSCert uploads the given CRT cert,
// UploadHTTPSCert implements the Configure interface.
func (c *C7000) UploadHTTPSCert(cert []byte, certFileName string, key []byte, keyFileName string) (bool, error) {
	return false, nil
}

// CurrentHTTPSCert returns the current x509 certficates configured on the BMC
// The bool value returned indicates if the BMC supports CSR generation.
// CurrentHTTPSCert implements the Configure interface.
func (c *C7000) CurrentHTTPSCert() (x []*x509.Certificate, b bool, e error) {
	return x, b, e
}
