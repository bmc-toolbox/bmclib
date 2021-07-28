package m1000e

import (
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/bmc-toolbox/bmclib/cfgresources"
	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/internal"
	"github.com/google/go-querystring/query"
)

// This ensures the compiler errors if this type is missing
// a method that should be implmented to satisfy the Configure interface.
var _ devices.Configure = (*M1000e)(nil)

// Resources returns a slice of supported resources and
// the order they are to be applied in.
// Resources implements the Configure interface
func (m *M1000e) Resources() []string {
	return []string{
		"user",
		"syslog",
		"ntp",
		"ldap",
		"ldap_group",
		//"ssl",
	}
}

// ResourcesSetup returns
// - slice of supported one time setup resources,
//   in the order they must be applied
// ResourcesSetup implements the CmcSetup interface
// see cfgresources.SetupChassis for list of setup resources.
func (m *M1000e) ResourcesSetup() []string {
	return []string{
		"setipmioverlan",
		"flexaddress",
		"dynamicpower",
		"bladespower",
	}
}

// ApplyCfg implements the Bmc interface
// this is to be deprecated.
func (m *M1000e) ApplyCfg(config *cfgresources.ResourcesConfig) (err error) {
	return err
}

// Power implemented the Configure interface
func (m *M1000e) Power(cfg *cfgresources.Power) (err error) {
	return err
}

// SetLicense implements the Configure interface.
func (m *M1000e) SetLicense(cfg *cfgresources.License) (err error) {
	return err
}

// Bios implements the Configure interface.
func (m *M1000e) Bios(cfg *cfgresources.Bios) (err error) {
	return err
}

// Network method implements the Configure interface
// applies various network parameters.
func (m *M1000e) Network(cfg *cfgresources.Network) (reset bool, err error) {
	return reset, err
}

// User applies the User configuration resource,
// if the user exists, it updates the users password,
// User implements the Configure interface.
// Iterate over iDrac users and adds/removes/modifies user accounts
func (m *M1000e) User(cfgUsers []*cfgresources.User) (err error) {
	err = m.httpLogin()
	if err != nil {
		return err
	}

	for id, cfgUser := range cfgUsers {

		userID := id + 1
		//setup params to post
		userParams := m.newUserCfg(cfgUser, userID)

		userParams.SessionToken = m.SessionToken
		path := fmt.Sprintf("user?id=%d", userID)
		form, _ := query.Values(userParams)
		err = m.post(path, &form)
		if err != nil {
			return err
		}

		m.log.V(1).Info("User account config parameters applied.", "IP", m.ip, "Model", m.HardwareType())

	}

	return err
}

// Syslog applies the Syslog configuration resource
// Syslog implements the Configure interface
// TODO: this currently applies network config as well,
//       figure a way to split the two.
func (m *M1000e) Syslog(cfg *cfgresources.Syslog) (err error) {

	interfaceParams := m.newInterfaceCfg(cfg)

	interfaceParams.SessionToken = m.SessionToken
	form, _ := query.Values(interfaceParams)
	err = m.post("interfaces", &form)
	if err != nil {
		return err
	}

	m.log.V(1).Info("Interface config parameters applied.", "IP", m.ip, "Model", m.HardwareType())
	return err
}

// Ntp applies NTP configuration params
// Ntp implements the Configure interface.
func (m *M1000e) Ntp(cfg *cfgresources.Ntp) (err error) {

	err = m.httpLogin()
	if err != nil {
		return err
	}

	datetimeParams := m.newDatetimeCfg(cfg)

	datetimeParams.SessionToken = m.SessionToken
	path := "datetime"
	form, _ := query.Values(datetimeParams)
	err = m.post(path, &form)
	if err != nil {
		return err
	}

	m.log.V(1).Info("DateTime config parameters applied.", "IP", m.ip, "Model", m.HardwareType())
	return err
}

