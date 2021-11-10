package idrac9

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/providers/dell"
	multierror "github.com/hashicorp/go-multierror"
)

func (i *IDrac9) httpLogin() (err error) {
	if i.httpClient != nil {
		return
	}

	httpClient, err := httpclient.Build()
	if err != nil {
		return err
	}

	i.log.V(1).Info("connecting to bmc", "step", "bmc connection", "vendor", dell.VendorID, "ip", i.ip)

	url := fmt.Sprintf("https://%s/sysmgmt/2015/bmc/session", i.ip)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("user", fmt.Sprintf("\"%s\"", i.username))
	req.Header.Add("password", fmt.Sprintf("\"%s\"", i.password))

	reqDump, _ := httputil.DumpRequestOut(req, true)
	i.log.V(2).Info("requestTrace", "requestDump", string(reqDump), "url", url)

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	switch resp.StatusCode {
	case 404:
		return errors.ErrPageNotFound
	case 503:
		return errors.ErrIdracMaxSessionsReached
	}

	i.xsrfToken = resp.Header.Get("XSRF-TOKEN")

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respDump, _ := httputil.DumpResponse(resp, true)
	i.log.V(2).Info("responseTrace", "responseDump", string(respDump))

	iDracAuth := &dell.IDracAuth{}
	err = json.Unmarshal(payload, iDracAuth)
	if err != nil {
		return err
	}

	// 0 = Login successful.
	// 7 = Login successful with default credentials.
	if iDracAuth.AuthResult != 0 && iDracAuth.AuthResult != 7 {
		return errors.ErrLoginFailed
	}

	i.httpClient = httpClient
	return nil
}

// loadHwData load the full hardware information from the iDrac
func (i *IDrac9) loadHwData() (err error) {
	err = i.httpLogin()
	if err != nil {
		return fmt.Errorf("IDrac9.loadHwData(): HTTP login problem: " + err.Error())
	}

	url := "sysmgmt/2012/server/inventory/hardware"
	statusCode, response, err := i.get(url, nil)
	if err != nil || statusCode != 200 {
		msg := fmt.Sprintf("Status code is %d", statusCode)
		if err != nil {
			msg += ", Error: " + err.Error()
		} else {
			msg += ", Response: " + string(response)
		}
		return fmt.Errorf("IDrac9.loadHwData(): GET request failed. " + msg)
	}

	iDracInventory := &dell.IDracInventory{}
	err = xml.Unmarshal(response, iDracInventory)
	if err != nil {
		return fmt.Errorf("IDrac9.loadHwData(): XML unmarshal problem: " + err.Error())
	}

	if iDracInventory.Component == nil {
		return fmt.Errorf("IDrac9.loadHwData(): HTTP login problem: " + errors.ErrUnableToReadData.Error())
	}

	i.iDracInventory = iDracInventory
	return nil
}

// Close closes the connection properly
func (i *IDrac9) Close(ctx context.Context) error {
	var multiErr error

	if i.httpClient != nil {
		if _, _, err := i.delete("sysmgmt/2015/bmc/session"); err != nil {
			multiErr = multierror.Append(multiErr, err)
		}
	}

	if err := i.sshClient.Close(); err != nil {
		multiErr = multierror.Append(multiErr, err)
	}

	return multiErr
}
