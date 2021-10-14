package asrockrack

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/providers"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
)

const (
	// ProviderName for the provider implementation
	ProviderName = "asrockrack"
	// ProviderProtocol for the provider implementation
	ProviderProtocol = "vendorapi"
)

var (
	// Features implemented by asrockrack https
	Features = registrar.Features{
		providers.FeatureBiosVersionRead,
		providers.FeatureBmcVersionRead,
		providers.FeatureBiosFirmwareUpdate,
		providers.FeatureBmcFirmwareUpdate,
	}
)

// ASRockRack holds the status and properties of a connection to a asrockrack bmc
type ASRockRack struct {
	ip            string
	username      string
	password      string
	loginSession  *loginSession
	httpClient    *http.Client
	resetRequired bool // Indicates if the BMC requires a reset
	skipLogout    bool // A Close() / httpsLogout() request is ignored if the BMC was just flashed - since the sessions are terminated either way
	log           logr.Logger
}

// New returns a new ASRockRack instance ready to be used
func New(ip string, username string, password string, log logr.Logger) (*ASRockRack, error) {
	client, err := httpclient.Build()
	if err != nil {
		return nil, err
	}

	return &ASRockRack{
		ip:           ip,
		username:     username,
		password:     password,
		log:          log,
		loginSession: &loginSession{},
		httpClient:   client,
	}, nil
}

// Compatible implements the registrar.Verifier interface
// returns true if the BMC is identified to be an asrockrack
func (a *ASRockRack) Compatible() bool {
	resp, statusCode, err := a.queryHTTPS("/", "GET", nil, nil, 0)
	if err != nil {
		return false
	}

	if statusCode != 200 {
		return false
	}

	return bytes.Contains(resp, []byte(`ASRockRack`))
}

// Open a connection to a BMC, implements the Opener interface
func (a *ASRockRack) Open(ctx context.Context) (err error) {
	return a.httpsLogin()
}

// Close a connection to a BMC, implements the Closer interface
func (a *ASRockRack) Close(ctx context.Context) (err error) {
	if a.skipLogout {
		return nil
	}

	return a.httpsLogout()
}

// CheckCredentials verify whether the credentials are valid or not
func (a *ASRockRack) CheckCredentials() (err error) {
	return a.httpsLogin()
}

// BiosVersion returns the BIOS version from the BMC
func (a *ASRockRack) GetBIOSVersion(ctx context.Context) (string, error) {
	fwInfo, err := a.firmwareInfo()
	if err != nil {
		return "", err
	}

	return fwInfo.BIOSVersion, nil
}

// BMCVersion returns the BMC version
func (a *ASRockRack) GetBMCVersion(ctx context.Context) (string, error) {
	fwInfo, err := a.firmwareInfo()
	if err != nil {
		return "", err
	}

	return fwInfo.BMCVersion, nil
}

