package ilo

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/internal/sshclient"
	"github.com/bmc-toolbox/bmclib/providers/hp"

	multierror "github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
)

// Login initiates the connection to a chassis device
func (i *Ilo) httpLogin() (err error) {
	if i.httpClient != nil {
		return
	}

	httpClient, err := httpclient.Build()
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{"step": "bmc connection", "vendor": hp.VendorID, "ip": i.ip}).Debug("connecting to bmc")

	data := fmt.Sprintf("{\"method\":\"login\", \"user_login\":\"%s\", \"password\":\"%s\" }", i.username, i.password)

	req, err := http.NewRequest("POST", i.loginURL.String(), bytes.NewBufferString(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	u, err := url.Parse(i.loginURL.String())
	if err != nil {
		return err
	}

	for _, cookie := range httpClient.Jar.Cookies(u) {
		if cookie.Name == "sessionKey" {
			i.sessionKey = cookie.Value
		}
	}

	if log.GetLevel() == log.DebugLevel {
		dump, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			log.Println(fmt.Sprintf("[Request] %s", i.loginURL.String()))
			log.Println(">>>>>>>>>>>>>>>")
			log.Printf("%s\n\n", dump)
			log.Println(">>>>>>>>>>>>>>>")
		}
	}

	if i.sessionKey == "" {
		log.WithFields(log.Fields{
			"step":  "Login()",
			"IP":    i.ip,
			"Model": i.BmcType(),
		}).Warn("Expected sessionKey cookie value not found.")
	}

	if resp.StatusCode == 404 {
		return errors.ErrPageNotFound
	}

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if log.GetLevel() == log.DebugLevel {
		dump, err := httputil.DumpResponse(resp, true)
		if err == nil {
			log.Println("[Response]")
			log.Println("<<<<<<<<<<<<<<")
			log.Printf("%s\n\n", dump)
			log.Println("<<<<<<<<<<<<<<")
		}
	}

	if strings.Contains(string(payload), "Invalid login attempt") {
		return errors.ErrLoginFailed
	}

	i.httpClient = httpClient

	serial, err := i.Serial()
	if err != nil {
		return err
	}
	i.serial = serial

	return err
}

// Login initiates the connection to a chassis device
func (i *Ilo) sshLogin() (err error) {
	if i.sshClient != nil {
		return
	}

	log.WithFields(log.Fields{"step": "chassis connection", "vendor": hp.VendorID, "ip": i.ip}).Debug("connecting to chassis")
	i.sshClient, err = sshclient.New(i.ip, i.username, i.password)
	if err != nil {
		return err
	}

	return err
}

// Close closes the connection properly
func (i *Ilo) Close() (err error) {
	if i.httpClient != nil {
		log.WithFields(log.Fields{"step": "bmc connection", "vendor": hp.VendorID, "ip": i.ip}).Debug("logout from bmc http")

		data := []byte(`{"method":"logout"}`)

		req, e := http.NewRequest("POST", i.loginURL.String(), bytes.NewBuffer(data))
		if e != nil {
			err = multierror.Append(e, err)
		} else {
			req.Header.Set("Content-Type", "application/json")

			resp, e := i.httpClient.Do(req)
			if e != nil {
				err = multierror.Append(e, err)
			} else {
				defer resp.Body.Close()
				defer io.Copy(ioutil.Discard, resp.Body)
			}
		}
	}

	if i.sshClient != nil {
		log.WithFields(log.Fields{"step": "bmc connection", "vendor": hp.VendorID, "ip": i.ip}).Debug("logout from bmc ssh")

		e := i.sshClient.Close()
		if e != nil {
			err = multierror.Append(e, err)
		}
	}

	return err
}
