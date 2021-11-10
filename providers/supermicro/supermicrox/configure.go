package supermicrox

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/internal/helper"

	"github.com/google/go-querystring/query"
)

// This ensures the compiler errors if this type is missing
// a method that should be implmented to satisfy the Configure interface.
var _ devices.Configure = (*SupermicroX)(nil)

// Resources returns a slice of supported resources and
// the order they are to be applied in.
func (s *SupermicroX) Resources() []string {
	return []string{
		"user",
		"syslog",
		"network",
		"ntp",
		//"ldap", - ldap configuration is applied as part of ldap_group.
		"ldap_group",
		"https_cert",
	}
}

// ApplyCfg implements the Bmc interface
// this is to be deprecated.
func (s *SupermicroX) ApplyCfg(config *cfgresources.ResourcesConfig) (err error) {
	return err
}

// Power implemented the Configure interface
func (s *SupermicroX) Power(cfg *cfgresources.Power) (err error) {
	return err
}

// SetLicense implements the Configure interface.
func (s *SupermicroX) SetLicense(cfg *cfgresources.License) (err error) {
	return err
}

// Bios implements the Configure interface.
func (s *SupermicroX) Bios(cfg *cfgresources.Bios) (err error) {
	return err
}

// Returns the UTC offset for a given timezone location
func timezoneToUtcOffset(location *time.Location) (offset int) {
	utcTime := time.Now().In(location)
	_, offset = utcTime.Zone()
	return offset
}

// Return bool value if the role is valid.
func (s *SupermicroX) isRoleValid(role string) bool {
	validRoles := []string{"admin", "user"}
	for _, v := range validRoles {
		if role == v {
			return true
		}
	}

	return false
}

// returns a map of user accounts and their ids
func (s *SupermicroX) queryUserAccounts() (userAccounts map[int]string, err error) {
	userAccounts = make(map[int]string)
	ipmi, err := s.query("CONFIG_INFO.XML=(0,0)")
	if err != nil {
		s.log.V(1).Error(err, "queryUserAccounts(): Error querying user accounts.")
		return userAccounts, err
	}

	for idx, account := range ipmi.ConfigInfo.UserAccounts {
		userAccounts[idx] = account.Name
	}

	return userAccounts, err
}

