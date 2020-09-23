package exterr

import (
	"fmt"
)

func (e Error) DumpStackWithID(requestID string) []string {
	var result []string

	if e.Cause != nil {
		if newErr, ok := e.Cause.(*Error); ok {
			result = append(result, newErr.DumpStackWithID(requestID)...)
		}

	}
	result = append(result, fmt.Sprintf("%s [%s:%d] %s", requestID, e.File, e.Line, e.Error()))
	return result
}

func (e Error) DumpStack() []string {
	var result []string

	if e.Cause != nil {
		if newErr, ok := e.Cause.(*Error); ok {
			result = append(result, newErr.DumpStack()...)
		}
	}
	result = append(result, fmt.Sprintf("[%s:%d] %s", e.File, e.Line, e.Error()))
	return result
}
