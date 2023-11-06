package supermicro

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bmc-toolbox/bmclib/v2/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/v2/internal/redfishwrapper"
	"github.com/bmc-toolbox/bmclib/v2/providers"

	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
	"github.com/pkg/errors"

	bmclibconsts "github.com/bmc-toolbox/bmclib/v2/constants"
	bmclibErrs "github.com/bmc-toolbox/bmclib/v2/errors"
)

const (
	// ProviderName for the provider Supermicro implementation
	ProviderName = "supermicro"
	// ProviderProtocol for the provider supermicro implementation
	ProviderProtocol = "vendorapi"
)

var (
	// Features implemented
	Features = registrar.Features{
		providers.FeatureScreenshot,
		providers.FeatureMountFloppyImage,
		providers.FeatureUnmountFloppyImage,
		providers.FeatureFirmwareUpload,
		providers.FeatureFirmwareInstallUploaded,
		providers.FeatureFirmwareTaskStatus,
	}
)

// supports
//
// product: SYS-5019C-MR, baseboard part number: X11SCM-F
//   - screen capture
//   - bios firmware install
//   - bmc firmware install
//
// product: SYS-510T-MR, baseboard part number: X12STH-SYS, X12SPO-NTF
//   - screen capture
//   - floppy image mount
// product: 6029P-E1CR12L, baseboard part number: X11DPH-T
// . - screen capture
//   - bios firmware install
//   - bmc firmware install
//   - floppy image mount

type Config struct {
	HttpClient           *http.Client
	Port                 string
	httpClientSetupFuncs []func(*http.Client)
}

// Option for setting optional Client values
type Option func(*Config)

func WithHttpClient(httpClient *http.Client) Option {
	return func(c *Config) {
		c.HttpClient = httpClient
	}
}

// WithSecureTLS returns an option that enables secure TLS with an optional cert pool.
func WithSecureTLS(rootCAs *x509.CertPool) Option {
	return func(c *Config) {
		c.httpClientSetupFuncs = append(c.httpClientSetupFuncs, httpclient.SecureTLSOption(rootCAs))
	}
}

func WithPort(port string) Option {
	return func(c *Config) {
		c.Port = port
	}
}

// Connection details
type Client struct {
	client    *http.Client
	redfish   *redfishwrapper.Client
	host      string
	user      string
	pass      string
	port      string
	csrfToken string
	model     string
	redfish   *redfishwrapper.Client
	log       logr.Logger
	_         [32]byte
}

// New returns connection with a Supermicro client initialized
func NewClient(host, user, pass string, log logr.Logger, opts ...Option) *Client {
	if !strings.HasPrefix(host, "https://") && !strings.HasPrefix(host, "http://") {
		host = "https://" + host
	}

	defaultConfig := &Config{
		Port: "443",
	}

	for _, opt := range opts {
		opt(defaultConfig)
	}

	return &Client{
		host:   host,
		user:   user,
		pass:   pass,
		port:   defaultConfig.Port,
		client: httpclient.Build(defaultConfig.httpClientSetupFuncs...),
		log:    log,
	}
}

func (c *Client) x12() *x12 { return &x12{c} }

func (c *Client) x11() *x11 { return &x11{c} }

func (c *Client) redfishSession(ctx context.Context) (err error) {
	c.redfish = redfishwrapper.NewClient(c.host, "", c.user, c.pass, redfishwrapper.WithHTTPClient(c.client))
	if err := c.redfish.Open(ctx); err != nil {
		return err
	}

	return nil
}

// Open a connection to a Supermicro BMC using the vendor API.
func (c *Client) Open(ctx context.Context) (err error) {
	data := fmt.Sprintf(
		"name=%s&pwd=%s&check=00",
		base64.StdEncoding.EncodeToString([]byte(c.user)),
		base64.StdEncoding.EncodeToString([]byte(c.pass)),
	)

	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}

	body, status, err := c.query(ctx, "cgi/login.cgi", http.MethodPost, bytes.NewBufferString(data), headers, 0)
	if err != nil {
		return errors.Wrap(bmclibErrs.ErrLoginFailed, err.Error())
	}

	if status != 200 {
		return errors.Wrap(bmclibErrs.ErrLoginFailed, strconv.Itoa(status))
	}

	if !bytes.Contains(body, []byte(`url_redirect.cgi?url_name=mainmenu`)) &&
		!bytes.Contains(body, []byte(`url_redirect.cgi?url_name=topmenu`)) {
		return errors.Wrap(bmclibErrs.ErrLoginFailed, "unexpected response contents")
	}

	contentsTopMenu, status, err := c.query(ctx, "cgi/url_redirect.cgi?url_name=topmenu", http.MethodGet, nil, nil, 0)
	if err != nil {
		return errors.Wrap(bmclibErrs.ErrLoginFailed, err.Error())
	}

	if status != 200 {
		return errors.Wrap(bmclibErrs.ErrLoginFailed, strconv.Itoa(status))
	}

	token := parseToken(contentsTopMenu)
	if token == "" {
		return errors.Wrap(bmclibErrs.ErrLoginFailed, "could not parse CSRF-TOKEN from page")
	}

	c.csrfToken = token

	c.model, err = c.deviceModel(ctx)
	if err != nil {
		return errors.Wrap(bmclibErrs.ErrLoginFailed, err.Error())
	}

	return nil
}

