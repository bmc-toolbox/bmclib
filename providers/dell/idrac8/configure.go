package idrac8

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/internal/helper"
	log "github.com/sirupsen/logrus"
)

// This ensures the compiler errors if this type is missing
// a method that should be implmented to satisfy the Configure interface.
var _ devices.Configure = (*IDrac8)(nil)

// Resources returns a slice of supported resources and
// the order they are to be applied in.
func (i *IDrac8) Resources() []string {
	return []string{
		"user",
		"syslog",
		"network",
		"ntp",
		"ldap",
		"ldap_group",
		"https_cert",
	}
}

// ApplyCfg implements the Bmc interface
// this is to be deprecated.
func (i *IDrac8) ApplyCfg(config *cfgresources.ResourcesConfig) (err error) {
	return err
}

// SetLicense implements the Configure interface.
func (i *IDrac8) SetLicense(cfg *cfgresources.License) (err error) {
	return err
}

// Bios implements the Configure interface.
func (i *IDrac8) Bios(cfg *cfgresources.Bios) (err error) {
	return err
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

// User applies the User configuration resource,
// if the user exists, it updates the users password,
// User implements the Configure interface.
// Iterate over iDrac users and adds/removes/modifies user accounts
func (i *IDrac8) User(cfgUsers []*cfgresources.User) (err error) {

	err = i.validateUserCfg(cfgUsers)
	if err != nil {
		msg := "User config validation failed."
		log.WithFields(log.Fields{
			"step":  "applyUserParams",
			"IP":    i.ip,
			"Model": i.BmcType(),
			"Error": err,
		}).Warn(msg)
		return errors.New(msg)
	}

	idracUsers, err := i.queryUsers()
	if err != nil {
		msg := "Unable to query existing users"
		log.WithFields(log.Fields{
			"step":  "applyUserParams",
			"IP":    i.ip,
			"Model": i.BmcType(),
			"Error": err,
		}).Warn(msg)
		return errors.New(msg)
	}

	////for each configuration user
	for _, cfgUser := range cfgUsers {

		userID, userInfo, uExists := userInIdrac(cfgUser.Name, idracUsers)
		//user to be added/updated
		if cfgUser.Enable {

			//new user to be added
			if uExists == false {
				userID, userInfo, err = getEmptyUserSlot(idracUsers)
				if err != nil {
					log.WithFields(log.Fields{
						"IP":    i.ip,
						"Model": i.BmcType(),
						"step":  helper.WhosCalling(),
						"User":  cfgUser.Name,
						"Error": err,
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

			err = i.putUser(userID, userInfo)
			if err != nil {
				log.WithFields(log.Fields{
					"IP":    i.ip,
					"Model": i.BmcType(),
					"step":  helper.WhosCalling(),
					"User":  cfgUser.Name,
					"Error": err,
				}).Warn("Add/Update user request failed.")
				continue
			}

		} // end if cfgUser.Enable

		//if the user exists but is disabled in our config, remove the user
		if cfgUser.Enable == false && uExists == true {

			userInfo.Enable = "Disabled"
			userInfo.SolEnable = "Disabled"
			userInfo.UserName = cfgUser.Name
			userInfo.Privilege = "0"
			userInfo.IpmiLanPrivilege = "No Access"

			err = i.putUser(userID, userInfo)
			if err != nil {
				log.WithFields(log.Fields{
					"IP":    i.ip,
					"Model": i.BmcType(),
					"step":  helper.WhosCalling(),
					"User":  cfgUser.Name,
					"Error": err,
				}).Warn("Disable user request failed.")
			}
		}

		log.WithFields(log.Fields{
			"IP":    i.ip,
			"Model": i.BmcType(),
			"User":  cfgUser.Name,
		}).Debug("User parameters applied.")
	}

	return err
}

// Syslog applies the Syslog configuration resource
// Syslog implements the Configure interface
func (i *IDrac8) Syslog(cfg *cfgresources.Syslog) (err error) {

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
	response, _, err := i.put(endpoint, payload)
	if err != nil {
		log.WithFields(log.Fields{
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"response": string(response),
		}).Warn("PUT request failed.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":    i.ip,
		"Model": i.BmcType(),
	}).Debug("Syslog parameters applied.")

	return err
}

// Ntp applies NTP configuration params
// Ntp implements the Configure interface.
func (i *IDrac8) Ntp(cfg *cfgresources.Ntp) (err error) {

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
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"response": string(response),
		}).Warn("GET request failed.")
	}

	log.WithFields(log.Fields{
		"IP":    i.ip,
		"Model": i.BmcType(),
	}).Debug("NTP servers param applied.")

}

// Ldap applies LDAP configuration params.
// Ldap implements the Configure interface.
func (i *IDrac8) Ldap(cfg *cfgresources.Ldap) error {

	if cfg.Server == "" {
		msg := "Ldap resource parameter Server required but not declared."
		log.WithFields(log.Fields{
			"step": "applyLdapServerParam",
		}).Warn(msg)
		return errors.New(msg)
	}

	endpoint := fmt.Sprintf("data?set=xGLServer:%s", cfg.Server)
	response, err := i.get(endpoint, nil)
	if err != nil {
		msg := "Request to set ldap server failed."
		log.WithFields(log.Fields{
			"IP":       i.ip,
			"Model":    i.BmcType(),
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"response": string(response),
		}).Warn(msg)
		return errors.New(msg)
	}

	err = i.applyLdapSearchFilterParam(cfg)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"IP":    i.ip,
		"Model": i.BmcType(),
	}).Debug("Ldap server param set.")
	return nil
}

