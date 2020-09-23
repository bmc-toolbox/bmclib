package idrac8

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/providers/dell"
	multierror "github.com/hashicorp/go-multierror"
)

func (i *IDrac8) httpLogin() (err error) {
	if i.httpClient != nil {
		return
	}

	httpClient, err := httpclient.Build()
	if err != nil {
		return err
	}

	i.log.V(1).Info("connecting to bmc", "step", "bmc connection", "vendor", dell.VendorID, "ip", i.ip)

	data := fmt.Sprintf("user=%s&password=%s", i.username, i.password)
	url := fmt.Sprintf("https://%s/data/login", i.ip)
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(data))
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

	iDracAuth := &dell.IDracAuth{}
	err = xml.Unmarshal(payload, iDracAuth)
	if err != nil {
		return err
	}

	stTemp := strings.Split(iDracAuth.ForwardURL, ",")
	if len(stTemp) != 2 {
		return errors.ErrLoginFailed
	}

	i.httpClient = httpClient

	i.st1 = strings.TrimLeft(stTemp[0], "index.html?ST1=")
	i.st2 = strings.TrimLeft(stTemp[1], "ST2=")

	return err
}

// loadHwData load the full hardware information from the iDrac
func (i *IDrac8) loadHwData() (err error) {
	err = i.httpLogin()
	if err != nil {
		return err
	}

	if i.iDracInventory != nil {
		return err
	}

	url := "sysmgmt/2012/server/inventory/hardware"
	payload, err := i.get(url, nil)
	if err != nil {
		return err
	}

	iDracInventory := &dell.IDracInventory{}
	err = xml.Unmarshal(payload, iDracInventory)
	if err != nil {
		return err
	}

	if iDracInventory.Component == nil {
		return errors.ErrUnableToReadData
	}

	i.iDracInventory = iDracInventory

	return err
}

// Close closes the connection properly
func (i *IDrac8) Close() error {
	var multiErr error

	if i.httpClient != nil {
		resp, err := i.httpClient.Get(fmt.Sprintf("https://%s/data/logout", i.ip))
		if err != nil {
			multiErr = multierror.Append(multiErr, err)
		} else {
			defer resp.Body.Close()
			defer io.Copy(ioutil.Discard, resp.Body)
		}
	}

	if err := i.sshClient.Close(); err != nil {
		multiErr = multierror.Append(multiErr, err)
	}

	return multiErr
}
