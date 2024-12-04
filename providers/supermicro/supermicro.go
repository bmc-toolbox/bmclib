package supermicro

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
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

	common "github.com/metal-toolbox/bmc-common"
	"github.com/metal-toolbox/bmclib/constants"
	"github.com/metal-toolbox/bmclib/internal/httpclient"
	"github.com/metal-toolbox/bmclib/internal/redfishwrapper"
	"github.com/metal-toolbox/bmclib/internal/sum"
	"github.com/metal-toolbox/bmclib/providers"
	"github.com/stmcginnis/gofish/redfish"

	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
	"github.com/pkg/errors"

	bmclibconsts "github.com/metal-toolbox/bmclib/constants"
	bmclibErrs "github.com/metal-toolbox/bmclib/errors"
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
		providers.FeatureFirmwareInstallSteps,
		providers.FeatureInventoryRead,
		providers.FeaturePowerSet,
		providers.FeaturePowerState,
		providers.FeatureBmcReset,
		providers.FeatureGetBiosConfiguration,
		providers.FeatureSetBiosConfiguration,
		providers.FeatureSetBiosConfigurationFromFile,
		providers.FeatureResetBiosConfiguration,
		providers.FeatureBootProgress,
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
	serviceClient *serviceClient
	bmc           bmcQueryor
	log           logr.Logger
}

type bmcQueryor interface {
	firmwareInstallSteps(component string) ([]constants.FirmwareInstallStep, error)
	firmwareUpload(ctx context.Context, component string, file *os.File) (taskID string, err error)
	firmwareInstallUploaded(ctx context.Context, component, uploadTaskID string) (installTaskID string, err error)
	firmwareTaskStatus(ctx context.Context, component, taskID string) (state constants.TaskState, status string, err error)
	// query device model from the bmc
	queryDeviceModel(ctx context.Context) (model string, err error)
	// returns the device model, that was queried previously with queryDeviceModel
	deviceModel() (model string)
	supportsInstall(component string) error
	getBootProgress() (*redfish.BootProgress, error)
	bootComplete() (bool, error)
}

// New returns connection with a Supermicro client initialized
func NewClient(host, user, pass string, log logr.Logger, opts ...Option) *Client {
	defaultConfig := &Config{
		Port: "443",
	}

	for _, opt := range opts {
		opt(defaultConfig)
	}

	serviceClient, err := newBmcServiceClient(
		host,
		defaultConfig.Port,
		user,
		pass,
		httpclient.Build(defaultConfig.httpClientSetupFuncs...),
	)

	// We probably want to treat this as a fatal error and/or pass the error back to the caller
	// I did not want to chase that thread atm, so we intentionally return nil here if
	// newBmcServiceClient returns an error.
	if err != nil {
		return nil
	}

	return &Client{
		serviceClient: serviceClient,
		log:           log,
	}
}

// Open a connection to a Supermicro BMC using the vendor API.
func (c *Client) Open(ctx context.Context) (err error) {
	data := fmt.Sprintf(
		"name=%s&pwd=%s&check=00",
		base64.StdEncoding.EncodeToString([]byte(c.serviceClient.user)),
		base64.StdEncoding.EncodeToString([]byte(c.serviceClient.pass)),
	)

	headers := map[string]string{"Content-Type": "application/x-www-form-urlencoded"}

	body, status, err := c.serviceClient.query(ctx, "cgi/login.cgi", http.MethodPost, bytes.NewBufferString(data), headers, 0)
	if err != nil {
		return errors.Wrap(bmclibErrs.ErrLoginFailed, err.Error())
	}

	if status != 200 {
		return errors.Wrap(bmclibErrs.ErrLoginFailed, strconv.Itoa(status))
	}

	// called after a session was opened but further login dependencies failed
	closeWithError := func(ctx context.Context, err error) error {
		_ = c.Close(ctx)
		return err
	}

	if !bytes.Contains(body, []byte(`url_redirect.cgi?url_name=mainmenu`)) &&
		!bytes.Contains(body, []byte(`url_redirect.cgi?url_name=topmenu`)) {
		return closeWithError(ctx, errors.Wrap(bmclibErrs.ErrLoginFailed, "unexpected response contents"))
	}

	contentsTopMenu, status, err := c.serviceClient.query(ctx, "cgi/url_redirect.cgi?url_name=topmenu", http.MethodGet, nil, nil, 0)
	if err != nil {
		return closeWithError(ctx, errors.Wrap(bmclibErrs.ErrLoginFailed, err.Error()))
	}

	if status != 200 {
		return closeWithError(ctx, errors.Wrap(bmclibErrs.ErrLoginFailed, strconv.Itoa(status)))
	}

	// Note: older firmware version on the X11s don't use a CSRF token
	// so here theres no explicit requirement for it to be found.
	//
	// X11DPH-T 01.71.11 10/25/2019
	csrfToken := parseToken(contentsTopMenu)
	c.serviceClient.setCsrfToken(csrfToken)

	c.bmc, err = c.bmcQueryor(ctx)
	if err != nil {
		return closeWithError(ctx, errors.Wrap(bmclibErrs.ErrLoginFailed, err.Error()))
	}

	if err := c.serviceClient.redfishSession(ctx); err != nil {
		return closeWithError(ctx, errors.Wrap(bmclibErrs.ErrLoginFailed, err.Error()))
	}

	return nil
}

