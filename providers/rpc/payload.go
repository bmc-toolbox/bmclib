package rpc

import "fmt"

type Method string

const (
	BootDeviceMethod   Method = "setBootDevice"
	PowerSetMethod     Method = "setPowerState"
	PowerGetMethod     Method = "getPowerState"
	VirtualMediaMethod Method = "setVirtualMedia"
)

// RequestPayload is the payload sent to the ConsumerURL.
type RequestPayload struct {
	ID     int64       `json:"id"`
	Host   string      `json:"host"`
	Method Method      `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

// BootDeviceParams are the parameters options used when setting a boot device.
type BootDeviceParams struct {
	Device     string `json:"device"`
	Persistent bool   `json:"persistent"`
	EFIBoot    bool   `json:"efiBoot"`
}

// PowerSetParams are the parameters options used when setting the power state.
type PowerSetParams struct {
	State string `json:"state"`
}

// PowerGetParams are the parameters options used when getting the power state.
type VirtualMediaParams struct {
	MediaURL string `json:"mediaUrl"`
	Kind     string `json:"kind"`
}

// ResponsePayload is the payload received from the ConsumerURL.
// The Result field is an interface{} so that different methods
// can define the contract according to their needs.
type ResponsePayload struct {
	// ID is the ID of the response. It should match the ID of the request but is not enforced.
	ID     int64          `json:"id"`
	Host   string         `json:"host"`
	Result interface{}    `json:"result,omitempty"`
	Error  *ResponseError `json:"error,omitempty"`
}

type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type PowerGetResult string

const (
	PoweredOn  PowerGetResult = "on"
	PoweredOff PowerGetResult = "off"
)

func (p PowerGetResult) String() string {
	return string(p)
}

func (r *ResponseError) String() string {
	return fmt.Sprintf("code: %v, message: %v", r.Code, r.Message)
}
