package asrockrack

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
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
	ProviderProtocol = "https"
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
	ip           string
	username     string
	password     string
	sid          *http.Cookie
	loginSession *loginSession
	httpClient   *http.Client
	fwInfo       *firmwareInfo
	flashModeSet bool
	ctx          context.Context
	log          logr.Logger
}

// New returns a new SupermicroX instance ready to be used
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

	err = a.httpsLogout()
	if err != nil {
		return err
	}

	return nil
}

// BiosVersion returns the BIOS version from the BMC
func (a *ASRockRack) BiosVersion() (string, error) {

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
func (a *ASRockRack) BMCVersion() (string, error) {

	var err error
	if a.fwInfo == nil {
		a.fwInfo, err = a.firmwareInfo()
		if err != nil {
			return "", err
		}
	}

	return a.fwInfo.BIOSVersion, nil
}

// nolint: gocyclo
// BMC firmware update is a multi step process
// this method initiates the upgrade process and waits in a loop until the device has been upgraded
func (a *ASRockRack) FirmwareUpdateBMC(filePath string) error {

	// The BMC needs to be reset once set into flash mode
	defer func() {
		if a.flashModeSet {
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
	a.log.V(1).Info("info", "step 1/5 - setting device into flash mode.. this takes a minute", "filePath", filePath)
	err = a.setFlashMode()
	if err != nil {
		return fmt.Errorf("failed in step 1/5 - set device in flash mode: " + err.Error())
	}

	// 2. upload firmware image file
	a.log.V(1).Info("info", "step 2/5 - uploading firmware image", "filePath", filePath)
	err = a.uploadFirmware(filePath)
	if err != nil {
		return fmt.Errorf("failed in step 2/5 - upload firmware image: " + err.Error())
	}

	// 3. BMC to verify the uploaded file
	err = a.verifyUploadedFirmware()
	a.log.V(1).Info("info", "step 3/5 - verify uploaded firmware", "filePath", filePath)
	if err != nil {
		return fmt.Errorf("failed in step 3/5 - verify uploaded firmware: " + err.Error())
	}

	startTS := time.Now()
	// 4. Run the upgrade - preserving current config
	a.log.V(1).Info("info", "step 4/5 - run the upgrade, preserving current configuration", "filePath", filePath)
	err = a.upgradeBMC()
	if err != nil {
		return fmt.Errorf("failed in step 4/5 - verify uploaded firmware: " + err.Error())
	}

	// progress check interval
	progressT := time.NewTicker(10 * time.Second).C
	// timeout interval
	timeoutT := time.NewTicker(30 * time.Minute).C
	maxErrors := 20
	var errorsCount int

	// 5.loop until firmware was updated - with a timeout
	for {
		select {
		case <-progressT:
			// check progress
			p, err := a.flashProgress()
			if err != nil {
				errorsCount++
				a.log.Error(err, "step 5/5 - error checking flash progress", "error count", errorsCount, "max errors", maxErrors, "elapsed time", time.Since(startTS).String())
			}

			a.log.V(1).Info("info", "step 5/5 - flash progress..", "progress", p.Progress, "action", p.Action, "elapsed time", time.Since(startTS).String())

			// all done!
			if strings.Contains(p.Progress, "100% done") {
				return nil
			}
		case <-timeoutT:
			return fmt.Errorf("timeout in step 5/5 - flash progress, error count: %d, elapsed time: %s", errorsCount, time.Since(startTS).String())
		}
	}

}

func (a *ASRockRack) FirmwareUpdateBIOS(filePath string) error {

	return nil
}