// nolint: gocyclo
// BMC firmware update is a multi step process
// this method initiates the upgrade process and waits in a loop until the device has been upgraded
// fileSize is required to setup the multipart form upload Content-Length
func (a *ASRockRack) FirmwareUpdateBMC(ctx context.Context, fileReader io.Reader, fileSize int64) error {
	defer func() {
		// The device needs to be reset to be removed from flash mode,
		// this is required once setFlashMode() is invoked.
		// The BMC resets itself once a firmware flash is successful or failed.
		if a.resetRequired {
			a.log.V(1).Info("info", "resetting BMC, this takes a few minutes")
			err := a.reset()
			if err != nil {
				a.log.Error(err, "failed to reset BMC")
			}
		}
	}()

	var err error

	// 1. set the device to flash mode - prepares the flash
	a.log.V(2).Info("info", "action", "set device to flash mode, takes a minute...", "step", "1/5")
	err = a.setFlashMode()
	if err != nil {
		return fmt.Errorf("failed in step 1/5 - set device to flash mode: " + err.Error())
	}

	// 2. upload firmware image file
	a.log.V(2).Info("info", "action", "upload BMC firmware image", "step", "2/5")
	err = a.uploadFirmware("api/maintenance/firmware", fileReader, fileSize)
	if err != nil {
		return fmt.Errorf("failed in step 2/5 - upload BMC firmware image: " + err.Error())
	}

	// 3. BMC to verify the uploaded file
	err = a.verifyUploadedFirmware()
	a.log.V(2).Info("info", "action", "BMC verify uploaded firmware", "step", "3/5")
	if err != nil {
		return fmt.Errorf("failed in step 3/5 - BMC verify uploaded firmware: " + err.Error())
	}

	startTS := time.Now()
	// 4. Run the upgrade - preserving current config
	a.log.V(2).Info("info", "action", "proceed with upgrade, preserve current configuration", "step", "4/5")
	err = a.upgradeBMC()
	if err != nil {
		return fmt.Errorf("failed in step 4/5 - proceed with upgrade: " + err.Error())
	}

	// progress check interval
	progressT := time.NewTicker(500 * time.Millisecond).C
	// timeout interval
	timeoutT := time.NewTicker(20 * time.Minute).C
	maxErrors := 20
	var errorsCount int

	// 5.loop until firmware was updated - with a timeout
	for {
		select {
		case <-progressT:
			// check progress
			endpoint := "api/maintenance/firmware/flash-progress"
			p, err := a.flashProgress(endpoint)
			if err != nil {
				errorsCount++
				a.log.V(2).Error(err, "step", "5/5 - error checking flash progress", "error count", errorsCount, "max errors", maxErrors, "elapsed time", time.Since(startTS).String())
				continue
			}

			a.log.V(2).Info("info", "action", p.Action, "step", "5/5", "progress", p.Progress, "elapsed time", time.Since(startTS).String())

			// all done!
			if p.State == 2 {
				a.log.V(2).Info("info", "action", "flash process complete", "step", "5/5", "progress", p.Progress, "elapsed time", time.Since(startTS).String())
				// The BMC resets by itself after a successful flash
				a.resetRequired = false
				// HTTP sessions are terminated once the BMC resets after an upgrade
				a.skipLogout = true
				return nil
			}
		case <-timeoutT:
			return fmt.Errorf("timeout in step 5/5 - flash progress, error count: %d, elapsed time: %s", errorsCount, time.Since(startTS).String())
		}
	}
}

func (a *ASRockRack) FirmwareUpdateBIOS(ctx context.Context, fileReader io.Reader, fileSize int64) error {
	defer func() {
		if a.resetRequired {
			a.log.V(2).Info("info", "resetting BMC, this takes a few minutes")
			err := a.reset()
			if err != nil {
				a.log.Error(err, "failed to reset BMC")
			}
		}
	}()

	var err error

	// 1. upload firmware image file
	a.log.V(2).Info("info", "action", "upload BIOS firmware image", "step", "1/4")
	err = a.uploadFirmware("api/asrr/maintenance/BIOS/firmware", fileReader, fileSize)
	if err != nil {
		return fmt.Errorf("failed in step 1/4 - upload firmware image: " + err.Error())
	}

	// 2. set update parameters to preserve configurations
	a.log.V(2).Info("info", "action", "set flash configuration", "step", "2/4")
	err = a.biosUpgradeConfiguration()
	if err != nil {
		return fmt.Errorf("failed in step 2/4 - set flash configuration: " + err.Error())
	}

	startTS := time.Now()
	// 3. run upgrade
	a.log.V(2).Info("info", "action", "proceed with upgrade", "step", "3/4")
	err = a.biosUpgrade()
	if err != nil {
		return fmt.Errorf("failed in step 3/4 - proceed with upgrade: " + err.Error())
	}

	// progress check interval
	progressT := time.NewTicker(2 * time.Second).C
	// timeout interval
	timeoutT := time.NewTicker(30 * time.Minute).C
	maxErrors := 20
	var errorsCount int

	// 5.loop until firmware was updated - with a timeout
	for {
		select {
		case <-progressT:
			// check progress
			endpoint := "api/asrr/maintenance/BIOS/flash-progress"
			p, err := a.flashProgress(endpoint)
			if err != nil {
				errorsCount++
				a.log.V(2).Error(err, "action", "check flash progress", "step", "4/4", "error count", errorsCount, "max errors", maxErrors, "elapsed time", time.Since(startTS).String())
				continue
			}

			a.log.V(2).Info("info", "action", "check flash progress", "step", "4/4", "progress", p.Progress, "action", p.Action, "elapsed time", time.Since(startTS).String())

			// all done!
			if p.State == 2 {
				a.log.V(2).Info("info", "action", "flash process complete", "step", "4/4", "progress", p.Progress, "elapsed time", time.Since(startTS).String())
				// Reset BMC after flash
				a.resetRequired = true
				return nil
			}
		case <-timeoutT:
			return fmt.Errorf("timeout in step 4/4 - flash progress, error count: %d, elapsed time: %s", errorsCount, time.Since(startTS).String())
		}
	}
}
