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

	// this make possible to setup logging and properties at any stage
	_ "github.com/bmc-toolbox/bmclib/logging"
	log "github.com/sirupsen/logrus"
)

func (i *IDrac8) httpLogin() (err error) {
	if i.httpClient != nil {
		return
	}	

	log.WithFields(log.Fields{"step": "bmc connection", "vendor": dell.VendorID, "ip": i.ip}).Debug("connecting to bmc")

	data := fmt.Sprintf("user=%s&password=%s", i.username, i.password)
	url := fmt.Sprintf("https://%s/data/login", i.ip)
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(data))
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := i.client.Do(req)
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

	i.st1 = strings.TrimLeft(stTemp[0], "index.html?ST1=")
	i.st2 = strings.TrimLeft(stTemp[1], "ST2=")

	err = i.loadHwData()
	if err != nil {
		return err
	}

	serial, err := i.Serial()
	if err != nil {
		return err
	}
	i.serial = serial

	return err
}

// loadHwData load the full hardware information from the iDrac
func (i *IDrac8) loadHwData() (err error) {
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

// sshLogin initiates the connection to a chassis device
func (i *IDrac8) sshLogin() (err error) {
	if m.sshClient != nil {
		return
	}

	log.WithFields(log.Fields{"step": "chassis connection", "vendor": dell.VendorID, "ip": m.ip}).Debug("connecting to chassis")
	m.sshClient, err = sshclient.New(m.ip, m.username, m.password)
	if err != nil {
		return err
	}

	return err
}


// Close logs out and close the bmc connection
func (i *IDrac8) Close() (err error) {
	log.WithFields(log.Fields{"step": "bmc connection", "vendor": dell.VendorID, "ip": i.ip}).Debug("logout from bmc")

	resp, err := i.client.Get(fmt.Sprintf("https://%s/data/logout", i.ip))
	if err != nil {
		return err
	}
	io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	return err
}