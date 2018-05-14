package c7000

import (
	"encoding/xml"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/ncode/bmc/cfgresources"
	"reflect"
	"strings"
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
					}).Warn("Unable to set Syslog config.")
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
					}).Warn("Unable to set NTP config.")
				}
			case "Ldap":
				ldapCfg := cfg.Field(r).Interface().(*cfgresources.Ldap)
				c.applyLdapParams(ldapCfg)
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
func (c *C7000) applyLdapParams(cfg *cfgresources.Ldap) {

	err := c.applyLdapGroupParams(cfg)
	if err != nil {
		log.WithFields(log.Fields{
			"step":     "applyLdapParams",
			"resource": "Ldap",
			"IP":       c.ip,
			"Error":    err,
		}).Warn("applyLdapParams returned error.")
		return
	}

	err = c.applysetLdapInfo4(cfg)
	if err != nil {
		log.WithFields(log.Fields{
			"step":     "applyLdapParams",
			"resource": "Ldap",
			"IP":       c.ip,
			"Error":    err,
		}).Warn("applyLdapParams returned error.")
		return
	}

	err = c.applyEnableLdapAuth(cfg.Enable)
	if err != nil {
		log.WithFields(log.Fields{
			"step":     "applyLdapParams",
			"resource": "Ldap",
			"IP":       c.ip,
			"Error":    err,
		}).Warn("applyLdapParams returned error.")
		return
	}

}

// Apply Ldap server config params
// <hpoa:setLdapInfo4>
//   <hpoa:directoryServerAddress>example.com</hpoa:directoryServerAddress>
//   <hpoa:directoryServerSslPort>636</hpoa:directoryServerSslPort>
//   <hpoa:directoryServerGCPort>0</hpoa:directoryServerGCPort>
//   <hpoa:userNtAccountNameMapping>false</hpoa:userNtAccountNameMapping>
//   <hpoa:enableServiceAccount>false</hpoa:enableServiceAccount>
//   <hpoa:serviceAccountName></hpoa:serviceAccountName>
//   <hpoa:serviceAccountPassword></hpoa:serviceAccountPassword>
//   <hpoa:searchContexts xmlns:hpoa="hpoa.xsd">
//    <hpoa:searchContext>ou=People,dc=example,dc=com</hpoa:searchContext>
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
			"step": "applysetLdapInfo4",
		}).Warn("Ldap resource parameter Server required but not declared.")
		return err
	}

	if cfg.Port == 0 {
		log.WithFields(log.Fields{
			"step": "applysetLdapInfo4",
		}).Fatal("Ldap resource parameter Port required but not declared.")
		return err
	}

	if cfg.Group == "" {
		log.WithFields(log.Fields{
			"step": "applysetLdapInfo4",
		}).Fatal("Ldap resource parameter Group required but not declared.")
		return err
	}

	if cfg.BaseDn == "" {
		log.WithFields(log.Fields{
			"step": "applysetLdapInfo4",
		}).Fatal("Ldap resource parameter GroupBaseDn required but not declared.")
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

	doc := wrapXML(payload, c.XmlToken)
	output, err := xml.MarshalIndent(doc, "  ", "    ")
	if err != nil {
		log.WithFields(log.Fields{
			"step":  "applysetLdapInfo4",
			"Error": err,
		}).Warn("Unable to marshal ldap payload.")
		return err
	}

	// A hack to declare self closing xml tags, until https://github.com/golang/go/issues/21399 is fixed.
	output = []byte(strings.Replace(string(output), "<hpoa:searchContext></hpoa:searchContext>", "<hpoa:searchContext/>", -1))
	statusCode, _, err := c.postXML(output)
	if statusCode != 200 || err != nil {
		log.WithFields(log.Fields{
			"step":       "applysetLdapInfo4",
			"statusCode": statusCode,
			"Error":      err,
		}).Warn("Ldap applysetLdapInfo4 apply request returned non 200.")
		return err
	}

	return err

}

// <hpoa:enableLdapAuthentication>
//  <hpoa:enableLdap>true</hpoa:enableLdap>
//  <hpoa:enableLocalUsers>true</hpoa:enableLocalUsers>
// </hpoa:enableLdapAuthentication>
func (c *C7000) applyEnableLdapAuth(enable bool) (err error) {

	payload := enableLdapAuthentication{EnableLdap: enable, EnableLocalUsers: true}
	doc := wrapXML(payload, c.XmlToken)
	output, err := xml.MarshalIndent(doc, "  ", "    ")
	if err != nil {
		log.WithFields(log.Fields{
			"step":  "applyEnableLdapAuth",
			"Error": err,
		}).Warn("Unable to marshal ldap payload.")
		return err
	}

	statusCode, _, err := c.postXML(output)
	if statusCode != 200 || err != nil {
		log.WithFields(log.Fields{
			"step":       "applyEnableLdapAuth",
			"statusCode": statusCode,
			"Error":      err,
		}).Warn("Ldap applyEnableLdapAuth apply request returned non 200.")
		return err
	}

	return err
}