// Ldap applies LDAP configuration params.
// Ldap implements the Configure interface.
func (m *M1000e) Ldap(cfg *cfgresources.Ldap) (err error) {

	directoryServicesParams := m.newDirectoryServicesCfg(cfg)

	directoryServicesParams.SessionToken = m.SessionToken
	path := "dirsvcs"
	form, _ := query.Values(directoryServicesParams)
	err = m.post(path, &form)
	if err != nil {
		return err
	}

	m.log.V(1).Info("Ldap config parameters applied.", "IP", m.ip, "Model", m.HardwareType())
	return err
}

// /cgi-bin/webcgi/ldaprg?index=1
// apply ldap role payload
func (m *M1000e) applyLdapRoleCfg(cfg LdapArgParams, roleID int) (err error) {
	err = m.httpLogin()
	if err != nil {
		return err
	}

	cfg.SessionToken = m.SessionToken
	path := fmt.Sprintf("ldaprg?index=%d", roleID)
	form, _ := query.Values(cfg)
	err = m.post(path, &form)
	if err != nil {
		return err
	}

	m.log.V(1).Info("Ldap Role group config parameters applied.", "IP", m.ip, "Model", m.HardwareType())
	return err
}

// LdapGroup applies LDAP Group/Role related configuration
// LdapGroup implements the Configure interface.
func (m *M1000e) LdapGroup(cfg []*cfgresources.LdapGroup, cfgLdap *cfgresources.Ldap) (err error) {

	roleID := 1
	for _, group := range cfgGroups {
		ldapRoleParams, err := m.newLdapRoleCfg(group, roleID)
		if err != nil {
			m.log.V(1).Error(err, "Unable to apply Ldap role group config.",
				"step", "applyLdapGroupParams",
				"Ldap role", group.Role,
				"IP", m.ip,
				"Model", m.HardwareType(),
				"Error", internal.ErrStringOrEmpty(err),
			)
			return err
		}

		err = m.applyLdapRoleCfg(ldapRoleParams, roleID)
		if err != nil {
			m.log.V(1).Error(err, "Unable to apply Ldap role group config.",
				"step", "applyLdapGroupParams",
				"Ldap role", group.Role,
				"IP", m.ip,
				"Model", m.HardwareType(),
				"Error", err,
			)
			return err
		}

		m.log.V(1).Info("Ldap group parameters applied.",
			"IP", m.ip,
			"Model", m.HardwareType(),
			"Role", group.Role,
			"Group", group.Group,
		)

		roleID++
	}

	return nil
}

// GenerateCSR generates a CSR request on the BMC.
// GenerateCSR implements the Configure interface.
func (m *M1000e) GenerateCSR(cert *cfgresources.HTTPSCertAttributes) ([]byte, error) {
	return []byte{}, nil
}

// UploadHTTPSCert uploads the given CRT cert,
// UploadHTTPSCert implements the Configure interface.
func (m *M1000e) UploadHTTPSCert(cert []byte, certFileName string, key []byte, keyFileName string) (bool, error) {
	return false, nil
}

// CurrentHTTPSCert returns the current x509 certficates configured on the BMC
// The bool value returned indicates if the BMC supports CSR generation.
// CurrentHTTPSCert implements the Configure interface.
func (m *M1000e) CurrentHTTPSCert() (c []*x509.Certificate, b bool, e error) {
	return c, b, e
}

