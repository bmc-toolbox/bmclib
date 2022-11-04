package supermicro

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/bmc-toolbox/bmclib/v2/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/v2/providers"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
	"github.com/pkg/errors"

	bmcerrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

const (
	// ProviderName for the provider implementation
	ProviderName = "supermicro"
	// ProviderProtocol for the provider implementation
	ProviderProtocol = "vendorapi"
)

var (
	// Features implemented by asrockrack https
	Features = registrar.Features{
		providers.FeatureInventoryRead,
		providers.FeatureBmcReset,
		providers.FeatureUserCreate,
		providers.FeatureUserUpdate,
	}
)

type Supermicro struct {
	ip                   string
	username             string
	password             string
	csrfToken            string
	sid                  *http.Cookie
	httpClient           *http.Client
	ctx                  context.Context
	log                  logr.Logger
	httpClientSetupFuncs []func(*http.Client)
}

// SupermicroOption is a type that can configure a *Supermicro
type SupermicroOption func(*Supermicro)

// WithSecureTLS enforces trusted TLS connections, with an optional CA certificate pool.
// Using this option with an nil pool uses the system CAs.
func WithSecureTLS(rootCAs *x509.CertPool) SupermicroOption {
	return func(i *Supermicro) {
		i.httpClientSetupFuncs = append(i.httpClientSetupFuncs, httpclient.SecureTLSOption(rootCAs))
	}
}

// New returns a new Supermicro instance ready to be used
func New(ctx context.Context, ip string, username string, password string, log logr.Logger) (sm *Supermicro, err error) {
	return NewWithOptions(ctx, ip, username, password, log)
}

// NewWithOptions returns a new Supermicro with options ready to be used
func NewWithOptions(ctx context.Context, ip string, username string, password string, log logr.Logger, opts ...SupermicroOption) (*Supermicro, error) {
	sm := &Supermicro{
		ip:       ip,
		username: username,
		password: password,
		ctx:      ctx,
		log:      log,
	}
	for _, opt := range opts {
		opt(sm)
	}
	return sm, nil
}

// Open initiates the connection to an Supermicro device
func (s *Supermicro) Open(ctx context.Context) (err error) {
	// login
	endpoint := "/cgi/login.cgi"

	userEncoded := base64.StdEncoding.EncodeToString([]byte(s.username))
	passEncoded := base64.StdEncoding.EncodeToString([]byte(s.password))

	payload := bytes.NewReader([]byte(fmt.Sprintf("name=%s&pwd=%s&check=00", userEncoded, passEncoded)))
	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}

	body, statusCode, err := s.queryHTTPS(ctx, endpoint, http.MethodPost, payload, headers)
	if err != nil {
		return errors.Wrap(bmcerrs.ErrLoginFailed, err.Error())
	}

	if statusCode == 404 || !strings.Contains(string(body), "../cgi/url_redirect.cgi?url_name=mainmenu") {
		return errors.Wrap(bmcerrs.ErrLoginFailed, "status code: "+strconv.Itoa(statusCode))
	}

	// query for the csrf token
	s.csrfToken, err = s.queryCSRFToken(ctx)
	if err != nil {
		return errors.Wrap(bmcerrs.ErrLoginFailed, err.Error())
	}

	return nil
}

// queryCSRFToken queries for the page containing the csrf token and returns the token.
func (s *Supermicro) queryCSRFToken(ctx context.Context) (token string, err error) {
	endpoint := "/cgi/url_redirect.cgi?url_name=topmenu"

	body, statusCode, err := s.queryHTTPS(ctx, endpoint, http.MethodGet, nil, nil)
	if err != nil {
		return "", errors.Wrap(bmcerrs.ErrLoginFailed, err.Error())
	}

	if statusCode != 200 {
		return "", errors.Wrap(bmcerrs.ErrNon200Response, "statusCode: "+strconv.Itoa(statusCode))
	}

	// The CSRF token is found at the bottom of the page, in the format..
	//
	// CSRF_TOKEN", "UnmFMu+KtlARoAkXSEVtKrcy4J41ygbH6uDq+hqDVQQ");</script></body>
	// </html>
	idx := bytes.Index(body, []byte(`CSRF_TOKEN`))
	if idx == -1 {

		return "", fmt.Errorf("failed to locate CSRF_TOKEN value")
	}

	// split out token from body
	bytesWithToken := body[idx:]
	// strip out trailing html tags
	strWithToken := strings.Trim(string(bytesWithToken), ");</script></body>\n</html>")
	// split on comma and space
	strWithToken = strings.Split(strWithToken, ", ")[1]
	// trim out quotes and return
	return strings.Trim(strWithToken, "\""), nil
}

