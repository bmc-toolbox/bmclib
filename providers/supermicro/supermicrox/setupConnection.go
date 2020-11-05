package supermicrox

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
	"github.com/bmc-toolbox/bmclib/providers/supermicro"
)

// httpLogin initiates the connection to an SupermicroX device
func (s *SupermicroX) httpLogin() (err error) {
	if s.httpClient != nil {
		return
	}

	httpClient, err := httpclient.Build()
	if err != nil {
		return err
	}

	s.log.V(1).Info("connecting to bmc", "step", "bmc connection", "vendor", supermicro.VendorID, "ip", s.ip)

	data := fmt.Sprintf("name=%s&pwd=%s", s.username, s.password)
	req, err := http.NewRequest("POST", fmt.Sprintf("https://%s/cgi/login.cgi", s.ip), bytes.NewBufferString(data))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode == 404 {
		return errors.ErrPageNotFound
	}

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if !strings.Contains(string(payload), "../cgi/url_redirect.cgi?url_name=mainmenu") {
		return errors.ErrLoginFailed
	}

	s.httpClient = httpClient

	return err
}

// Close closes the connection properly
func (s *SupermicroX) Close() (err error) {
	if s.httpClient != nil {
		bmcURL := fmt.Sprintf("https://%s/cgi/logout.cgi", s.ip)
		s.log.V(1).Info("logout from bmc", "step", "bmc connection", "vendor", supermicro.VendorID, "ip", s.ip)

		req, err := http.NewRequest("POST", bmcURL, nil)
		if err != nil {
			return err
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		u, err := url.Parse(bmcURL)
		if err != nil {
			return err
		}
		for _, cookie := range s.httpClient.Jar.Cookies(u) {
			if cookie.Name == "SID" && cookie.Value != "" {
				req.AddCookie(cookie)
			}
		}
		reqDump, _ := httputil.DumpRequestOut(req, true)
		s.log.V(2).Info("request", "url", fmt.Sprintf("https://%s/cgi/%s", bmcURL, s.ip), "requestDump", reqDump)

		resp, err := s.httpClient.Do(req)
		if err != nil {
			return err
		}

		defer resp.Body.Close()
		defer io.Copy(ioutil.Discard, resp.Body) // nolint

		return err
	}
	return err
}
