package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrLoginFailed is returned when we fail to login to a bmc
	ErrLoginFailed = errors.New("failed to login")

	// ErrLogoutFailed is returned when we fail to logout from a bmc
	ErrLogoutFailed = errors.New("failed to logout")

	// ErrNotAuthenticated is returned when the session is not active.
	ErrNotAuthenticated = errors.New("not authenticated")

	// ErrNon200Response is returned when bmclib recieves an unexpected non-200 status code for a query
	ErrNon200Response = errors.New("non-200 response returned for the endpoint")

	// ErrNotImplemented is returned for not implemented methods called
	ErrNotImplemented = errors.New("this feature hasn't been implemented yet")

	// ErrRetrievingUserAccounts is returned when bmclib is unable to retrieve user accounts from the BMC
	ErrRetrievingUserAccounts = errors.New("error retrieving user accounts")

	// ErrInvalidUserRole is returned when the given user account role is not valid
	ErrInvalidUserRole = errors.New("invalid user account role")

	// ErrUserParamsRequired is returned when all the required user parameters are not provided - username, password, role
	ErrUserParamsRequired = errors.New("username, password and role are required parameters")

	// ErrUserAccountExists is returned when a user account with the username is already present
	ErrUserAccountExists = errors.New("user account already exists")

	// ErrNoUserSlotsAvailable is returned when there are no user account slots available
	ErrNoUserSlotsAvailable = errors.New("no user account slots available")

	// ErrUserAccountNotFound is returned when the user account is not present
	ErrUserAccountNotFound = errors.New("given user account does not exist")

	// ErrUserAccountUpdate is returned when the user account failed to be updated
	ErrUserAccountUpdate = errors.New("user account attributes could not be updated")

	// ErrRedfishChassisOdataID is returned when no compatible Chassis Odata IDs were identified
	ErrRedfishChassisOdataID = errors.New("no compatible Chassis Odata IDs identified")

	// ErrRedfishSystemOdataID is returned when no compatible System Odata IDs were identified
	ErrRedfishSystemOdataID = errors.New("no compatible System Odata IDs identified")

	// ErrRedfishManagerOdataID is returned when no compatible Manager Odata IDs were identified
	ErrRedfishManagerOdataID = errors.New("no compatible Manager Odata IDs identified")

	// ErrRedfishServiceNil is returned when a redfish method is invoked on a nil redfish (gofish) Service object
	ErrRedfishServiceNil = errors.New("redfish connection returned a nil redfish Service object")

	// ErrRedfishSoftwareInventory is returned when software inventory could not be collected over redfish
	ErrRedfishSoftwareInventory = errors.New("error collecting redfish software inventory")

	// ErrFirmwareUpload is returned when a firmware upload method fails
	ErrFirmwareUpload = errors.New("error uploading firmware")

	// ErrFirmwareInstall is returned for firmware install failures
	ErrFirmwareInstall = errors.New("error updating firmware")

	// ErrFirmwareInstallStatus is returned for firmware install status read
	ErrFirmwareInstallStatus = errors.New("error querying firmware install status")

	// ErrRedfishUpdateService is returned on redfish update service errors
	ErrRedfishUpdateService = errors.New("redfish update service error")

	// ErrTaskNotFound is returned when the (redfish) task could not be found
	ErrTaskNotFound = errors.New("task not found")

	// ErrTaskPurge is returned when a (redfish) task could not be purged
	ErrTaskPurge = errors.New("unable to purge task")

	// ErrPowerStatusRead is returned when a power status read query fails
	ErrPowerStatusRead = errors.New("error returning power status")

	// ErrPowerStatusSet is returned when a power status set query fails
	ErrPowerStatusSet = errors.New("error setting power status")

	// ErrProviderImplementation is returned when theres an error in the BMC provider implementation
	ErrProviderImplementation = errors.New("error in provider implementation")

	// ErrCompatibilityCheck is returned when the compatibility probe failed to complete successfully.
	ErrCompatibilityCheck = errors.New("compatibility check failed")

	// ErrNoBiosAttributes is returned when no bios attributes are available from the BMC.
	ErrNoBiosAttributes = errors.New("no BIOS attributes available")

	// ErrScreenshot is returned when screen capture fails.
	ErrScreenshot = errors.New("error in capturing screen")

	// ErrIncompatibleProvider is returned by Open() when the device is not compatible with the provider
	ErrIncompatibleProvider = errors.New("provider not compatible with device")
)

type ErrUnsupportedHardware struct {
	msg string
}

func (e *ErrUnsupportedHardware) Error() string {
	return fmt.Sprintf("Hardware not supported: %s", e.msg)
}

func NewErrUnsupportedHardware(s string) error {
	return &ErrUnsupportedHardware{s}
}