// Applies ldap search filter param.
// set=xGLSearchFilter:objectClass\=posixAccount
func (i *IDrac8) applyLdapSearchFilterParam(cfg *cfgresources.Ldap) error {

	if cfg.SearchFilter == "" {
		msg := "Ldap resource parameter SearchFilter required but not declared."
		log.WithFields(log.Fields{
			"step": "applyLdapSearchFilterParam",
		}).Warn(msg)
		return errors.New(msg)
	}

	endpoint := fmt.Sprintf("data?set=xGLSearchFilter:%s", escapeLdapString(cfg.SearchFilter))
	response, err := i.get(endpoint, nil)
	if err != nil {
		msg := "request to set ldap search filter failed."
		log.WithFields(log.Fields{
			"IP":       i.ip,
			"Model":    i.BmcType(),
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"response": string(response),
		}).Warn(msg)
		return errors.New(msg)
	}

	log.WithFields(log.Fields{
		"IP":    i.ip,
		"Model": i.BmcType(),
	}).Debug("Ldap search filter param applied.")
	return nil
}

// LdapGroup applies LDAP Group/Role related configuration
// LdapGroup implements the Configure interface.
// nolint: gocyclo
func (i *IDrac8) LdapGroup(cfgGroup []*cfgresources.LdapGroup, cfgLdap *cfgresources.Ldap) (err error) {

	groupID := 1

	//set to decide what privileges the group should have
	//497 == operator
	//511 == administrator (full privileges)
	privID := "0"

	//groupPrivilegeParam is populated per group and is passed to i.applyLdapRoleGroupPrivParam
	groupPrivilegeParam := ""

	//first some preliminary checks
	if cfgLdap.Port == 0 {
		msg := "Ldap resource parameter Port required but not declared"
		log.WithFields(log.Fields{
			"step": "applyLdapRoleGroupPrivParam",
		}).Warn(msg)
		return errors.New(msg)
	}

	if cfgLdap.BaseDn == "" {
		msg := "Ldap resource parameter BaseDn required but not declared."
		log.WithFields(log.Fields{
			"step": "applyLdapRoleGroupPrivParam",
		}).Warn(msg)
		return errors.New(msg)
	}

	if cfgLdap.UserAttribute == "" {
		msg := "Ldap resource parameter userAttribute required but not declared."
		log.WithFields(log.Fields{
			"step": "applyLdapRoleGroupPrivParam",
		}).Warn(msg)
		return errors.New(msg)
	}

	if cfgLdap.GroupAttribute == "" {
		msg := "Ldap resource parameter groupAttribute required but not declared."
		log.WithFields(log.Fields{
			"step": "applyLdapRoleGroupPrivParam",
		}).Warn(msg)
		return errors.New(msg)
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

		endpoint := fmt.Sprintf("data?set=xGLGroup%dName:%s", groupID, groupDn)
		response, err := i.get(endpoint, nil)
		if err != nil {
			log.WithFields(log.Fields{
				"IP":       i.ip,
				"Model":    i.BmcType(),
				"endpoint": endpoint,
				"step":     "applyLdapGroupParams",
				"response": string(response),
			}).Warn("GET request failed.")
			return err
		}

		log.WithFields(log.Fields{
			"IP":    i.ip,
			"Model": i.BmcType(),
			"Role":  group.Role,
		}).Debug("Ldap GroupDN config applied.")

		switch group.Role {
		case "user":
			privID = "497"
		case "admin":
			privID = "511"
		}

		groupPrivilegeParam += fmt.Sprintf("xGLGroup%dPriv:%s,", groupID, privID)
		groupID++

	}

	//set the rest of the group privileges to 0
	for i := groupID + 1; i <= 5; i++ {
		groupPrivilegeParam += fmt.Sprintf("xGLGroup%dPriv:0,", i)
	}

	err = i.applyLdapRoleGroupPrivParam(cfgLdap, groupPrivilegeParam)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":    i.ip,
			"Model": i.BmcType(),
			"step":  "applyLdapGroupParams",
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
	responseCode, responseBody, err := i.post(endpoint, []byte(payload), "")
	if err != nil || responseCode != 200 {
		log.WithFields(log.Fields{
			"IP":           i.ip,
			"Model":        i.BmcType(),
			"endpoint":     endpoint,
			"step":         helper.WhosCalling(),
			"responseCode": responseCode,
			"response":     string(responseBody),
		}).Warn("POST request failed.")
		return err
	}

	log.WithFields(log.Fields{
		"IP":    i.ip,
		"Model": i.BmcType(),
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
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"response": string(response),
		}).Warn("GET request failed.")
	}

	log.WithFields(log.Fields{
		"IP":    i.ip,
		"Model": i.BmcType(),
	}).Debug("Timezone param applied.")

}