// User applies the User configuration resource,
// if the user exists, it updates the users password,
// User implements the Configure interface.
// supermicro user accounts start with 1, account 0 which is a large empty string :\.
// nolint: gocyclo
func (s *SupermicroX) User(users []*cfgresources.User) (err error) {
	currentUsers, err := s.queryUserAccounts()
	if err != nil {
		msg := "Unable to query existing users."
		s.log.V(1).Error(err, msg,
			"ip", s.ip,
			"HardwareType", s.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return errors.New(msg)
	}

	for _, user := range users {
		if user.Name == "" {
			msg := "User resource expects parameter: Name."
			s.log.V(1).Info(msg, "step", "applyUserParams")
			return errors.New(msg)
		}

		if user.Password == "" {
			msg := "User resource expects parameter: Password."
			s.log.V(1).Info(msg, "step", "applyUserParams", "username", user.Name)
			return errors.New(msg)
		}

		if !s.isRoleValid(user.Role) {
			msg := "User resource Role must be declared and a must be a valid role: 'admin' OR 'user'."
			s.log.V(1).Info(msg, "step", "applyUserParams", "username", user.Name)
			return errors.New(msg)
		}

		configUser := ConfigUser{
			Username:     user.Name,
			Password:     user.Password,
			NewPrivilege: 3,
			UserID:       1,
		}
		if user.Role == "admin" {
			configUser.NewPrivilege = 4
		}
		var userID int
		comparisonNum := 10
		for id, name := range currentUsers {
			if name == user.Name {
				userID = id
				break
			} else if name == "" {
				if id < comparisonNum {
					userID = id
					comparisonNum = id
				}
			}
		}
		if userID == 0 {
			return errors.New("no user slots available")
		}
		configUser.UserID = userID

		if !user.Enable {
			configUser.Username = ""
		}

		endpoint := "config_user.cgi"
		form, _ := query.Values(configUser)
		statusCode, err := s.post(endpoint, &form, []byte{}, "")
		if err != nil || statusCode != 200 {
			if err == nil {
				err = fmt.Errorf("Received a non-200 status code from the POST request to %s.", endpoint)
			} else {
				err = fmt.Errorf("POST request to %s failed with error: %s", endpoint, err.Error())
			}

			s.log.V(1).Error(err, "POST request to set User config failed.",
				"ip", s.ip,
				"HardwareType", s.HardwareType(),
				"endpoint", endpoint,
				"StatusCode", statusCode,
				"step", helper.WhosCalling(),
			)
			return err
		}

		s.log.V(1).Info("User parameters applied.", "ip", s.ip, "HardwareType", s.HardwareType(), "user", user.Name)
	}

	return err
}

// Network method implements the Configure interface
// applies various network parameters.
func (s *SupermicroX) Network(cfg *cfgresources.Network) (reset bool, err error) {
	sshPort := 22

	if cfg.SSHPort != 0 && cfg.SSHPort != sshPort {
		sshPort = cfg.SSHPort
	}

	configPort := ConfigPort{
		Op:                "config_port",
		HTTPPort:          80,
		HTTPSPort:         443,
		IkvmPort:          5900,
		VMPort:            623,
		SSHPort:           sshPort,
		WsmanPort:         5985,
		SnmpPort:          161,
		httpEnable:        true,
		httpsEnable:       true,
		IkvmEnable:        true,
		VMEnable:          true,
		SSHEnable:         cfg.SSHEnable,
		SnmpEnable:        false,
		WsmanEnable:       false,
		SslRedirectEnable: true,
	}

	endpoint := "op.cgi"
	form, _ := query.Values(configPort)
	statusCode, err := s.post(endpoint, &form, []byte{}, "")
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a non-200 status code from the POST request to %s.", endpoint)
		} else {
			err = fmt.Errorf("POST request to %s failed with error: %s", endpoint, err.Error())
		}

		s.log.V(1).Error(err, "POST request to set Port config failed.",
			"ip", s.ip,
			"HardwareType", s.HardwareType(),
			"endpoint", endpoint,
			"statusCode", statusCode,
			"step", helper.WhosCalling(),
		)
		return false, err
	}

	s.log.V(1).Info("Network config parameters applied.", "ip", s.ip, "HardwareType", s.HardwareType())
	return reset, err
}

// Ntp applies NTP configuration params
// Ntp implements the Configure interface.
func (s *SupermicroX) Ntp(cfg *cfgresources.Ntp) (err error) {
	var enable string
	if cfg.Server1 == "" {
		s.log.V(1).Info("NTP resource expects parameter: server1.",
			"step", "applyNtpParams",
			"HardwareType", s.HardwareType())
		return
	}

	if cfg.Timezone == "" {
		s.log.V(1).Info("NTP resource expects parameter: timezone.",
			"step", "applyNtpParams",
			"HardwareType", s.HardwareType())
		return
	}

	tzLocation, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		s.log.V(1).Error(err, "Ntp(): Invalid timezone parameter.",
			"step", "applyNtpParams",
			"HardwareType", s.HardwareType(),
			"Timezone", cfg.Timezone,
		)
		return
	}

	tzUtcOffset := timezoneToUtcOffset(tzLocation)

	if !cfg.Enable {
		s.log.V(1).Info("Ntp resource declared with enable: false.",
			"step", "applyNtpParams",
			"HardwareType", s.HardwareType())
		return
	}

	enable = "on"

	t := time.Now().In(tzLocation)
	// Fri Jun 06 2018 14:28:25 GMT+0100 (CET)
	ts := fmt.Sprintf("%s %d %d:%d:%d %s (%s)",
		t.Format("Fri Jun 01"),
		t.Year(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		t.Format("GMT+0200"),
		tzLocation)

	configDateTime := ConfigDateTime{
		Op:                 "config_date_time",
		Timezone:           tzUtcOffset,
		DstEn:              false,
		Enable:             enable,
		NtpServerPrimary:   cfg.Server1,
		NtpServerSecondary: cfg.Server2,
		Year:               t.Year(),
		Month:              int(t.Month()),
		Day:                int(t.Day()),
		Hour:               int(t.Hour()),
		Minute:             int(t.Minute()),
		Second:             int(t.Second()),
		TimeStamp:          ts,
	}

	endpoint := "op.cgi"
	form, _ := query.Values(configDateTime)
	statusCode, err := s.post(endpoint, &form, []byte{}, "")
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a non-200 status code from the POST request to %s.", endpoint)
		} else {
			err = fmt.Errorf("POST request to %s failed with error: %s", endpoint, err.Error())
		}

		s.log.V(1).Error(err, "POST request to set NTP config failed.",
			"ip", s.ip,
			"HardwareType", s.HardwareType(),
			"endpoint", endpoint,
			"statusCode", statusCode,
			"step", helper.WhosCalling(),
		)
		return err
	}

	s.log.V(1).Info("NTP config parameters applied.",
		"ip", s.ip,
		"HardwareType", s.HardwareType())
	return nil
}

