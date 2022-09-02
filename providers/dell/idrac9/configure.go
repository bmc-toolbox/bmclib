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
	"github.com/bmc-toolbox/bmclib/internal"
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

	validate := validator.New()
	err = validate.Struct(newBiosSettings)
	if err != nil {
		i.log.V(1).Error(err, "Bios(): Config validation failed.", "step", "applyBiosParams")
		return err
	}

	currentBiosSettings, err := i.getBiosSettings()
	if err != nil || currentBiosSettings == nil {
		if err == nil {
			err = fmt.Errorf("Call to getBiosSettings() returned nil.")
		}

		msg := "Bios(): Unable to get current BIOS settings through RedFish."
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return errors.New(msg)
	}

	// Compare current BIOS settings with our declared config.
	if *newBiosSettings != *currentBiosSettings {
		toApplyBiosSettings, err := diffBiosSettings(newBiosSettings, currentBiosSettings)
		if err != nil {
			i.log.V(1).Error(err, "diffBiosSettings returned error.",
				"IP", i.ip,
				"HardwareType", i.HardwareType(),
				"step", helper.WhosCalling(),
			)
			return err
		}

		i.log.V(0).Info("BIOS configuration to be applied...",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
			"Changes (Ignore empty fields)", fmt.Sprintf("%+v", toApplyBiosSettings),
		)

		// Purge any existing pending BIOS setting jobs (otherwise, we won't be able to set any params):
		err = i.purgeJobsForBiosSettings()
		if err != nil {
			i.log.V(1).Error(err, "Bios(): Unable to purge pending BIOS setting jobs.",
				"step", "applyBiosParams",
				"resource", "Bios",
				"IP", i.ip,
				"HardwareType", i.HardwareType(),
			)
		}

		err = i.setBiosSettings(toApplyBiosSettings)
		if err != nil {
			msg := "setBiosAttributes() returned error."
			i.log.V(1).Error(err, msg,
				"IP", i.ip,
				"HardwareType", i.HardwareType(),
				"step", helper.WhosCalling(),
			)
			return errors.New(msg)
		}

		i.log.V(0).Info("BIOS configuration update job queued in IDRAC.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
	} else {
		i.log.V(0).Info("Bios configuration is up to date.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
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
	err = i.httpLogin()
	if err != nil {
		msg := "IDRAC9 User(): HTTP login failed: " + err.Error()
		i.log.V(1).Error(err, msg,
			"step", "applyUserParams",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
		)
		return errors.New(msg)
	}

	err = internal.ValidateUserConfig(cfgUsers)
	if err != nil {
		msg := "IDRAC9 User(): User config validation failed: " + err.Error()
		i.log.V(1).Error(err, msg,
			"step", "applyUserParams",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
		)
		return errors.New(msg)
	}

	idracUsers, err := i.queryUsers()
	if err != nil {
		msg := "IDRAC9 User(): Unable to query existing users."
		i.log.V(1).Error(err, msg,
			"step", "applyUserParams",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"Model", i.HardwareType(),
		)
		return errors.New(msg + " Error: " + err.Error())
	}

	// This user is reserved for IDRAC usage, we can't delete it.
	delete(idracUsers, 1)

	// Start from a clean slate.
	for id := range idracUsers {
		statusCode, payload, err := i.delete(fmt.Sprintf("sysmgmt/2017/server/user?userid=%d", id))
		if err != nil {
			msg := fmt.Sprintf("IDRAC9 User(): Unable to remove existing user (ID %d): %s", id, err.Error())
			i.log.V(1).Error(err, msg,
				"step", "applyUserParams",
				"IP", i.ip,
				"HardwareType", i.HardwareType(),
			)
			return err
		}

		if statusCode > 299 {
			err = fmt.Errorf("Request failed with status code %d and payload %s.", statusCode, string(payload))
			msg := fmt.Sprintf("IDRAC9 User(): Unable to remove existing user (ID %d): %s", id, err.Error())
			i.log.V(1).Error(err, msg,
				"step", "applyUserParams",
				"IP", i.ip,
				"HardwareType", i.HardwareType(),
			)
			return err
		}
	}

	// As mentioned before, user ID 1 is reserved for IDRAC usage.
	userID := 2

	for _, cfgUser := range cfgUsers {
		// If the user is not enabled in the config, just skip.
		if !cfgUser.Enable {
			continue
		}

		user := UserInfo{}
		user.Enable = "Enabled"
		user.UserName = cfgUser.Name
		user.Password = cfgUser.Password
		if cfgUser.SolEnable {
			user.SolEnable = "Enabled"
		} else {
			user.SolEnable = "Disabled"
		}
		if cfgUser.SNMPv3Enable {
			user.ProtocolEnable = "Enabled"
		} else {
			user.ProtocolEnable = "Disabled"
		}
		if cfgUser.Role == "admin" {
			user.Privilege = "511"
			user.IpmiLanPrivilege = "Administrator"
		} else {
			user.Privilege = "499"
			user.IpmiLanPrivilege = "Operator"
		}

		err = i.putUser(userID, user)
		if err != nil {
			i.log.V(1).Error(err, "User(): Add/Update user request failed.",
				"IP", i.ip,
				"HardwareType", i.HardwareType(),
				"step", helper.WhosCalling(),
				"User", cfgUser.Name,
			)
			continue
		}
		i.log.V(1).Info("User parameters applied.", "IP", i.ip, "HardwareType", i.HardwareType(), "User", cfgUser.Name)
		userID++
	}

	return nil
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
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling,
		)
		return err
	}

	if cfg.BaseDn == "" {
		msg := "Ldap resource parameter BaseDn required but not declared."
		err = errors.New(msg)
		i.log.V(1).Error(err, msg,
			"HardwareType", i.HardwareType(),
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
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return err
	}

	return err
}

// Applies LDAP Group/Role-related configuration.
// Implements the Configure interface.
func (i *IDrac9) LdapGroups(cfgGroups []*cfgresources.LdapGroup, cfgLdap *cfgresources.Ldap) (err error) {
	roleID := 0
	for _, cfgRole := range cfgGroups {
		if !cfgRole.Enable {
			continue
		}

		// Use the next slot.
		roleID++

		// The distinguished name of the group:
		//   e.g. If `GroupBaseDn` is ou=Group,dc=example,dc=com and `Group` is cn=fooUsers;
		//        `groupDN` will be cn=fooUsers,ou=Group,dc=example,dc=com.
		role := LdapRoleGroup{
			DN:        fmt.Sprintf("%s,%s", cfgRole.Group, cfgRole.GroupBaseDn),
			Privilege: "0",
		}

		if cfgRole.Role == "admin" {
			role.Privilege = "511"
		} else if cfgRole.Role == "user" {
			role.Privilege = "499"
		}

		// Actual query:
		err = i.putLdapRoleGroup(fmt.Sprintf("%d", roleID), role)
		if err == nil {
			i.log.V(1).Info("LDAP Role Group parameters applied.",
				"IP", i.ip,
				"HardwareType", i.HardwareType(),
				"Step", helper.WhosCalling(),
				"Ldap Role Group", cfgRole.Role,
				"Role Group DN", cfgRole.Role,
			)
		} else {
			i.log.V(1).Error(err, "Add/Update LDAP Role Group request failed.",
				"IP", i.ip,
				"HardwareType", i.HardwareType(),
				"step", helper.WhosCalling(),
				"Ldap Role Group", cfgRole.Group,
				"Role Group DN", cfgRole.Role,
			)
			continue
		}
	}

	// Remove all the rest.
	for roleID++; roleID <= 5; roleID++ {
		role := LdapRoleGroup{
			DN:        "",
			Privilege: "0",
		}
		err = i.putLdapRoleGroup(fmt.Sprintf("%d", roleID), role)
		if err != nil {
			i.log.V(1).Error(err, "Remove LDAP Role Group request failed.",
				"IP", i.ip,
				"HardwareType", i.HardwareType(),
				"step", helper.WhosCalling(),
			)
		}
	}

	return nil
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
			"HardwareType", i.HardwareType(),
			"Step", helper.WhosCalling(),
		)
		return err
	}

	if cfg.Timezone == "" {
		msg := "NTP resource expects parameter: timezone."
		err = errors.New(msg)
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
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
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
			"Unknown Timezone", cfg.Timezone,
		)
		return err
	}

	err = i.putTimezone(Timezone{Timezone: cfg.Timezone})
	if err != nil {
		i.log.V(1).Error(err, "PUT timezone request failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
			"Timezone", cfg.Timezone,
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
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return err
	}

	i.log.V(1).Info("NTP servers param applied.",
		"IP", i.ip,
		"HardwareType", i.HardwareType(),
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
			"HardwareType", i.HardwareType(),
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

	if !cfg.Enable {
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
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return err
	}

	// Enable alerts
	err = i.putAlertEnable(AlertEnable{"Enabled"})
	if err != nil {
		i.log.V(1).Error(err, "PUT to enable Alerts failed request failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return err
	}

	// Configure alerts
	err = i.putAlertConfig()
	if err != nil {
		i.log.V(1).Error(err, "PUT to configure alerts failed request failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return err
	}

	i.log.V(1).Info("Syslog and alert parameters applied.", "IP", i.ip, "HardwareType", i.HardwareType())
	return err
}

func (i *IDrac9) setSnmp(cfg cfgresources.Network) error {
	enableSNMP := 1
	if !cfg.SNMPEnable {
		enableSNMP = 0
	}

	sshSnmpCommand := fmt.Sprint("racadm set iDRAC.SNMP.AgentEnable ", enableSNMP)

	_, err := i.sshClient.Run(sshSnmpCommand)

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

	if !cfg.DNSFromDHCP {
		params["DNSFromDHCP"] = "Disabled"
	}

	if !cfg.SolEnable {
		params["EnableSerialOverLan"] = "Disabled"
		params["EnableSerialRedirection"] = "Disabled"
	}

	if !cfg.IpmiEnable {
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
		i.log.V(1).Error(err, "PUT IPv4 request failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
	}

	err = i.putSerialOverLan(serialOverLan)
	if err != nil {
		i.log.V(1).Error(err, "PUT SerialOverLan request failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
	}

	err = i.putSerialRedirection(serialRedirection)
	if err != nil {
		i.log.V(1).Error(err, "PUT SerialRedirection request failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
	}

	err = i.putIpmiOverLan(ipmiOverLan)
	if err != nil {
		i.log.V(1).Error(err, "PUT IpmiOverLan request failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
	}

	i.log.V(1).Info("Network config parameters applied.",
		"IP", i.ip,
		"HardwareType", i.HardwareType())

	// SNMP section

	err = i.setSnmp(*cfg)
	if err != nil {
		msg := "Unable to set SNMP settings"
		i.log.V(1).Error(err, msg,
			"step", "SNMPEnable",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
		)
		return reset, err
	} else {
		i.log.V(1).Info("SNMP parameters applied.", "IP",
			i.ip, "HardwareType",
			i.HardwareType())
	}

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
		if err == nil {
			err = fmt.Errorf("Cert form upload POST request to %s failed with status code %d.", endpoint, status)
		}

		i.log.V(1).Error(err, "UploadHTTPSCert(): Cert form upload POST request failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"endpoint", endpoint,
			"step", helper.WhosCalling(),
			"StatusCode", status,
		)
		return false, err
	}

	// extract resourceURI from response
	certStore := new(certStore)
	err = json.Unmarshal(body, certStore)
	if err != nil {
		i.log.V(1).Error(err, "Unable to unmarshal cert store response payload.",
			"step", helper.WhosCalling(),
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
		)
		return false, err
	}

	resourceURI, err := json.Marshal(certStore.File)
	if err != nil {
		i.log.V(1).Error(err, "Unable to marshal cert store resource URI.",
			"step", helper.WhosCalling(),
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
		)
		return false, err
	}

	// 2. POST resource URI
	endpoint = "sysmgmt/2012/server/network/ssl/cert"
	status, _, err = i.post(endpoint, []byte(resourceURI), "")
	if err != nil || status != 201 {
		if err == nil {
			err = fmt.Errorf("Cert form upload POST request to %s failed with status code %d.", endpoint, status)
		}

		i.log.V(1).Error(err, "UploadHTTPSCert(): Cert form upload POST request failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"endpoint", endpoint,
			"step", helper.WhosCalling(),
			"StatusCode", status,
		)
		return false, err
	}

	return true, nil
}