func (c *Client) openRedfish(ctx context.Context) error {
	if c.redfish != nil && c.redfish.SessionActive() == nil {
		return nil
	}

	rfclient := redfishwrapper.NewClient(c.host, "", c.user, c.pass)
	if err := rfclient.Open(ctx); err != nil {
		return err
	}

	c.redfish = rfclient

	return nil
}

func (c *Client) closeRedfish(ctx context.Context) {
	if c.redfish != nil {
		// error not checked on purpose
		_ = c.redfish.Close(ctx)

		c.redfish = nil
	}
}

func parseToken(body []byte) string {
	var key string
	if bytes.Contains(body, []byte(`CSRF-TOKEN`)) {
		key = "CSRF-TOKEN"
	}

	if bytes.Contains(body, []byte(`CSRF_TOKEN`)) {
		key = "CSRF_TOKEN"
	}

	if key == "" {
		return ""
	}

	re, err := regexp.Compile(`"CSRF_TOKEN", "(?P<token>.*)"`)
	if err != nil {
		return ""
	}

	found := re.FindSubmatch(body)
	if len(found) == 0 {
		return ""
	}

	return string(found[1])
}

// Close a connection to a Supermicro BMC using the vendor API.
func (c *Client) Close(ctx context.Context) error {
	if c.client == nil {
		return nil
	}

	_, status, err := c.query(ctx, "cgi/logout.cgi", http.MethodGet, nil, nil, 0)
	if err != nil {
		return errors.Wrap(bmclibErrs.ErrLogoutFailed, err.Error())
	}

	if status != 200 {
		return errors.Wrap(bmclibErrs.ErrLogoutFailed, strconv.Itoa(status))
	}

	if c.redfish != nil {
		err = c.redfish.Close(ctx)
		if err != nil {
			return errors.Wrap(bmclibErrs.ErrLogoutFailed, err.Error())
		}
	}


	return nil
}

// Name returns the client provider name.
func (c *Client) Name() string {
	return ProviderName
}

func (c *Client) Screenshot(ctx context.Context) (image []byte, fileType string, err error) {
	fileType = "jpg"

	// request screen preview to be saved
	if err := c.initScreenPreview(ctx); err != nil {
		return nil, "", err
	}

	// give the bmc a few seconds to store the screen preview
	time.Sleep(2 * time.Second)

	// retrieve screen preview
	image, errFetch := c.fetchScreenPreview(ctx)
	if errFetch != nil {
		return nil, "", err
	}

	return image, fileType, nil
}

func (c *Client) fetchScreenPreview(ctx context.Context) ([]byte, error) {
	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}

	endpoint := "cgi/url_redirect.cgi?url_name=Snapshot&url_type=img"
	body, status, err := c.query(ctx, endpoint, http.MethodGet, nil, headers, 0)
	if err != nil {
		return nil, errors.Wrap(bmclibErrs.ErrScreenshot, strconv.Itoa(status))
	}

	if status != 200 {
		return nil, errors.Wrap(bmclibErrs.ErrScreenshot, strconv.Itoa(status))
	}

	return body, nil
}

func (c *Client) initScreenPreview(ctx context.Context) error {
	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}

	data := "op=sys_preview&_="

	body, status, err := c.query(ctx, "cgi/op.cgi", http.MethodPost, bytes.NewBufferString(data), headers, 0)
	if err != nil {
		return errors.Wrap(bmclibErrs.ErrScreenshot, err.Error())
	}

	if status != 200 {
		return errors.Wrap(bmclibErrs.ErrScreenshot, strconv.Itoa(status))
	}

	if !bytes.Contains(body, []byte(`<IPMI>`)) {
		return errors.Wrap(bmclibErrs.ErrScreenshot, "unexpected response: "+string(body))
	}

	return nil
}

// PowerSet sets the power state of a server
func (c *Client) PowerSet(ctx context.Context, state string) (ok bool, err error) {
	switch strings.ToLower(state) {
	case "cycle":
		return c.powerCycle(ctx)
	default:
		return false, errors.New("action not implemented for provider")
	}
}

