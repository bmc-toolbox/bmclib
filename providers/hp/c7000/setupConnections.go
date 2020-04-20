package c7000

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	multierror "github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
)

// Login initiates the connection to a chassis device
func (c *C7000) httpLogin() (err error) {
	if c.httpClient != nil {
		return
	}

	httpClient, err := httpclient.Build()
	if err != nil {
		return err
	}

	//setup the login payload
	username := Username{Text: c.username}
	password := Password{Text: c.password}
	userlogin := UserLogIn{Username: username, Password: password}

	//wrap the XML doc in the SOAP envelope
	doc := wrapXML(userlogin, "")

	payload, err := xml.MarshalIndent(doc, "  ", "    ")
	if err != nil {
		return err
	}

	u, err := url.Parse(fmt.Sprintf("https://%s/hpoa", c.ip))
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u.String(), bytes.NewReader(payload))
	if err != nil {
		return err
	}

	//req.Header.Add("Content-Type", "application/soap+xml; charset=utf-8")
	req.Header.Add("Content-Type", "text/plain;charset=UTF-8")
	if log.GetLevel() == log.TraceLevel {
		log.Println(fmt.Sprintf("https://%s/hpoa", c.ip))
		dump, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			log.Printf("%s\n\n", dump)
		}
	}

	resp, err := httpClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		return errors.ErrLoginFailed
	}
	defer resp.Body.Close()

	if log.GetLevel() == log.TraceLevel {
		dump, err := httputil.DumpResponse(resp, true)
		if err == nil {
			log.Printf("%s\n\n", dump)
		}
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var loginResponse EnvelopeLoginResponse
	err = xml.Unmarshal(responseBody, &loginResponse)
	if err != nil {
		return err
	}

	c.XMLToken = loginResponse.Body.UserLogInResponse.HpOaSessionKeyToken.OaSessionKey.Text
	if c.XMLToken == "" {
		return errors.ErrLoginFailed
	}

	c.httpClient = httpClient

	return err
}

// Close closes the connection properly
func (c *C7000) Close() error {
	var miltiErr error

	if c.httpClient != nil {
		payload := UserLogout{}
		_, _, err := c.postXML(payload)
		if err != nil {
			miltiErr = multierror.Append(miltiErr, err)
		}
	}

	if err := c.sshClient.Close(); err != nil {
		miltiErr = multierror.Append(miltiErr, err)
	}

	return miltiErr
}
