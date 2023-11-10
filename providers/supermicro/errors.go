package supermicro

import (
	"fmt"
	"strconv"

	"github.com/pkg/errors"
)

var (
	ErrQueryFRUInfo      = errors.New("FRU information query returned error")
	ErrXMLAPIUnsupported = errors.New("XML API is unsupported")
	ErrModelUnknown      = errors.New("Model number unknown")
	ErrModelUnsupported  = errors.New("Model not supported")
)

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
