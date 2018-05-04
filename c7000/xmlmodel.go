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
	Server  Server
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
// <hpoa:configureNtp>
//   <hpoa:ntpPrimary>ntp0.example.com</hpoa:ntpPrimary>
//   <hpoa:ntpSecondary>ntp1.example.com</hpoa:ntpSecondary>
//   <hpoa:ntpPoll>720</hpoa:ntpPoll>
//  </hpoa:configureNtp>
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

// <hpoa:setEnclosureTimeZone>
//  <hpoa:timeZone>CET</hpoa:timeZone>
// </hpoa:setEnclosureTimeZone>
type setEnclosureTimeZone struct {
	XMLName  xml.Name `xml:"hpoa:setEnclosureTimeZone"`
	Timezone timeZone
}

type timeZone struct {
	XMLName xml.Name `xml:"hpoa:timeZone"`
	Text    string   `xml:",chardata"`
}
