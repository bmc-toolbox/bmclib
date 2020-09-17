package discover

import (
	"bytes"
	"context"
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
	"github.com/bmc-toolbox/bmclib/providers/supermicro/supermicrox11"
	"github.com/go-logr/logr"
)

var (
	idrac8SysDesc = []string{"PowerEdge M630", "PowerEdge R630"}
	idrac9SysDesc = []string{"PowerEdge M640", "PowerEdge R640", "PowerEdge R6415", "PowerEdge R6515"}
	m1000eSysDesc = []string{"PowerEdge M1000e"}
)

type Probe struct {
	client   *http.Client
	host     string
	username string
	password string
}

func (p *Probe) hpIlo(ctx context.Context, log logr.Logger) (bmcConnection interface{}, err error) {

	resp, err := p.client.Get(fmt.Sprintf("https://%s/xmldata?item=all", p.host))
	if err != nil {
		return bmcConnection, err
	}
	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body) // nolint

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
				log.V(1).Info("step", "ScanAndConnect", "host", p.host, "vendor", string(devices.HP), "msg", "it's a HP with iLo")
				return ilo.New(ctx, p.host, p.username, p.password, log)
			}

			return bmcConnection, fmt.Errorf("it's an HP, but I cound't not identify the hardware type. Please verify")
		}
	}

	return bmcConnection, errors.ErrDeviceNotMatched
}

func (p *Probe) hpC7000(ctx context.Context, log logr.Logger) (bmcConnection interface{}, err error) {

	resp, err := p.client.Get(fmt.Sprintf("https://%s/xmldata?item=all", p.host))
	if err != nil {
		return bmcConnection, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body) // nolint

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
			log.V(1).Info("step", "ScanAndConnect", "host", p.host, "vendor", string(devices.HP), "msg", "it's a chassis")
			return c7000.New(ctx, p.host, p.username, p.password, log)
		}

	}
	return bmcConnection, errors.ErrDeviceNotMatched
}

// hpCl100 attempts to identify a cloudline device
func (p *Probe) hpCl100(ctx context.Context, log logr.Logger) (bmcConnection interface{}, err error) {

	// HPE Cloudline CL100
	resp, err := p.client.Get(fmt.Sprintf("https://%s/res/ok.png", p.host))
	if err != nil {
		return bmcConnection, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body) // nolint

	var firstBytes = make([]byte, 8)
	_, err = io.ReadFull(resp.Body, firstBytes)
	if err != nil {
		return bmcConnection, err
	}
	// ensure the response we got included a png
	if resp.StatusCode == 200 && bytes.Contains(firstBytes, []byte("PNG")) {
		log.V(1).Info("step", "ScanAndConnect", "host", p.host, "vendor", string(devices.Cloudline), "msg", "it's a discrete")
		return bmcConnection, errors.NewErrUnsupportedHardware("hpe cl100 not supported")
	}

	return bmcConnection, errors.ErrDeviceNotMatched

}

func (p *Probe) idrac8(ctx context.Context, log logr.Logger) (bmcConnection interface{}, err error) {

	resp, err := p.client.Get(fmt.Sprintf("https://%s/session?aimGetProp=hostname,gui_str_title_bar,OEMHostName,fwVersion,sysDesc", p.host))
	if err != nil {
		return bmcConnection, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body) // nolint

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bmcConnection, err
	}

	if resp.StatusCode == 200 && containsAnySubStr(payload, idrac8SysDesc) {
		log.V(1).Info("step", "connection", "host", p.host, "vendor", string(devices.Dell), "msg", "it's a idrac8")
		return idrac8.New(ctx, p.host, p.username, p.password, log)
	}

	return bmcConnection, errors.ErrDeviceNotMatched
}

