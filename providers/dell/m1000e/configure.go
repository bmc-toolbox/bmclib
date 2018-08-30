package m1000e

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/google/go-querystring/query"
	log "github.com/sirupsen/logrus"
)

func (m *M1000e) ApplyCfg(config *cfgresources.ResourcesConfig) (err error) {
	//for each field in the struct that is not nil
	//call the respective getCfg, then post the data to the respective pages.

	cfg := reflect.ValueOf(config).Elem()

	//Each Field in ResourcesConfig struct is a ptr to a resource,
	//Here we figure the resources to be configured, i.e the ptr is not nil
	for r := 0; r < cfg.NumField(); r++ {
		resourceName := cfg.Type().Field(r).Name
		if cfg.Field(r).Pointer() != 0 {
			switch resourceName {
			case "User":
				//fmt.Printf("%s: %v : %s\n", resourceName, reflect.ValueOf(cfg.Field(r)), cfg.Field(r).Kind())
				//retrieve users resource values as an interface
				userAccounts := cfg.Field(r).Interface()

				//assert userAccounts interface to its actual type - A slice of ptrs to User
				for id, user := range userAccounts.([]*cfgresources.User) {
					userId := id + 1
					//setup params to post
					userParams := m.newUserCfg(user, userId)
					//post params
					err := m.applyUserCfg(userParams, userId)
					if err != nil {
						log.WithFields(log.Fields{
							"step":     "ApplyCfg",
							"Resource": cfg.Field(r).Kind(),
							"IP":       m.ip,
							"Serial":   m.serial,
						}).Warn("Unable to set user config.")
					}

				}

			case "Syslog":
				// interface of values of config struct field and type assert
				interfaceParams := m.newInterfaceCfg(cfg.Field(r).Interface().(*cfgresources.Syslog))
				err := m.applyInterfaceCfg(interfaceParams)
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "ApplyCfg",
						"resource": cfg.Field(r).Kind(),
						"IP":       m.ip,
						"Model":    m.BmcType(),
						"Serial":   m.serial,
					}).Warn("Unable to set Syslog config.")
				}
			case "Network":
			case "Ntp":
				datetimeParams := m.newDatetimeCfg(cfg.Field(r).Interface().(*cfgresources.Ntp))
				err = m.applyDatetimeCfg(datetimeParams)
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "ApplyCfg",
						"resource": cfg.Field(r).Kind(),
						"IP":       m.ip,
						"Model":    m.BmcType(),
						"Serial":   m.serial,
					}).Warn("Unable to set Ntp config.")
				}

			case "Ldap":
				//configure ldap service parameters
				directoryServicesParams := m.newDirectoryServicesCfg(cfg.Field(r).Interface().(*cfgresources.Ldap))
				err = m.applyDirectoryServicesCfg(directoryServicesParams)
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "ApplyCfg",
						"resource": cfg.Field(r).Kind(),
						"IP":       m.ip,
						"Model":    m.BmcType(),
						"Serial":   m.serial,
					}).Warn("Unable to set Ldap config.")
				}
			case "LdapGroup":
				ldapGroups := cfg.Field(r).Interface()
				err := m.applyLdapGroupParams(ldapGroups.([]*cfgresources.LdapGroup))
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "applyLdapParams",
						"resource": "Ldap",
						"IP":       m.ip,
						"Model":    m.BmcType(),
						"Serial":   m.serial,
						"Error":    err,
					}).Warn("applyLdapGroupParams returned error.")
				}
			case "Ssl":
				err := m.applySslCfg(cfg.Field(r).Interface().(*cfgresources.Ssl))
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "ApplyCfg",
						"resource": cfg.Field(r).Kind(),
						"IP":       m.ip,
						"Model":    m.BmcType(),
						"Serial":   m.serial,
					}).Warn("Unable to set SSL config.")
				}
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

