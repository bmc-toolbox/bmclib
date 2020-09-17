package supermicrox11

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"time"

	"github.com/google/go-querystring/query"

	"github.com/bmc-toolbox/bmclib/errors"
)

// CurrentHTTPSCert returns the current x509 certficates configured on the BMC
// the bool value returned is set to true if the BMC support CSR generation.
// CurrentHTTPSCert implements the Configure interface.
func (s *SupermicroX) CurrentHTTPSCert() ([]*x509.Certificate, bool, error) {

	dialer := &net.Dialer{
		Timeout: time.Duration(10) * time.Second,
	}

	conn, err := tls.DialWithDialer(dialer, "tcp", s.ip+":"+"443", &tls.Config{InsecureSkipVerify: true})

	if err != nil {
		return []*x509.Certificate{{}}, false, err
	}

	defer conn.Close()

	return conn.ConnectionState().PeerCertificates, false, nil

}

// Screenshot returns a thumbnail of video display from the bmc.
// 1. request capture preview.
// 2. sleep for 3 seconds to give ikvm time to ensure preview was captured
// 3. request for preview.
func (s *SupermicroX) Screenshot() (response []byte, extension string, err error) {

	postEndpoint := "CapturePreview.cgi"
	getEndpoint := "cgi/url_redirect.cgi?"

	extension = "bmp"

	// allow thumbnails only for supermicro x10s.
	if s.HardwareType() != BmcType {
		return response, extension, errors.ErrFeatureUnavailable
	}

	tzLocation, err := time.LoadLocation("CET")
	t := time.Now().In(tzLocation)

	//Fri Jun 06 2018 14:28:25 GMT+0100 (CET)
	ts := fmt.Sprintf("%s %d %d:%d:%d %s (%s)",
		t.Format("Fri Jun 01"),
		t.Year(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		t.Format("GMT+0200"),
		tzLocation)

	capturePreview := CapturePreview{
		IkvmPreview: "(0,0)",
		TimeStamp:   ts,
	}

	form, _ := query.Values(capturePreview)
	_, statusCode, err := s.post(postEndpoint, &form, []byte{}, "")
	if err != nil {
		return response, extension, err
	}

	if statusCode != 200 {
		return response, extension, fmt.Errorf("Non 200 response from endpoint")
	}

	time.Sleep(3 * time.Second)
	//Fri Jun 06 2018 14:28:25 GMT+0100 (CET)
	ts = fmt.Sprintf("%s %d %d:%d:%d %s (%s)",
		t.Format("Fri Jun 01"),
		t.Year(),
		t.Hour(),
		t.Minute(),
		t.Second(),
		t.Format("GMT+0200"),
		tzLocation)

	urlRedirect := URLRedirect{
		URLName:   "Snapshot",
		URLType:   "img",
		TimeStamp: ts,
	}

	queryString, _ := query.Values(urlRedirect)
	getEndpoint += queryString.Encode()

	response, err = s.get(getEndpoint)
	if err != nil {
		return []byte{}, extension, err
	}

	return response, extension, err
}