// Actions carried out in order
// 1.  addLdapGroup
// 2.  setLdapGroupBayAcl
// 3.  addLdapGroupBayAccess (done)
func (c *C7000) applyLdapGroupParams(cfg *cfgresources.Ldap) (err error) {

	if cfg.Role == "" {
		log.WithFields(log.Fields{
			"step": "apply-ldap-cfg",
		}).Warn("Ldap resource parameter Role required but not declared.")
		return
	}

	if !c.isRoleValid(cfg.Role) {
		log.WithFields(log.Fields{
			"step": "apply-ldap-cfg",
			"role": cfg.Role,
		}).Warn("Ldap resource Role must be a valid role: admin OR user.")
		return
	}

	if cfg.Group == "" {
		log.WithFields(log.Fields{
			"step": "apply-ldap-cfg",
		}).Warn("Ldap resource parameter Group required but not declared.")
		return
	}

	//1. addLdapGroup
	err = c.applyAddLdapGroup(cfg.Group)
	if err != nil {
		log.WithFields(log.Fields{
			"step":     "addLdapGroup",
			"resource": "Ldap",
			"IP":       c.ip,
			"Error":    err,
		}).Warn("addLdapGroup returned error.")
		return
	}

	//2. setLdapGroupBayAcl
	err = c.applyLdapGroupBayAcl(cfg.Role, cfg.Group)
	if err != nil {
		log.WithFields(log.Fields{
			"step":     "setLdapGroupBayAcl",
			"resource": "Ldap",
			"IP":       c.ip,
			"Error":    err,
		}).Warn("addLdapGroup returned error.")
		return
	}

	//3. applyAddLdapGroupBayAccess
	err = c.applyAddLdapGroupBayAccess(cfg.Group)
	if err != nil {
		log.WithFields(log.Fields{
			"step":     "applyAddLdapGroupBayAccess",
			"resource": "Ldap",
			"IP":       c.ip,
			"Error":    err,
		}).Warn("addLdapGroup returned error.")
		return
	}

	return
}

// LDAP setup group, soap actions in order.
// <hpoa:addLdapGroup>
//  <hpoa:ldapGroup>bmcAdmins</hpoa:ldapGroup>
// </hpoa:addLdapGroup>
func (c *C7000) applyAddLdapGroup(group string) (err error) {

	payload := addLdapGroup{LdapGroup: ldapGroup{Text: group}}
	doc := wrapXML(payload, c.XmlToken)
	output, err := xml.MarshalIndent(doc, "  ", "    ")
	if err != nil {
		log.WithFields(log.Fields{
			"step":  "applyAddLdapGroup",
			"Error": err,
		}).Warn("Unable to marshal ldap payload.")
		return err
	}

	statusCode, _, err := c.postXML(output)
	if statusCode == 200 || statusCode == 500 { // 500 indicates the group exists.
		log.WithFields(log.Fields{
			"step":       "applyAddLdapGroup",
			"statusCode": statusCode,
		}).Debug("Ldap applyAddLdapGroup applied.")
		return nil
	}

	if statusCode >= 300 || err != nil {
		log.WithFields(log.Fields{
			"step":       "applyAddLdapGroup",
			"statusCode": statusCode,
		}).Warn("Ldap applyAddLdapGroup request returned non 200.")
		return err
	}

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

	payload := setLdapGroupBayAcl{LdapGroup: ldapGroup{Text: group}, Acl: acl{Text: userAcl}}

	doc := wrapXML(payload, c.XmlToken)
	output, err := xml.MarshalIndent(doc, "  ", "    ")
	if err != nil {
		log.WithFields(log.Fields{
			"step":  "applyLdapGroupBayAcl",
			"Error": err,
		}).Fatal("Unable to marshal ldap payload.")
	}

	statusCode, _, err := c.postXML(output)
	if statusCode != 200 || err != nil {
		log.WithFields(log.Fields{
			"step":       "applyLdapGroupBayAcl",
			"statusCode": statusCode,
			"Error":      err,
		}).Warn("LDAP applyLdapGroupBayAcl request returned non 200.")
		return err
	}

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

	doc := wrapXML(payload, c.XmlToken)
	output, err := xml.MarshalIndent(doc, "  ", "    ")
	if err != nil {
		log.WithFields(log.Fields{
			"step":  "applyAddLdapGroupBayAccess",
			"Error": err,
		}).Fatal("Unable to marshal ldap payload.")
	}

	statusCode, _, err := c.postXML(output)
	if statusCode != 200 || err != nil {
		log.WithFields(log.Fields{
			"step":       "applyAddLdapGroupBayAccess",
			"statusCode": statusCode,
			"Error":      err,
		}).Warn("LDAP applyAddLdapGroupBayAccess apply request returned non 200.")
		return err
	}

	return err
}