func (m *M1000e) applyLdapGroupParams(cfg []*cfgresources.LdapGroup) (err error) {
	err = m.httpLogin()
	if err != nil {
		return err
	}

	roleId := 1
	for _, group := range cfg {
		ldapRoleParams, err := m.newLdapRoleCfg(group, roleId)
		if err != nil {
			log.WithFields(log.Fields{
				"step":      "applyLdapGroupParams",
				"Ldap role": group.Role,
				"IP":        m.ip,
				"Model":     m.BmcType(),
				"Serial":    m.serial,
				"Error":     err,
			}).Warn("Unable to apply Ldap role group config.")
			return err
		}

		err = m.applyLdapRoleCfg(ldapRoleParams, roleId)
		if err != nil {
			log.WithFields(log.Fields{
				"step":      "applyLdapGroupParams",
				"Ldap role": group.Role,
				"IP":        m.ip,
				"Model":     m.BmcType(),
				"Serial":    m.serial,
				"Error":     err,
			}).Warn("Unable to apply Ldap role group config.")
			return err
		}

		log.WithFields(log.Fields{
			"IP":     m.ip,
			"Model":  m.BmcType(),
			"Serial": m.serial,
			"Role":   group.Role,
			"Group":  group.Group,
		}).Debug("Ldap group parameters applied.")

		roleId += 1
	}

	return nil
}

//  /cgi-bin/webcgi/datetime
// apply datetime payload
func (m *M1000e) applyDatetimeCfg(cfg DatetimeParams) (err error) {
	err = m.httpLogin()
	if err != nil {
		return err
	}

	cfg.SessionToken = m.SessionToken
	path := fmt.Sprintf("datetime")
	form, _ := query.Values(cfg)
	err = m.post(path, &form)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"IP":     m.ip,
		"Model":  m.BmcType(),
		"Serial": m.serial,
	}).Debug("DateTime config parameters applied.")
	return err
}

//  /cgi-bin/webcgi/dirsvcs
// apply directoryservices payload
func (m *M1000e) applyDirectoryServicesCfg(cfg DirectoryServicesParams) (err error) {
	err = m.httpLogin()
	if err != nil {
		return err
	}

	cfg.SessionToken = m.SessionToken
	path := fmt.Sprintf("dirsvcs")
	form, _ := query.Values(cfg)
	err = m.post(path, &form)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"IP":     m.ip,
		"Model":  m.BmcType(),
		"Serial": m.serial,
	}).Debug("Ldap config parameters applied.")
	return err
}

// /cgi-bin/webcgi/ldaprg?index=1
// apply ldap role payload
func (m *M1000e) applyLdapRoleCfg(cfg LdapArgParams, roleId int) (err error) {
	err = m.httpLogin()
	if err != nil {
		return err
	}

	cfg.SessionToken = m.SessionToken
	path := fmt.Sprintf("ldaprg?index=%d", roleId)
	form, _ := query.Values(cfg)
	err = m.post(path, &form)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"IP":     m.ip,
		"Model":  m.BmcType(),
		"Serial": m.serial,
	}).Debug("Ldap Role group config parameters applied.")
	return err
}

// Configures various interface params - syslog, snmp etc.
func (m *M1000e) ApplySecurityCfg(cfg LoginSecurityParams) (err error) {
	err = m.httpLogin()
	if err != nil {
		return err
	}

	cfg.SessionToken = m.SessionToken
	form, _ := query.Values(cfg)
	err = m.post("loginSecurity", &form)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"IP":     m.ip,
		"Model":  m.BmcType(),
		"Serial": m.serial,
	}).Debug("Security config parameters applied.")
	return err

}

// Configures various interface params - syslog, snmp etc.
func (m *M1000e) applyInterfaceCfg(cfg InterfaceParams) (err error) {
	err = m.httpLogin()
	if err != nil {
		return err
	}

	cfg.SessionToken = m.SessionToken
	form, _ := query.Values(cfg)
	err = m.post("interfaces", &form)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"IP":     m.ip,
		"Model":  m.BmcType(),
		"Serial": m.serial,
	}).Debug("Interface config parameters applied.")
	return err
}

// call the cgi-bin/webcgi/user?id=<> endpoint
// with the user account payload
func (m *M1000e) applyUserCfg(cfg UserParams, userId int) (err error) {
	err = m.httpLogin()
	if err != nil {
		return err
	}

	cfg.SessionToken = m.SessionToken
	path := fmt.Sprintf("user?id=%d", userId)
	form, _ := query.Values(cfg)
	err = m.post(path, &form)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"IP":     m.ip,
		"Model":  m.BmcType(),
		"Serial": m.serial,
	}).Debug("User account config parameters applied.")
	return err
}

