package ilo

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/internal/helper"
	log "github.com/sirupsen/logrus"
)

// This ensures the compiler errors if this type is missing
// a method that should be implmented to satisfy the Configure interface.
var _ devices.Configure = (*Ilo)(nil)

// Resources returns a slice of supported resources and
// the order they are to be applied in.
func (i *Ilo) Resources() []string {
	return []string{
		"user",
		"syslog",
		"license",
		"ntp",
		"ldap_group",
		"ldap",
		"network",
		"power",
		"https_cert",
	}
}

// ApplyCfg applies configuration
// To be deprecated once the Configure interface is ready.
func (i *Ilo) ApplyCfg(config *cfgresources.ResourcesConfig) (err error) {

	//check sessionKey is available
	if i.sessionKey == "" {
		msg := "Expected sessionKey not found, unable to configure BMC."
		log.WithFields(log.Fields{
			"step":  "Login()",
			"IP":    i.ip,
			"Model": i.HardwareType(),
		}).Warn(msg)
		return errors.New(msg)
	}

	return nil
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
		if userInfo.UserName == user || userInfo.LoginName == user {
			return userInfo, true
		}
	}

	return userInfo, false
}

// checks if a ldap group is present in a given list
func ldapGroupExists(group string, directoryGroups []DirectoryGroups) (directoryGroup DirectoryGroups, exists bool) {

	for _, directoryGroup := range directoryGroups {
		if directoryGroup.Dn == group {
			return directoryGroup, true
		}
	}

	return directoryGroup, false
}

