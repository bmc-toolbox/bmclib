package ilo

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/internal/helper"
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
	if i.sessionKey == "" {
		msg := "Expected sessionKey not found, unable to configure BMC."
		i.log.V(1).Info(msg,
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", "Login()",
		)
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

// Checks if a user is present in a given list.
func userExists(user string, usersInfo []UserInfo) (userInfo UserInfo, exists bool) {
	for _, userInfo := range usersInfo {
		if userInfo.UserName == user || userInfo.LoginName == user {
			return userInfo, true
		}
	}

	return userInfo, false
}

// User applies the User configuration resource.
// If the user exists, it updates the password.
// User implements the Configure interface.
func (i *Ilo) User(users []*cfgresources.User) (err error) {
	existingUsers, err := i.queryUsers()
	if err != nil {
		msg := "ILO User(): Unable to query existing users."
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", "applyUserParams",
		)
		return err
	}

	// Validation cycle.
	for _, user := range users {
		if user.Name == "" {
			msg := "User resource expects parameter: Name."
			i.log.V(1).Info(msg, "step", "applyUserParams")
			return errors.New(msg)
		}

		if user.Password == "" {
			msg := "User resource expects parameter: Password."
			i.log.V(1).Info(msg, "step", "applyUserParams", "Username", user.Name)
			return errors.New(msg)
		}

		if !i.isRoleValid(user.Role) {
			msg := "User resource Role must be declared and a must be a valid role: 'admin' OR 'user'."
			i.log.V(1).Info(msg, "step", "applyUserParams", "Username", user.Name)
			return errors.New(msg)
		}
	}

	for _, user := range users {
		var postPayload bool

		userinfo, uexists := userExists(user.Name, existingUsers)
		userinfo.SessionKey = i.sessionKey

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

		// If the user is disabled, remove them.
		if !user.Enable && uexists {
			userinfo.Method = "del_user"
			userinfo.UserID = userinfo.ID
			msg := "User disabled in config, will be removed."
			i.log.V(1).Info(msg,
				"IP", i.ip,
				"HardwareType", i.HardwareType(),
				"User", user.Name,
			)
			postPayload = true
		}

		if postPayload {
			payload, err := json.Marshal(userinfo)
			if err != nil {
				msg := "User(): Unable to marshal userInfo payload to set User config."
				i.log.V(1).Error(err, msg,
					"IP", i.ip,
					"HardwareType", i.HardwareType(),
					"step", helper.WhosCalling(),
					"User", user.Name,
				)
				continue
			}

			endpoint := "json/user_info"
			statusCode, response, err := i.post(endpoint, payload)
			if err != nil || statusCode != 200 {
				if err == nil {
					err = fmt.Errorf("Received a %d status code from the POST request to %s.", statusCode, endpoint)
				} else {
					err = fmt.Errorf("POST request to %s failed with error: %s", endpoint, err.Error())
				}

				i.log.V(1).Error(err, "POST request to set User config failed.",
					"IP", i.ip,
					"HardwareType", i.HardwareType(),
					"endpoint", endpoint,
					"step", helper.WhosCalling(),
					"User", user.Name,
					"StatusCode", statusCode,
					"response", string(response),
				)
				continue
			}

			i.log.V(1).Info("User parameters applied.",
				"IP", i.ip,
				"HardwareType", i.HardwareType(),
				"User", user.Name,
			)
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
		i.log.V(1).Info(msg, "step", helper.WhosCalling())
		return errors.New(msg)
	}

	if cfg.Port == 0 {
		i.log.V(1).Info("Syslog resource port set to default: 514.", "step", helper.WhosCalling())
		port = 514
	} else {
		port = cfg.Port
	}

	if !cfg.Enable {
		enable = 0
		i.log.V(1).Info("Syslog resource declared with disable.", "step", helper.WhosCalling())
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
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return errors.New(msg)
	}

	endpoint := "json/remote_syslog"
	statusCode, response, err := i.post(endpoint, payload)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the POST request to %s.", statusCode, endpoint)
		} else {
			err = fmt.Errorf("POST request to %s failed with error: %s", endpoint, err.Error())
		}

		i.log.V(1).Error(err, "POST request to set Syslog config failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"endpoint", endpoint,
			"step", helper.WhosCalling(),
			"StatusCode", statusCode,
			"response", string(response),
		)
		return err
	}

	i.log.V(1).Info("Syslog parameters applied.", "IP", i.ip, "HardwareType", i.HardwareType())

	return err
}

