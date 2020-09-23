package idrac9

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"

	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/internal/helper"

	"gopkg.in/go-playground/validator.v9"
)

// This ensures the compiler errors if this type is missing
// a method that should be implmented to satisfy the Configure interface.
var _ devices.Configure = (*IDrac9)(nil)

// Resources returns a slice of supported resources and
// the order they are to be applied in.
func (i *IDrac9) Resources() []string {
	return []string{
		"user",
		"syslog",
		"network",
		"ntp",
		"ldap",
		"ldap_group",
		"bios",
		"https_cert",
	}
}

// ApplyCfg implements the Bmc interface
func (i *IDrac9) ApplyCfg(config *cfgresources.ResourcesConfig) (err error) {
	return err
}

// Power implemented the Configure interface
func (i *IDrac9) Power(cfg *cfgresources.Power) (err error) {
	return err
}

// Bios sets up Bios configuration
// Bios implements the Configure interface
func (i *IDrac9) Bios(cfg *cfgresources.Bios) (err error) {

	newBiosSettings := cfg.Dell.Idrac9BiosSettings

	//validate config
	validate := validator.New()
	err = validate.Struct(newBiosSettings)
	if err != nil {
		i.log.V(1).Error(err, "Config validation failed.", "step", "applyBiosParams", "Error", err.Error())
		return err
	}

	//GET current settings
	currentBiosSettings, err := i.getBiosSettings()
	if err != nil || currentBiosSettings == nil {
		msg := "Unable to get current bios settings through redfish."
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"Model", i.HardwareType(),
			"step", helper.WhosCalling(),
			"Error", err.Error(),
		)
		return errors.New(msg)
	}

	//Compare current bios settings with our declared config.
	if *newBiosSettings != *currentBiosSettings {

		//retrieve fields that is the config to be applied
		toApplyBiosSettings, err := diffBiosSettings(newBiosSettings, currentBiosSettings)
		if err != nil {
			i.log.V(1).Error(err, "diffBiosSettings returned error.",
				"IP", i.ip,
				"Model", i.HardwareType(),
				"step", helper.WhosCalling(),
				"Error", err.Error(),
			)
			return err
		}

		i.log.V(0).Info("Bios configuration to be applied",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"step", helper.WhosCalling(),
			"Changes (Ignore empty fields)", fmt.Sprintf("%+v", toApplyBiosSettings),
		)

		//purge any existing pending bios setting jobs
		//or we will not be able to set any params
		err = i.purgeJobsForBiosSettings()
		if err != nil {
			i.log.V(1).Info("Unable to purge pending bios setting jobs.",
				"step", "applyBiosParams",
				"resource", "Bios",
				"IP", i.ip,
				"Model", i.HardwareType(),
				"Bios settings pending", err.Error(),
			)
		}

		err = i.setBiosSettings(toApplyBiosSettings)
		if err != nil {
			msg := "setBiosAttributes returned error."
			i.log.V(1).Error(err, msg,
				"IP", i.ip,
				"Model", i.HardwareType(),
				"step", helper.WhosCalling(),
				"Error", err.Error(),
			)
			return errors.New(msg)
		}

		i.log.V(0).Info("Bios configuration update job queued in iDrac.",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"step", helper.WhosCalling(),
		)

	} else {

		i.log.V(0).Info("Bios configuration is up to date.",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
	}

	return err
}

