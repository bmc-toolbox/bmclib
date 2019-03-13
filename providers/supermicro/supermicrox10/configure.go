package supermicrox10

import (
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/internal/helper"

	"github.com/google/go-querystring/query"
)

// This ensures the compiler errors if this type is missing
// a method that should be implmented to satisfy the Configure interface.
var _ devices.Configure = (*SupermicroX10)(nil)

// Resources returns a slice of supported resources and
// the order they are to be applied in.
func (s *SupermicroX10) Resources() []string {
	return []string{
		"user",
		"syslog",
		"network",
		"ntp",
		//"ldap", - ldap configuration is applied as part of ldap_group.
		"ldap_group",
	}
}

// ApplyCfg implements the Bmc interface
// this is to be deprecated.
func (s *SupermicroX10) ApplyCfg(config *cfgresources.ResourcesConfig) (err error) {
	return err
}

// SetLicense implements the Configure interface.
func (s *SupermicroX10) SetLicense(cfg *cfgresources.License) (err error) {
	return err
}

// Bios implements the Configure interface.
func (s *SupermicroX10) Bios(cfg *cfgresources.Bios) (err error) {
	return err
}

// Returns the UTC offset for a given timezone location
func timezoneToUtcOffset(location *time.Location) (offset int) {
	utcTime := time.Now().In(location)
	_, offset = utcTime.Zone()
	return offset
}

// Return bool value if the role is valid.
func (s *SupermicroX10) isRoleValid(role string) bool {

	validRoles := []string{"admin", "user"}
	for _, v := range validRoles {
		if role == v {
			return true
		}
	}

	return false
}

// returns a map of user accounts and their ids
func (s *SupermicroX10) queryUserAccounts() (userAccounts map[string]int, err error) {

	userAccounts = make(map[string]int)
	ipmi, err := s.query("CONFIG_INFO.XML=(0,0)")
	if err != nil {
		fmt.Println(err)
		return userAccounts, err
	}

	for idx, account := range ipmi.ConfigInfo.UserAccounts {
		if account.Name != "" {
			userAccounts[account.Name] = idx
		}
	}

	return userAccounts, err
}

// User applies the User configuration resource,
// if the user exists, it updates the users password,
// User implements the Configure interface.
// supermicro user accounts start with 1, account 0 which is a large empty string :\.
// nolint: gocyclo
func (s *SupermicroX10) User(users []*cfgresources.User) (err error) {

	currentUsers, err := s.queryUserAccounts()
	if err != nil {
		msg := "Unable to query current user accounts."
		log.WithFields(log.Fields{
			"IP":    s.ip,
			"Model": s.BmcType(),
			"Step":  helper.WhosCalling(),
			"Error": err,
		}).Warn(msg)
		return errors.New(msg)
	}

	userID := 1
	for _, user := range users {

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

		if !s.isRoleValid(user.Role) {
			msg := "User resource Role must be declared and a must be a valid role: 'admin' OR 'user'."
			log.WithFields(log.Fields{
				"step":     "applyUserParams",
				"Username": user.Name,
			}).Warn(msg)
			return errors.New(msg)
		}

		configUser := ConfigUser{}

		//if the user is enabled setup parameters
		if user.Enable {
			configUser.Username = user.Name
			configUser.Password = user.Password
			configUser.UserID = userID

			if user.Role == "admin" {
				configUser.NewPrivilege = 4
			} else if user.Role == "user" {
				configUser.NewPrivilege = 3
			}
		} else {
			_, uexists := currentUsers[user.Name]
			//if the user exists, delete it
			//this is done by sending an empty username along with,
			//the respective userid
			if uexists {
				configUser.Username = ""
				configUser.UserID = currentUsers[user.Name]
			} else {
				userID++
				continue
			}
		}

		endpoint := "config_user.cgi"
		form, _ := query.Values(configUser)
		statusCode, err := s.post(endpoint, &form)
		if err != nil || statusCode != 200 {
			msg := "POST request to set User config returned error."
			log.WithFields(log.Fields{
				"IP":         s.ip,
				"Model":      s.BmcType(),
				"Endpoint":   endpoint,
				"StatusCode": statusCode,
				"Step":       helper.WhosCalling(),
				"Error":      err,
			}).Warn(msg)
			return errors.New(msg)
		}

		log.WithFields(log.Fields{
			"IP":    s.ip,
			"Model": s.BmcType(),
			"User":  user.Name,
		}).Debug("User parameters applied.")

		userID++
	}

	return err
}