// Network method implements the Configure interface
// applies various network parameters.
func (i *IDrac8) Network(cfg *cfgresources.Network) (reset bool, err error) {

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

	responseCode, responseBody, err := i.post(endpoint, []byte(payload), "")
	if err != nil || responseCode != 200 {
		log.WithFields(log.Fields{
			"IP":           i.ip,
			"Model":        i.BmcType(),
			"endpoint":     endpoint,
			"step":         helper.WhosCalling(),
			"responseCode": responseCode,
			"response":     string(responseBody),
		}).Warn("POST request to set Network params failed.")
		return reset, err
	}

	log.WithFields(log.Fields{
		"IP":    i.ip,
		"Model": i.BmcType(),
	}).Debug("Network config parameters applied.")
	return reset, err
}

// GenerateCSR generates a CSR request on the BMC.
func (i *IDrac8) GenerateCSR(cert *cfgresources.HTTPSCertAttributes) ([]byte, error) {

	var payload []string

	endpoint := "bindata?set"
	payload = []string{
		cert.CommonName,
		cert.OrganizationName,
		cert.OrganizationUnit,
		cert.Locality,
		cert.StateName,
		cert.CountryCode,
		strings.Join(strings.Split(cert.Email, "@"), "@040"), // heh, don't ask.
		cert.SubjectAltName,
	}

	queryString := url.QueryEscape(fmt.Sprintf("%s=serverCSR(%s)", endpoint, strings.Join(payload, ",")))

	body, err := i.get(queryString, nil)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":       i.ip,
			"Model":    i.BmcType(),
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"Error":    err,
		}).Warn("GET request failed.")
		return []byte{}, err
	}

	return body, nil
}

// UploadHTTPSCert uploads the given CRT cert,
// returns true if the BMC needs a reset.
// 1. POST upload signed x509 cert in multipart form.
// 2. POST returned resource URI
func (i *IDrac8) UploadHTTPSCert(cert []byte, fileName string) (bool, error) {

	endpoint := "sysmgmt/2012/server/transient/filestore?fileupload=true"
	endpoint += fmt.Sprintf("&ST1=%s", i.st1)

	// form params
	params := make(map[string]string)
	params["caller"] = ""
	params["pageCode"] = ""
	params["pageId"] = "2"
	params["pageName"] = ""
	params["index"] = "8"

	// setup a buffer for our multipart form
	var form bytes.Buffer
	w := multipart.NewWriter(&form)

	// write params to form
	for k, v := range params {
		_ = w.WriteField(k, v)
	}

	// setup the ssl cert part
	formWriter, err := w.CreateFormFile("serverSSLCertificate", fileName)
	if err != nil {
		return false, err
	}

	_, err = io.Copy(formWriter, bytes.NewReader(cert))
	if err != nil {
		return false, err
	}

	_ = w.WriteField("CertType", "2")

	// close multipart writer - adds the teminating boundary.
	w.Close()

	// 1. POST upload x509 cert
	status, body, err := i.post(endpoint, form.Bytes(), w.FormDataContentType())
	if err != nil || status != 201 {
		log.WithFields(log.Fields{
			"IP":       i.ip,
			"Model":    i.BmcType(),
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"status":   status,
		}).Warn("Cert form upload POST request failed, expected 201.")
		return false, err
	}

	// extract resourceURI from response
	var certStore = new(certStore)
	err = json.Unmarshal(body, certStore)
	if err != nil {
		log.WithFields(log.Fields{
			"step":  helper.WhosCalling(),
			"IP":    i.ip,
			"Model": i.BmcType(),
			"Error": err,
		}).Warn("Unable to unmarshal cert store response payload.")
		return false, err
	}

	resourceURI, err := json.Marshal(certStore.File)
	if err != nil {
		log.WithFields(log.Fields{
			"step":  helper.WhosCalling(),
			"IP":    i.ip,
			"Model": i.BmcType(),
			"Error": err,
		}).Warn("Unable to marshal cert store resource URI.")
		return false, err
	}

	// 2. POST resource URI
	endpoint = "sysmgmt/2012/server/network/ssl/cert"
	status, _, err = i.post(endpoint, []byte(resourceURI), "")
	if err != nil || status != 201 {
		log.WithFields(log.Fields{
			"IP":       i.ip,
			"Model":    i.BmcType(),
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"status":   status,
		}).Warn("Cert form upload POST request failed, expected 201.")
		return false, err
	}

	return true, err

}