// User applies the User configuration resource,
// if the user exists, it updates the users password,
// User implements the Configure interface.
// nolint: gocyclo
func (i *Ilo) User(users []*cfgresources.User) (err error) {

	existingUsers, err := i.queryUsers()
	if err != nil {
		msg := "Unable to query existing users"
		log.WithFields(log.Fields{
			"step":  "applyUserParams",
			"IP":    i.ip,
			"Model": i.HardwareType(),
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
				userinfo.UserID = userinfo.ID
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
			userinfo.UserID = userinfo.ID
			log.WithFields(log.Fields{
				"IP":    i.ip,
				"Model": i.HardwareType(),
				"User":  user.Name,
			}).Debug("User disabled in config, will be removed.")
			postPayload = true
		}

		if postPayload {
			payload, err := json.Marshal(userinfo)
			if err != nil {
				log.WithFields(log.Fields{
					"IP":    i.ip,
					"Model": i.HardwareType(),
					"step":  helper.WhosCalling(),
					"User":  user.Name,
					"Error": err,
				}).Warn("Unable to marshal userInfo payload to set User config.")
				continue
			}

			endpoint := "json/user_info"
			statusCode, response, err := i.post(endpoint, payload)
			if err != nil || statusCode != 200 {
				log.WithFields(log.Fields{
					"IP":         i.ip,
					"Model":      i.HardwareType(),
					"endpoint":   endpoint,
					"step":       helper.WhosCalling(),
					"User":       user.Name,
					"StatusCode": statusCode,
					"response":   string(response),
					"Error":      err,
				}).Warn("POST request to set User config returned error.")
				continue
			}

			log.WithFields(log.Fields{
				"IP":    i.ip,
				"Model": i.HardwareType(),
				"User":  user.Name,
			}).Debug("User parameters applied.")

		}
	}

	return err
}

// Syslog applies the Syslog configuration resource
// Syslog implements the Configure interface
func (i *Ilo) Syslog(cfg *cfgresources.Syslog) (err error) {

	var port int
	enable := 1

	if cfg.Server == "" {
		msg := "Syslog resource expects parameter: Server."
		log.WithFields(log.Fields{
			"step": helper.WhosCalling(),
		}).Warn(msg)
		return errors.New(msg)
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
		enable = 0
		log.WithFields(log.Fields{
			"step": helper.WhosCalling(),
		}).Debug("Syslog resource declared with disable.")
	}

	remoteSyslog := RemoteSyslog{
		SyslogEnable: enable,
		SyslogPort:   port,
		Method:       "syslog_save",
		SyslogServer: cfg.Server,
		SessionKey:   i.sessionKey,
	}

	payload, err := json.Marshal(remoteSyslog)
	if err != nil {
		msg := "Unable to marshal RemoteSyslog payload to set Syslog config."
		log.WithFields(log.Fields{
			"IP":    i.ip,
			"Model": i.HardwareType(),
			"step":  helper.WhosCalling(),
			"Error": err,
		}).Warn(msg)
		return errors.New(msg)
	}

	endpoint := "json/remote_syslog"
	statusCode, response, err := i.post(endpoint, payload)
	if err != nil || statusCode != 200 {
		msg := "POST request to set User config returned error."
		log.WithFields(log.Fields{
			"IP":         i.ip,
			"Model":      i.HardwareType(),
			"endpoint":   endpoint,
			"step":       helper.WhosCalling(),
			"StatusCode": statusCode,
			"response":   string(response),
			"Error":      err,
		}).Warn(msg)
		return errors.New(msg)
	}

	log.WithFields(log.Fields{
		"IP":    i.ip,
		"Model": i.HardwareType(),
	}).Debug("Syslog parameters applied.")

	return err
}

// SetLicense applies license configuration params
// SetLicense implements the Configure interface.
func (i *Ilo) SetLicense(cfg *cfgresources.License) (err error) {

	if cfg.Key == "" {
		msg := "License resource expects parameter: Key."
		log.WithFields(log.Fields{
			"step": helper.WhosCalling(),
		}).Warn(msg)
		return errors.New(msg)
	}

	license := LicenseInfo{
		Key:        cfg.Key,
		Method:     "activate",
		SessionKey: i.sessionKey,
	}

	payload, err := json.Marshal(license)
	if err != nil {
		msg := "Unable to marshal License payload to activate License."
		log.WithFields(log.Fields{
			"IP":    i.ip,
			"Model": i.HardwareType(),
			"step":  helper.WhosCalling(),
			"Error": err,
		}).Warn(msg)
		return errors.New(msg)
	}

	endpoint := "json/license_info"
	statusCode, response, err := i.post(endpoint, payload)
	if err != nil || statusCode != 200 {
		msg := "POST request to set User config returned error."
		log.WithFields(log.Fields{
			"IP":         i.ip,
			"Model":      i.HardwareType(),
			"endpoint":   endpoint,
			"step":       helper.WhosCalling(),
			"StatusCode": statusCode,
			"response":   string(response),
			"Error":      err,
		}).Warn(msg)
		return errors.New(msg)
	}

	log.WithFields(log.Fields{
		"IP":    i.ip,
		"Model": i.HardwareType(),
	}).Debug("License activated.")

	return err
}

// Ntp applies NTP configuration params
// Ntp implements the Configure interface.
func (i *Ilo) Ntp(cfg *cfgresources.Ntp) (err error) {

	enable := 1
	if cfg.Server1 == "" {
		msg := "NTP resource expects parameter: server1."
		log.WithFields(log.Fields{
			"step": helper.WhosCalling(),
		}).Warn(msg)
		return errors.New(msg)
	}

	if cfg.Timezone == "" {
		msg := "NTP resource expects parameter: timezone."
		log.WithFields(log.Fields{
			"step": helper.WhosCalling(),
		}).Warn(msg)
		return errors.New(msg)
	}

	// supported timezone based on device.
	var timezones map[string]int

	// ideally ilo5 ilo4 should be split up into its own device
	// instead of depending on HardwareType.
	if i.HardwareType() == "ilo5" {
		timezones = TimezonesIlo5
	} else {
		timezones = TimezonesIlo4
	}
	_, validTimezone := timezones[cfg.Timezone]
	if !validTimezone {
		msg := "NTP resource a valid timezone parameter, for valid timezones see hp/ilo/model.go"
		log.WithFields(log.Fields{
			"step":             helper.WhosCalling(),
			"Unknown Timezone": cfg.Timezone,
		}).Warn(msg)
		return errors.New(msg)
	}

	if cfg.Enable != true {
		enable = 0
		log.WithFields(log.Fields{
			"step": helper.WhosCalling(),
		}).Debug("NTP resource declared with disable.")
	}

	existingConfig, err := i.queryNetworkSntp()
	if err != nil {
		msg := "Unable to query existing config"
		log.WithFields(log.Fields{
			"step":  helper.WhosCalling(),
			"IP":    i.ip,
			"Model": i.HardwareType(),
			"Error": err,
		}).Warn(msg)
		return errors.New(msg)
	}

	networkSntp := NetworkSntp{
		Interface:                   existingConfig.Interface,
		PendingChange:               existingConfig.PendingChange,
		NicWcount:                   existingConfig.NicWcount,
		TzWcount:                    existingConfig.TzWcount,
		Ipv4Disabled:                0,
		Ipv6Disabled:                0,
		DhcpEnabled:                 enable,
		Dhcp6Enabled:                enable,
		UseDhcpSuppliedTimeServers:  0, //we probably want to expose these as params
		UseDhcp6SuppliedTimeServers: 0,
		Sdn1WCount:                  existingConfig.Sdn1WCount,
		Sdn2WCount:                  existingConfig.Sdn2WCount,
		TimePropagate:               existingConfig.TimePropagate,
		SntpServer1:                 cfg.Server1,
		SntpServer2:                 cfg.Server2,
		OurZone:                     timezones[cfg.Timezone],
		Method:                      "set_sntp",
		SessionKey:                  i.sessionKey,
	}

	payload, err := json.Marshal(networkSntp)
	if err != nil {
		msg := "Unable to marshal NetworkSntp payload to set NTP config."
		log.WithFields(log.Fields{
			"IP":    i.ip,
			"Model": i.HardwareType(),
			"step":  helper.WhosCalling(),
			"Error": err,
		}).Warn(msg)
		return errors.New(msg)
	}

	endpoint := "json/network_sntp"
	statusCode, response, err := i.post(endpoint, payload)
	if err != nil || statusCode != 200 {
		msg := "POST request to set NTP config returned error."
		log.WithFields(log.Fields{
			"IP":         i.ip,
			"Model":      i.HardwareType(),
			"endpoint":   endpoint,
			"step":       helper.WhosCalling(),
			"StatusCode": statusCode,
			"response":   string(response),
			"Error":      err,
		}).Warn(msg)
		return errors.New(msg)
	}

	log.WithFields(log.Fields{
		"IP":    i.ip,
		"Model": i.HardwareType(),
	}).Debug("NTP parameters applied.")

	return err
}

// LdapGroup applies LDAP Group/Role related configuration
// LdapGroup implements the Configure interface.
// nolint: gocyclo
func (i *Ilo) LdapGroup(cfg []*cfgresources.LdapGroup, cfgLdap *cfgresources.Ldap) (err error) {

	directoryGroups, err := i.queryDirectoryGroups()
	if err != nil {
		msg := "Unable to query existing Ldap groups"
		log.WithFields(log.Fields{
			"IP":    i.ip,
			"Model": i.HardwareType(),
			"Step":  helper.WhosCalling(),
			"Error": err,
		}).Warn(msg)
		return errors.New(msg)
	}

	for _, group := range cfg {

		var postPayload bool
		if group.Group == "" {
			msg := "Ldap resource parameter Group required but not declared."
			log.WithFields(log.Fields{
				"Model":     i.HardwareType(),
				"step":      helper.WhosCalling,
				"Ldap role": group.Role,
			}).Warn(msg)
			return errors.New(msg)
		}

		if !i.isRoleValid(group.Role) {
			msg := "Ldap resource Role must be a valid role: admin OR user."
			log.WithFields(log.Fields{
				"Model":     i.HardwareType(),
				"step":      helper.WhosCalling(),
				"Ldap role": group.Role,
			}).Warn(msg)
			return errors.New(msg)
		}

		groupDn := group.Group
		directoryGroup, gexists := ldapGroupExists(groupDn, directoryGroups)

		directoryGroup.Dn = groupDn
		directoryGroup.SessionKey = i.sessionKey

		//if the group is enabled setup parameters
		if group.Enable {

			directoryGroup.LoginPriv = 1
			directoryGroup.RemoteConsPriv = 1
			directoryGroup.VirtualMediaPriv = 1
			directoryGroup.ResetPriv = 1

			if group.Role == "admin" {
				directoryGroup.ConfigPriv = 1
				directoryGroup.UserPriv = 1
			} else if group.Role == "user" {
				directoryGroup.ConfigPriv = 0
				directoryGroup.UserPriv = 0
			}

			//if the group exists, modify it
			if gexists {
				directoryGroup.Method = "mod_group"
			} else {

				directoryGroup.Method = "add_group"
			}

			postPayload = true
		}

		//if the group is disabled remove it
		if group.Enable == false && gexists {
			directoryGroup.Method = "del_group"
			log.WithFields(log.Fields{
				"IP":    i.ip,
				"Model": i.HardwareType(),
				"User":  group.Group,
			}).Debug("Ldap role group disabled in config, will be removed.")
			postPayload = true
		}

		if postPayload {
			payload, err := json.Marshal(directoryGroup)
			if err != nil {
				log.WithFields(log.Fields{
					"IP":    i.ip,
					"Model": i.HardwareType(),
					"Step":  helper.WhosCalling(),
					"Group": group.Group,
					"Error": err,
				}).Warn("Unable to marshal directoryGroup payload to set LdapGroup config.")
				continue
			}

			endpoint := "json/directory_groups"
			statusCode, response, err := i.post(endpoint, payload)
			if err != nil || statusCode != 200 {
				log.WithFields(log.Fields{
					"IP":         i.ip,
					"Model":      i.HardwareType(),
					"endpoint":   endpoint,
					"step":       helper.WhosCalling(),
					"Group":      group.Group,
					"StatusCode": statusCode,
					"response":   string(response),
					"Error":      err,
				}).Warn("POST request to set User config returned error.")
				continue
			}

			log.WithFields(log.Fields{
				"IP":    i.ip,
				"Model": i.HardwareType(),
				"User":  group.Group,
			}).Debug("LdapGroup parameters applied.")

		}

	}

	return err
}

// Ldap applies LDAP configuration params.
// Ldap implements the Configure interface.
func (i *Ilo) Ldap(cfg *cfgresources.Ldap) (err error) {

	if cfg.Server == "" {
		msg := "Ldap resource parameter Server required but not declared."
		log.WithFields(log.Fields{
			"Model": i.HardwareType(),
			"step":  helper.WhosCalling,
		}).Warn(msg)
		return errors.New(msg)
	}

	if cfg.Port == 0 {
		msg := "Ldap resource parameter Port required but not declared."
		log.WithFields(log.Fields{
			"Model": i.HardwareType(),
			"step":  helper.WhosCalling,
		}).Warn(msg)
		return errors.New(msg)
	}

	if cfg.BaseDn == "" {
		msg := "Ldap resource parameter BaseDn required but not declared."
		log.WithFields(log.Fields{
			"Model": i.HardwareType(),
			"step":  helper.WhosCalling,
		}).Warn(msg)
		return errors.New(msg)
	}

	var enable int
	if cfg.Enable == false {
		enable = 0
	} else {
		enable = 1
	}

	directory := Directory{
		ServerAddress:         cfg.Server,
		ServerPort:            cfg.Port,
		UserContexts:          []string{cfg.BaseDn},
		AuthenticationEnabled: enable,
		LocalUserAcct:         1,
		EnableGroupAccount:    1,
		EnableKerberos:        0,
		EnableGenericLdap:     enable,
		Method:                "mod_dir_config",
		SessionKey:            i.sessionKey,
	}

	payload, err := json.Marshal(directory)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":    i.ip,
			"Model": i.HardwareType(),
			"Step":  helper.WhosCalling(),
			"Error": err,
		}).Warn("Unable to marshal directory payload to set Ldap config.")
		return err
	}

	endpoint := "json/directory"
	statusCode, response, err := i.post(endpoint, payload)
	if err != nil || statusCode != 200 {
		msg := "POST request to set Ldap config returned error."
		log.WithFields(log.Fields{
			"IP":         i.ip,
			"Model":      i.HardwareType(),
			"endpoint":   endpoint,
			"step":       helper.WhosCalling(),
			"StatusCode": statusCode,
			"response":   string(response),
			"Error":      err,
		}).Warn(msg)
		return err
	}

	log.WithFields(log.Fields{
		"IP":    i.ip,
		"Model": i.HardwareType(),
	}).Debug("Ldap parameters applied.")

	return err

}

