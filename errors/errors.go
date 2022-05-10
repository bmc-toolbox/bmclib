package errors

import (
	"errors"
	"fmt"
)

var (
	// ErrLoginFailed is returned when we fail to login to a bmc
	ErrLoginFailed = errors.New("failed to login")

	// ErrBiosNotFound is returned when we are not able to find the server bios version
	ErrBiosNotFound = errors.New("bios version not found")

	// ErrVendorUnknown is returned when we are unable to identify the redfish vendor
	ErrVendorUnknown = errors.New("unable to identify the vendor")

	// ErrInvalidSerial is returned when the serial number for the device is invalid
	ErrInvalidSerial = errors.New("unable to find the serial number")

	// ErrPageNotFound is used to inform the http request that we couldn't find the expected page and/or endpoint
	ErrPageNotFound = errors.New("requested page couldn't be found in the server")

	// ErrRedFishNotSupported is returned when redfish isn't supported by the vendor
	ErrRedFishNotSupported = errors.New("redfish not supported")

	// ErrUnableToReadData is returned when we fail to read data from a chassis or bmc
	ErrUnableToReadData = errors.New("unable to read data from this device")

	// ErrVendorNotSupported is returned when we are able to identify a vendor but we won't support it
	ErrVendorNotSupported = errors.New("vendor not supported")

	// ErrUnableToGetSessionToken is returned when we are unable to retrieve ST2 which is required to set configuration parameters
	ErrUnableToGetSessionToken = errors.New("unable to get ST2 session token")

	// Err500 is returned when we receive a 500 response from an endpoint.
	Err500 = errors.New("we've received 500 calling this endpoint")

	// ErrNon200Response is returned when bmclib recieves an unexpected non-200 status code for a query
	ErrNon200Response = errors.New("non-200 response returned for the endpoint")

	// ErrNotImplemented is returned for not implemented methods called
	ErrNotImplemented = errors.New("this feature hasn't been implemented yet")

	// ErrFeatureUnavailable is returned for features not available/supported.
	ErrFeatureUnavailable = errors.New("this feature isn't supported/available for this hardware")

	// ErrIdracMaxSessionsReached indicates the bmc has reached the max number of login sessions.
	ErrIdracMaxSessionsReached = errors.New("the maximum number of user sessions is reached")

	// Err401Redfish indicates auth failure
	Err401Redfish = errors.New("redfish authorization failed")

	// ErrDeviceNotMatched is the error returned when the device was not a type it was probed for
	ErrDeviceNotMatched = errors.New("the vendor device did not match the probe")

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
