package rpc

import "fmt"

// Method is the RPC method name invoked against the ConsumerURL.
type Method string

const (
	// BootDeviceMethod sets the next boot device.
	BootDeviceMethod Method = "setBootDevice"
	// PowerSetMethod sets the power state.
	PowerSetMethod Method = "setPowerState"
	// PowerGetMethod gets the power state.
	PowerGetMethod Method = "getPowerState"
	// VirtualMediaMethod sets virtual media.
	VirtualMediaMethod Method = "setVirtualMedia"
	// PingMethod pings the ConsumerURL.
	PingMethod Method = "ping"
)

// RequestPayload is the payload sent to the ConsumerURL.
type RequestPayload struct {
	ID     int64  `json:"id"`
	Host   string `json:"host"`
	Method Method `json:"method"`
	Params any    `json:"params,omitempty"`
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

// VirtualMediaParams are the parameters options used when setting virtual media.
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
	Result any            `json:"result,omitempty"`
	Error  *ResponseError `json:"error,omitempty"`
}

// ResponseError describes an error returned in a ResponsePayload.
type ResponseError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// PowerGetResult is the power state returned by a getPowerState response.
type PowerGetResult string

const (
	// PoweredOn indicates the host is powered on.
	PoweredOn PowerGetResult = "on"
	// PoweredOff indicates the host is powered off.
	PoweredOff PowerGetResult = "off"
)

func (p PowerGetResult) String() string {
	return string(p)
}

func (r *ResponseError) String() string {
	return fmt.Sprintf("code: %v, message: %v", r.Code, r.Message)
}
