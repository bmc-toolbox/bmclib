package m1000e

import (
	"errors"
	"fmt"

	"github.com/bmc-toolbox/bmclib/cfgresources"
)

//func (m *M1000e) newSslCfg(ssl *cfgresources.Ssl) (MFormParams map[string]string) {
//
//	//params for the multipart form.
//	MformParams := make(map[string]string)
//
//	MformParams["ST2"] = m.SessionToken
//	MformParams["caller"] = ""
//	MformParams["pageCode"] = ""
//	MformParams["pageId"] = "2"
//	MformParams["pageName"] = ""
//
//	return MformParams
//}

// Given the Ntp resource,
// populate the required Datetime params
func (m *M1000e) newDatetimeCfg(ntp *cfgresources.Ntp) DatetimeParams {
	if ntp.Timezone == "" {
		// TODO update method with error return and return err in this if, was doing logrus.Fatal here
		msg := "ntp resource parameter timezone required but not declared"
		err := errors.New(msg)
		m.log.V(0).Error(err, msg, "step", "apply-ntp-cfg")
	}

	if ntp.Server1 == "" {
		// TODO update method with error return and return err in this if, was doing logrus.Fatal here
		msg := "ntp resource parameter server1 required but not declared."
		err := errors.New(msg)
		m.log.V(0).Error(err, msg, "step", "apply-ntp-cfg")
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
		m.log.V(1).Info("Ldap resource parameter Server required but not declared.", "step", "newDirectoryServicesCfg")
	}

	if ldap.Port == 0 {
		m.log.V(1).Info("Ldap resource parameter Port required but not declared.", "step", "newDirectoryServicesCfg")
	}

	if ldap.BaseDn == "" {
		m.log.V(1).Info("Ldap resource parameter baseDn required but not declared.", "step", "newDirectoryServicesCfg")
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
		GenLdapBindDn:                ldap.BindDn,
		GenLdapBindPasswd:            "PASSWORD",
		GenLdapBindPasswdChanged:     false,
		GenLdapBaseDn:                ldap.BaseDn,
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

// Given the Ldap resource, populate required LdapArgParams
func (m *M1000e) newLdapRoleCfg(cfgGroups *cfgresources.LdapGroup, roleID int) (ldapArgCfg LdapArgParams, err error) {
	var privBitmap, genLdapRolePrivilege int

	if cfgGroups.Group == "" {
		msg := "Ldap resource parameter Group required but not declared."
		err = errors.New(msg)
		m.log.V(1).Error(err, msg, "Role", cfgGroups.Role, "step", "newLdapRoleCfg")
		return ldapArgCfg, err
	}

	if cfgGroups.GroupBaseDn == "" && cfgGroups.Enable {
		msg := "Ldap resource parameter GroupBaseDn required but not declared."
		err = errors.New(msg)
		m.log.V(1).Error(err, msg, "Role", cfgGroups.Role, "step", "newLdapRoleCfg")
		return ldapArgCfg, err
	}

	if !m.isRoleValid(cfgGroups.Role) {
		msg := "Ldap resource Role must be a valid role: admin OR user."
		err = errors.New(msg)
		m.log.V(1).Error(err, msg, "Role", cfgGroups.Role, "step", "newLdapRoleCfg")
		return ldapArgCfg, err
	}

	groupDn := fmt.Sprintf("cn=%s,%s", cfgGroups.Group, cfgGroups.GroupBaseDn)

	switch cfgGroups.Role {
	case "admin":
		privBitmap = 4095
		genLdapRolePrivilege = privBitmap
	case "user":
		privBitmap = 1
		genLdapRolePrivilege = privBitmap
	}

	ldapArgCfg = LdapArgParams{
		SessionToken:         m.SessionToken,
		PrivBitmap:           privBitmap,
		Index:                roleID,
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

	return ldapArgCfg, err
}

// Given the syslog resource, populate the required InterfaceParams
// check for missing params
func (m *M1000e) newInterfaceCfg(syslog *cfgresources.Syslog) InterfaceParams {
	var syslogPort int

	if syslog.Server == "" {
		// TODO update method with error return and return err in this if, was doing logrus.Fatal here
		msg := "syslog resource expects parameter: Server"
		err := errors.New(msg)
		m.log.V(0).Error(err, msg, "step", "apply-interface-cfg")
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
		WebserverHTTPPort:                80,
		WebserverHTTPSPort:               443,
		SSHEnable:                        true,
		SSHMaxSessions:                   4,
		SSHTimeout:                       1800,
		SSHPort:                          22,
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
func (m *M1000e) newUserCfg(user *cfgresources.User, userID int) UserParams {
	var cmcGroup, privilege int

	if user.Name == "" {
		// TODO update method with error return and return err in this if, was doing logrus.Fatal here
		msg := "user resource expects parameter: Name"
		err := errors.New(msg)
		m.log.V(0).Error(err, msg, "step", "apply-user-cfg")
	}

	if user.Password == "" {
		// TODO update method with error return and return err in this if, was doing logrus.Fatal here
		msg := "user resource expects parameter: Password"
		err := errors.New(msg)
		m.log.V(0).Error(err, msg, "step", "apply-user-cfg")
	}

	if !m.isRoleValid(user.Role) {
		// TODO update method with error return and return err in this if, was doing logrus.Fatal here
		msg := "user resource Role must be declared and a valid role: admin"
		err := errors.New(msg)
		m.log.V(0).Error(err, msg, "step", "apply-user-cfg", "role", user.Role)
	}

	if user.Role == "admin" {
		cmcGroup = 4095
		privilege = cmcGroup
	}

	userCfg := UserParams{
		SessionToken:    m.SessionToken,
		Privilege:       privilege,
		UserID:          userID,
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