// PowerStateGet gets the power state of a BMC machine
func (c *Client) PowerStateGet(ctx context.Context) (state string, err error) {
	if c.serviceClient == nil || c.serviceClient.redfish == nil {
		return "", errors.Wrap(bmclibErrs.ErrLoginFailed, "client not initialized")
	}

	return c.serviceClient.redfish.SystemPowerStatus(ctx)
}

// PowerSet sets the power state of a server
func (c *Client) PowerSet(ctx context.Context, state string) (ok bool, err error) {
	if c.serviceClient == nil || c.serviceClient.redfish == nil {
		return false, errors.Wrap(bmclibErrs.ErrLoginFailed, "client not initialized")
	}

	return c.serviceClient.redfish.PowerSet(ctx, state)
}

// BmcReset power cycles the BMC
func (c *Client) BmcReset(ctx context.Context, resetType string) (ok bool, err error) {
	if c.serviceClient == nil || c.serviceClient.redfish == nil {
		return false, errors.Wrap(bmclibErrs.ErrLoginFailed, "client not initialized")
	}

	return c.serviceClient.redfish.BMCReset(ctx, resetType)
}

// Inventory collects hardware inventory and install firmware information
func (c *Client) Inventory(ctx context.Context) (device *common.Device, err error) {
	if c.serviceClient == nil || c.serviceClient.redfish == nil {
		return nil, errors.Wrap(bmclibErrs.ErrLoginFailed, "client not initialized")
	}

	return c.serviceClient.redfish.Inventory(ctx, false)
}

// GetBiosConfiguration return bios configuration
func (c *Client) GetBiosConfiguration(ctx context.Context) (biosConfig map[string]string, err error) {
	if c.serviceClient == nil || c.serviceClient.sum == nil {
		return nil, errors.Wrap(bmclibErrs.ErrLoginFailed, "client not initialized")
	}

	return c.serviceClient.sum.GetBiosConfiguration(ctx)
}

// SetBiosConfiguration set bios configuration
func (c *Client) SetBiosConfiguration(ctx context.Context, biosConfig map[string]string) (err error) {
	if c.serviceClient == nil || c.serviceClient.sum == nil {
		return errors.Wrap(bmclibErrs.ErrLoginFailed, "client not initialized")
	}

	return c.serviceClient.sum.SetBiosConfiguration(ctx, biosConfig)
}

// SetBiosConfigurationFromFile sets the bios configuration from a raw vendor config file
func (c *Client) SetBiosConfigurationFromFile(ctx context.Context, cfg string) (err error) {
	if c.serviceClient == nil || c.serviceClient.sum == nil {
		return errors.Wrap(bmclibErrs.ErrLoginFailed, "client not initialized")
	}

	return c.serviceClient.sum.SetBiosConfigurationFromFile(ctx, cfg)
}

// ResetBiosConfiguration sets the bios configuration back to "factory" defaults
func (c *Client) ResetBiosConfiguration(ctx context.Context) (err error) {
	if c.serviceClient == nil || c.serviceClient.sum == nil {
		return errors.Wrap(bmclibErrs.ErrLoginFailed, "client not initialized")
	}

	return c.serviceClient.sum.ResetBiosConfiguration(ctx)
}

