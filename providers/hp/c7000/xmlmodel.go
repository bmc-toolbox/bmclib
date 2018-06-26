package c7000

import (
	"encoding/xml"
)

type Username struct {
	XMLName xml.Name `xml:"hpoa:username"`
	Text    string   `xml:",chardata"`
}

type Password struct {
	XMLName xml.Name `xml:"hpoa:password"`
	Text    string   `xml:",chardata"`
}

type UserLogIn struct {
	XMLName  xml.Name `xml:"hpoa:userLogIn"`
	Text     string   `xml:",chardata"`
	Username Username
	Password Password
}

type Body struct {
	XMLName xml.Name    `xml:"SOAP-ENV:Body"`
	Text    string      `xml:",chardata"`
	Content interface{} `xml:",any"`
}

type EnvelopeLoginResponse struct {
	XMLName xml.Name `xml:"Envelope"`
	Text    string   `xml:",chardata"`
	SOAPENV string   `xml:"SOAP-ENV,attr"`
	SOAPENC string   `xml:"SOAP-ENC,attr"`
	Xsi     string   `xml:"xsi,attr"`
	Xsd     string   `xml:"xsd,attr"`
	Wsu     string   `xml:"wsu,attr"`
	Wsse    string   `xml:"wsse,attr"`
	Hpoa    string   `xml:"hpoa,attr"`
	Body    struct {
		UserLogInResponse struct {
			HpOaSessionKeyToken struct {
				OaSessionKey struct {
					Text string `xml:",chardata"`
				} `xml:"oaSessionKey"`
			} `xml:"HpOaSessionKeyToken"`
		} `xml:"userLogInResponse"`
	} `xml:"Body"`
}

type OaSessionKey struct {
	XMLName xml.Name `xml:"hpoa:oaSessionKey"`
	Text    string   `xml:",chardata"`
}

type HpOaSessionKeyToken struct {
	XMLName      xml.Name `xml:"hpoa:HpOaSessionKeyToken"`
	OaSessionKey OaSessionKey
}

type Security struct {
	XMLName             xml.Name `xml:"wsse:Security"`
	MustUnderstand      string   `xml:"SOAP-ENV:mustUnderstand,attr"`
	HpOaSessionKeyToken HpOaSessionKeyToken
}

type Header struct {
	XMLName  xml.Name `xml:"SOAP-ENV:Header,omitempty"`
	Security Security
}

type Server struct {
	XMLName xml.Name `xml:"hpoa:server"`
	Text    string   `xml:",chardata"`
}

type SetRemoteSyslogServer struct {
	XMLName xml.Name `xml:"hpoa:setRemoteSyslogServer"`
	Server  string   `xml:"hpoa:server"`
}

type SetRemoteSyslogPort struct {
	XMLName xml.Name `xml:"hpoa:setRemoteSyslogPort"`
	Port    int      `xml:"hpoa:port"`
}

type SetRemoteSyslogEnabled struct {
	XMLName xml.Name `xml:"hpoa:setRemoteSyslogEnabled"`
	Enabled bool     `xml:"hpoa:enabled"`
}

type Envelope struct {
	XMLName xml.Name `xml:"SOAP-ENV:Envelope"`
	Text    string   `xml:",chardata"`
	SOAPENV string   `xml:"xmlns:SOAP-ENV,attr"`
	Xsi     string   `xml:"xmlns:xsi,attr"`
	Xsd     string   `xml:"xmlns:xsd,attr"`
	Wsu     string   `xml:"xmlns:wsu,attr"`
	Wsse    string   `xml:"xmlns:wsse,attr"`
	Hpoa    string   `xml:"xmlns:hpoa,attr"`
	Header  Header
	Body    Body
}

// <hpoa:addUser>
//  <hpoa:username>Test</hpoa:username>
//   <hpoa:password>foobar</hpoa:password>
// </hpoa:addUser>
type AddUser struct {
	XMLName  xml.Name `xml:"hpoa:addUser"`
	Username Username
	Password Password
}

// <hpoa:setUserPassword>
//  <hpoa:username>Administrator</hpoa:username>
//  <hpoa:password>foobar</hpoa:password>
// </hpoa:setUserPassword>
type SetUserPassword struct {
	XMLName  xml.Name `xml:"hpoa:setUserPassword"`
	Username Username
	Password Password
}

// Ntp payload - minus the body, envelope
type configureNtp struct {
	XMLName      xml.Name `xml:"hpoa:configureNtp"`
	NtpPrimary   NtpPrimary
	NtpSecondary NtpSecondary
	NtpPoll      NtpPoll
}

type NtpPrimary struct {
	XMLName xml.Name `xml:"hpoa:ntpPrimary"`
	Text    string   `xml:",chardata"`
}

type NtpSecondary struct {
	XMLName xml.Name `xml:"hpoa:ntpSecondary"`
	Text    string   `xml:",chardata"`
}

