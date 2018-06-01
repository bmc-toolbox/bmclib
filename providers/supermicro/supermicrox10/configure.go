package supermicrox10

import (
	"errors"
	"fmt"
	"github.com/google/go-querystring/query"
	"github.com/ncode/bmclib/cfgresources"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// returns the calling function.
func funcName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
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

// returns a map of user accounts and thier ids
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

func (s *SupermicroX10) ApplyCfg(config *cfgresources.ResourcesConfig) (err error) {

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
				err := s.applyUserParams(userAccounts.([]*cfgresources.User))
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "ApplyCfg",
						"resource": cfg.Field(r).Kind(),
						"IP":       s.ip,
						"Model":    s.ModelId(),
						"Error":    err,
					}).Warn("Unable to set User config.")
				}

			case "Syslog":
				syslogCfg := cfg.Field(r).Interface().(*cfgresources.Syslog)
				err := s.applySyslogParams(syslogCfg)
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "ApplyCfg",
						"resource": cfg.Field(r).Kind(),
						"IP":       s.ip,
						"Model":    s.ModelId(),
						"Error":    err,
					}).Warn("Unable to set Syslog config.")
				}
			case "Network":
				fmt.Printf("%s: %v : %s\n", resourceName, cfg.Field(r), cfg.Field(r).Kind())
			case "Ntp":
				ntpCfg := cfg.Field(r).Interface().(*cfgresources.Ntp)
				err := s.applyNtpParams(ntpCfg)
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "ApplyCfg",
						"resource": cfg.Field(r).Kind(),
						"IP":       s.ip,
						"Model":    s.ModelId(),
						"Error":    err,
					}).Warn("Unable to set NTP config.")
				}
			case "LdapGroup":
				continue
			case "Ldap":
				ldapCfg := cfg.Field(r).Interface()
				err := s.applyLdapParams(ldapCfg.(*cfgresources.Ldap))
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "applyLdapParams",
						"resource": "Ldap",
						"IP":       s.ip,
						"Model":    s.ModelId(),
						"Error":    err,
					}).Warn("applyLdapParams returned error.")
				}
			case "Ssl":
				fmt.Printf("%s: %v : %s\n", resourceName, cfg.Field(r), cfg.Field(r).Kind())
			default:
				log.WithFields(log.Fields{
					"step":     "ApplyCfg",
					"Resource": cfg.Field(r).Kind(),
				}).Warn("Unknown resource definition.")
				//fmt.Printf("%v\n", cfg.Field(r))

			}
		}
	}

	return err
}

// attempts to add the user
// if the user exists, update the users password.
// supermicro user accounts start with 1, account 0 which is a large empty string :\.
func (s *SupermicroX10) applyUserParams(users []*cfgresources.User) (err error) {

	currentUsers, err := s.queryUserAccounts()
	if err != nil {
		msg := "Unable to query current user accounts."
		log.WithFields(log.Fields{
			"IP":    s.ip,
			"Model": s.ModelId(),
			"Step":  funcName(),
			"Error": err,
		}).Warn(msg)
		return errors.New(msg)
	}

	userId := 1
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
			configUser.UserId = userId

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
				configUser.UserId = currentUsers[user.Name]
			} else {
				userId += 1
				continue
			}
		}

		endpoint := "config_user.cgi"
		form, _ := query.Values(configUser)
		statusCode, err := s.post(endpoint, &form, false)
		if err != nil || statusCode != 200 {
			msg := "POST request to set User config returned error."
			log.WithFields(log.Fields{
				"IP":         s.ip,
				"Model":      s.ModelId(),
				"Endpoint":   endpoint,
				"StatusCode": statusCode,
				"Step":       funcName(),
				"Error":      err,
			}).Warn(msg)
			return errors.New(msg)
		}

		log.WithFields(log.Fields{
			"IP":    s.ip,
			"Model": s.ModelId(),
			"User":  user.Name,
		}).Info("User parameters applied.")

		userId += 1
	}

	return err
}