// GenerateCSR generates a CSR request on the BMC.
// If its the first CSR attempt - the BMC is going to take a while to generate the CSR,
// the response will be a 500 with the body	{"message":"JS_CERT_NOT_AVAILABLE","details":null}
// If the configuration for the Subject has not changed and the CSR is ready a CSR is returned.
func (i *Ilo) GenerateCSR(cert *cfgresources.HTTPSCertAttributes) ([]byte, error) {

	csrConfig := &csr{
		Country:          cert.CountryCode,
		State:            cert.StateName,
		Locality:         cert.Locality,
		CommonName:       cert.CommonName,
		OrganizationName: cert.OrganizationName,
		OrganizationUnit: cert.OrganizationUnit,
		IncludeIP:        1, // Use IP as SAN
		Method:           "create_csr",
		SessionKey:       i.sessionKey,
	}

	payload, err := json.Marshal(csrConfig)
	if err != nil {
		return []byte{}, err
	}

	endpoint := "json/csr"
	statusCode, response, err := i.post(endpoint, payload)
	if statusCode == 500 {
		return []byte{}, fmt.Errorf("CSR being generated, retry later")
	}

	// if its a not a 200 at this point,
	// something else went wrong.
	if statusCode != 200 {
		return []byte{}, fmt.Errorf("Unexpected return code: %d", statusCode)
	}

	// Some other error
	if err != nil {
		return []byte{}, err
	}

	var r = new(csrResponse)
	err = json.Unmarshal(response, r)
	if err != nil {
		return []byte{}, err
	}

	return []byte(r.CsrPEM), nil
}

