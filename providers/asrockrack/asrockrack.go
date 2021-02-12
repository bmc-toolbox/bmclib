package asrockrack

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/bmc-toolbox/bmclib/internal/httpclient"
	"github.com/bmc-toolbox/bmclib/providers"
	"github.com/go-logr/logr"
	"github.com/jacobweinstock/registrar"
)

const (
	// BmcType defines the bmc model that is supported by this package
	BmcType = "asrockrack"
	// VendorID represents the id of the vendor across all packages
	VendorID = "ASRockRack"
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
	fwInfo        *firmwareInfo
	resetRequired bool // Indicates if the BMC requires a reset
	skipLogout    bool // A Close() / httpsLogout() request is ignored if the BMC was just flashed - since the sessions are terminated either way
	ctx           context.Context
	log           logr.Logger
}

// New returns a new ASRockRack instance ready to be used
func New(ctx context.Context, ip string, username string, password string, log logr.Logger) (*ASRockRack, error) {

	client, err := httpclient.Build()
	if err != nil {
		return nil, err
	}

	return &ASRockRack{
		ip:           ip,
		username:     username,
		password:     password,
		ctx:          ctx,
		log:          log,
		loginSession: &loginSession{},
		httpClient:   client,
	}, nil
}

// Open a connection to a BMC
func (a *ASRockRack) Open(ctx context.Context) (err error) {

	err = a.httpsLogin()
	if err != nil {
		return err
	}

	return nil
}

// Close a connection to a BMC
func (a *ASRockRack) Close(ctx context.Context) (err error) {

	if a.skipLogout {
		return nil
	}

	err = a.httpsLogout()
	if err != nil {
		return err
	}

	return nil
}

// CheckCredentials verify whether the credentials are valid or not
func (a *ASRockRack) CheckCredentials() (err error) {
	err = a.httpsLogin()
	if err != nil {
		return err
	}
	return err
}

// BiosVersion returns the BIOS version from the BMC
func (a *ASRockRack) GetBIOSVersion(ctx context.Context) (string, error) {

	var err error
	if a.fwInfo == nil {
		a.fwInfo, err = a.firmwareInfo()
		if err != nil {
			return "", err
		}
	}

	return a.fwInfo.BIOSVersion, nil
}

// BMCVersion returns the BMC version
func (a *ASRockRack) GetBMCVersion(ctx context.Context) (string, error) {

	var err error
	if a.fwInfo == nil {
		a.fwInfo, err = a.firmwareInfo()
		if err != nil {
			return "", err
		}
	}

	return a.fwInfo.BMCVersion, nil
}

// nolint: gocyclo
// BMC firmware update is a multi step process
// this method initiates the upgrade process and waits in a loop until the device has been upgraded
func (a *ASRockRack) FirmwareUpdateBMC(ctx context.Context, filePath string) error {

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

	// 0. validate the given filePath exists
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return err
	}

	// 1. set the device to flash mode - prepares the flash
	a.log.V(1).Info("info", "step", "1/5 - setting device into flash mode.. this takes a minute")
	err = a.setFlashMode()
	if err != nil {
		return fmt.Errorf("failed in step 1/5 - set device in flash mode: " + err.Error())
	}

	// 2. upload firmware image file
	a.log.V(1).Info("info", "step", "2/5 uploading firmware image", "filePath", filePath)
	err = a.uploadFirmware("api/maintenance/firmware", filePath)
	if err != nil {
		return fmt.Errorf("failed in step 2/5 - upload firmware image: " + err.Error())
	}

	// 3. BMC to verify the uploaded file
	err = a.verifyUploadedFirmware()
	a.log.V(1).Info("info", "step", "3/5 - bmc to verify uploaded firmware")
	if err != nil {
		return fmt.Errorf("failed in step 3/5 - verify uploaded firmware: " + err.Error())
	}

	startTS := time.Now()
	// 4. Run the upgrade - preserving current config
	a.log.V(1).Info("info", "step 4/5", "run the upgrade, preserving current configuration")
	err = a.upgradeBMC()
	if err != nil {
		return fmt.Errorf("failed in step 4/5 - verify uploaded firmware: " + err.Error())
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
			endpoint := "api/maintenance/firmware/flash-progress"
			p, err := a.flashProgress(endpoint)
			if err != nil {
				errorsCount++
				a.log.V(1).Error(err, "step", "5/5 - error checking flash progress", "error count", errorsCount, "max errors", maxErrors, "elapsed time", time.Since(startTS).String())
				continue
			}

			a.log.V(1).Info("info", "step", "5/5 - check flash progress..", "progress", p.Progress, "action", p.Action, "elapsed time", time.Since(startTS).String())

			// all done!
			if p.State == 2 {
				a.log.V(1).Info("info", "step", "5/5 - firmware flash complete!", "progress", p.Progress, "action", p.Action, "elapsed time", time.Since(startTS).String())
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

func (a *ASRockRack) FirmwareUpdateBIOS(ctx context.Context, filePath string) error {

	defer func() {
		if a.resetRequired {
			a.log.V(1).Info("info", "resetting BMC, this takes a few minutes")
			err := a.reset()
			if err != nil {
				a.log.Error(err, "failed to reset BMC")
			}
		}
	}()

	// 0. validate the given filePath exists
	_, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return err
	}

	// 1. upload firmware image file
	a.log.V(1).Info("info", "step", "1/4 uploading firmware image", "filePath", filePath)
	err = a.uploadFirmware("api/asrr/maintenance/BIOS/firmware", filePath)
	if err != nil {
		return fmt.Errorf("failed in step 1/4 - upload firmware image: " + err.Error())
	}

	// 2. set update parameters to preserve configuratin
	a.log.V(1).Info("info", "step", "2/4 set preserve configuration")
	err = a.biosUpgradeConfiguration()
	if err != nil {
		return fmt.Errorf("failed in step 2/4 - set preserve configuration: " + err.Error())
	}

	startTS := time.Now()
	// 3. run upgrade
	a.log.V(1).Info("info", "step", "3/4 run upgrade")
	err = a.biosUpgrade()
	if err != nil {
		return fmt.Errorf("failed in step 3/4 - run upgrade: " + err.Error())
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
				a.log.V(1).Error(err, "step", "4/4 - error checking flash progress", "error count", errorsCount, "max errors", maxErrors, "elapsed time", time.Since(startTS).String())
				continue
			}

			a.log.V(1).Info("info", "step", "4/4 - check flash progress..", "progress", p.Progress, "action", p.Action, "elapsed time", time.Since(startTS).String())

			// all done!
			if p.State == 2 {
				a.log.V(1).Info("info", "step", "4/4 - firmware flash complete!", "progress", p.Progress, "action", p.Action, "elapsed time", time.Since(startTS).String())
				// Reset BMC after flash
				a.resetRequired = true
				return nil
			}
		case <-timeoutT:
			return fmt.Errorf("timeout in step 5/5 - flash progress, error count: %d, elapsed time: %s", errorsCount, time.Since(startTS).String())
		}
	}

}