func (s *SupermicroX10) applyNtpParams(cfg *cfgresources.Ntp) (err error) {

	var enable string
	if cfg.Server1 == "" {
		log.WithFields(log.Fields{
			"step":  "applyNtpParams",
			"Model": s.ModelId(),
		}).Warn("NTP resource expects parameter: server1.")
		return
	}

	if cfg.Timezone == "" {
		log.WithFields(log.Fields{
			"step":  "applyNtpParams",
			"Model": s.ModelId(),
		}).Warn("NTP resource expects parameter: timezone.")
		return
	}

	tzLocation, err := time.LoadLocation(cfg.Timezone)
	if err != nil {
		log.WithFields(log.Fields{
			"step":              "applyNtpParams",
			"Model":             s.ModelId(),
			"Declared timezone": cfg.Timezone,
			"Error":             err,
		}).Warn("NTP resource declared parameter timezone invalid.")
		return
	}

	tzUtcOffset := timezoneToUtcOffset(tzLocation)

	if cfg.Enable != true {
		enable = "off"
		log.WithFields(log.Fields{
			"step":  "applyNtpParams",
			"Model": s.ModelId(),
		}).Debug("Ntp resource declared with enable: false.")
		return
	} else {
		enable = "on"
	}

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
	statusCode, err := s.post(endpoint, &form, false)
	if err != nil || statusCode != 200 {
		msg := "POST request to set Syslog config returned error."
		log.WithFields(log.Fields{
			"IP":         s.ip,
			"Model":      s.ModelId(),
			"Endpoint":   endpoint,
			"StatusCode": statusCode,
			"Step":       funcName(),
			"Error":      err,
		}).Warn(msg)
		return errors.New(msg)
	}

	//
	log.WithFields(log.Fields{
		"IP":    s.ip,
		"Model": s.ModelId(),
	}).Info("NTP config parameters applied.")
	return err
}

func (s *SupermicroX10) applyLdapParams(cfg *cfgresources.Ldap) (err error) {
	return err
}

func (s *SupermicroX10) applySyslogParams(cfg *cfgresources.Syslog) (err error) {

	var port int

	if cfg.Server == "" {
		msg := "Syslog resource expects parameter: Server."
		log.WithFields(log.Fields{
			"step":  funcName(),
			"Model": s.ModelId(),
		}).Warn(msg)
		return errors.New(msg)
	}

	if cfg.Port == 0 {
		log.WithFields(log.Fields{
			"step":  funcName(),
			"Model": s.ModelId(),
		}).Debug("Syslog resource port set to default: 514.")
		port = 514
	} else {
		port = cfg.Port
	}

	if cfg.Enable != true {
		log.WithFields(log.Fields{
			"step":  funcName(),
			"Model": s.ModelId(),
		}).Debug("Syslog resource declared with disable.")
	}

	serverIp, err := net.LookupIP(cfg.Server)
	if err != nil || serverIp == nil {
		msg := "Unable to lookup IP for syslog server hostname, yes supermicros requires the Syslog server IP :|."
		log.WithFields(log.Fields{
			"step":  funcName(),
			"Model": s.ModelId(),
		}).Warn(msg)
		return errors.New(msg)
	}

	configSyslog := ConfigSyslog{
		Op:          "config_syslog",
		SyslogIp1:   fmt.Sprintf("%s", serverIp[0]),
		SyslogPort1: port,
		Enable:      cfg.Enable,
	}

	endpoint := fmt.Sprintf("op.cgi")
	form, _ := query.Values(configSyslog)
	statusCode, err := s.post(endpoint, &form, false)
	if err != nil || statusCode != 200 {
		msg := "POST request to set Syslog config returned error."
		log.WithFields(log.Fields{
			"IP":         s.ip,
			"Model":      s.ModelId(),
			"Endpoint":   endpoint,
			"StatusCode": statusCode,
			"step":       funcName(),
			"Error":      err,
		}).Warn(msg)
		return errors.New(msg)
	}
	//returns okStarting Syslog daemon if successful

	log.WithFields(log.Fields{
		"IP":    s.ip,
		"Model": s.ModelId(),
	}).Info("Syslog config parameters applied.")
	return err
}

// posts a urlencoded form to the given endpoint
func (s *SupermicroX10) post(endpoint string, form *url.Values, debug bool) (statusCode int, err error) {

	u, err := url.Parse(fmt.Sprintf("https://%s/cgi/%s", s.ip, endpoint))
	if err != nil {
		return statusCode, err
	}
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return statusCode, err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	for _, cookie := range s.client.Jar.Cookies(u) {
		if cookie.Name == "SID" && cookie.Value != "" {
			req.AddCookie(cookie)
		}
	}
	//XXX to debug
	//fmt.Printf("--> %+v\n", form.Encode())
	//return err
	if debug {
		fmt.Println(fmt.Sprintf("https://%s/cgi/%s", s.ip, endpoint))
		dump, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			fmt.Printf("%s\n\n", dump)
		}
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return statusCode, err
	}
	defer resp.Body.Close()
	if debug {
		dump, err := httputil.DumpResponse(resp, true)
		if err == nil {
			fmt.Printf("%s\n\n", dump)
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