// User applies the User configuration resource,
// if the user exists, it updates the users password,
// User implements the Configure interface.
// Iterate over iDrac users and adds/removes/modifies user accounts
// nolint: gocyclo
func (i *IDrac9) User(cfgUsers []*cfgresources.User) (err error) {

	err = i.validateCfg(cfgUsers)
	if err != nil {
		msg := "Config validation failed."
		i.log.V(1).Error(err, msg,
			"step", "applyUserParams",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"Error", err.Error(),
		)
		return errors.New(msg)
	}

	idracUsers, err := i.queryUsers()
	if err != nil {
		msg := "Unable to query existing users"
		i.log.V(1).Error(err, msg,
			"step", "applyUserParams",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"Error", err.Error(),
		)
		return errors.New(msg)
	}

	//for each configuration user
	for _, cfgUser := range cfgUsers {

		userID, userInfo, uExists := userInIdrac(cfgUser.Name, idracUsers)

		//user to be added/updated
		if cfgUser.Enable {

			//new user to be added
			if uExists == false {
				userID, userInfo, err = getEmptyUserSlot(idracUsers)
				if err != nil {
					i.log.V(1).Error(err, "Unable to add new User.",
						"IP", i.ip,
						"Model", i.HardwareType(),
						"step", helper.WhosCalling(),
						"User", cfgUser.Name,
						"Error", err.Error(),
					)
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
				i.log.V(1).Error(err, "Add/Update user request failed.",
					"IP", i.ip,
					"Model", i.HardwareType(),
					"step", helper.WhosCalling(),
					"User", cfgUser.Name,
					"Error", err.Error(),
				)
				continue
			}

		} // end if cfgUser.Enable

		//if the user exists but is disabled in our config, remove the user
		if cfgUser.Enable == false && uExists == true {
			endpoint := fmt.Sprintf("sysmgmt/2017/server/user?userid=%d", userID)
			statusCode, response, err := i.delete(endpoint)
			if err != nil {
				i.log.V(1).Error(err, "Delete user request failed.",
					"IP", i.ip,
					"Model", i.HardwareType(),
					"step", helper.WhosCalling(),
					"User", cfgUser.Name,
					"Error", err.Error(),
					"StatusCode", statusCode,
					"Response", response,
				)
				continue
			}
		}

		i.log.V(1).Info("User parameters applied.",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"User", cfgUser.Name,
		)

	}

	return err
}

// Ldap applies LDAP configuration params.
// Ldap implements the Configure interface.
func (i *IDrac9) Ldap(cfg *cfgresources.Ldap) (err error) {
	params := map[string]string{
		"Enable":               "Disabled",
		"Port":                 "636",
		"UserAttribute":        "uid",
		"GroupAttribute":       "memberUid",
		"GroupAttributeIsDN":   "Disabled",
		"CertValidationEnable": "Disabled",
		"SearchFilter":         "objectClass=posixAccount",
	}

	if cfg.Server == "" {
		msg := "Ldap resource parameter Server required but not declared."
		err = errors.New(msg)
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"Model", i.HardwareType(),
			"step", helper.WhosCalling,
		)
		return err
	}

	if cfg.BaseDn == "" {
		msg := "Ldap resource parameter BaseDn required but not declared."
		err = errors.New(msg)
		i.log.V(1).Error(err, msg,
			"Model", i.HardwareType(),
			"step", helper.WhosCalling,
		)
		return err
	}

	if cfg.Enable {
		params["Enable"] = "Enabled"
	}

	if cfg.Port == 0 {
		params["Port"] = fmt.Sprint(cfg.Port)
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
		msg := "ldap params PUT request failed."
		err = errors.New(msg)
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"Model", i.HardwareType(),
			"step", helper.WhosCalling(),
			"Error", err.Error(),
		)
		return err
	}

	return err
}

