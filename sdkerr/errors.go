package sdkerr

import (
	"errors"
	"fmt"
	"strings"
)

// http
var (
	// ErrValidation indicates a validation error.
	ErrValidation = errors.New("validation error")
	// ErrConfiguration indicates a configuration error.
	ErrConfiguration = errors.New("configuration error")
	// ErrRequestFailed indicates a request failed.
	ErrRequestFailed = errors.New("request failed")
	// ErrAPIError indicates an API error.
	ErrAPIError = errors.New("api error")
	// ErrDecodeError indicates a decode error.
	ErrDecodeError = errors.New("decode error")
)

// ws
var (
	// ErrWSConnection indicates a websocket connection error.
	ErrWSConnection = errors.New("websocket connection error")
	// ErrWSWrite indicates a websocket write failed.
	ErrWSWrite = errors.New("websocket write failed")
	// ErrWSRead indicates a websocket read failed.
	ErrWSRead = errors.New("websocket read failed")
	// ErrWSPing indicates a websocket ping failed.
	ErrWSPing = errors.New("websocket ping failed")
	// ErrWSMessageTimeout indicates a websocket message response timeout.
	ErrWSMessageTimeout = errors.New("websocket message response timeout")
	// ErrWSServerError indicates a websocket server error.
	ErrWSServerError = errors.New("websocket server error")
	// ErrWSClose indicates a websocket close failed.
	ErrWSClose = errors.New("websocket close failed")
	// ErrWSUnknown indicates a websocket unknown error.
	ErrWSUnknown = errors.New("websocket unknown error")
)

// SDKError is a custom error type for the SDK.
type SDKError struct {
	kind    error
	message string
	cause   error
	op      string
	subsys  string
}

// Error returns the error message.
func (e *SDKError) Error() string {
	var parts []string

	if e.subsys != "" {
		parts = append(parts, fmt.Sprintf("subsys: %s", e.subsys))
	}
	if e.op != "" {
		parts = append(parts, fmt.Sprintf("op: %s", e.op))
	}
	if e.kind != nil {
		parts = append(parts, fmt.Sprintf("kind: %s", e.kind))
	}
	if e.message != "" {
		parts = append(parts, fmt.Sprintf("msg: %s", e.message))
	}
	if e.cause != nil {
		parts = append(parts, fmt.Sprintf("cause: %s", e.cause))
	}

	return strings.Join(parts, " | ")
}

// Is reports whether any error in an SDKError's chain matches target.
func (e *SDKError) Is(target error) bool {
	if e.kind != nil && errors.Is(e.kind, target) {
		return true
	}
	if e.cause != nil && errors.Is(e.cause, target) {
		return true
	}
	return false
}

// As finds the first error in an SDKError's chain that matches target, and if so, sets target to that error value and returns true.
func (e *SDKError) As(target any) bool {
	if e.kind != nil && errors.As(e.kind, target) {
		return true
	}
	if e.cause != nil && errors.As(e.cause, target) {
		return true
	}
	return false
}

// Unwrap returns the cause of the error.
func (e *SDKError) Unwrap() error {
	return e.cause
}

// Kind returns the kind of the error.
func (e *SDKError) Kind() error {
	return e.kind
}

// Message returns the message of the error.
func (e *SDKError) Message() string {
	return e.message
}

// Cause returns the cause of the error.
func (e *SDKError) Cause() error {
	return e.cause
}

// Op returns the operation of the error.
func (e *SDKError) Op() string {
	return e.op
}

// Subsys returns the subsystem of the error.
func (e *SDKError) Subsys() string {
	return e.subsys
}

// NewSDKError creates a new SDKError.
func NewSDKError() *SDKError {
	return &SDKError{}
}

// WithKind sets the kind of the error.
func (e *SDKError) WithKind(kind error) *SDKError {
	e.kind = kind
	return e
}

// WithMessage sets the message of the error.
func (e *SDKError) WithMessage(msg string) *SDKError {
	e.message = msg
	return e
}

// WithCause sets the cause of the error.
func (e *SDKError) WithCause(err error) *SDKError {
	e.cause = err
	return e
}

// WithOp sets the operation of the error.
func (e *SDKError) WithOp(op string) *SDKError {
	e.op = op
	return e
}

// WithSubsys sets the subsystem of the error.
func (e *SDKError) WithSubsys(subsys string) *SDKError {
	e.subsys = subsys
	return e
}
