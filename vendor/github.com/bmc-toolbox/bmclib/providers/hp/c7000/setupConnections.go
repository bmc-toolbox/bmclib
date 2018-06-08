package c7000

import (
	"encoding/xml"

	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/sshclient"
	"github.com/bmc-toolbox/bmclib/providers/hp"
	log "github.com/sirupsen/logrus"
)

// Login initiates the connection to a chassis device
func (c *C7000) httpLogin() (err error) {
	//setup the login payload
	username := Username{Text: c.username}
	password := Password{Text: c.password}
	userlogin := UserLogIn{Username: username, Password: password}

	//wrap the XML doc in the SOAP envelope
	doc := wrapXML(userlogin, "")

	output, err := xml.MarshalIndent(doc, "  ", "    ")
	if err != nil {
		return err
	}

	statusCode, responseBody, err := c.postXML(output)

	if err != nil || statusCode != 200 {
		return errors.ErrLoginFailed
	}

	var loginResponse EnvelopeLoginResponse
	err = xml.Unmarshal(responseBody, &loginResponse)
	if err != nil {
		return errors.ErrLoginFailed
	}

	c.XMLToken = loginResponse.Body.UserLogInResponse.HpOaSessionKeyToken.OaSessionKey.Text

	serial, err := c.Serial()
	if err != nil {
		return err
	}
	c.serial = serial

	return err
}

// Login initiates the connection to a chassis device
func (c *C7000) sshLogin() (err error) {
	if c.sshClient != nil {
		return
	}

	log.WithFields(log.Fields{"step": "chassis connection", "vendor": hp.VendorID, "ip": c.ip}).Debug("connecting to chassis")
	c.sshClient, err = sshclient.New(c.ip, c.username, c.password)
	if err != nil {
		return err
	}

	return err
}