// LdapGroup applies LDAP Group/Role related configuration
// LdapGroup implements the Configure interface.
// nolint: gocyclo
func (i *IDrac9) LdapGroup(cfg []*cfgresources.LdapGroup, cfgLdap *cfgresources.Ldap) (err error) {

	idracLdapRoleGroups, err := i.queryLdapRoleGroups()
	if err != nil {
		msg := "Unable to query existing users"
		err = errors.New(msg)
		i.log.V(1).Error(err, msg,
			"step", "applyUserParams",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"Error", err.Error(),
		)
		return err
	}

	//for each configuration ldap role group
	for _, cfgRole := range cfg {

		// the distinguished name of the group
		// example, if the GroupBaseDn is ou=Group,dc=example,dc=com and the Group is cn=fooUsers
		// the groupDN will be Group+GroupBaseDn = cn=fooUsers,ou=Group,dc=example,dc=com
		var groupDN = fmt.Sprintf("%s,%s", cfgRole.Group, cfgRole.GroupBaseDn)

		roleID, role, rExists := ldapRoleGroupInIdrac(groupDN, idracLdapRoleGroups)

		//role to be added/updated
		if cfgRole.Enable {

			//new role to be added
			if rExists == false {
				roleID, role, err = getEmptyLdapRoleGroupSlot(idracLdapRoleGroups)
				if err != nil {
					i.log.V(1).Error(err, "Unable to add new Ldap Role Group.",
						"IP", i.ip,
						"Model", i.HardwareType(),
						"step", helper.WhosCalling(),
						"Ldap Role Group", cfgRole.Group,
						"Role Group DN", cfgRole.Role,
						"Error", err.Error(),
					)
					continue
				}
			}

			role.DN = groupDN

			//set appropriate privileges
			if cfgRole.Role == "admin" {
				role.Privilege = "511"
			} else {
				role.Privilege = "499"
			}

			err = i.putLdapRoleGroup(roleID, role)
			if err != nil {
				i.log.V(1).Error(err, "Add/Update LDAP Role Group request failed.",
					"IP", i.ip,
					"Model", i.HardwareType(),
					"step", helper.WhosCalling(),
					"Ldap Role Group", cfgRole.Group,
					"Role Group DN", cfgRole.Role,
					"Error", err.Error(),
				)
				continue
			}

		} // end if cfgUser.Enable

		//if the role exists but is disabled in our config, remove the role
		if cfgRole.Enable == false && rExists == true {

			role.DN = ""
			role.Privilege = "0"
			err = i.putLdapRoleGroup(roleID, role)
			if err != nil {
				i.log.V(1).Error(err, "Remove LDAP Role Group request failed.",
					"IP", i.ip,
					"Model", i.HardwareType(),
					"step", helper.WhosCalling(),
					"Ldap Role Group", cfgRole.Group,
					"Role Group DN", cfgRole.Role,
					"Error", err.Error(),
				)
				continue
			}
		}

		i.log.V(1).Info("Ldap Role Group parameters applied.",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"Step", helper.WhosCalling(),
			"Ldap Role Group", cfgRole.Role,
			"Role Group DN", cfgRole.Role,
		)

	}

	return err
}

// Ntp applies NTP configuration params
// Ntp implements the Configure interface.
func (i *IDrac9) Ntp(cfg *cfgresources.Ntp) (err error) {

	var enable string

	if cfg.Enable {
		enable = "Enabled"
	} else {
		enable = "Disabled"
	}

	if cfg.Server1 == "" {
		msg := "NTP resource expects parameter: server1."
		err = errors.New(msg)
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"Model", i.HardwareType(),
			"Step", helper.WhosCalling(),
		)
		return err
	}

	if cfg.Timezone == "" {
		msg := "NTP resource expects parameter: timezone."
		err = errors.New(msg)
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"Model", i.HardwareType(),
			"Step", helper.WhosCalling(),
		)
		return err
	}

	_, validTimezone := Timezones[cfg.Timezone]
	if !validTimezone {
		msg := "NTP resource a valid timezone parameter, for valid timezones see dell/idrac9/model.go"
		err = errors.New(msg)
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"Model", i.HardwareType(),
			"step", helper.WhosCalling(),
			"Unknown Timezone", cfg.Timezone,
		)
		return err
	}

	err = i.putTimezone(Timezone{Timezone: cfg.Timezone})
	if err != nil {
		i.log.V(1).Error(err, "PUT timezone request failed.",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"step", helper.WhosCalling(),
			"Timezone", cfg.Timezone,
			"Error", err.Error(),
		)
		return err
	}

	payload := NtpConfig{
		Enable: enable,
		NTP1:   cfg.Server1,
		NTP2:   cfg.Server2,
		NTP3:   cfg.Server3,
	}

	err = i.putNtpConfig(payload)
	if err != nil {
		i.log.V(1).Error(err, "PUT Ntp request failed.",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"step", helper.WhosCalling(),
			"Error", err.Error(),
		)
		return err
	}

	i.log.V(1).Info("NTP servers param applied.",
		"IP", i.ip,
		"Model", i.HardwareType(),
	)

	return err
}

