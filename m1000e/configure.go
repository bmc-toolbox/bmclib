package m1000e

import (
	"fmt"
	"github.com/google/go-querystring/query"
	log "github.com/sirupsen/logrus"
	"github.com/ncode/bmc/cfgresources"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
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
					}).Warn("Unable to set Syslog config.")
				}
			case "Network":
				fmt.Printf("%s: %v : %s\n", resourceName, cfg.Field(r), cfg.Field(r).Kind())
			case "Ntp":
				fmt.Printf("%s: %v : %s\n", resourceName, cfg.Field(r), cfg.Field(r).Kind())
			case "Ldap":
				//configure ldap service parameters
				directoryServicesParams := m.newDirectoryServicesCfg(cfg.Field(r).Interface().(*cfgresources.Ldap))
				err = m.applyDirectoryServicesCfg(directoryServicesParams)
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "ApplyCfg",
						"resource": cfg.Field(r).Kind(),
						"IP":       m.ip,
					}).Warn("Unable to set Ldap services config.")
				}

				//configure ldap role groups
				ldapRoleParams := m.newLdapRoleCfg(cfg.Field(r).Interface().(*cfgresources.Ldap))
				err := m.applyLdapRoleCfg(ldapRoleParams, 1)
				if err != nil {
					log.WithFields(log.Fields{
						"step":     "ApplyCfg",
						"resource": cfg.Field(r).Kind(),
						"IP":       m.ip,
					}).Warn("Unable to set Ldap role group config.")
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

// TODO:
// support Certificate Validation Enabled
// A multipart form would be required to upload the cacert
// Given the Ldap resource, populate required DirectoryServicesParams
func (m *M1000e) newDirectoryServicesCfg(ldap *cfgresources.Ldap) DirectoryServicesParams {

	var userAttribute, groupAttribute string
	if ldap.Server == "" {
		log.WithFields(log.Fields{
			"step": "apply-ldap-cfg",
		}).Fatal("Ldap resource parameter Server required but not declared.")
	}

	if ldap.Port == 0 {
		log.WithFields(log.Fields{
			"step": "apply-ldap-cfg",
		}).Fatal("Ldap resource parameter Port required but not declared.")
	}

	if ldap.GroupDn == "" {
		log.WithFields(log.Fields{
			"step": "apply-ldap-cfg",
		}).Fatal("Ldap resource parameter baseDn required but not declared.")
	}

	if ldap.UserAttribute == "" {
		userAttribute = "uid"
	} else {
		userAttribute = ldap.UserAttribute
	}

	if ldap.GroupAttribute == "" {
		groupAttribute = "memberUid"
	} else {
		groupAttribute = ldap.GroupAttribute
	}

	directoryServicesParams := DirectoryServicesParams{
		SessionToken:                 m.SessionToken,
		SeviceSelected:               "ldap",
		CertType:                     5,
		Action:                       1,
		Choose:                       2,
		GenLdapEnableCk:              true,
		GenLdapEnable:                true,
		GenLdapGroupAttributeIsDnCk:  true,
		GenLdapGroupAttributeIsDn:    true,
		GenLdapCertValidateEnableCk:  true,
		GenLdapCertValidateEnable:    false,
		GenLdapBindDn:                "",
		GenLdapBindPasswd:            "PASSWORD", //we
		GenLdapBindPasswdChanged:     false,
		GenLdapBaseDn:                ldap.GroupDn,
		GenLdapUserAttribute:         userAttribute,
		GenLdapGroupAttribute:        groupAttribute,
		GenLdapSearchFilter:          ldap.SearchFilter,
		GenLdapConnectTimeoutSeconds: 30,
		GenLdapSearchTimeoutSeconds:  120,
		LdapServers:                  1,
		GenLdapServerAddr:            ldap.Server,
		GenLdapServerPort:            ldap.Port,
		GenLdapSrvLookupEnable:       false,
		AdEnable:                     false,
		AdTfaSsoEnableBitmask1:       0,
		AdTfaSsoEnableBitmask2:       0,
		AdCertValidateEnableCk:       false,
		AdCertValidateEnable:         false,
		AdRootDomain:                 "",
		AdTimeout:                    120,
		AdFilterEnable:               false,
		AdDcFilter:                   "",
		AdGcFilter:                   "",
		AdSchemaExt:                  1,
		RoleGroupFlag:                0,
		RoleGroupIndex:               "",
		AdCmcName:                    "",
		AdCmcdomain:                  "",
	}

	return directoryServicesParams
}

// Given the Ldap resource, populate required LdapArgParams
func (m *M1000e) newLdapRoleCfg(ldap *cfgresources.Ldap) LdapArgParams {

	// as of now we care to only set the admin role.
	// this needs to be updated to support various roles.

	roleId := 1

	validRole := "admin"
	var privBitmap, genLdapRolePrivilege int

	if ldap.Role != validRole {
		log.WithFields(log.Fields{
			"step": "apply-ldap-cfg",
		}).Fatal("Ldap resource Role must be declared and a valid role: admin.")
	}

	if ldap.GroupDn == "" {
		log.WithFields(log.Fields{
			"step": "apply-ldap-cfg",
		}).Fatal("Ldap resource GroupDn must be declared.")
	}

	if ldap.Role == "admin" {
		privBitmap = 4095
		genLdapRolePrivilege = privBitmap
	}

	ldapArgCfg := LdapArgParams{
		SessionToken:         m.SessionToken,
		PrivBitmap:           privBitmap,
		Index:                roleId,
		GenLdapRoleDn:        ldap.GroupDn,
		GenLdapRolePrivilege: genLdapRolePrivilege,
		Login:                true,
		Cfg:                  true,
		Cfguser:              true,
		Clearlog:             true,
		Chassiscontrol:       true,
		Superuser:            true,
		Serveradmin:          true,
		Testalert:            true,
		Debug:                true,
		Afabricadmin:         true,
		Bfabricadmin:         true,
	}

	return ldapArgCfg

}

// Given the syslog resource, populate the required InterfaceParams
// check for missing params
func (m *M1000e) newInterfaceCfg(syslog *cfgresources.Syslog) InterfaceParams {

	var syslogPort int

	if syslog.Server == "" {
		log.WithFields(log.Fields{
			"step": "apply-interface-cfg",
		}).Fatal("Syslog resource expects parameter: Server.")
	}

	if syslog.Port == 0 {
		syslogPort = syslog.Port
	} else {
		syslogPort = 514
	}

	interfaceCfg := InterfaceParams{
		SessionToken:                     m.SessionToken,
		SerialEnable:                     true,
		SerialRedirect:                   true,
		SerialTimeout:                    1800,
		SerialBaudrate:                   115200,
		SerialQuitKey:                    "^\\",
		SerialHistoryBufSize:             8192,
		SerialLoginCommand:               "",
		WebserverEnable:                  true,
		WebserverMaxSessions:             4,
		WebserverTimeout:                 1800,
		WebserverHttpPort:                80,
		WebserverHttpsPort:               443,
		SshEnable:                        true,
		SshMaxSessions:                   4,
		SshTimeout:                       1800,
		SshPort:                          22,
		TelnetEnable:                     true,
		TelnetMaxSessions:                4,
		TelnetTimeout:                    1800,
		TelnetPort:                       23,
		RacadmEnable:                     true,
		RacadmMaxSessions:                4,
		RacadmTimeout:                    60,
		SnmpEnable:                       true,
		SnmpCommunityNameGet:             "public",
		SnmpProtocol:                     0,
		SnmpDiscoveryPortSet:             161,
		ChassisLoggingRemoteSyslogEnable: syslog.Enable,
		ChassisLoggingRemoteSyslogHost1:  syslog.Server,
		ChassisLoggingRemoteSyslogHost2:  "",
		ChassisLoggingRemoteSyslogHost3:  "",
		ChassisLoggingRemoteSyslogPort:   syslogPort,
	}

	return interfaceCfg
}

// Given the user resource, populate the required UserParams
// check for missing params
func (m *M1000e) newUserCfg(user *cfgresources.User, userId int) UserParams {

	// as of now we care to only set the admin role.
	// this needs to be updated to support various roles.
	validRole := "admin"
	var cmcGroup, privilege int

	if user.Name == "" {
		log.WithFields(log.Fields{
			"step": "apply-user-cfg",
		}).Fatal("User resource expects parameter: Name.")
	}

	if user.Password == "" {
		log.WithFields(log.Fields{
			"step": "apply-user-cfg",
		}).Fatal("User resource expects parameter: Password.")
	}

	if user.Role != validRole {
		log.WithFields(log.Fields{
			"step": "apply-user-cfg",
		}).Fatal("User resource Role must be declared and a valid role: admin.")
	}

	if user.Role == "admin" {
		cmcGroup = 4095
		privilege = cmcGroup
	}

	userCfg := UserParams{
		SessionToken:    m.SessionToken,
		Privilege:       privilege,
		UserId:          userId,
		EnableUser:      user.Enable,
		UserName:        user.Name,
		ChangePassword:  true,
		Password:        user.Password,
		ConfirmPassword: user.Password,
		CmcGroup:        cmcGroup,
		Login:           true,
		Cfg:             true,
		CfgUser:         true,
		ClearLog:        true,
		ChassisControl:  true,
		SuperUser:       true,
		ServerAdmin:     true,
		TestAlert:       true,
		Debug:           true,
		AFabricAdmin:    true,
		BFabricAdmin:    true,
		CFabricAcminc:   true,
	}

	return userCfg
}

//  /cgi-bin/webcgi/dirsvcs
// apply directoryservices payload
func (m *M1000e) applyDirectoryServicesCfg(cfg DirectoryServicesParams) (err error) {

	cfg.SessionToken = m.SessionToken
	path := fmt.Sprintf("dirsvcs")
	form, _ := query.Values(cfg)
	err = m.post(path, &form)
	if err != nil {
		return err
	}

	return err
}

// /cgi-bin/webcgi/ldaprg?index=1
// apply ldap role payload
func (m *M1000e) applyLdapRoleCfg(cfg LdapArgParams, roleId int) (err error) {

	cfg.SessionToken = m.SessionToken
	path := fmt.Sprintf("ldaprg?index=%d", roleId)
	form, _ := query.Values(cfg)
	err = m.post(path, &form)
	if err != nil {
		return err
	}

	return err
}

// Configures various interface params - syslog, snmp etc.
func (m *M1000e) ApplySecurityCfg(cfg LoginSecurityParams) (err error) {

	cfg.SessionToken = m.SessionToken
	form, _ := query.Values(cfg)
	err = m.post("loginSecurity", &form)
	if err != nil {
		return err
	}

	return err

}
func (m *M1000e) applyInterfaceCfg(cfg InterfaceParams) (err error) {

	cfg.SessionToken = m.SessionToken
	form, _ := query.Values(cfg)
	err = m.post("interfaces", &form)
	if err != nil {
		return err
	}

	return err
}

// call the cgi-bin/webcgi/user?id=<> endpoint
// with the user account payload
func (m *M1000e) applyUserCfg(cfg UserParams, userId int) (err error) {

	cfg.SessionToken = m.SessionToken
	path := fmt.Sprintf("user?id=%d", userId)
	form, _ := query.Values(cfg)
	err = m.post(path, &form)
	if err != nil {
		return err
	}

	return err
}

// posts a urlencoded form to the given endpoint
func (m *M1000e) post(endpoint string, form *url.Values) (err error) {

	u, err := url.Parse(fmt.Sprintf("https://%s/cgi-bin/webcgi/%s", m.ip, endpoint))
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	//XXX to debug
	//fmt.Printf("--> %+v\n", form.Encode())
	//return err
	resp, err := m.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	//fmt.Printf("-->> %d\n", resp.StatusCode)
	//fmt.Printf("%s\n", body)
	return err
}

//Implement a constructor to ensure required values are set
//func (m *M1000e) setSecurityCfg(cfg LoginSecurityParams) (cfg LoginSecurityParams, err error) {
//	return cfg, err
//}
