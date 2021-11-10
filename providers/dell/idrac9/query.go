package idrac9

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/bmc-toolbox/bmclib/internal/helper"
)

// CurrentHTTPSCert returns the current x509 certficates configured on the BMC
// The bool value returned indicates if the BMC supports CSR generation.
// CurrentHTTPSCert implements the Configure interface.
func (i *IDrac9) CurrentHTTPSCert() ([]*x509.Certificate, bool, error) {
	dialer := &net.Dialer{
		Timeout: time.Duration(10) * time.Second,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", i.ip+":"+"443", &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return []*x509.Certificate{{}}, true, err
	}

	defer conn.Close()

	return conn.ConnectionState().PeerCertificates, true, nil
}

// Screenshot grab screen preview.
func (i *IDrac9) Screenshot() (response []byte, extension string, err error) {
	err = i.httpLogin()
	if err != nil {
		return nil, "", err
	}

	extension = "png"
	url := "sysmgmt/2015/server/preview"
	statusCode, _, err := i.get(url, &map[string]string{})
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a non-200 status code from the GET request to %s.", url)
		}

		return nil, "", err
	}

	url = "capconsole/scapture0.png"
	statusCode, response, err = i.get(url, &map[string]string{})
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a non-200 status code from the GET request to %s.", url)
		}

		return nil, "", err
	}

	return response, extension, nil
}

func (i *IDrac9) queryUsers() (users map[int]User, err error) {
	url := "sysmgmt/2012/server/configgroup/iDRAC.Users"

	statusCode, response, err := i.get(url, &map[string]string{})
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a non-200 status code from the GET request to %s.", url)
		}

		i.log.V(1).Error(err, "queryUsers(): GET request failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"endpoint", url,
			"step", helper.WhosCalling(),
		)
		return users, err
	}

	userData := make(idracUsers)
	err = json.Unmarshal(response, &userData)
	if err != nil {
		i.log.V(1).Error(err, "Unable to unmarshal payload.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"resource", "User",
			"step", "queryUserInfo",
		)
		return users, err
	}

	return userData["iDRAC.Users"], err
}