// UploadHTTPSCert uploads the given CRT cert,
// UploadHTTPSCert implements the Configure interface.
// return true if the bmc requires a reset.
func (i *Ilo) UploadHTTPSCert(cert []byte, certFileName string, key []byte, keyFileName string) (bool, error) {

	certPayload := &certImport{
		Method:          "import_certificate",
		CertificateData: string(cert),
		SessionKey:      i.sessionKey,
	}

	payload, err := json.Marshal(certPayload)
	if err != nil {
		return false, err
	}

	endpoint := "json/certificate"
	statusCode, _, err := i.post(endpoint, payload)
	if err != nil {
		return false, err
	}

	if statusCode != 200 {
		return false, fmt.Errorf("Unexpected return code: %d", statusCode)
	}

	// ILOs need a reset after cert upload.
	return true, nil
}

// Network method implements the Configure interface
// nolint: gocyclo
func (i *Ilo) Network(cfg *cfgresources.Network) (reset bool, err error) {

	// check if AccessSettings configuration update is required.
	accessSettings, updateAccessSettings, err := i.cmpAccessSettings(cfg)
	if err != nil {
		return reset, err
	}

	if updateAccessSettings {
		payload, err := json.Marshal(accessSettings)
		if err != nil {
			return reset, fmt.Errorf("Error marshaling AccessSettings payload: %s", err)
		}

		endpoint := "json/access_settings"
		statusCode, _, err := i.post(endpoint, payload)
		if err != nil || statusCode != 200 {
			return reset, fmt.Errorf("Error/non 200 response calling access_settings, status: %d, error: %s", statusCode, err)
		}

		reset = true
	}

	// check the current network IPv4 config
	networkIPv4Settings, updateIPv4Settings, err := i.cmpNetworkIPv4Settings(cfg)
	if err != nil {
		return reset, err
	}

	if updateIPv4Settings {

		payload, err := json.Marshal(networkIPv4Settings)
		if err != nil {
			return reset, fmt.Errorf("Error marshaling NetworkIPv4 payload: %s", err)
		}

		endpoint := "json/network_ipv4/interface/0"
		statusCode, _, err := i.post(endpoint, payload)
		if err != nil || statusCode != 200 {
			return reset, fmt.Errorf("Error/non 200 response calling access_settings, status: %d, error: %s", statusCode, err)
		}

		reset = true
	}

	return reset, nil
}

