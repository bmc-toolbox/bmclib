package errors

import "errors"

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
	// Err500 is returned when we recieve a 500 response from an endpoint.
	Err500 = errors.New("we've received 500 calling this endpoint")
	// ErrNotImplemented is returned for not implemented methods called
	ErrNotImplemented = errors.New("this feature hasn't been implemented yet")

	// ErriDracMaxSessionsReached indicates the bmc has reached the max number of login sessions.
	ErrIdracMaxSessionsReached = errors.New("The maximum number of user sessions is reached.")

	// Err401Redfish indicates auth failure
	Err401Redfish = errors.New("Redfish authorization failed.")
)