// SetLicense applies license configuration params
// SetLicense implements the Configure interface.
func (i *Ilo) SetLicense(cfg *cfgresources.License) (err error) {
	if cfg.Key == "" {
		msg := "License resource expects parameter: Key."
		i.log.V(1).Info(msg, "step", helper.WhosCalling())
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
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return errors.New(msg + ": " + err.Error())
	}

	endpoint := "json/license_info"
	statusCode, response, err := i.post(endpoint, payload)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the POST request to %s.", statusCode, endpoint)
		} else {
			err = fmt.Errorf("POST request to %s failed with error: %s", endpoint, err.Error())
		}

		msg := "POST request to set License failed."
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"endpoint", endpoint,
			"step", helper.WhosCalling(),
			"StatusCode", statusCode,
			"response", string(response),
		)
		return err
	}

	i.log.V(1).Info("License activated.", "IP", i.ip, "HardwareType", i.HardwareType())

	return err
}

// Ntp applies NTP configuration params
// Ntp implements the Configure interface.
func (i *Ilo) Ntp(cfg *cfgresources.Ntp) (err error) {
	enable := 1
	if cfg.Server1 == "" {
		msg := "NTP resource expects parameter: server1."
		i.log.V(1).Info(msg, "step", helper.WhosCalling())
		return errors.New(msg)
	}

	if cfg.Timezone == "" {
		msg := "NTP resource expects parameter: timezone."
		i.log.V(1).Info(msg, "step", helper.WhosCalling())
		return errors.New(msg)
	}

	// supported timezone based on device.
	var timezones map[string]int

	// ideally ilo5 ilo4 should be split up into its own device
	// instead of depending on HardwareType.
	if i.HardwareType() == Ilo5 {
		timezones = TimezonesIlo5
	} else {
		timezones = TimezonesIlo4
	}
	_, validTimezone := timezones[cfg.Timezone]
	if !validTimezone {
		msg := "NTP resource a valid timezone parameter, for valid timezones see hp/ilo/model.go"
		i.log.V(1).Info(msg, "step", helper.WhosCalling(), "UnknownTimezone", cfg.Timezone)
		return errors.New(msg)
	}

	if !cfg.Enable {
		enable = 0
		i.log.V(1).Info("NTP resource declared with disable.", "step", helper.WhosCalling())
	}

	existingConfig, err := i.queryNetworkSntp()
	if err != nil {
		msg := "Unable to query existing config"
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return err
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
		UseDhcpSuppliedTimeServers:  0, // TODO: Maybe expose these as params?
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
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return err
	}

	endpoint := "json/network_sntp"
	statusCode, response, err := i.post(endpoint, payload)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the POST request to %s.", statusCode, endpoint)
		} else {
			err = fmt.Errorf("POST request to %s failed with error: %s", endpoint, err.Error())
		}

		i.log.V(1).Error(err, "POST request to set NTP config failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"endpoint", endpoint,
			"step", helper.WhosCalling(),
			"StatusCode", statusCode,
			"response", string(response),
		)
		return err
	}

	i.log.V(1).Info("NTP parameters applied.", "IP", i.ip, "HardwareType", i.HardwareType())

	return err
}