// Ldap applies LDAP configuration params.
// Ldap implements the Configure interface.
// Configuration for LDAP is applied in the LdapGroup method,
// since supermicros just support a single LDAP group.
func (s *SupermicroX) Ldap(cfgLdap *cfgresources.Ldap) error {
	return nil
}

// LdapGroups applies LDAP and LDAP Group/Role related configuration,
// LdapGroups implements the Configure interface.
// Supermicro does not have any separate configuration for Ldap groups just for generic ldap
// nolint: gocyclo
func (s *SupermicroX) LdapGroups(cfgGroups []*cfgresources.LdapGroup, cfgLdap *cfgresources.Ldap) (err error) {
	if cfgLdap.Server == "" {
		msg := "Ldap resource parameter Server required but not declared."
		s.log.V(1).Info(msg, "step", helper.WhosCalling(), "HardwareType", s.HardwareType())
		return errors.New(msg)
	}

	if cfgLdap.Port == 0 {
		msg := "Ldap resource parameter Port required but not declared"
		s.log.V(1).Info(msg,
			"step", helper.WhosCalling(),
			"HardwareType", s.HardwareType())
		return errors.New(msg)
	}

	if !cfgLdap.Enable {
		s.log.V(1).Info("Ldap resource declared with enable: false.",
			"step", helper.WhosCalling(),
			"HardwareType", s.HardwareType())
		return
	}

	if cfgLdap.BaseDn == "" {
		msg := "Ldap resource parameter BaseDn required but not declared."
		s.log.V(1).Info(msg, "step", helper.WhosCalling(), "HardwareType", s.HardwareType())
		return errors.New(msg)
	}

	serverIP, err := net.LookupIP(cfgLdap.Server)
	if err != nil || serverIP == nil {
		msg := "Unable to lookup the IP for ldap server hostname."
		s.log.V(1).Info(msg, "step", helper.WhosCalling(), "HardwareType", s.HardwareType())
		return errors.New(msg)
	}

	// Since SuperMicro can work with just one search base, we go with the "user" role group.
	for _, group := range cfgGroups {
		if !group.Enable {
			continue
		}

		if group.Role == "" {
			msg := "Ldap resource parameter Role required but not declared."
			s.log.V(1).Info(msg, "step", helper.WhosCalling(), "role", group.Role)
			continue
		}

		if strings.Contains(group.Role, "admin") {
			continue
		}

		if group.Group == "" {
			msg := "Ldap resource parameter Group required but not declared."
			s.log.V(1).Info(msg, "step", helper.WhosCalling(), "role", group.Role)
			return errors.New(msg)
		}

		if group.GroupBaseDn == "" {
			msg := "Ldap resource parameter GroupBaseDn required but not declared."
			s.log.V(1).Info(msg,
				"step", helper.WhosCalling(),
				"group", group.Group,
				"role", group.Role)
			return errors.New(msg)
		}

		if !s.isRoleValid(group.Role) {
			msg := "Ldap resource Role must be a valid role: admin OR user."
			s.log.V(1).Info(msg, "step", helper.WhosCalling(), "role", group.Role)
			return errors.New(msg)
		}

		configLdap := ConfigLdap{
			Op:           "config_ldap",
			Enable:       "on",
			EnableSsl:    true,
			LdapIP:       string(serverIP[0]),
			BaseDn:       group.Group,
			LdapPort:     cfgLdap.Port,
			BindDn:       cfgLdap.BindDn,
			BindPassword: "********", // default value
		}

		endpoint := "op.cgi"
		form, _ := query.Values(configLdap)
		statusCode, err := s.post(endpoint, &form, []byte{}, "")
		if err != nil || statusCode != 200 {
			if err == nil {
				err = fmt.Errorf("Received a non-200 status code from the POST request to %s.", endpoint)
			} else {
				err = fmt.Errorf("POST request to %s failed with error: %s", endpoint, err.Error())
			}

			s.log.V(1).Error(err, "POST request to set LDAP group config failed.",
				"step", helper.WhosCalling(),
				"ip", s.ip,
				"HardwareType", s.HardwareType(),
				"endpoint", endpoint,
				"StatusCode", statusCode,
				"Group", group.Group,
			)
			return err
		}
	}

	s.log.V(1).Info("LDAP config parameters applied.", "ip", s.ip, "HardwareType", s.HardwareType())
	return err
}

