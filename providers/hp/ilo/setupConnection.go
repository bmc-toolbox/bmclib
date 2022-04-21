package ilo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/providers/hp"

	multierror "github.com/hashicorp/go-multierror"
)

// Login initiates the connection to a bmc device
func (i *Ilo) httpLogin() (err error) {
	if i.httpClient != nil {
		return
	}

	httpClient, err := httpclient.Build(i.httpClientSetupFuncs...)
	if err != nil {
		return err
	}

	i.log.V(1).Info("connecting to bmc", "step", "bmc connection", "vendor", hp.VendorID, "ip", i.ip)

	data := fmt.Sprintf("{\"method\":\"login\", \"user_login\":\"%s\", \"password\":\"%s\" }", i.username, i.password)

	req, err := http.NewRequest("POST", i.loginURL.String(), bytes.NewBufferString(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	reqDump, _ := httputil.DumpRequestOut(req, true)
	i.log.V(2).Info("requestTrace", "requestDump", string(reqDump), "url", i.loginURL.String())

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

	if i.sessionKey == "" {
		i.log.V(1).Info("Expected sessionKey cookie value not found.", "step", "Login()", "IP", i.ip, "HardwareType", i.HardwareType())
	}

	if resp.StatusCode == 404 {
		return errors.ErrPageNotFound
	}

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respDump, _ := httputil.DumpResponse(resp, true)
	i.log.V(2).Info("responseTrace", "responseDump", string(respDump))

	if strings.Contains(string(payload), "Invalid login attempt") {
		return errors.ErrLoginFailed
	}

	i.httpClient = httpClient

	return err
}

// Close closes the connection properly
func (i *Ilo) Close(ctx context.Context) error {
	var multiErr error

	if i.httpClient != nil {
		i.log.V(1).Info("logout from bmc http", "step", "bmc connection", "vendor", hp.VendorID, "ip", i.ip)

		data := []byte(fmt.Sprintf(`{"method":"logout", "session_key": "%s"}`, i.sessionKey))

		req, err := http.NewRequest("POST", i.loginURL.String(), bytes.NewBuffer(data))
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		} else {
			req.Header.Set("Content-Type", "application/json")

			reqDump, _ := httputil.DumpRequestOut(req, true)
			i.log.V(2).Info("requestTrace", "requestDump", string(reqDump), "url", i.loginURL.String())

			resp, err := i.httpClient.Do(req)
			if err != nil {
				multiErr = multierror.Append(multiErr, err)
			} else {
				defer resp.Body.Close()
				defer io.Copy(ioutil.Discard, resp.Body) // nolint

				respDump, _ := httputil.DumpResponse(resp, true)
				i.log.V(2).Info("responseTrace", "responseDump", string(respDump))
			}
		}
	}

	if err := i.sshClient.Close(); err != nil {
		multiErr = multierror.Append(multiErr, err)
	}

	return multiErr
}