// Close closes the connection properly
func (s *Supermicro) Close(ctx context.Context) (err error) {
	endpoint := "/cgi/logout.cgi"
	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}

	_, _, err = s.queryHTTPS(ctx, endpoint, http.MethodPost, nil, headers)
	if err != nil {
		return errors.Wrap(bmcerrs.ErrLogoutFailed, err.Error())
	}

	return nil
}

// queryCGI queries the ipmi.cgi endpoint with the given command.
func (s *Supermicro) queryCGI(ctx context.Context, cmd string) (ipmi *IPMI, err error) {
	endpoint := "/cgi/ipmi.cgi"
	headers := map[string]string{
		"Referer": "https://" + s.ip,
	}

	payload := bytes.NewReader([]byte(cmd))

	body, statusCode, err := s.queryHTTPS(ctx, endpoint, http.MethodPost, payload, headers)
	if err != nil {
		return nil, err
	}

	if statusCode != 200 {
		return nil, errors.Wrap(bmcerrs.ErrNon200Response, "got: "+strconv.Itoa(statusCode))
	}

	return s.unmarshalXML(body)
}

// unmarshalXML unpacks given XML and returns the IPMI struct
func (s *Supermicro) unmarshalXML(body []byte) (*IPMI, error) {

	fmt.Println(string(body))
	ipmi := &IPMI{}
	err := xml.Unmarshal(body, ipmi)
	if err != nil {
		return nil, err
	}

	return ipmi, nil
}

// queryHTTPS runs the https query and returns the response body with the status code
func (s *Supermicro) queryHTTPS(ctx context.Context, endpoint, method string, payload io.Reader, headers map[string]string) (response []byte, statusCode int, err error) {
	bmcURL := fmt.Sprintf("https://%s"+endpoint, s.ip)

	if s.httpClient == nil {
		s.httpClient, err = httpclient.Build(s.httpClientSetupFuncs...)
		if err != nil {
			return nil, 0, err
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, bmcURL, payload)
	if err != nil {
		return nil, 0, err
	}

	if len(headers) > 0 {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}

	s.log.V(2).Info("Querying BMC", "URL", bmcURL, "vendor", VendorID, "ip", s.ip)

	u, err := url.Parse(bmcURL)
	if err != nil {
		return nil, 0, err
	}

	for _, cookie := range s.httpClient.Jar.Cookies(u) {
		if cookie.Name == "SID" && cookie.Value != "" {
			req.AddCookie(cookie)
		}
	}

	if s.csrfToken != "" {
		req.Header.Add("CSRF_TOKEN", s.csrfToken)
	}

	// debug dump request
	if os.Getenv("BMCLIB_LOG_LEVEL") == "trace" {
		reqDump, _ := httputil.DumpRequestOut(req, true)
		s.log.V(3).Info("trace", "url", bmcURL, "requestDump", string(reqDump))
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	// grab session cookie when on login
	if strings.HasSuffix(bmcURL, "login.cgi") {
		for _, cookie := range resp.Cookies() {
			if cookie.Name == "SID" && cookie.Value != "" {
				s.sid = cookie
			} else {
				s.sid = &http.Cookie{}
			}
		}
	}

	// debug dump response
	if os.Getenv("BMCLIB_LOG_LEVEL") == "trace" {
		respDump, _ := httputil.DumpResponse(resp, true)
		s.log.V(3).Info("trace", "url", bmcURL, "responseDump", string(respDump))
	}

	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return responseBody, resp.StatusCode, nil
}
