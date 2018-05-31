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
)

// returns the calling function.
func funcName() string {
	pc, _, _, _ := runtime.Caller(1)
	return runtime.FuncForPC(pc).Name()
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

func (s *SupermicroX10) applyUserParams(users []*cfgresources.User) (err error) {
	return err
}

func (s *SupermicroX10) applyNtpParams(cfg *cfgresources.Ntp) (err error) {
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
	err = s.post(endpoint, &form, false)
	if err != nil {
		msg := "POST request to set Syslog config returned error."
		log.WithFields(log.Fields{
			"IP":       s.ip,
			"Model":    s.ModelId(),
			"endpoint": endpoint,
			"step":     funcName(),
			"Error":    err,
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
func (s *SupermicroX10) post(endpoint string, form *url.Values, debug bool) (err error) {

	u, err := url.Parse(fmt.Sprintf("https://%s/cgi/%s", s.ip, endpoint))
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return err
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
		return err
	}
	defer resp.Body.Close()
	if debug {
		dump, err := httputil.DumpResponse(resp, true)
		if err == nil {
			fmt.Printf("%s\n\n", dump)
		}
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	//fmt.Printf("-->> %d\n", resp.StatusCode)
	//fmt.Printf("%s\n", body)
	return err
}