// call cgi-bin/webcgi/certuploadext
// with the ssl cert payload
func (m *M1000e) applySslCfg(ssl *cfgresources.Ssl) (err error) {
	err = m.httpLogin()
	if err != nil {
		return err
	}

	endpoint := fmt.Sprintf("certuploadext")

	formParams := make(map[string]string)
	formParams["ST2"] = m.SessionToken
	formParams["caller"] = ""
	formParams["pageCode"] = ""
	formParams["pageId"] = "2"
	formParams["pageName"] = ""

	err = m.NewSslMultipartUpload(endpoint, formParams, ssl.CertFile, ssl.KeyFile)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"IP":     m.ip,
		"Model":  m.BmcType(),
		"Serial": m.serial,
	}).Debug("SSL certs uploaded.")
	return err
}

// setup a multipart form with the expected order of form parameters
// for the payload format see  https://github.com/bmc-toolbox/bmclib/issues/3
func (m *M1000e) NewSslMultipartUpload(endpoint string, params map[string]string, SslCert string, SslKey string) (err error) {
	err = m.httpLogin()
	if err != nil {
		return err
	}

	file, err := os.Open(SslKey)
	if err != nil {
		log.WithFields(log.Fields{
			"step": "ssl-multipart-upload",
		}).Fatal("Declared SSL key file doesnt exist: ", SslKey)
		return err
	}
	defer file.Close()

	//given a map of key values, post the payload as a multipart form
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	//first we write the form params
	for key, val := range params {
		_ = writer.WriteField(key, val)
	}

	//create a form part with the ssl key
	keyPart, err := writer.CreateFormFile("file_key", filepath.Base(SslKey))
	if err != nil {
		return err
	}

	//copy the ssl key into the keypart
	_, err = io.Copy(keyPart, file)

	//write cert type into the form after the ssl key file
	_ = writer.WriteField("certType", "6")

	file, err = os.Open(SslCert)
	if err != nil {
		log.WithFields(log.Fields{
			"step": "ssl-multipart-upload",
		}).Fatal("Declared SSL cert file doesnt exist: ", SslCert)
		return err
	}
	defer file.Close()

	//create a form part with the ssl cert
	certPart, err := writer.CreateFormFile("file", filepath.Base(SslCert))
	if err != nil {
		return err
	}
	_, err = io.Copy(certPart, file)

	//write cert type into the form after the ssl key file
	_ = writer.WriteField("certType", "6")

	err = writer.Close()
	if err != nil {
		return err
	}

	url := fmt.Sprintf("https://%s/cgi-bin/webcgi/%s", m.ip, endpoint)
	req, err := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if log.GetLevel() == log.DebugLevel {
		dump, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			log.Println(fmt.Sprintf("[Request] https://%s/cgi-bin/webcgi/%s", m.ip, endpoint))
			log.Println(">>>>>>>>>>>>>>>")
			log.Printf("%s\n\n", dump)
			log.Println(">>>>>>>>>>>>>>>")
		}
	}

	resp, err := m.httpClient.Do(req)
	if err != nil {
		return err
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

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	//fmt.Printf("%s\n", body)
	return err
}

// posts a urlencoded form to the given endpoint
func (m *M1000e) post(endpoint string, form *url.Values) (err error) {
	err = m.httpLogin()
	if err != nil {
		return err
	}

	u, err := url.Parse(fmt.Sprintf("https://%s/cgi-bin/webcgi/%s", m.ip, endpoint))
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if log.GetLevel() == log.DebugLevel {
		dump, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			log.Println(fmt.Sprintf("[Request] https://%s/cgi-bin/webcgi/%s", m.ip, endpoint))
			log.Println(">>>>>>>>>>>>>>>")
			log.Printf("%s\n\n", dump)
			log.Println(">>>>>>>>>>>>>>>")
		}
	}

	//XXX to debug
	//fmt.Printf("--> %+v\n", form.Encode())
	//return err
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return err
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

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return err
}

//Implement a constructor to ensure required values are set
//func (m *M1000e) setSecurityCfg(cfg LoginSecurityParams) (cfg LoginSecurityParams, err error) {
//	return cfg, err
//}