type NtpPoll struct {
	XMLName xml.Name `xml:"hpoa:ntpPoll"`
	Text    string   `xml:",chardata"`
}

type setEnclosureTimeZone struct {
	XMLName  xml.Name `xml:"hpoa:setEnclosureTimeZone"`
	Timezone timeZone
}

type timeZone struct {
	XMLName xml.Name `xml:"hpoa:timeZone"`
	Text    string   `xml:",chardata"`
}

type addLdapGroup struct {
	XMLName   xml.Name `xml:"hpoa:addLdapGroup"`
	LdapGroup ldapGroup
}

type ldapGroup struct {
	XMLName xml.Name `xml:"hpoa:ldapGroup"`
	Text    string   `xml:",chardata"`
}

type setLdapGroupBayAcl struct {
	XMLName   xml.Name `xml:"hpoa:setLdapGroupBayAcl"`
	LdapGroup ldapGroup
	Acl       acl
}

type acl struct {
	XMLName xml.Name `xml:"hpoa:acl"`
	Text    string   `xml:",chardata"`
}

type addLdapGroupBayAccess struct {
	XMLName   xml.Name `xml:"hpoa:addLdapGroupBayAccess"`
	LdapGroup ldapGroup
	Bays      bays
}

type bays struct {
	XMLName              xml.Name `xml:"hpoa:bays"`
	Hpoa                 string   `xml:"xmlns:hpoa,attr"`
	OaAccess             oaAccess
	BladeBays            bladeBays
	InterconnectTrayBays interconnectTrayBays
}

type oaAccess struct {
	XMLName xml.Name `xml:"hpoa:oaAccess"`
	Text    bool     `xml:",chardata"` //bool
}

type bladeBays struct {
	XMLName xml.Name `xml:"hpoa:bladeBays"`
	Blade   []blade
}

type blade struct {
	XMLName   xml.Name `xml:"hpoa:blade"`
	Hpoa      string   `xml:"xmlns:hpoa,attr"`
	BayNumber bayNumber
	Access    access
}

type bayNumber struct {
	XMLName xml.Name `xml:"hpoa:bayNumber"`
	Text    int      `xml:",chardata"`
}

type access struct {
	XMLName xml.Name `xml:"hpoa:access"`
	Text    bool     `xml:",chardata"`
}

type interconnectTrayBays struct {
	XMLName          xml.Name `xml:"hpoa:interconnectTrayBays"`
	InterconnectTray []interconnectTray
}

type interconnectTray struct {
	XMLName   xml.Name `xml:"hpoa:interconnectTray"`
	Hpoa      string   `xml:"xmlns:hpoa,attr"`
	BayNumber bayNumber
	Access    access
}

type enableLdapAuthentication struct {
	XMLName          xml.Name `xml:"hpoa:enableLdapAuthentication"`
	EnableLdap       bool     `xml:"hpoa:enableLdap"`
	EnableLocalUsers bool     `xml:"hpoa:enableLocalUsers"`
}

type setLdapInfo4 struct {
	XMLName                  xml.Name `xml:"hpoa:setLdapInfo4"`
	DirectoryServerAddress   string   `xml:"hpoa:directoryServerAddress"`
	DirectoryServerSslPort   int      `xml:"hpoa:directoryServerSslPort"`
	DirectoryServerGCPort    int      `xml:"hpoa:directoryServerGCPort"`
	UserNtAccountNameMapping bool     `xml:"hpoa:userNtAccountNameMapping"`
	EnableServiceAccount     bool     `xml:"hpoa:enableServiceAccount"`
	ServiceAccountName       string   `xml:"hpoa:serviceAccountName"`
	ServiceAccountPassword   string   `xml:"hpoa:serviceAccountPassword"`
	SearchContexts           SearchContexts
}

type SearchContexts struct {
	XMLName       xml.Name `xml:"hpoa:searchContexts"`
	Hpoa          string   `xml:"xmlns:hpoa,attr"`
	SearchContext []SearchContext
}

type SearchContext struct {
	XMLName xml.Name `xml:"hpoa:searchContext"`
	Text    string   `xml:",chardata"`
}

// manage power config
//<hpoa:setPowerConfigInfo><hpoa:redundancyMode>AC_REDUNDANT</hpoa:redundancyMode><hpoa:powerCeiling>0</hpoa:powerCeiling><hpoa:dynamicPowerSaverEnabled>false</hpoa:dynamicPowerSaverEnabled></hpoa:setPowerConfigInfo>

//mark setup wizard complete - required if the chassis was reset.
//<hpoa:setWizardComplete><hpoa:wizardStatus>WIZARD_SETUP_COMPLETE</hpoa:wizardStatus></hpoa:setWizardComplete>

// UserLogout is requied to logout the session
type UserLogout struct {
	XMLName xml.Name `xml:"hpoa:userLogOut"`
}
