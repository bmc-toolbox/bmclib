package m1000e

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	multierror "github.com/hashicorp/go-multierror"

	"github.com/bmc-toolbox/bmclib/providers/dell"
	log "github.com/sirupsen/logrus"
)

// retrieves ST2 which is required to submit form data
func (m *M1000e) getSessionToken() (token string, err error) {
	data, err := m.get("general")
	if err != nil {
		return token, err
	}
	//<input xmlns="" type="hidden" value="2a17b6d37baa526b75e06243d34763da" name="ST2" id="ST2" />
	re := regexp.MustCompile("<input.*value=\\\"(\\w+)\\\" name=\"ST2\"")
	match := re.FindSubmatch(data)
	if len(match) == 0 {
		return token, errors.ErrUnableToGetSessionToken
	}
	return string(match[1]), err
}

func (m *M1000e) loadHwData() (err error) {
	url := "json?method=groupinfo"
	payload, err := m.get(url)
	if err != nil {
		return err
	}

	m.cmcJSON = &dell.CMC{}
	err = json.Unmarshal(payload, m.cmcJSON)
	if err != nil {
		return err
	}

	if m.cmcJSON.Chassis == nil {
		return errors.ErrUnableToReadData
	}

	url = "json?method=blades-wwn-info"
	payload, err = m.get(url)
	if err != nil {
		return err
	}

	m.cmcWWN = &dell.CMCWWN{}
	err = json.Unmarshal(payload, m.cmcWWN)
	if err != nil {
		return err
	}

	return err
}

// Login initiates the connection to a chassis device
func (m *M1000e) httpLogin() (err error) {
	if m.httpClient != nil {
		return
	}
	log.WithFields(log.Fields{"step": "chassis connection", "vendor": dell.VendorID, "ip": m.ip}).Debug("connecting to chassis")

	form := url.Values{}
	form.Add("user", m.username)
	form.Add("password", m.password)

	u, err := url.Parse(fmt.Sprintf("https://%s/cgi-bin/webcgi/login", m.ip))
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", u.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	httpClient, err := httpclient.Build()
	if err != nil {
		return err
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	auth, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if strings.Contains(string(auth), "Try Again") {
		return errors.ErrLoginFailed
	}

	if resp.StatusCode == 404 {
		return errors.ErrPageNotFound
	}

	m.httpClient = httpClient

	err = m.loadHwData()
	if err != nil {
		return err
	}

	// retrieve session token to set config params.
	token, err := m.getSessionToken()
	if err != nil {
		return err
	}
	m.SessionToken = token

	return err
}

// Close closes the connection properly
func (m *M1000e) Close() error {
	var multiErr error
	if m.httpClient != nil {
		_, err := m.httpClient.Get(fmt.Sprintf("https://%s/cgi-bin/webcgi/logout", m.ip))
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}

	if err := m.sshClient.Close(); err != nil {
		multiErr = multierror.Append(multiErr, err)
	}

	return multiErr
}
