package discover

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bmc-toolbox/bmclib/devices"
	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/providers/dell/idrac8"
	"github.com/bmc-toolbox/bmclib/providers/dell/idrac9"
	"github.com/bmc-toolbox/bmclib/providers/dell/m1000e"
	"github.com/bmc-toolbox/bmclib/providers/hp"
	"github.com/bmc-toolbox/bmclib/providers/hp/c7000"
	"github.com/bmc-toolbox/bmclib/providers/hp/ilo"
	"github.com/bmc-toolbox/bmclib/providers/supermicro/supermicrox"
	log "github.com/sirupsen/logrus"
)

type Probe struct {
	client   *http.Client
	host     string
	username string
	password string
}

func (p *Probe) hpIlo() (bmcConnection interface{}, err error) {

	resp, err := p.client.Get(fmt.Sprintf("https://%s/xmldata?item=all", p.host))
	if err != nil {
		return bmcConnection, err
	}
	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body)

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bmcConnection, err
	}

	if resp.StatusCode != 200 {
		return bmcConnection, errors.ErrDeviceNotMatched
	}

	if bytes.Contains(payload[:6], []byte("RIMP")) {

		iloXMLC := &hp.Rimp{}
		err = xml.Unmarshal(payload, iloXMLC)
		if err != nil {
			return bmcConnection, err
		}

		iloXML := &hp.RimpBlade{}
		err = xml.Unmarshal(payload, iloXML)
		if err != nil {
			return bmcConnection, err
		}

		if iloXML.HSI != nil {
			if strings.HasPrefix(iloXML.MP.Pn, "Integrated Lights-Out") {
				return ilo.New(p.host, p.username, p.password)
			}

			return bmcConnection, fmt.Errorf("it's an HP, but I cound't not identify the hardware type. Please verify")
		}
	}

	return bmcConnection, errors.ErrDeviceNotMatched
}

func (p *Probe) hpC7000() (bmcConnection interface{}, err error) {

	resp, err := p.client.Get(fmt.Sprintf("https://%s/xmldata?item=all", p.host))
	if err != nil {
		return bmcConnection, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body)

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bmcConnection, err
	}

	if resp.StatusCode != 200 {
		return bmcConnection, errors.ErrDeviceNotMatched
	}

	if bytes.Contains(payload[:6], []byte("RIMP")) {

		iloXMLC := &hp.Rimp{}
		err = xml.Unmarshal(payload, iloXMLC)
		if err != nil {
			return bmcConnection, err
		}

		if iloXMLC.Infra2 != nil {
			log.WithFields(log.Fields{"step": "ScanAndConnect", "host": p.host, "vendor": devices.HP}).Debug("it's a chassis")
			return c7000.New(p.host, p.username, p.password)
		}

	}
	return bmcConnection, errors.ErrDeviceNotMatched
}

// hpCl100 attempts to identify a cloudline device
func (p *Probe) hpCl100() (bmcConnection interface{}, err error) {

	// HPE Cloudline CL100
	resp, err := p.client.Get(fmt.Sprintf("https://%s/res/ok.png", p.host))
	if err != nil {
		return bmcConnection, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body)

	var firstBytes = make([]byte, 8)
	_, err = io.ReadFull(resp.Body, firstBytes)
	if err != nil {
		return bmcConnection, err
	}
	// ensure the response we got included a png
	if resp.StatusCode == 200 && bytes.Contains(firstBytes, []byte("PNG")) {
		log.WithFields(log.Fields{"step": "ScanAndConnect", "host": p.host, "vendor": devices.Cloudline}).Debug("it's a discrete")
		return bmcConnection, errors.NewErrUnsupportedHardware("hpe cl100 not supported")
	}

	return bmcConnection, errors.ErrDeviceNotMatched

}

func (p *Probe) idrac8() (bmcConnection interface{}, err error) {

	resp, err := p.client.Get(fmt.Sprintf("https://%s/session?aimGetProp=hostname,gui_str_title_bar,OEMHostName,fwVersion,sysDesc", p.host))
	if err != nil {
		return bmcConnection, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body)

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bmcConnection, err
	}

	if resp.StatusCode == 200 && bytes.Contains(payload, []byte("PowerEdge M630")) {
		log.WithFields(log.Fields{"step": "connection", "host": p.host, "vendor": devices.Dell}).Debug("it's a idrac8")
		return idrac8.New(p.host, p.username, p.password)
	}

	return bmcConnection, errors.ErrDeviceNotMatched
}

func (p *Probe) idrac9() (bmcConnection interface{}, err error) {

	resp, err := p.client.Get(fmt.Sprintf("https://%s/sysmgmt/2015/bmc/info", p.host))
	if err != nil {
		return bmcConnection, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body)

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bmcConnection, err
	}

	if resp.StatusCode == 200 && bytes.Contains(payload, []byte("PowerEdge M640")) {
		log.WithFields(log.Fields{"step": "connection", "host": p.host, "vendor": devices.Dell}).Debug("it's a idrac9")
		return idrac9.New(p.host, p.username, p.password)
	}

	return bmcConnection, errors.ErrDeviceNotMatched
}

func (p *Probe) m1000e() (bmcConnection interface{}, err error) {
	resp, err := p.client.Get(fmt.Sprintf("https://%s/cgi-bin/webcgi/login", p.host))
	if err != nil {
		return bmcConnection, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body)

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bmcConnection, err
	}

	if resp.StatusCode == 200 && bytes.Contains(payload, []byte("PowerEdge M1000e")) {
		log.WithFields(log.Fields{"step": "connection", "host": p.host, "vendor": devices.Dell}).Debug("it's a m1000e chassis")
		return m1000e.New(p.host, p.username, p.password)
	}

	return bmcConnection, errors.ErrDeviceNotMatched
}

func (p *Probe) supermicrox() (bmcConnection interface{}, err error) {
	resp, err := p.client.Get(fmt.Sprintf("https://%s/cgi/login.cgi", p.host))
	if err != nil {
		return bmcConnection, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body)

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bmcConnection, err
	}

	// looking for ATEN in the response payload isn't the most ideal way, although it is unique to Supermicros
	if resp.StatusCode == 200 && bytes.Contains(payload, []byte("ATEN International")) {
		log.WithFields(log.Fields{"step": "connection", "host": p.host, "vendor": devices.Supermicro}).Debug("it's a supermicro")
		return supermicrox.New(p.host, p.username, p.password)
	}

	return bmcConnection, errors.ErrDeviceNotMatched
}

func (p *Probe) quanta() (bmcConnection interface{}, err error) {
	resp, err := p.client.Get(fmt.Sprintf("https://%s/page/login.html", p.host))
	if err != nil {
		return bmcConnection, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body)

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bmcConnection, err
	}

	// ensure the response we got included a png
	if resp.StatusCode == 200 && bytes.Contains(payload, []byte("Quanta")) {
		log.WithFields(log.Fields{"step": "ScanAndConnect", "host": p.host, "vendor": devices.Quanta}).Debug("it's a quanta")
		return bmcConnection, errors.NewErrUnsupportedHardware("quanta hardware not supported")
	}

	return bmcConnection, errors.ErrDeviceNotMatched
}