// Syslog applies the Syslog configuration resource
// Syslog implements the Configure interface
// this also enables alerts from the BMC
func (s *SupermicroX) Syslog(cfg *cfgresources.Syslog) (err error) {
	var port int

	if cfg.Server == "" {
		msg := "Syslog resource expects parameter: Server."
		s.log.V(1).Info(msg, "step", helper.WhosCalling(), "HardwareType", s.HardwareType())
		return errors.New(msg)
	}

	if cfg.Port == 0 {
		msg := "Syslog resource port set to default: 514."
		s.log.V(1).Info(msg, "step", helper.WhosCalling(), "HardwareType", s.HardwareType())
		port = 514
	} else {
		port = cfg.Port
	}

	if !cfg.Enable {
		msg := "Syslog resource declared with disable."
		s.log.V(1).Info(msg, "step", helper.WhosCalling(), "HardwareType", s.HardwareType())
	}

	serverIP, err := net.LookupIP(cfg.Server)
	if err != nil || serverIP == nil {
		msg := "Unable to lookup IP for syslog server hostname, yes supermicros requires the Syslog server IP :|."
		s.log.V(1).Info(msg, "step", helper.WhosCalling(), "HardwareType", s.HardwareType())
		return errors.New(msg)
	}

	configSyslog := ConfigSyslog{
		Op:          "config_syslog",
		SyslogIP1:   string(serverIP[0]),
		SyslogPort1: port,
		Enable:      cfg.Enable,
	}

	endpoint := "op.cgi"
	form, _ := query.Values(configSyslog)

	statusCode, err := s.post(endpoint, &form, []byte{}, "")
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a non-200 status code from the POST request to %s.", endpoint)
		} else {
			err = fmt.Errorf("POST request to %s failed with error: %s", endpoint, err.Error())
		}

		s.log.V(1).Error(err, "POST request to set Syslog config returned error.",
			"step", helper.WhosCalling(),
			"ip", s.ip,
			"HardwareType", s.HardwareType(),
			"endpoint", endpoint,
			"StatusCode", statusCode,
		)
		return err
	}

	// enable maintenance events
	endpoint = "system_event_log.cgi"

	form = make(url.Values)
	form.Add("enable", "1")

	statusCode, err = s.post(endpoint, &form, []byte{}, "")
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a non-200 status code from the POST request to %s.", endpoint)
		} else {
			err = fmt.Errorf("POST request to %s failed with error: %s", endpoint, err.Error())
		}

		s.log.V(1).Error(err, "POST request to enable maintenance alerts failed.",
			"step", helper.WhosCalling(),
			"ip", s.ip,
			"HardwareType", s.HardwareType(),
			"endpoint", endpoint,
			"StatusCode", statusCode,
		)
		return err
	}

	s.log.V(1).Info("Syslog config parameters applied.", "ip", s.ip, "HardwareType", s.HardwareType())
	return err
}