func (c *Client) fruInfo(ctx context.Context) (*FruInfo, error) {
	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded; charset=UTF-8"}

	payload := "op=FRU_INFO.XML&r=(0,0)&_="

	body, status, err := c.query(ctx, "cgi/ipmi.cgi", http.MethodPost, bytes.NewBufferString(payload), headers, 0)
	if err != nil {
		return nil, errors.Wrap(ErrQueryFRUInfo, err.Error())
	}

	if status != 200 {
		return nil, unexpectedResponseErr([]byte(payload), body, status)
	}

	if !bytes.Contains(body, []byte(`<IPMI>`)) {
		return nil, unexpectedResponseErr([]byte(payload), body, status)
	}

	data := &IPMI{}
	if err := xml.Unmarshal(body, data); err != nil {
		return nil, errors.Wrap(ErrQueryFRUInfo, err.Error())
	}

	return data.FruInfo, nil
}

// powerCycle using SMC XML API
//
// This method is only here for the case when firmware updates are being applied using this provider.
func (c *Client) powerCycle(ctx context.Context) (bool, error) {
	payload := []byte(`op=SET_POWER_INFO.XML&r=(1,3)&_=`)

	headers := map[string]string{
		"Content-type": "application/x-www-form-urlencoded; charset=UTF-8",
	}

	body, status, err := c.query(ctx, "cgi/ipmi.cgi", http.MethodPost, bytes.NewBuffer(payload), headers, 0)
	if err != nil {
		return false, err
	}

	if status != http.StatusOK {
		return false, unexpectedResponseErr(payload, body, status)
	}

	return true, nil
}

func (c *Client) deviceModel(ctx context.Context) (string, error) {
	deviceModel, err := c.x11().deviceModel(ctx)
	if err != nil {
		if errors.Is(err, ErrXMLAPIUnsupported) {
			var errRF error

			deviceModel, errRF = c.x12().deviceModel(ctx)
			if errRF != nil {
				return "", errors.Wrap(errRF, "XML API not supported and failed to identify model over Redfish")
			}
		}
	}

	c.model = deviceModel

	return deviceModel, nil
}

func (c *Client) query(ctx context.Context, endpoint, method string, payload io.Reader, headers map[string]string, contentLength int64) ([]byte, int, error) {
	var body []byte
	var err error
	var req *http.Request

	host := c.host

	if c.port != "" {
		host = c.host + ":" + c.port
	}

	hostEndpoint := fmt.Sprintf("%s/%s", host, endpoint)

	req, err = http.NewRequestWithContext(ctx, method, hostEndpoint, payload)
	if err != nil {
		return nil, 0, err
	}

	if c.csrfToken != "" {
		req.Header.Add("Csrf-Token", c.csrfToken)
		// because old firmware
		req.Header.Add("CSRF_TOKEN", c.csrfToken)
	}

	// required on  X11SCM-F with 1.23.06 and older BMC firmware
	// https://go.googlesource.com/go/+/go1.20/src/net/http/request.go#124
	req.Host, err = hostIP(c.host)
	if err != nil {
		return nil, 0, err
	}

	// required on  X11SCM-F with 1.23.06 and older BMC firmware
	req.Header.Add("Referer", c.host)

	for k, v := range headers {
		req.Header.Add(k, v)
	}

	// Content-Length headers are ignored, unless defined in this manner
	// https://go.googlesource.com/go/+/go1.20/src/net/http/request.go#165
	// https://go.googlesource.com/go/+/go1.20/src/net/http/request.go#91
	if contentLength > 0 {
		req.ContentLength = contentLength
	}

	endpointURL, err := url.Parse(hostEndpoint)
	if err != nil {
		return nil, 0, err
	}

	// include session cookie
	for _, cookie := range c.client.Jar.Cookies(endpointURL) {
		if cookie.Name == "SID" && cookie.Value != "" {
			req.AddCookie(cookie)
		}

	}

	var reqDump []byte

	if os.Getenv(bmclibconsts.EnvEnableDebug) == "true" {
		reqDump, _ = httputil.DumpRequestOut(req, true)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return body, 0, err
	}

	// cookies are visible after the request has been made, so we dump the request and cookies here
	// https://github.com/golang/go/issues/22745
	if os.Getenv(bmclibconsts.EnvEnableDebug) == "true" {
		fmt.Println(string(reqDump))

		for _, v := range req.Cookies() {
			header := "Cookie: " + v.String() + "\r"
			fmt.Println(header)
		}
	}

	// debug dump response
	if os.Getenv(bmclibconsts.EnvEnableDebug) == "true" {
		respDump, _ := httputil.DumpResponse(resp, true)

		fmt.Println(string(respDump))
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		return body, 0, err
	}

	defer resp.Body.Close()

	return body, resp.StatusCode, nil
}

func hostIP(hostURL string) (string, error) {
	hostURLParsed, err := url.Parse(hostURL)
	if err != nil {
		return "", err
	}

	if strings.Contains(hostURLParsed.Host, ":") {
		return strings.Split(hostURLParsed.Host, ":")[0], nil

	}

	return hostURLParsed.Host, nil
}
