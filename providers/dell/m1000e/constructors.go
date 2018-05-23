package m1000e

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/ncode/bmclib/cfgresources"
)

func (m *M1000e) newSslCfg(ssl *cfgresources.Ssl) (MFormParams map[string]string) {

	//params for the multipart form.
	MformParams := make(map[string]string)

	MformParams["ST2"] = m.SessionToken
	MformParams["caller"] = ""
	MformParams["pageCode"] = ""
	MformParams["pageId"] = "2"
	MformParams["pageName"] = ""

	return MformParams
}

// Given the Ntp resource,
// populate the required Datetime params
func (m *M1000e) newDatetimeCfg(ntp *cfgresources.Ntp) DatetimeParams {

	if ntp.Timezone == "" {
		log.WithFields(log.Fields{
			"step": "apply-ntp-cfg",
		}).Fatal("Ntp resource parameter timezone required but not declared.")
	}

	if ntp.Server1 == "" {
		log.WithFields(log.Fields{
			"step": "apply-ntp-cfg",
		}).Fatal("Ntp resource parameter server1 required but not declared.")
	}

	dateTime := DatetimeParams{
		SessionToken:          m.SessionToken,
		NtpEnable:             ntp.Enable,
		NtpServer1:            ntp.Server1,
		NtpServer2:            ntp.Server2,
		NtpServer3:            ntp.Server3,
		DateTimeChanged:       true,
		CmcTimeTimezoneString: ntp.Timezone,
		TzChanged:             true,
	}

	return dateTime
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
		}).Warn("Ldap resource parameter Server required but not declared.")
	}

	if ldap.Port == 0 {
		log.WithFields(log.Fields{
			"step": "apply-ldap-cfg",
		}).Fatal("Ldap resource parameter Port required but not declared.")
	}
	if ldap.Group == "" {
		log.WithFields(log.Fields{
			"step": "apply-ldap-cfg",
		}).Fatal("Ldap resource parameter Group required but not declared.")
	}

	if ldap.GroupBaseDn == "" {
		log.WithFields(log.Fields{
			"step": "apply-ldap-cfg",
		}).Fatal("Ldap resource parameter GroupBaseDn required but not declared.")
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

	groupDn := fmt.Sprintf("cn=%s,%s", ldap.Group, ldap.GroupBaseDn)
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
		GenLdapBaseDn:                groupDn,
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

// Return bool value if the role is valid.
func (m *M1000e) isRoleValid(role string) bool {

	validRoles := []string{"admin", "user"}
	for _, v := range validRoles {
		if role == v {
			return true
		}
	}

	return false
}

// TODO: the code should not Fatal, but return so configuration continues.
// Given the Ldap resource, populate required LdapArgParams
func (m *M1000e) newLdapRoleCfg(ldap *cfgresources.Ldap) LdapArgParams {

	// this needs to be updated to support various roles.
	roleId := 1

	var privBitmap, genLdapRolePrivilege int

	if !m.isRoleValid(ldap.Role) {
		log.WithFields(log.Fields{
			"step": "apply-ldap-cfg",
			"role": ldap.Role,
		}).Fatal("Ldap resource Role must be a valid role: admin OR user.")
	}

	if ldap.Group == "" {
		log.WithFields(log.Fields{
			"step": "apply-ldap-cfg",
		}).Fatal("Ldap resource Group required but not declared.")
	}

	if ldap.GroupBaseDn == "" {
		log.WithFields(log.Fields{
			"step": "apply-ldap-cfg",
		}).Fatal("Ldap resource GroupBaseDn required but not declared.")
	}

	groupDn := fmt.Sprintf("cn=%s,%s", ldap.Group, ldap.GroupBaseDn)

	//TODO
	//this needs more work, the resource declaration needs to support a list of roles,
	//and the appropriate permissions need to be set below.
	switch ldap.Role {
	case "admin":
		privBitmap = 4095
		genLdapRolePrivilege = privBitmap
	case "user":
		privBitmap = 1
		genLdapRolePrivilege = privBitmap
	}

	ldapArgCfg := LdapArgParams{
		SessionToken:         m.SessionToken,
		PrivBitmap:           privBitmap,
		Index:                roleId,
		GenLdapRoleDn:        groupDn,
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
