package supermicro

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

var (
	// ErrQueryFRUInfo is returned when querying FRU information fails.
	ErrQueryFRUInfo = errors.New("FRU information query returned error")
	// ErrXMLAPIUnsupported is returned when the device does not support the XML API.
	ErrXMLAPIUnsupported = errors.New("XML API is unsupported")
	// ErrModelUnknown is returned when the device model number is unknown.
	ErrModelUnknown = errors.New("Model number unknown")
	// ErrModelUnsupported is returned when the device model is not supported.
	ErrModelUnsupported = errors.New("Model not supported")
	// ErrBoardIDUnknown is returned when the board ID cannot be identified.
	ErrBoardIDUnknown = errors.New("BoardID could not be identified")
	// ErrUnexpectedResponse is returned when the BMC returns unexpected response content.
	ErrUnexpectedResponse = errors.New("Unexpected response content")
	// ErrUnexpectedStatusCode is returned when the BMC returns an unexpected HTTP status code.
	ErrUnexpectedStatusCode = errors.New("Unexpected status code")
)

// UnexpectedResponseError describes an unexpected HTTP response returned by the BMC.
type UnexpectedResponseError struct {
	payload    string
	response   string
	statusCode string
}

func (e *UnexpectedResponseError) Error() string {
	return fmt.Sprintf(
		"unexpected response - statusCode: %s, payload: %s, response: %s",
		e.statusCode,
		e.payload,
		e.response,
	)
}

func unexpectedResponseErr(payload, response []byte, statusCode int) error {
	return &UnexpectedResponseError{
		string(payload),
		string(response),
		strconv.Itoa(statusCode),
	}
}