// attempts to add the user
// if the user exists, update the users password.
func (c *C7000) applyUserParams(cfg *cfgresources.User) (err error) {
	// as of now we care to only set the admin role.
	// this needs to be updated to support various roles.
	validRole := "admin"

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

	if cfg.Role != validRole {
		log.WithFields(log.Fields{
			"step": "apply-user-cfg",
		}).Fatal("User resource Role must be declared and a valid role: admin.")
	}

	username := Username{Text: cfg.Name}
	password := Password{Text: cfg.Password}
	adduser := AddUser{Username: username, Password: password}

	//wrap the XML payload in the SOAP envelope
	doc := wrapXML(adduser, c.XmlToken)
	output, err := xml.MarshalIndent(doc, "  ", "    ")
	if err != nil {
		log.WithFields(log.Fields{
			"step":  "apply-user-cfg",
			"user":  cfg.Name,
			"Error": err,
		}).Fatal("Unable to marshal user payload.")
	}

	statusCode, _, err := c.postXML(output)
	if err != nil {
		return err
	}

	//user exists
	if statusCode == 400 {
		log.WithFields(log.Fields{
			"step":        "apply-user-cfg",
			"user":        cfg.Name,
			"Return code": statusCode,
		}).Debug("User already exists, setting password.")

		//update user password
		err := c.setUserPassword(cfg.Name, cfg.Password)
		if err != nil {
			return err
		}
	}

	log.WithFields(log.Fields{
		"step": "apply-user-cfg",
		"user": cfg.Name,
	}).Debug("User cfg applied.")

	return err
}

func (c *C7000) setUserPassword(user string, password string) (err error) {

	u := Username{Text: user}
	p := Password{Text: password}
	setuserpassword := SetUserPassword{Username: u, Password: p}

	//wrap the XML payload in the SOAP envelope
	doc := wrapXML(setuserpassword, c.XmlToken)
	output, err := xml.MarshalIndent(doc, "  ", "    ")
	if err != nil {
		log.WithFields(log.Fields{
			"step":  "set-user-password",
			"user":  user,
			"Error": err,
		}).Fatal("Unable to set user password.")
	}

	//fmt.Printf("-->> %d\n", statusCode)
	statusCode, _, err := c.postXML(output)
	if err != nil {
		log.WithFields(log.Fields{
			"step":        "apply-user-cfg",
			"user":        user,
			"return code": statusCode,
			"Error":       err,
		}).Warn("Unable to set user password.")
		return err
	}

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

	if cfg.Enable != true {
		log.WithFields(log.Fields{
			"step": "apply-ntp-cfg",
		}).Debug("Ntp resource declared with enable: false.")
		return
	}

	//setup ntp XML payload
	ntppoll := NtpPoll{Text: "720"} //default period to poll the NTP server
	primaryServer := NtpPrimary{Text: cfg.Server1}
	secondaryServer := NtpSecondary{Text: cfg.Server2}
	ntp := configureNtp{NtpPrimary: primaryServer, NtpSecondary: secondaryServer, NtpPoll: ntppoll}

	//wrap the XML payload in the SOAP envelope
	doc := wrapXML(ntp, c.XmlToken)
	output, err := xml.MarshalIndent(doc, "  ", "    ")
	if err != nil {
		log.WithFields(log.Fields{
			"step": "apply-ntp-cfg",
		}).Warn("Unable to marshal ntp payload.")
		return err
	}

	//fmt.Printf("%s\n", output)
	statusCode, _, err := c.postXML(output)
	if err != nil || statusCode != 200 {
		log.WithFields(log.Fields{
			"step": "apply-ntp-cfg",
		}).Warn("NTP apply request returned non 200.")
		return err
	}

	err = c.applyNtpTimezoneParam(cfg.Timezone)
	if err != nil {
		log.WithFields(log.Fields{
			"step": "apply-ntp-timezone-cfg",
		}).Warn("Unable to apply cfg.")
		return err
	}

	return err
}