// Syslog applies the Syslog configuration resource
// Syslog implements the Configure interface
//
// As part of Syslog we enable alerts and alert filters to syslog,
// the iDrac will not send out any messages over syslog unless this is enabled,
// and since not all BMCs currently support configuring filtering for alerts,
// for now the configuration for alert filters/enabling is managed through this method.
func (i *IDrac9) Syslog(cfg *cfgresources.Syslog) (err error) {

	var port int
	enable := "Enabled"

	if cfg.Server == "" {
		i.log.V(1).Info("Syslog resource expects parameter: Server.",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return
	}

	if cfg.Port == 0 {
		i.log.V(1).Info("Syslog resource port set to default: 514.", "step", helper.WhosCalling())
		port = 514
	} else {
		port = cfg.Port
	}

	if cfg.Enable != true {
		enable = "Disabled"
		i.log.V(1).Info("Syslog resource declared with enable: false.", "step", helper.WhosCalling())
	}

	payload := Syslog{
		Port:    strconv.Itoa(port),
		Server1: cfg.Server,
		Server2: "",
		Server3: "",
		Enable:  enable,
	}
	err = i.putSyslog(payload)
	if err != nil {
		i.log.V(1).Error(err, "PUT Syslog request failed.",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"step", helper.WhosCalling(),
			"Error", err.Error(),
		)
		return err
	}

	// Enable alerts
	err = i.putAlertEnable(AlertEnable{"Enabled"})
	if err != nil {
		i.log.V(1).Error(err, "PUT to enable Alerts failed request failed.",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"step", helper.WhosCalling(),
			"Error", err.Error(),
		)
		return err
	}

	// Configure alerts
	err = i.putAlertConfig()
	if err != nil {
		i.log.V(1).Error(err, "PUT to configure alerts failed request failed.",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"step", helper.WhosCalling(),
			"Error", err.Error(),
		)
		return err
	}

	i.log.V(1).Info("Syslog and alert parameters applied.", "IP", i.ip, "Model", i.HardwareType())
	return err
}

// Network method implements the Configure interface
// applies various network parameters.
func (i *IDrac9) Network(cfg *cfgresources.Network) (reset bool, err error) {

	params := map[string]string{
		"EnableIPv4":              "Enabled",
		"DHCPEnable":              "Enabled",
		"DNSFromDHCP":             "Enabled",
		"EnableSerialOverLan":     "Enabled",
		"EnableSerialRedirection": "Enabled",
		"EnableIpmiOverLan":       "Enabled",
	}

	if cfg.DNSFromDHCP == false {
		params["DNSFromDHCP"] = "Disabled"
	}

	if cfg.SolEnable == false {
		params["EnableSerialOverLan"] = "Disabled"
		params["EnableSerialRedirection"] = "Disabled"
	}

	if cfg.IpmiEnable == false {
		params["EnableIpmiOverLan"] = "Disabled"
	}

	ipv4 := Ipv4{
		Enable:      params["EnableIPv4"],
		DHCPEnable:  params["DHCPEnable"],
		DNSFromDHCP: params["DNSFromDHCP"],
	}

	serialOverLan := SerialOverLan{
		Enable:       params["EnableSerialOverLan"],
		BaudRate:     "115200",
		MinPrivilege: "Administrator",
	}

	serialRedirection := SerialRedirection{
		Enable:  params["EnableSerialRedirection"],
		QuitKey: "^\\",
	}

	ipmiOverLan := IpmiOverLan{
		Enable:        params["EnableIpmiOverLan"],
		PrivLimit:     "Administrator",
		EncryptionKey: "0000000000000000000000000000000000000000",
	}

	err = i.putIPv4(ipv4)
	if err != nil {
		i.log.V(1).Info("PUT IPv4 request failed.",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"step", helper.WhosCalling(),
			"Error", err.Error(),
		)
	}

	err = i.putSerialOverLan(serialOverLan)
	if err != nil {
		i.log.V(1).Info("PUT SerialOverLan request failed.",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"step", helper.WhosCalling(),
			"Error", err.Error(),
		)
	}

	err = i.putSerialRedirection(serialRedirection)
	if err != nil {
		i.log.V(1).Info("PUT SerialRedirection request failed.",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"step", helper.WhosCalling(),
			"Error", err.Error(),
		)
	}

	err = i.putIpmiOverLan(ipmiOverLan)
	if err != nil {
		i.log.V(1).Info("PUT IpmiOverLan request failed.",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"step", helper.WhosCalling(),
			"Error", err.Error(),
		)
	}

	i.log.V(1).Info("Network config parameters applied.", "IP", i.ip, "Model", i.HardwareType())
	return reset, err
}

// SetLicense implements the Configure interface.
func (i *IDrac9) SetLicense(cfg *cfgresources.License) (err error) {
	return err
}

// GenerateCSR generates a CSR request on the BMC and returns the CSR.
// GenerateCSR implements the Configure interface.
// 1. PUT CSR info based on configuration
// 2. POST sysmgmt/2012/server/network/ssl/csr which returns a base64encoded CSR.
func (i *IDrac9) GenerateCSR(cert *cfgresources.HTTPSCertAttributes) ([]byte, error) {

	c := CSRInfo{
		CommonName:       cert.CommonName,
		CountryCode:      cert.CountryCode,
		LocalityName:     cert.Locality,
		OrganizationName: cert.OrganizationName,
		OrganizationUnit: cert.OrganizationUnit,
		StateName:        cert.StateName,
		EmailAddr:        cert.Email,
		SubjectAltName:   cert.SubjectAltName,
	}

	// 1. PUT CSR params
	err := i.putCSR(c)
	if err != nil {
		return []byte{}, err
	}

	// 2. POST request for CSR file data
	status, body, _ := i.post("sysmgmt/2012/server/network/ssl/csr", []byte{}, "")
	if status != 200 {
		return []byte{}, fmt.Errorf("Non 200 response when requesting for CSR : %d", status)
	}

	return body, nil
}

// UploadHTTPSCert implements the Configure interface.
// UploadHTTPSCert uploads the given CRT cert,
// returns true if the BMC needs a reset.
// 1. POST upload signed x509 cert in multipart form.
// 2. POST returned resource URI
func (i *IDrac9) UploadHTTPSCert(cert []byte, certFileName string, key []byte, keyFileName string) (bool, error) {

	endpoint := "sysmgmt/2012/server/transient/filestore"

	// setup a buffer for our multipart form
	var form bytes.Buffer
	w := multipart.NewWriter(&form)

	// setup the ssl cert part
	formWriter, err := w.CreateFormFile("fileName", certFileName)
	if err != nil {
		return false, err
	}

	_, err = io.Copy(formWriter, bytes.NewReader(cert))
	if err != nil {
		return false, err
	}

	// close multipart writer - adds the teminating boundary.
	w.Close()

	// 1. POST upload x509 cert
	status, body, err := i.post(endpoint, form.Bytes(), w.FormDataContentType())
	if err != nil || status != 201 {
		i.log.V(1).Error(err, "Cert form upload POST request failed, expected 201.",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"endpoint", endpoint,
			"step", helper.WhosCalling(),
			"status", status,
		)
		return false, err
	}

	// extract resourceURI from response
	var certStore = new(certStore)
	err = json.Unmarshal(body, certStore)
	if err != nil {
		i.log.V(1).Error(err, "Unable to unmarshal cert store response payload.",
			"step", helper.WhosCalling(),
			"IP", i.ip,
			"Model", i.HardwareType(),
			"Error", err.Error(),
		)
		return false, err
	}

	resourceURI, err := json.Marshal(certStore.File)
	if err != nil {
		i.log.V(1).Error(err, "Unable to marshal cert store resource URI.",
			"step", helper.WhosCalling(),
			"IP", i.ip,
			"Model", i.HardwareType(),
			"Error", err.Error(),
		)
		return false, err
	}

	// 2. POST resource URI
	endpoint = "sysmgmt/2012/server/network/ssl/cert"
	status, _, err = i.post(endpoint, []byte(resourceURI), "")
	if err != nil || status != 201 {
		i.log.V(1).Error(err, "Cert form upload POST request failed, expected 201.",
			"IP", i.ip,
			"Model", i.HardwareType(),
			"endpoint", endpoint,
			"step", helper.WhosCalling(),
			"status", status,
		)
		return false, err
	}

	return true, err
}
