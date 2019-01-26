package idrac9

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"

	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/internal/sshclient"
	"github.com/bmc-toolbox/bmclib/providers/dell"
	multierror "github.com/hashicorp/go-multierror"

	// this make possible to setup logging and properties at any stage
	_ "github.com/bmc-toolbox/bmclib/logging"
	log "github.com/sirupsen/logrus"
)

func (i *IDrac9) httpLogin() (err error) {
	if i.httpClient != nil {
		return
	}

	httpClient, err := httpclient.Build()
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{"step": "bmc connection", "vendor": dell.VendorID, "ip": i.ip}).Debug("connecting to bmc")

	url := fmt.Sprintf("https://%s/sysmgmt/2015/bmc/session", i.ip)
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("user", fmt.Sprintf("\"%s\"", i.username))
	req.Header.Add("password", fmt.Sprintf("\"%s\"", i.password))

	if log.GetLevel() == log.DebugLevel {
		dump, err := httputil.DumpRequestOut(req, true)
		if err == nil {
			log.Println(fmt.Sprintf("[Request] %s", url))
			log.Println(">>>>>>>>>>>>>>>")
			log.Printf("%s\n\n", dump)
			log.Println(">>>>>>>>>>>>>>>")
		}
	}

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

	if log.GetLevel() == log.DebugLevel {
		dump, err := httputil.DumpResponse(resp, false)
		if err == nil {
			log.Println("[Response]")
			log.Println("<<<<<<<<<<<<<<")
			log.Printf("%s\n\n", dump)
			log.Println("<<<<<<<<<<<<<<")
		}
	}

	iDracAuth := &dell.IDracAuth{}
	err = json.Unmarshal(payload, iDracAuth)
	if err != nil {
		httpclient.DumpInvalidPayload(url, i.ip, payload)
		return err
	}

	//0 = login success.
	//7 = login success with default credentials.
	if iDracAuth.AuthResult != 0 && iDracAuth.AuthResult != 7 {
		return errors.ErrLoginFailed
	}

	i.httpClient = httpClient

	return err
}

// loadHwData load the full hardware information from the iDrac
func (i *IDrac9) loadHwData() (err error) {
	err = i.httpLogin()
	if err != nil {
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
func (i *IDrac9) sshLogin() (err error) {
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
func (i *IDrac9) Close() (err error) {
	if i.httpClient != nil {
		_, _, e := i.delete("sysmgmt/2015/bmc/session")
		if e != nil {
			err = multierror.Append(e, err)
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