//applies timezone
// TODO: validate timezone string.
func (c *C7000) applyNtpTimezoneParam(timezone string) (err error) {

	//setup timezone XML payload
	tz := setEnclosureTimeZone{Timezone: timeZone{Text: timezone}}

	//wrap the XML payload in the SOAP envelope
	doc := wrapXML(tz, c.XmlToken)
	output, err := xml.MarshalIndent(doc, "  ", "    ")
	if err != nil {
		log.WithFields(log.Fields{
			"step": "apply-ntp-timezone-cfg",
		}).Warn("Unable to marshal ntp timezone payload.")
		return err
	}

	//fmt.Printf("%s\n", output)
	statusCode, _, err := c.postXML(output)
	if err != nil || statusCode != 200 {
		log.WithFields(log.Fields{
			"step": "apply-ntp-timezone-cfg",
		}).Warn("NTP applyNtpTimezoneParam request returned non 200.")
		return err
	}
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
			"step": "apply-syslog-cfg",
		}).Warn("Syslog resource expects parameter: Server.")
		return
	}

	if cfg.Port == 0 {
		log.WithFields(log.Fields{
			"step": "apply-syslog-cfg",
		}).Debug("Syslog resource port set to default: 514.")
		port = 514
	} else {
		port = cfg.Port
	}

	if cfg.Enable != true {
		log.WithFields(log.Fields{
			"step": "apply-syslog-cfg",
		}).Debug("Syslog resource declared with enable: false.")
	}

	c.applySyslogServer(cfg.Server)
	c.applySyslogPort(port)
	c.applySyslogEnabled(cfg.Enable)

	return err
}

// Sets syslog server
// <hpoa:setRemoteSyslogServer>
//  <hpoa:server>foobar</hpoa:server>
// </hpoa:setRemoteSyslogServer>
func (c *C7000) applySyslogServer(server string) {

	payload := SetRemoteSyslogServer{Server: server}

	//wrap the XML payload in the SOAP envelope
	doc := wrapXML(payload, c.XmlToken)
	output, err := xml.MarshalIndent(doc, "  ", "    ")
	if err != nil {
		log.WithFields(log.Fields{
			"step": "applySyslogServer",
		}).Warn("Unable to marshal syslog payload.")
		return
	}

	statusCode, _, err := c.postXML(output)
	if err != nil || statusCode != 200 {
		log.WithFields(log.Fields{
			"step": "applySyslogServer",
		}).Warn("Syslog set server request returned non 200.")
		return
	}

	return
}

// Sets syslog port
// <hpoa:setRemoteSyslogPort>
//  <hpoa:port>514</hpoa:port>
// </hpoa:setRemoteSyslogPort>
func (c *C7000) applySyslogPort(port int) {
	payload := SetRemoteSyslogPort{Port: port}

	//wrap the XML payload in the SOAP envelope
	doc := wrapXML(payload, c.XmlToken)
	output, err := xml.MarshalIndent(doc, "  ", "    ")
	if err != nil {
		log.WithFields(log.Fields{
			"step": "applySyslogPort",
		}).Warn("Unable to marshal syslog payload.")
		return
	}

	statusCode, _, err := c.postXML(output)
	if err != nil || statusCode != 200 {
		log.WithFields(log.Fields{
			"step": "applySyslogPort",
		}).Warn("Syslog set port request returned non 200.")
		return
	}

	return
}

// Enables syslogging
// <hpoa:setRemoteSyslogEnabled>
//  <hpoa:enabled>true</hpoa:enabled>
// </hpoa:setRemoteSyslogEnabled>
func (c *C7000) applySyslogEnabled(enabled bool) {

	payload := SetRemoteSyslogEnabled{Enabled: enabled}

	//wrap the XML payload in the SOAP envelope
	doc := wrapXML(payload, c.XmlToken)
	output, err := xml.MarshalIndent(doc, "  ", "    ")
	if err != nil {
		log.WithFields(log.Fields{
			"step": "SetRemoteSyslogEnabled",
		}).Warn("Unable to marshal syslog payload.")
		return
	}

	statusCode, _, err := c.postXML(output)
	if err != nil || statusCode != 200 {
		log.WithFields(log.Fields{
			"step": "SetRemoteSyslogEnabled",
		}).Warn("Syslog enable request returned non 200.")
		return
	}

	return

}
