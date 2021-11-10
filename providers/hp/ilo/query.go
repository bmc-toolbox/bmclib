package ilo

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/helper"
)

// CurrentHTTPSCert returns the current x509 certficates configured on the BMC
// The bool value returned indicates if the BMC supports CSR generation.
// CurrentHTTPSCert implements the Configure interface
func (i *Ilo) CurrentHTTPSCert() ([]*x509.Certificate, bool, error) {
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

// Screenshot returns a thumbnail of video display from the bmc.
func (i *Ilo) Screenshot() (response []byte, extension string, err error) {
	// Screen thumbnails are only available in ILO5.
	if i.HardwareType() != "ilo5" {
		return nil, "", errors.ErrFeatureUnavailable
	}

	err = i.httpLogin()
	if err != nil {
		return nil, "", err
	}

	endpoint := "images/thumbnail.bmp"
	statusCode, response, err := i.get(endpoint, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		return nil, "", err
	}

	return response, "bmp", nil
}

func (i *Ilo) queryDirectoryGroups() ([]DirectoryGroups, error) {
	endpoint := "json/directory_groups"
	statusCode, payload, err := i.get(endpoint, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		i.log.V(1).Error(err, "queryDirectoryGroups(): GET request failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"endpoint", endpoint,
			"step", helper.WhosCalling(),
		)
		return nil, err
	}

	var directoryGroupAccts DirectoryGroupAccts
	err = json.Unmarshal(payload, &directoryGroupAccts)
	if err != nil {
		msg := "queryDirectoryGroups(): Unable to unmarshal payload."
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return nil, err
	}

	return directoryGroupAccts.Groups, nil
}

func (i *Ilo) queryUsers() ([]UserInfo, error) {
	endpoint := "json/user_info"
	statusCode, payload, err := i.get(endpoint, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}
		i.log.V(1).Error(err, "queryUsers(): GET request failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"endpoint", endpoint,
			"step", helper.WhosCalling(),
		)
		return nil, err
	}

	var users Users
	err = json.Unmarshal(payload, &users)
	if err != nil {
		msg := "Unable to unmarshal payload."
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"resource", "User",
			"step", "queryUserInfo",
		)
		return nil, err
	}

	return users.UsersInfo, nil
}

func (i *Ilo) queryNetworkSntp() (networkSntp NetworkSntp, err error) {
	endpoint := "json/network_sntp/interface/0"
	statusCode, payload, err := i.get(endpoint, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		i.log.V(1).Error(err, "queryNetworkSntp(): GET request failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"endpoint", endpoint,
			"step", helper.WhosCalling(),
		)
		return networkSntp, err
	}

	err = json.Unmarshal(payload, &networkSntp)
	if err != nil {
		msg := "queryNetworkSntp(): Unable to unmarshal payload."
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return networkSntp, err
	}

	return networkSntp, nil
}

func (i *Ilo) queryAccessSettings() (accessSettings AccessSettings, err error) {
	endpoint := "json/access_settings"
	statusCode, payload, err := i.get(endpoint, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		i.log.V(1).Error(err, "queryAccessSettings(): GET request failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"endpoint", endpoint,
			"step", helper.WhosCalling(),
		)
		return accessSettings, err
	}

	err = json.Unmarshal(payload, &accessSettings)
	if err != nil {
		msg := "queryAccessSettings(): Unable to unmarshal payload."
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return accessSettings, err
	}

	return accessSettings, nil
}

func (i *Ilo) queryNetworkIPv4() (networkIPv4 NetworkIPv4, err error) {
	endpoint := "json/network_ipv4/interface/0"
	statusCode, payload, err := i.get(endpoint, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		i.log.V(1).Error(err, "queryNetworkIPv4(): GET request failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"endpoint", endpoint,
			"step", helper.WhosCalling(),
		)
		return networkIPv4, err
	}

	err = json.Unmarshal(payload, &networkIPv4)
	if err != nil {
		i.log.V(1).Error(err, "queryNetworkIPv4(): Unable to unmarshal payload.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return networkIPv4, err
	}

	return networkIPv4, nil
}

func (i *Ilo) queryPowerRegulator() (PowerRegulator, error) {
	endpoint := "json/power_regulator"

	var powerRegulator PowerRegulator

	statusCode, payload, err := i.get(endpoint, true)
	if err != nil || statusCode != 200 {
		if err == nil {
			err = fmt.Errorf("Received a %d status code from the GET request to %s.", statusCode, endpoint)
		}

		i.log.V(1).Error(err, "queryPowerRegulator(): GET request failed.",
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"endpoint", endpoint,
			"step", helper.WhosCalling(),
		)
		return PowerRegulator{}, err
	}

	err = json.Unmarshal(payload, &powerRegulator)
	if err != nil {
		msg := "queryPowerRegulator(): Unable to unmarshal payload."
		i.log.V(1).Error(err, msg,
			"IP", i.ip,
			"HardwareType", i.HardwareType(),
			"step", helper.WhosCalling(),
		)
		return PowerRegulator{}, err
	}

	return powerRegulator, err
}