// Network method implements the Configure interface
// applies various network parameters.
func (s *SupermicroX10) Network(cfg *cfgresources.Network) (reset bool, err error) {

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

	endpoint := fmt.Sprintf("op.cgi")
	form, _ := query.Values(configPort)
	statusCode, err := s.post(endpoint, &form)
	if err != nil || statusCode != 200 {
		msg := "POST request to set Port config returned error."
		log.WithFields(log.Fields{
			"IP":         s.ip,
			"Model":      s.BmcType(),
			"Endpoint":   endpoint,
			"StatusCode": statusCode,
			"Step":       helper.WhosCalling(),
			"Error":      err,
		}).Warn(msg)
		return reset, errors.New(msg)
	}

	log.WithFields(log.Fields{
		"IP":    s.ip,
		"Model": s.BmcType(),
	}).Debug("Network config parameters applied.")
	return reset, err
}

// Ntp applies NTP configuration params
// Ntp implements the Configure interface.
func (s *SupermicroX10) Ntp(cfg *cfgresources.Ntp) (err error) {

	var enable string
	if cfg.Server1 == "" {
		log.WithFields(log.Fields{
			"step":  "applyNtpParams",
			"Model": s.BmcType(),
		}).Warn("NTP resource expects parameter: server1.")
		return
	}

	if cfg.Timezone == "" {
		log.WithFields(log.Fields{
			"step":  "applyNtpParams",
			"Model": s.BmcType(),
		}).Warn("NTP resource expects parameter: timezone.")
		return
	}

	tzLocation, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		log.WithFields(log.Fields{
			"step":              "applyNtpParams",
			"Model":             s.BmcType(),
			"Declared timezone": cfg.Timezone,
			"Error":             err,
		}).Warn("NTP resource declared parameter timezone invalid.")
		return
	}

	tzUtcOffset := timezoneToUtcOffset(tzLocation)

	if cfg.Enable != true {
		log.WithFields(log.Fields{
			"step":  "applyNtpParams",
			"Model": s.BmcType(),
		}).Debug("Ntp resource declared with enable: false.")
		return
	}

	enable = "on"

	t := time.Now().In(tzLocation)
	//Fri Jun 06 2018 14:28:25 GMT+0100 (CET)
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
		DstEn:              false, //daylight savings
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

	endpoint := fmt.Sprintf("op.cgi")
	form, _ := query.Values(configDateTime)
	statusCode, err := s.post(endpoint, &form)
	if err != nil || statusCode != 200 {
		msg := "POST request to set Syslog config returned error."
		log.WithFields(log.Fields{
			"IP":         s.ip,
			"Model":      s.BmcType(),
			"Endpoint":   endpoint,
			"StatusCode": statusCode,
			"Step":       helper.WhosCalling(),
			"Error":      err,
		}).Warn(msg)
		return errors.New(msg)
	}

	//
	log.WithFields(log.Fields{
		"IP":    s.ip,
		"Model": s.BmcType(),
	}).Debug("NTP config parameters applied.")
	return err
}

// Ldap applies LDAP configuration params.
// Ldap implements the Configure interface.
// Configuration for LDAP is applied in the LdapGroup method,
// since supermicros just support a single LDAP group.
func (s *SupermicroX10) Ldap(cfgLdap *cfgresources.Ldap) error {
	return nil
}

