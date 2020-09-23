package rz

import "time"

const (
	// DefaultTimestampFieldName is the default field name used for the timestamp field.
	DefaultTimestampFieldName = "timestamp"

	// DefaultLevelFieldName is the default field name used for the level field.
	DefaultLevelFieldName = "level"

	// DefaultMessageFieldName is the default field name used for the message field.
	DefaultMessageFieldName = "message"

	// DefaultErrorFieldName is the default field name used for error fields.
	DefaultErrorFieldName = "error"

	// DefaultCallerFieldName is the default field name used for caller field.
	DefaultCallerFieldName = "caller"

	// DefaultCallerSkipFrameCount is the default number of stack frames to skip to find the caller.
	DefaultCallerSkipFrameCount = 3

	// DefaultErrorStackFieldName is the default field name used for error stacks.
	DefaultErrorStackFieldName = "stack"

	// DefaultTimeFieldFormat defines the time format of the Time field type.
	// If set to an empty string, the time is formatted as an UNIX timestamp
	// as integer.
	DefaultTimeFieldFormat = time.RFC3339
)

var (
	// DefaultTimestampFunc defines default the function called to generate a timestamp.
	DefaultTimestampFunc func() time.Time = func() time.Time { return time.Now().UTC() }
)