// LdapGroups applies LDAP Group/Role related configuration
// LdapGroups implements the Configure interface.
// nolint: gocyclo
func (i *Ilo) LdapGroups(cfgGroups []*cfgresources.LdapGroup, cfgLdap *cfgresources.Ldap) (err error) {
	directoryGroups, err := i.queryDirectoryGroups()
	if err != nil {
		msg := "Unable to query existing LDAP groups."
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return err
	}

	endpoint := "json/directory_groups"

	// Let's start from a clean slate.
	for _, group := range directoryGroups {
		group.Method = "del_group"
		group.SessionKey = i.sessionKey

		payload, err := json.Marshal(group)
		if err != nil {
			i.log.V(1).Error(err, "Unable to marshal directoryGroup payload to set LdapGroup config.",
				"IP", i.ip,
				"HardwareType", i.HardwareType(),
				"step", helper.WhosCalling(),
				"Group", group,
			)
			continue
		}

		statusCode, response, err := i.post(endpoint, payload)
		if err != nil || statusCode != 200 {
			if err == nil {
				err = fmt.Errorf("Received a %d status code from the POST request to %s.", statusCode, endpoint)
			} else {
				err = fmt.Errorf("POST request to %s failed with error: %s", endpoint, err.Error())
			}

			i.log.V(1).Error(err, "POST request to delete LDAP groups failed.",
				"IP", i.ip,
				"HardwareType", i.HardwareType(),
				"endpoint", endpoint,
				"step", helper.WhosCalling(),
				"Group", group,
				"StatusCode", statusCode,
				"response", string(response),
			)
			continue
		}

		i.log.V(1).Info("Old LDAP group deleted successfully.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"Group", group,
		)
	}

	// Verify we have good configuration.
	for _, group := range cfgGroups {
		if !group.Enable {
			continue
		}

		if group.Group == "" {
			msg := "LDAP resource parameter Group required but not declared."
			i.log.V(1).Info(msg,
				"HardwareType", i.HardwareType(),
				"step", helper.WhosCalling(),
				"Ldap role", group.Role,
			)
			return errors.New(msg)
		}

		if !i.isRoleValid(group.Role) {
			msg := "LDAP resource Role must be a valid role: admin OR user."
			i.log.V(1).Info(msg,
				"HardwareType", i.HardwareType(),
				"step", helper.WhosCalling(),
				"Ldap role", group.Role,
			)
			return errors.New(msg)
		}
	}

	// Now, let's add what we have.
	for _, group := range cfgGroups {
		if !group.Enable {
			continue
		}

		var directoryGroup DirectoryGroups
		directoryGroup.Dn = fmt.Sprintf("%s,%s", group.Group, group.GroupBaseDn)
		directoryGroup.Method = "add_group"
		directoryGroup.SessionKey = i.sessionKey

		// Privileges
		directoryGroup.LoginPriv = 1
		directoryGroup.RemoteConsPriv = 1
		directoryGroup.VirtualMediaPriv = 1
		directoryGroup.ResetPriv = 1

		if group.Role == "admin" {
			directoryGroup.ConfigPriv = 1
			directoryGroup.UserPriv = 1
		} else {
			directoryGroup.ConfigPriv = 0
			directoryGroup.UserPriv = 0
		}

		payload, err := json.Marshal(directoryGroup)
		if err != nil {
			i.log.V(1).Error(err, "LdapGroups(): Unable to marshal directoryGroup payload to set LdapGroup config.",
				"IP", i.ip,
				"HardwareType", i.HardwareType(),
				"step", helper.WhosCalling(),
				"Group", group.Group,
			)
			continue
		}

		statusCode, response, err := i.post(endpoint, payload)
		if err != nil || statusCode != 200 {
			if err == nil {
				err = fmt.Errorf("Received a %d status code from the POST request to %s.", statusCode, endpoint)
			} else {
				err = fmt.Errorf("POST request to %s failed with error: %s", endpoint, err.Error())
			}

			i.log.V(1).Error(err, "POST request to set LDAP group failed.",
				"IP", i.ip,
				"HardwareType", i.HardwareType(),
				"endpoint", endpoint,
				"step", helper.WhosCalling(),
				"Group", group.Group,
				"StatusCode", statusCode,
				"response", string(response),
			)
			continue
		}

		i.log.V(1).Info("LdapGroup parameters applied.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"Group", group.Group,
		)
	}

	return nil
}

