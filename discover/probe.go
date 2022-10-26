package discover

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

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
	idrac8SysDesc = []string{"PowerEdge M630", "PowerEdge R630", "PowerEdge C6320"}
	idrac9SysDesc = []string{"PowerEdge M640", "PowerEdge R640", "PowerEdge R6415", "PowerEdge R6515", "PowerEdge R740xd"}
	m1000eSysDesc = []string{"PowerEdge M1000e"}
)

type Probe struct {
	client    *http.Client
	host      string
	username  string
	password  string
	certPool  *x509.CertPool
	secureTLS bool
}

func (p *Probe) hpIlo(ctx context.Context, log logr.Logger) (bmcConnection interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*60))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s/xmldata?item=all", p.host), nil)
	if err != nil {
		return bmcConnection, err
	}
	resp, err := p.client.Do(req)
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
				opts := []ilo.IloOption{}
				if p.secureTLS {
					opts = append(opts, ilo.WithSecureTLS(p.certPool))
				}
				log.V(1).Info("step", "ScanAndConnect", "host", p.host, "vendor", string(devices.HP), "msg", "it's a HP with iLo")
				return ilo.NewWithOptions(ctx, p.host, p.username, p.password, log, opts...)
			}

			return bmcConnection, fmt.Errorf("it's an HP, but I cound't not identify the hardware type. Please verify")
		}
	}

	return bmcConnection, errors.ErrDeviceNotMatched
}

func (p *Probe) hpC7000(ctx context.Context, log logr.Logger) (bmcConnection interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*120))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s/xmldata?item=all", p.host), nil)
	if err != nil {
		return bmcConnection, err
	}
	resp, err := p.client.Do(req)
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
			opts := []c7000.C7000Option{}
			if p.secureTLS {
				opts = append(opts, c7000.WithSecureTLS(p.certPool))
			}
			log.V(1).Info("step", "ScanAndConnect", "host", p.host, "vendor", string(devices.HP), "msg", "it's a chassis")
			return c7000.NewWithOptions(ctx, p.host, p.username, p.password, log, opts...)
		}
	}
	return bmcConnection, errors.ErrDeviceNotMatched
}

// Attempts to identify an HPE Cloudline CL100 device.
func (p *Probe) hpCl100(ctx context.Context, log logr.Logger) (bmcConnection interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*60))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s/res/ok.png", p.host), nil)
	if err != nil {
		return bmcConnection, err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return bmcConnection, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body) // nolint

	firstBytes := make([]byte, 8)
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
	ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*60))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s/session?aimGetProp=hostname,gui_str_title_bar,OEMHostName,fwVersion,sysDesc", p.host), nil)
	if err != nil {
		return bmcConnection, err
	}
	resp, err := p.client.Do(req)
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
		opts := []idrac8.IDrac8Option{}
		if p.secureTLS {
			opts = append(opts, idrac8.WithSecureTLS(p.certPool))
		}
		log.V(1).Info("step", "connection", "host", p.host, "vendor", string(devices.Dell), "msg", "it's a idrac8")
		return idrac8.NewWithOptions(ctx, p.host, p.username, p.password, log, opts...)
	}

	return bmcConnection, errors.ErrDeviceNotMatched
}

func (p *Probe) idrac9(ctx context.Context, log logr.Logger) (bmcConnection interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*60))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s/sysmgmt/2015/bmc/info", p.host), nil)
	if err != nil {
		return bmcConnection, err
	}
	resp, err := p.client.Do(req)
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
		opts := []idrac9.IDrac9Option{}
		if p.secureTLS {
			opts = append(opts, idrac9.WithSecureTLS(p.certPool))
		}
		log.V(1).Info("step", "connection", "host", p.host, "vendor", string(devices.Dell), "msg", "it's a idrac9")
		return idrac9.NewWithOptions(ctx, p.host, p.host, p.username, p.password, log, opts...)
	}

	return bmcConnection, errors.ErrDeviceNotMatched
}

func (p *Probe) m1000e(ctx context.Context, log logr.Logger) (bmcConnection interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*60))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s/cgi-bin/webcgi/login", p.host), nil)
	if err != nil {
		return bmcConnection, err
	}
	resp, err := p.client.Do(req)
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
		opts := []m1000e.M1000eOption{}
		if p.secureTLS {
			opts = append(opts, m1000e.WithSecureTLS(p.certPool))
		}
		log.V(1).Info("step", "connection", "host", p.host, "vendor", string(devices.Dell), "msg", "it's a m1000e chassis")
		return m1000e.NewWithOptions(ctx, p.host, p.username, p.password, log, opts...)
	}

	return bmcConnection, errors.ErrDeviceNotMatched
}

func (p *Probe) supermicrox(ctx context.Context, log logr.Logger) (bmcConnection interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*60))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s/cgi/login.cgi", p.host), nil)
	if err != nil {
		return bmcConnection, err
	}
	resp, err := p.client.Do(req)
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
		opts := []supermicrox.SupermicroXOption{}
		if p.secureTLS {
			opts = append(opts, supermicrox.WithSecureTLS(p.certPool))
		}
		log.V(1).Info("it's a supermicro", "step", "connection", "host", p.host, "vendor", devices.Supermicro, "hardwareType", supermicrox.X10)

		conn, err := supermicrox.NewWithOptions(ctx, p.host, p.username, p.password, log, opts...)
		if err != nil {
			return bmcConnection, err
		}
		// empty string means that either HardwareType() was unable to get the model or the model returned was empty
		// if HardwareType() returned something more than empty than the call worked, we'll assume other calls will
		// also work
		if conn.HardwareType() != "" {
			return conn, err
		}
	}

	return bmcConnection, errors.ErrDeviceNotMatched
}

func (p *Probe) supermicrox11(ctx context.Context, log logr.Logger) (bmcConnection interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*60))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s/cgi/login.cgi", p.host), nil)
	if err != nil {
		return bmcConnection, err
	}
	resp, err := p.client.Do(req)
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
		opts := []supermicrox11.SupermicroXOption{}
		if p.secureTLS {
			opts = append(opts, supermicrox11.WithSecureTLS(p.certPool))
		}
		log.V(1).Info("it's a supermicrox11", "step", "connection", "host", p.host, "vendor", devices.Supermicro, "hardwareType", supermicrox11.X11)

		conn, err := supermicrox11.NewWithOptions(ctx, p.host, p.username, p.password, log, opts...)
		if err != nil {
			return bmcConnection, err
		}
		// empty string means that either HardwareType() was unable to get the model or the model returned was empty
		// if HardwareType() returned something more than empty than the call worked, we'll assume other calls will
		// also work
		if conn.HardwareType() != "" {
			return conn, err
		}
	}

	return bmcConnection, errors.ErrDeviceNotMatched
}

func (p *Probe) quanta(ctx context.Context, log logr.Logger) (bmcConnection interface{}, err error) {
	ctx, cancel := context.WithTimeout(ctx, time.Duration(time.Second*60))
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("https://%s/page/login.html", p.host), nil)
	if err != nil {
		return bmcConnection, err
	}
	resp, err := p.client.Do(req)
	if err != nil {
		return bmcConnection, err
	}

	defer resp.Body.Close()
	defer io.Copy(ioutil.Discard, resp.Body) // nolint

	payload, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return bmcConnection, err
	}

	// Ensure the response we got includes a PNG.
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
