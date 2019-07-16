package ilo

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net"
	"time"

	"github.com/bmc-toolbox/bmclib/errors"
	"github.com/bmc-toolbox/bmclib/internal/helper"

	log "github.com/sirupsen/logrus"
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
	err = i.httpLogin()
	if err != nil {
		return response, extension, err
	}

	endpoint := "images/thumbnail.bmp"
	extension = "bmp"

	// screen thumbnails are only available in ilo5.
	if i.HardwareType() != "ilo5" {
		return response, extension, errors.ErrFeatureUnavailable
	}

	response, err = i.get(endpoint)
	if err != nil {
		return []byte{}, extension, err
	}

	return response, extension, err
}

func (i *Ilo) queryDirectoryGroups() (directoryGroups []DirectoryGroups, err error) {

	endpoint := "json/directory_groups"

	payload, err := i.get(endpoint)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":       i.ip,
			"Model":    i.HardwareType(),
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"Error":    err,
		}).Warn("GET request failed.")
		return directoryGroups, err
	}

	var directoryGroupAccts DirectoryGroupAccts
	//fmt.Printf("--> %+v\n", userinfo["users"])
	err = json.Unmarshal(payload, &directoryGroupAccts)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":    i.ip,
			"step":  helper.WhosCalling(),
			"Model": i.HardwareType(),
			"Error": err,
		}).Warn("Unable to unmarshal payload.")
		return directoryGroups, err
	}

	return directoryGroupAccts.Groups, err
}

func (i *Ilo) queryUsers() (usersInfo []UserInfo, err error) {

	endpoint := "json/user_info"

	payload, err := i.get(endpoint)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":       i.ip,
			"Model":    i.HardwareType(),
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"Error":    err,
		}).Warn("GET request failed.")
		return usersInfo, err
	}

	var users Users
	//fmt.Printf("--> %+v\n", userinfo["users"])
	err = json.Unmarshal(payload, &users)
	if err != nil {
		log.WithFields(log.Fields{
			"step":     "queryUserInfo",
			"resource": "User",
			"IP":       i.ip,
			"Model":    i.HardwareType(),
			"Error":    err,
		}).Warn("Unable to unmarshal payload.")
		return usersInfo, err
	}

	return users.UsersInfo, err
}

func (i *Ilo) queryNetworkSntp() (networkSntp NetworkSntp, err error) {

	endpoint := "json/network_sntp/interface/0"

	payload, err := i.get(endpoint)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":       i.ip,
			"Model":    i.HardwareType(),
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"Error":    err,
		}).Warn("GET request failed.")
		return networkSntp, err
	}

	err = json.Unmarshal(payload, &networkSntp)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":    i.ip,
			"step":  helper.WhosCalling(),
			"Model": i.HardwareType(),
			"Error": err,
		}).Warn("Unable to unmarshal payload.")
		return networkSntp, err
	}

	return networkSntp, err
}

func (i *Ilo) queryAccessSettings() (AccessSettings, error) {

	endpoint := "json/access_settings"

	var accessSettings AccessSettings

	payload, err := i.get(endpoint)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":       i.ip,
			"Model":    i.HardwareType(),
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"Error":    err,
		}).Warn("GET request failed.")
		return accessSettings, err
	}

	err = json.Unmarshal(payload, &accessSettings)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":    i.ip,
			"step":  helper.WhosCalling(),
			"Model": i.HardwareType(),
			"Error": err,
		}).Warn("Unable to unmarshal payload.")
		return accessSettings, err
	}

	return accessSettings, err
}

func (i *Ilo) queryNetworkIPv4() (NetworkIPv4, error) {

	endpoint := "json/network_ipv4/interface/0"

	var networkIPv4 NetworkIPv4

	payload, err := i.get(endpoint)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":       i.ip,
			"Model":    i.HardwareType(),
			"endpoint": endpoint,
			"step":     helper.WhosCalling(),
			"Error":    err,
		}).Warn("GET request failed.")
		return networkIPv4, err
	}

	err = json.Unmarshal(payload, &networkIPv4)
	if err != nil {
		log.WithFields(log.Fields{
			"IP":    i.ip,
			"step":  helper.WhosCalling(),
			"Model": i.HardwareType(),
			"Error": err,
		}).Warn("Unable to unmarshal payload.")
		return networkIPv4, err
	}

	return networkIPv4, err
}