func (p *Probe) idrac9(ctx context.Context, log logr.Logger) (bmcConnection interface{}, err error) {

	resp, err := p.client.Get(fmt.Sprintf("https://%s/sysmgmt/2015/bmc/info", p.host))
	if err != nil {
		return bmcConnection, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body) // nolint

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bmcConnection, err
	}

	if resp.StatusCode == 200 && containsAnySubStr(payload, idrac9SysDesc) {
		log.V(1).Info("step", "connection", "host", p.host, "vendor", string(devices.Dell), "msg", "it's a idrac9")
		return idrac9.New(ctx, p.host, p.host, p.username, p.password, log)
	}

	return bmcConnection, errors.ErrDeviceNotMatched
}

func (p *Probe) m1000e(ctx context.Context, log logr.Logger) (bmcConnection interface{}, err error) {
	resp, err := p.client.Get(fmt.Sprintf("https://%s/cgi-bin/webcgi/login", p.host))
	if err != nil {
		return bmcConnection, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body) // nolint

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bmcConnection, err
	}

	if resp.StatusCode == 200 && containsAnySubStr(payload, m1000eSysDesc) {
		log.V(1).Info("step", "connection", "host", p.host, "vendor", string(devices.Dell), "msg", "it's a m1000e chassis")
		return m1000e.New(ctx, p.host, p.username, p.password, log)
	}

	return bmcConnection, errors.ErrDeviceNotMatched
}

func (p *Probe) supermicrox(ctx context.Context, log logr.Logger) (bmcConnection interface{}, err error) {
	resp, err := p.client.Get(fmt.Sprintf("https://%s/cgi/login.cgi", p.host))
	if err != nil {
		return bmcConnection, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body) // nolint

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bmcConnection, err
	}

	// looking for ATEN in the response payload isn't the most ideal way, although it is unique to Supermicros
	if resp.StatusCode == 200 && bytes.Contains(payload, []byte("ATEN International")) {
		log.V(1).Info("it's a supermicro", "step", "connection", "host", p.host, "vendor", devices.Supermicro, "hardwareType", supermicrox.X10)
		conn, err := supermicrox.New(ctx, p.host, p.username, p.password, log)
		if err != nil {
			return bmcConnection, err
		}
		hwType := conn.HardwareType()
		if hwType == supermicrox.X10 {
			return conn, err
		}
	}

	return bmcConnection, errors.ErrDeviceNotMatched
}

func (p *Probe) supermicrox11(ctx context.Context, log logr.Logger) (bmcConnection interface{}, err error) {
	resp, err := p.client.Get(fmt.Sprintf("https://%s/cgi/login.cgi", p.host))
	if err != nil {
		return bmcConnection, err
	}
	defer resp.Body.Close()

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bmcConnection, err
	}

	// looking for ATEN in the response payload isn't the most ideal way, although it is unique to Supermicros
	if resp.StatusCode == 200 && bytes.Contains(payload, []byte("ATEN International")) {
		log.V(1).Info("it's a supermicrox11", "step", "connection", "host", p.host, "vendor", devices.Supermicro, "hardwareType", supermicrox11.X11)

		conn, err := supermicrox11.New(ctx, p.host, p.username, p.password, log)
		if err != nil {
			return bmcConnection, err
		}
		hwType := conn.HardwareType()
		if hwType == supermicrox11.X11 {
			return conn, err
		}
	}

	return bmcConnection, errors.ErrDeviceNotMatched
}

func (p *Probe) quanta(ctx context.Context, log logr.Logger) (bmcConnection interface{}, err error) {
	resp, err := p.client.Get(fmt.Sprintf("https://%s/page/login.html", p.host))
	if err != nil {
		return bmcConnection, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body) // nolint

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bmcConnection, err
	}

	// ensure the response we got included a png
	if resp.StatusCode == 200 && bytes.Contains(payload, []byte("Quanta")) {
		log.V(1).Info("step", "ScanAndConnect", "host", p.host, "vendor", string(devices.Quanta), "msg", "it's a quanta")
		return bmcConnection, errors.NewErrUnsupportedHardware("quanta hardware not supported")
	}

	return bmcConnection, errors.ErrDeviceNotMatched
}

func containsAnySubStr(data []byte, subStrs []string) bool {
	for _, subStr := range subStrs {
		if bytes.Contains(data, []byte(subStr)) {
			return true
		}
	}
	return false
}
