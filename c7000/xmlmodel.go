package c7000

import (
	"encoding/xml"
)

type Username struct {
	Text string `xml:",chardata"`
}

type Password struct {
	Text string `xml:",chardata"`
}

type UserLogIn struct {
	XMLName  xml.Name `xml:"hpoa:userLogIn"`
	Text     string   `xml:",chardata"`
	Username Username `xml:"hpoa:username"`
	Password Password `xml:"hpoa:password"`
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