// Ldap applies LDAP configuration params.
// Ldap implements the Configure interface.
func (i *Ilo) Ldap(cfg *cfgresources.Ldap) (err error) {
	if cfg.Server == "" {
		msg := "Ldap(): LDAP resource parameter Server required but not declared."
		err = errors.New(msg)
		i.log.V(1).Error(err, msg,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return err
	}

	if cfg.Port == 0 {
		msg := "Ldap(): LDAP resource parameter Port required but not declared."
		err = errors.New(msg)
		i.log.V(1).Error(err, msg,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return err
	}

	if cfg.BaseDn == "" {
		msg := "Ldap(): LDAP resource parameter BaseDn required but not declared."
		err = errors.New(msg)
		i.log.V(1).Error(err, msg,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return err
	}

	var enable int
	if cfg.Enable {
		enable = 1
	} else {
		enable = 0
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
		i.log.V(1).Error(err, "Ldap(): Unable to marshal directory payload to set LDAP config.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return err
	}

	endpoint := "json/directory"
	statusCode, response, err := i.post(endpoint, payload)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the POST request to %s.", statusCode, endpoint)
		} else {
			err = fmt.Errorf("POST request to %s failed with error: %s", endpoint, err.Error())
		}

		i.log.V(1).Error(err, "POST request to set Ldap config failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"endpoint", endpoint,
			"step", helper.WhosCalling(),
			"StatusCode", statusCode,
			"response", string(response),
		)
		return err
	}

	i.log.V(1).Info("Ldap parameters applied.", "IP", i.ip, "HardwareType", i.HardwareType())

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
	// Some general error?
	if err != nil {
		return nil, err
	}

	if statusCode == 500 {
		return []byte{}, fmt.Errorf("CSR being generated, retry later")
	}

	// If it's a not a 200 at this point, something else went wrong.
	if statusCode != 200 {
		return []byte{}, fmt.Errorf("Unexpected return code %d calling %s!", statusCode, endpoint)
	}

	r := new(csrResponse)
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
	type RedfishCertificatePayload struct {
		Certificate string `json:"Certificate"`
	}
	endpoint := "redfish/v1/Managers/1/SecurityService/HttpsCert/Actions/HpeHttpsCert.ImportCertificate/"
	certPayload := RedfishCertificatePayload{string(cert)}

	payload, err := json.Marshal(certPayload)
	if err != nil {
		return false, err
	}

	statusCode, _, err := i.post(endpoint, payload)
	if err != nil {
		return false, err
	}

	if statusCode != 200 {
		return false, fmt.Errorf("Unexpected return code %d calling %s!", statusCode, endpoint)
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
		return false, err
	}

	if updateAccessSettings {
		payload, err := json.Marshal(accessSettings)
		if err != nil {
			return false, fmt.Errorf("Error marshaling AccessSettings payload: %s", err)
		}

		endpoint := "json/access_settings"
		statusCode, _, err := i.post(endpoint, payload)
		if err != nil {
			return false, fmt.Errorf("Error calling access_settings: %s", err)
		}
		if statusCode != 200 {
			return false, fmt.Errorf("Non-200 response calling access_settings: %d", statusCode)
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
		if err != nil {
			return reset, fmt.Errorf("Error calling access_settings: %s", err)
		}
		if statusCode != 200 {
			return reset, fmt.Errorf("Non-200 response calling access_settings: %d", statusCode)
		}

		reset = true
	}

	return reset, nil
}

// Power settings
func (i *Ilo) Power(cfg *cfgresources.Power) error {
	if cfg.HPE == nil {
		return nil
	}

	// map of valid power_settings attributes to params passed to the iLO API
	powerRegulatorModes := map[string]string{
		"dynamic":     "dyn",
		"static_low":  "min",
		"static_high": "max",
		"os_control":  "osc",
	}

	configMode, exists := powerRegulatorModes[cfg.HPE.PowerRegulator]
	if cfg.HPE.PowerRegulator == "" || !exists {
		return fmt.Errorf("power_regulator parameter must be one of dynamic, static_log, static_high, os_control")
	}

	// check if a configuration update is required based on current setting
	config, changeRequired, err := i.cmpPowerSettings(configMode)
	if err != nil {
		return err
	}

	if !changeRequired {
		i.log.V(2).Info("power_regulator config - no change required.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"current mode", config.PowerMode,
			"expected mode", configMode,
		)
		return nil
	}

	i.log.V(2).Info("power_regulator change to be applied.",
		"IP", i.ip,
		"HardwareType", i.HardwareType(),
		"current mode", config.PowerMode,
		"expected mode", configMode,
	)

	config.SessionKey = i.sessionKey
	config.Method = "set"
	config.PowerMode = configMode

	payload, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("Error marshaling power_regulator payload: %s", err)
	}

	endpoint := "json/power_regulator"
	statusCode, _, err := i.post(endpoint, payload)
	if err != nil {
		return fmt.Errorf("Error calling power_regulator: %s", err)
	}
	if statusCode != 200 {
		return fmt.Errorf("Non-200 response calling power_regulator: %d", statusCode)
	}

	i.log.V(1).Info("power_regulator config applied.",
		"IP", i.ip,
		"HardwareType", i.HardwareType(),
	)

	return nil
}

// Bios method implements the Configure interface
func (i *Ilo) Bios(cfg *cfgresources.Bios) error {
	return nil
}
