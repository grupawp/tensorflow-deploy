package exterr

import (
	"encoding/json"
	"fmt"
	"runtime"
)

// Error is an extended error main struct
type Error struct {
	Err     string `json:"message,omitempty"`
	ErrCode int    `json:"code,omitempty"`
	Comp    string `json:"component,omitempty"`
	File    string `json:"file,omitempty"`
	Line    int    `json:"line,omitempty"`
	Cause   error
}

// NewErrorWithMessage creates Error type with string in parameter
func NewErrorWithMessage(msg string) *Error {
	err := Error{Err: msg}
	err.setFrame()
	return &err
}

// NewErrorWithErr creates Error type with error in parameter
func NewErrorWithErr(err error) *Error {
	if err == nil {
		panic("err parameter is nil")
	}

	newErr := Error{Err: err.Error()}
	newErr.setFrame()
	return &newErr
}

func (e *Error) setFrameWithSkip(skip int) {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return
	}
	e.File = file
	e.Line = line
}

func (e *Error) setFrame() {
	e.setFrameWithSkip(3)
}

// WithComponent sets component name of Error
func (e *Error) WithComponent(component string) *Error {
	e.Comp = component
	return e
}

// WithCode sets error code of Error
func (e *Error) WithCode(code int) *Error {
	e.ErrCode = code
	return e
}

// Error prints component and error code
func (e *Error) Error() string {
	var errorString string

	if len(e.Comp) > 0 {
		errorString = e.Comp + "-"
	}

	if e.ErrCode != 0 {
		errorString = errorString + fmt.Sprintf("%04d", e.ErrCode) + " "
	}

	return errorString + e.Err
}

// Code is getter to error code
func (e *Error) Code() int {
	return e.ErrCode
}

// Message is getter to message
func (e *Error) Message() string {
	return e.Err
}

// Component is getter to component
func (e *Error) Component() string {
	return e.Comp
}

// JSON returns json representation
func (e *Error) JSON() ([]byte, error) {
	return json.Marshal(e)
}