func (i *Ilo) Power(cfg *cfgresources.Power) error {

	if cfg.HPE == nil {
		return nil
	}

	// map of valid power_settings attributes to params passed to the iLO API
	var powerRegulatorModes = map[string]string{
		"dynamic":     "dyn",
		"static_low":  "min",
		"static_high": "max",
		"os_control":  "osc",
	}

	configMode, exists := powerRegulatorModes[cfg.HPE.PowerRegulator]
	if cfg.HPE.PowerRegulator == "" || !exists {
		return fmt.Errorf("power regulator parameter must be one of dynamic, static_log, static_high, os_control")
	}

	// check if a configuration update is required based on current setting
	config, changeRequired, err := i.cmpPowerSettings(configMode)
	if err != nil {
		return err
	}

	if !changeRequired {
		log.WithFields(log.Fields{
			"IP":            i.ip,
			"current mode":  config.PowerMode,
			"expected mode": configMode,
			"Model":         i.HardwareType(),
		}).Trace("Power regulator config - no change required.")
		return nil
	}

	log.WithFields(log.Fields{
		"IP":            i.ip,
		"current mode":  config.PowerMode,
		"to apply mode": configMode,
		"Model":         i.HardwareType(),
	}).Trace("Power regulator change to be applied.")

	config.SessionKey = i.sessionKey
	config.Method = "set"
	config.PowerMode = configMode

	payload, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("Error marshaling PowerRegulator payload: %s", err)
	}

	endpoint := "json/power_regulator"
	statusCode, _, err := i.post(endpoint, payload)
	if err != nil || statusCode != 200 {
		return fmt.Errorf("Error/non 200 response calling power_regulator, status: %d, error: %s", statusCode, err)
	}

	log.WithFields(log.Fields{
		"IP":    i.ip,
		"Model": i.HardwareType(),
	}).Debug("Power regulator config applied.")

	return nil
}

// Bios method implements the Configure interface
func (i *Ilo) Bios(cfg *cfgresources.Bios) error {
	return nil
}