// GenerateCSR generates a CSR request on the BMC.
// GenerateCSR implements the Configure interface.
func (s *SupermicroX) GenerateCSR(cert *cfgresources.HTTPSCertAttributes) ([]byte, error) {
	return []byte{}, nil
}

// UploadHTTPSCert uploads the given CRT cert,
// UploadHTTPSCert implements the Configure interface.
// 1. Upload the certificate and key pair
// 2. delay for a second (to let the BMC process the certificate)
// 3. Get the BMC to validate the certificate: SSL_VALIDATE.XML	(0,0)
// 4. delay for a second
// 5. Request for the current: SSL_STATUS.XML	(0,0)
func (s *SupermicroX) UploadHTTPSCert(cert []byte, certFileName string, key []byte, keyFileName string) (bool, error) {
	endpoint := "upload_ssl.cgi"

	// setup a buffer for our multipart form
	var form bytes.Buffer
	w := multipart.NewWriter(&form)

	// setup the ssl cert part
	certWriter, err := w.CreateFormFile("/tmp/cert.pem", "cert.pem")
	if err != nil {
		return false, err
	}

	_, err = io.Copy(certWriter, bytes.NewReader(cert))
	if err != nil {
		return false, err
	}

	// setup the ssl key part
	keyWriter, err := w.CreateFormFile("/tmp/key.pem", "key.pem")
	if err != nil {
		return false, err
	}
	_, err = io.Copy(keyWriter, bytes.NewReader(key))
	if err != nil {
		return false, err
	}

	// close multipart writer - adds the teminating boundary.
	w.Close()

	// 1. upload
	status, err := s.post(endpoint, &url.Values{}, form.Bytes(), w.FormDataContentType())
	if err != nil || status != 200 {
		msg := "Cert form upload POST request failed, expected 200."
		s.log.V(1).Info(msg,
			"step", helper.WhosCalling(),
			"ip", s.ip,
			"HardwareType", s.HardwareType(),
			"endpoint", endpoint,
			"statusCode", status,
			"error", internal.ErrStringOrEmpty(err))
		return false, err
	}

	// 2. delay
	time.Sleep(1 * time.Second)

	// 3. Get BMC to validate uploaded cert
	err = s.validateSSL()
	if err != nil {
		return false, err
	}

	// 4. delay
	time.Sleep(1 * time.Second)

	// 5. Get cert status
	err = s.statusSSL()
	if err != nil {
		return false, err
	}

	return true, nil
}

// The second part of the certificate upload process,
// we get the BMC to validate the uploaded SSL certificate.
func (s *SupermicroX) validateSSL() error {
	v := url.Values{}
	v.Set("SSL_VALIDATE.XML", "(0,0)")

	endpoint := "ipmi.cgi"
	status, err := s.post(endpoint, &v, []byte{}, "")
	if err != nil || status != 200 {
		msg := "Cert validate POST request failed, expected 200."
		s.log.V(1).Info(msg,
			"step", helper.WhosCalling(),
			"ip", s.ip,
			"HardwareType", s.HardwareType(),
			"endpoint", endpoint,
			"statusCode", status,
			"error", internal.ErrStringOrEmpty(err))
		return err
	}

	return nil
}

// The third part of the certificate upload process
// Get the current status of the certificate.
// POST https://10.193.251.43/cgi/ipmi.cgi SSL_STATUS.XML: (0,0)
func (s *SupermicroX) statusSSL() error {
	v := url.Values{}
	v.Add("SSL_STATUS.XML", "(0,0)")

	endpoint := "ipmi.cgi"
	status, err := s.post(endpoint, &v, []byte{}, "")
	if err != nil || status != 200 {
		msg := "Cert status POST request failed, expected 200."
		s.log.V(1).Info(msg,
			"step", helper.WhosCalling(),
			"ip", s.ip,
			"HardwareType", s.HardwareType(),
			"endpoint", endpoint,
			"statusCode", status,
			"error", internal.ErrStringOrEmpty(err))
		return err
	}

	return nil
}
