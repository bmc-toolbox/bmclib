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
	"github.com/bmc-toolbox/bmclib/internal/sshclient"
	"github.com/bmc-toolbox/bmclib/providers/dell"
	multierror "github.com/hashicorp/go-multierror"

	// this make possible to setup logging and properties at any stage
	_ "github.com/bmc-toolbox/bmclib/logging"
	log "github.com/sirupsen/logrus"
)

func (i *IDrac8) httpLogin() (err error) {
	if i.httpClient != nil {
		return
	}

	httpClient, err := httpclient.Build()
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{"step": "bmc connection", "vendor": dell.VendorID, "ip": i.ip}).Debug("connecting to bmc")

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
		httpclient.DumpInvalidPayload(url, i.ip, payload)
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
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return err
	}

	if iDracInventory == nil || iDracInventory.Component == nil {
		return errors.ErrUnableToReadData
	}

	i.iDracInventory = iDracInventory

	return err
}

// sshLogin initiates the connection to a bmc device
func (i *IDrac8) sshLogin() (err error) {
	if i.sshClient != nil {
		return
	}

	log.WithFields(log.Fields{"step": "bmc connection", "vendor": dell.VendorID, "ip": i.ip}).Debug("connecting to bmc")
	i.sshClient, err = sshclient.New(i.ip, i.username, i.password)
	if err != nil {
		return err
	}

	return err
}

// Close closes the connection properly
func (i *IDrac8) Close() (err error) {
	if i.httpClient != nil {
		resp, e := i.httpClient.Get(fmt.Sprintf("https://%s/data/logout", i.ip))
		if e != nil {
			err = multierror.Append(e, err)
		} else {
			defer resp.Body.Close()
			defer io.Copy(ioutil.Discard, resp.Body)
		}
	}

	if i.sshClient != nil {
		e := i.sshClient.Close()
		if e != nil {
			err = multierror.Append(e, err)
		}
	}

	return err
}