// LdapGroup applies LDAP and LDAP Group/Role related configuration,
// LdapGroup implements the Configure interface.
// Supermicro does not have any separate configuration for Ldap groups just for generic ldap
// nolint: gocyclo
func (s *SupermicroX10) LdapGroup(cfgGroup []*cfgresources.LdapGroup, cfgLdap *cfgresources.Ldap) (err error) {

	var enable string

	if cfgLdap.Server == "" {
		msg := "Ldap resource parameter Server required but not declared."
		log.WithFields(log.Fields{
			"step":  helper.WhosCalling(),
			"Model": s.BmcType(),
		}).Warn(msg)
		return errors.New(msg)
	}

	//first some preliminary checks
	if cfgLdap.Port == 0 {
		msg := "Ldap resource parameter Port required but not declared"
		log.WithFields(log.Fields{
			"step":  helper.WhosCalling(),
			"Model": s.BmcType(),
		}).Warn(msg)
		return errors.New(msg)
	}

	if cfgLdap.Enable != true {
		log.WithFields(log.Fields{
			"step":  helper.WhosCalling(),
			"Model": s.BmcType(),
		}).Debug("Ldap resource declared with enable: false.")
		return
	}

	enable = "on"

	if cfgLdap.BaseDn == "" {
		msg := "Ldap resource parameter BaseDn required but not declared."
		log.WithFields(log.Fields{
			"step":  helper.WhosCalling(),
			"Model": s.BmcType(),
		}).Warn(msg)
		return errors.New(msg)
	}

	serverIP, err := net.LookupIP(cfgLdap.Server)
	if err != nil || serverIP == nil {
		msg := "Unable to lookup the IP for ldap server hostname."
		log.WithFields(log.Fields{
			"step":  helper.WhosCalling(),
			"Model": s.BmcType(),
		}).Warn(msg)
		return errors.New(msg)
	}

	//for each ldap group setup config
	//since supermicro can work with just one Searchbase, we go with the 'user' role group
	for _, group := range cfgGroup {

		if !group.Enable {
			continue
		}

		if group.Role == "" {
			msg := "Ldap resource parameter Role required but not declared."
			log.WithFields(log.Fields{
				"Role": group.Role,
				"step": helper.WhosCalling(),
			}).Warn(msg)
			continue
		}

		if strings.Contains(group.Role, "admin") {
			continue
		}

		if group.Group == "" {
			msg := "Ldap resource parameter Group required but not declared."
			log.WithFields(log.Fields{
				"Role": group.Role,
				"step": helper.WhosCalling(),
			}).Warn(msg)
			return errors.New(msg)
		}

		if group.GroupBaseDn == "" {
			msg := "Ldap resource parameter GroupBaseDn required but not declared."
			log.WithFields(log.Fields{
				"Role":  group.Role,
				"Group": group.Group,
				"step":  helper.WhosCalling(),
			}).Warn(msg)
			return errors.New(msg)
		}

		if !s.isRoleValid(group.Role) {
			msg := "Ldap resource Role must be a valid role: admin OR user."
			log.WithFields(log.Fields{
				"Role": group.Role,
				"step": "applyLdapParams",
			}).Warn(msg)
			return errors.New(msg)
		}

		configLdap := ConfigLdap{
			Op:           "config_ldap",
			Enable:       enable,
			EnableSsl:    true,
			LdapIP:       fmt.Sprintf("%s", serverIP[0]),
			BaseDn:       group.Group,
			LdapPort:     cfgLdap.Port,
			BindDn:       cfgLdap.BindDn,
			BindPassword: "********", //default value
		}

		endpoint := fmt.Sprintf("op.cgi")
		form, _ := query.Values(configLdap)
		statusCode, err := s.post(endpoint, &form)
		if err != nil || statusCode != 200 {
			msg := "POST request to set Ldap config returned error."
			log.WithFields(log.Fields{
				"IP":         s.ip,
				"Model":      s.BmcType(),
				"Endpoint":   endpoint,
				"StatusCode": statusCode,
				"Step":       helper.WhosCalling(),
				"Error":      err,
			}).Warn(msg)
			return errors.New(msg)
		}
	}

	log.WithFields(log.Fields{
		"IP":    s.ip,
		"Model": s.BmcType(),
	}).Debug("Ldap config parameters applied.")
	return err
}