func (c *Client) bmcQueryor(ctx context.Context) (bmcQueryor, error) {
	bmcModels := []struct {
		bmc         bmcQueryor
		modelFamily string
	}{
		{
			newX11Client(c.serviceClient, c.log),
			"x11",
		},
		{
			newX12Client(c.serviceClient, c.log),
			"x12",
		},
		{
			newX13Client(c.serviceClient, c.log),
			"x13",
		},
	}

	var model string
	for _, bmcModel := range bmcModels {
		var err error

		// Note to maintainers: x12 and x13 lacks support for the ipmi.cgi endpoint,
		// which will lead to our graceful handling of ErrXMLAPIUnsupported below.
		tempModel, err := bmcModel.bmc.queryDeviceModel(ctx)
		if err != nil {
			if errors.Is(err, ErrXMLAPIUnsupported) {
				continue
			}

			return nil, errors.Wrap(ErrModelUnknown, err.Error())
		}

		if strings.HasPrefix(strings.ToLower(tempModel), bmcModel.modelFamily) {
			return bmcModel.bmc, nil
		}

		// For returning more informative error bellow
		if tempModel != "" {
			model = tempModel
		}
	}

	return nil, errors.Wrapf(ErrModelUnknown, "failed to setup query client, had unsupported model: %s", model)
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

	re, err := regexp.Compile(fmt.Sprintf(`"%s", "(?P<token>.*)"`, key))
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
	if c.serviceClient.client == nil {
		return nil
	}

	_, status, err := c.serviceClient.query(ctx, "cgi/logout.cgi", http.MethodGet, nil, nil, 0)
	if err != nil {
		return errors.Wrap(bmclibErrs.ErrLogoutFailed, err.Error())
	}

	if status != 200 {
		return errors.Wrap(bmclibErrs.ErrLogoutFailed, strconv.Itoa(status))
	}

	if c.serviceClient.redfish != nil {
		err = c.serviceClient.redfish.Close(ctx)
		if err != nil {
			return errors.Wrap(bmclibErrs.ErrLogoutFailed, err.Error())
		}

		c.serviceClient.redfish = nil
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
	body, status, err := c.serviceClient.query(ctx, endpoint, http.MethodGet, nil, headers, 0)
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

	body, status, err := c.serviceClient.query(ctx, "cgi/op.cgi", http.MethodPost, bytes.NewBufferString(data), headers, 0)
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

type serviceClient struct {
	host      string
	port      string
	user      string
	pass      string
	csrfToken string
	client    *http.Client
	redfish   *redfishwrapper.Client
	sum       *sum.Sum
}

func newBmcServiceClient(host, port, user, pass string, client *http.Client) (*serviceClient, error) {
	sc := &serviceClient{host: host, port: port, user: user, pass: pass, client: client}

	if !strings.HasPrefix(host, "https://") && !strings.HasPrefix(host, "http://") {
		sc.host = "https://" + host
	}

	s, err := sum.New(host, user, pass)
	if err != nil {
		return nil, err
	}
	sc.sum = s

	return sc, nil
}

func (c *serviceClient) setCsrfToken(t string) {
	c.csrfToken = t
}

func (c *serviceClient) redfishSession(ctx context.Context) (err error) {
	if c.redfish != nil && c.redfish.SessionActive() == nil {
		return nil
	}

	c.redfish = redfishwrapper.NewClient(
		c.host,
		c.port,
		c.user,
		c.pass,
		redfishwrapper.WithHTTPClient(c.client),
	)
	if err := c.redfish.Open(ctx); err != nil {
		return err
	}

	return nil
}

func (c *serviceClient) supportsFirmwareInstall(model string) error {
	if model == "" {
		return errors.Wrap(ErrModelUnknown, "unable to determine firmware install compatibility")
	}

	for _, s := range supportedModels {
		if strings.EqualFold(s, model) {
			return nil
		}
	}

	return errors.Wrap(ErrModelUnsupported, "firmware install not supported for: "+model)
}

func (c *serviceClient) query(ctx context.Context, endpoint, method string, payload io.Reader, headers map[string]string, contentLength int64) ([]byte, int, error) {
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

// SendNMI tells the BMC to issue an NMI to the device
func (c *Client) SendNMI(ctx context.Context) error {
	return c.serviceClient.redfish.SendNMI(ctx)
}

// GetBootProgress allows a caller to follow along as the system goes through its boot sequence
func (c *Client) GetBootProgress() (*redfish.BootProgress, error) {
	return c.bmc.getBootProgress()
}

// BootComplete checks if this system has reached the last state for boot
func (c *Client) BootComplete() (bool, error) {
	return c.bmc.bootComplete()
}