// Ssl applies the SSL configuration
// TODO: add to the configure interface.
// call cgi-bin/webcgi/certuploadext
// with the ssl cert payload
//func (m *M1000e) Ssl(ssl *cfgresources.Ssl) (err error) {
//	err = m.httpLogin()
//	if err != nil {
//		return err
//	}
//
//	endpoint := fmt.Sprintf("certuploadext")
//
//	formParams := make(map[string]string)
//	formParams["ST2"] = m.SessionToken
//	formParams["caller"] = ""
//	formParams["pageCode"] = ""
//	formParams["pageId"] = "2"
//	formParams["pageName"] = ""
//
//	err = m.newSslMultipartUpload(endpoint, formParams, ssl.CertFile, ssl.KeyFile)
//	if err != nil {
//		return err
//	}
//
//	m.log.V(1).Info("",
//		"IP":    m.ip,
//		"Model": m.HardwareType(),
//	}).Debug("SSL certs uploaded.")
//	return err
//}
//
//// setup a multipart form with the expected order of form parameters
//// for the payload format see  https://github.com/bmc-toolbox/bmclib/issues/3
//func (m *M1000e) newSslMultipartUpload(endpoint string, params map[string]string, SslCert string, SslKey string) (err error) {
//	err = m.httpLogin()
//	if err != nil {
//		return err
//	}
//
//	file, err := os.Open(SslKey)
//	if err != nil {
//		m.log.V(1).Info("",
//			"step": "ssl-multipart-upload",
//		}).Fatal("Declared SSL key file doesnt exist: ", SslKey)
//		return err
//	}
//	defer file.Close()
//
//	//given a map of key values, post the payload as a multipart form
//	body := &bytes.Buffer{}
//	writer := multipart.NewWriter(body)
//
//	//first we write the form params
//	for key, val := range params {
//		_ = writer.WriteField(key, val)
//	}
//
//	//create a form part with the ssl key
//	keyPart, err := writer.CreateFormFile("file_key", filepath.Base(SslKey))
//	if err != nil {
//		return err
//	}
//
//	//copy the ssl key into the keypart
//	_, err = io.Copy(keyPart, file)
//
//	//write cert type into the form after the ssl key file
//	_ = writer.WriteField("certType", "6")
//
//	file, err = os.Open(SslCert)
//	if err != nil {
//		m.log.V(1).Info("",
//			"step": "ssl-multipart-upload",
//		}).Fatal("Declared SSL cert file doesnt exist: ", SslCert)
//		return err
//	}
//	defer file.Close()
//
//	//create a form part with the ssl cert
//	certPart, err := writer.CreateFormFile("file", filepath.Base(SslCert))
//	if err != nil {
//		return err
//	}
//	_, err = io.Copy(certPart, file)
//
//	//write cert type into the form after the ssl key file
//	_ = writer.WriteField("certType", "6")
//
//	err = writer.Close()
//	if err != nil {
//		return err
//	}
//
//	url := fmt.Sprintf("https://%s/cgi-bin/webcgi/%s", m.ip, endpoint)
//	req, err := http.NewRequest("POST", url, body)
//	req.Header.Set("Content-Type", writer.FormDataContentType())
//	if log.GetLevel() == log.TraceLevel {
//		dump, err := httputil.DumpRequestOut(req, true)
//		if err == nil {
//			log.Println(fmt.Sprintf("[Request] https://%s/cgi-bin/webcgi/%s", m.ip, endpoint))
//			log.Println(">>>>>>>>>>>>>>>")
//			log.Printf("%s\n\n", dump)
//			log.Println(">>>>>>>>>>>>>>>")
//		}
//	}
//
//	resp, err := m.httpClient.Do(req)
//	if err != nil {
//		return err
//	}
//	defer resp.Body.Close()
//	if log.GetLevel() == log.TraceLevel {
//		dump, err := httputil.DumpResponse(resp, true)
//		if err == nil {
//			log.Println("[Response]")
//			log.Println("<<<<<<<<<<<<<<")
//			log.Printf("%s\n\n", dump)
//			log.Println("<<<<<<<<<<<<<<")
//		}
//	}
//
//	_, err = ioutil.ReadAll(resp.Body)
//	if err != nil {
//		return err
//	}
//
//	//fmt.Printf("%s\n", body)
//	return err
//}

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

	reqDump, _ := httputil.DumpRequestOut(req, true)
	m.log.V(2).Info("requestTrace", "requestDump", string(reqDump), "url", fmt.Sprintf("https://%s/cgi-bin/webcgi/%s", m.ip, endpoint))

	//XXX to debug
	//fmt.Printf("--> %+v\n", form.Encode())
	//return err
	resp, err := m.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respDump, _ := httputil.DumpResponse(resp, true)
	m.log.V(2).Info("responseTrace", "responseDump", string(respDump))

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	return err
}

// ApplySecurityCfg configures various interface params - syslog, snmp etc.
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

	m.log.V(1).Info("Security config parameters applied.", "IP", m.ip, "Model", m.HardwareType())
	return err

}
