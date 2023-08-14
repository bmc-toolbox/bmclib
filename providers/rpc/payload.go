package rpc

import "fmt"

type Method string

const (
	BootDeviceMethod   Method = "setBootDevice"
	PowerSetMethod     Method = "setPowerState"
	PowerGetMethod     Method = "getPowerState"
	VirtualMediaMethod Method = "setVirtualMedia"
)

type RequestPayload struct {
	ID     int64       `json:"id"`
	Host   string      `json:"host"`
	Method Method      `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

type BootDeviceParams struct {
	Device     string `json:"device"`
	Persistent bool   `json:"persistent"`
	EFIBoot    bool   `json:"efiBoot"`
}

type PowerSetParams struct {
	State string `json:"state"`
}

type VirtualMediaParams struct {
	MediaURL string `json:"mediaUrl"`
	Kind     string `json:"kind"`
}

type ResponsePayload struct {
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