// Syslog applies the Syslog configuration resource
// Syslog implements the Configure interface
func (s *SupermicroX10) Syslog(cfg *cfgresources.Syslog) (err error) {

	var port int

	if cfg.Server == "" {
		msg := "Syslog resource expects parameter: Server."
		log.WithFields(log.Fields{
			"step":  helper.WhosCalling(),
			"Model": s.BmcType(),
		}).Warn(msg)
		return errors.New(msg)
	}

	if cfg.Port == 0 {
		log.WithFields(log.Fields{
			"step":  helper.WhosCalling(),
			"Model": s.BmcType(),
		}).Debug("Syslog resource port set to default: 514.")
		port = 514
	} else {
		port = cfg.Port
	}

	if cfg.Enable != true {
		log.WithFields(log.Fields{
			"step":  helper.WhosCalling(),
			"Model": s.BmcType(),
		}).Debug("Syslog resource declared with disable.")
	}

	serverIP, err := net.LookupIP(cfg.Server)
	if err != nil || serverIP == nil {
		msg := "Unable to lookup IP for syslog server hostname, yes supermicros requires the Syslog server IP :|."
		log.WithFields(log.Fields{
			"step":  helper.WhosCalling(),
			"Model": s.BmcType(),
		}).Warn(msg)
		return errors.New(msg)
	}

	configSyslog := ConfigSyslog{
		Op:          "config_syslog",
		SyslogIP1:   fmt.Sprintf("%s", serverIP[0]),
		SyslogPort1: port,
		Enable:      cfg.Enable,
	}

	endpoint := fmt.Sprintf("op.cgi")
	form, _ := query.Values(configSyslog)
	statusCode, err := s.post(endpoint, &form)
	if err != nil || statusCode != 200 {
		msg := "POST request to set Syslog config returned error."
		log.WithFields(log.Fields{
			"IP":         s.ip,
			"Model":      s.BmcType(),
			"Endpoint":   endpoint,
			"StatusCode": statusCode,
			"step":       helper.WhosCalling(),
			"Error":      err,
		}).Warn(msg)
		return errors.New(msg)
	}
	//returns okStarting Syslog daemon if successful

	log.WithFields(log.Fields{
		"IP":    s.ip,
		"Model": s.BmcType(),
	}).Debug("Syslog config parameters applied.")
	return err
}

// posts a urlencoded form to the given endpoint
// nolint: gocyclo
func (s *SupermicroX10) post(endpoint string, form *url.Values) (statusCode int, err error) {
	err = s.httpLogin()
	if err != nil {
		return statusCode, err
	}

	u, err := url.Parse(fmt.Sprintf("https://%s/cgi/%s", s.ip, endpoint))
	if err != nil {
		return statusCode, err
	}
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return statusCode, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	for _, cookie := range s.httpClient.Jar.Cookies(u) {
		if cookie.Name == "SID" && cookie.Value != "" {
			req.AddCookie(cookie)
		}
	}

	if log.GetLevel() == log.DebugLevel {
		fmt.Println(fmt.Sprintf("https://%s/cgi/%s", s.ip, endpoint))
		dump, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			log.Println("[Request]")
			log.Println(">>>>>>>>>>>>>>>")
			log.Printf("%s\n\n", dump)
			log.Println(">>>>>>>>>>>>>>>")
		}
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return statusCode, err
	}
	defer resp.Body.Close()

	if log.GetLevel() == log.DebugLevel {
		dump, err := httputil.DumpResponse(resp, true)
		if err == nil {
			log.Println("[Response]")
			log.Println("<<<<<<<<<<<<<<")
			log.Printf("%s\n\n", dump)
			log.Println("<<<<<<<<<<<<<<")
		}
	}

	statusCode = resp.StatusCode
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return statusCode, err
	}
	//fmt.Printf("-->> %d\n", resp.StatusCode)
	//fmt.Printf("%s\n", body)
	return statusCode, err
}

// GenerateCSR generates a CSR request on the BMC.
// GenerateCSR implements the Configure interface.
func (s *SupermicroX10) GenerateCSR(cert *cfgresources.HTTPSCertAttributes) ([]byte, error) {
	return []byte{}, nil
}

// UploadHTTPSCert uploads the given CRT cert,
// UploadHTTPSCert implements the Configure interface.
func (s *SupermicroX10) UploadHTTPSCert(cert []byte, fileName string) (bool, error) {
	return false, nil
}

// CurrentHTTPSCert returns the current x509 certficates configured on the BMC
// CurrentHTTPSCert implements the Configure interface.
func (s *SupermicroX10) CurrentHTTPSCert() (c []*x509.Certificate, e error) {
	return c, e
}
