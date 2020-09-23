package rz

import (
	"io"
	"os"
	"time"
)

// LoggerOption is used to configure a logger.
type LoggerOption func(logger *Logger)

// Writer update logger's writer.
func Writer(writer io.Writer) LoggerOption {
	return func(logger *Logger) {
		if writer == nil {
			writer = os.Stdout
		}
		lw, ok := writer.(LevelWriter)
		if !ok {
			lw = levelWriterAdapter{writer}
		}
		logger.writer = lw
	}
}

// Level update logger's level.
func Level(lvl LogLevel) LoggerOption {
	return func(logger *Logger) {
		logger.level = lvl
	}
}

// Sampler update logger's sampler.
func Sampler(sampler LogSampler) LoggerOption {
	return func(logger *Logger) {
		logger.sampler = sampler
	}
}

// AddHook appends hook to logger's hook
func AddHook(hook LogHook) LoggerOption {
	return func(logger *Logger) {
		logger.hooks = append(logger.hooks, hook)
	}
}

// Hooks replaces logger's hooks
func Hooks(hooks ...LogHook) LoggerOption {
	return func(logger *Logger) {
		logger.hooks = hooks
	}
}

// Fields update logger's context fields
func Fields(fields ...Field) LoggerOption {
	return func(logger *Logger) {
		e := newEvent(logger.writer, logger.level)
		e.buf = nil
		copyInternalLoggerFieldsToEvent(logger, e)
		for i := range fields {
			fields[i](e)
		}
		if e.stack != logger.stack {
			logger.stack = e.stack
		}
		if e.caller != logger.caller {
			logger.caller = e.caller
		}
		if e.timestamp != logger.timestamp {
			logger.timestamp = e.timestamp
		}
		if e.buf != nil {
			logger.context = enc.AppendObjectData(logger.context, e.buf)
		}
	}
}

// Formatter update logger's formatter.
func Formatter(formatter LogFormatter) LoggerOption {
	return func(logger *Logger) {
		logger.formatter = formatter
	}
}

// TimestampFieldName update logger's timestampFieldName.
func TimestampFieldName(timestampFieldName string) LoggerOption {
	return func(logger *Logger) {
		logger.timestampFieldName = timestampFieldName
	}
}

// LevelFieldName update logger's levelFieldName.
func LevelFieldName(levelFieldName string) LoggerOption {
	return func(logger *Logger) {
		logger.levelFieldName = levelFieldName
	}
}

// MessageFieldName update logger's messageFieldName.
func MessageFieldName(messageFieldName string) LoggerOption {
	return func(logger *Logger) {
		logger.messageFieldName = messageFieldName
	}
}

// ErrorFieldName update logger's errorFieldName.
func ErrorFieldName(errorFieldName string) LoggerOption {
	return func(logger *Logger) {
		logger.errorFieldName = errorFieldName
	}
}

// CallerFieldName update logger's callerFieldName.
func CallerFieldName(callerFieldName string) LoggerOption {
	return func(logger *Logger) {
		logger.callerFieldName = callerFieldName
	}
}

// CallerSkipFrameCount update logger's callerSkipFrameCount.
func CallerSkipFrameCount(callerSkipFrameCount int) LoggerOption {
	return func(logger *Logger) {
		logger.callerSkipFrameCount = callerSkipFrameCount
	}
}

// ErrorStackFieldName update logger's errorStackFieldName.
func ErrorStackFieldName(errorStackFieldName string) LoggerOption {
	return func(logger *Logger) {
		logger.errorStackFieldName = errorStackFieldName
	}
}

// TimeFieldFormat update logger's timeFieldFormat.
func TimeFieldFormat(timeFieldFormat string) LoggerOption {
	return func(logger *Logger) {
		logger.timeFieldFormat = timeFieldFormat
	}
}

// TimestampFunc update logger's timestampFunc.
func TimestampFunc(timestampFunc func() time.Time) LoggerOption {
	return func(logger *Logger) {
		logger.timestampFunc = timestampFunc
	}
}

var (
	// DurationFieldUnit defines the unit for time.Duration type fields added
	// using the Duration method.
	DurationFieldUnit = time.Millisecond

	// DurationFieldInteger renders Dur fields as integer instead of float if
	// set to true.
	DurationFieldInteger = false

	// ErrorHandler is called whenever rz fails to write an event on its
	// output. If not set, an error is printed on the stderr. This handler must
	// be thread safe and non-blocking.
	ErrorHandler func(err error)

	// ErrorStackMarshaler extract the stack from err if any.
	ErrorStackMarshaler func(err error) interface{}

	// ErrorMarshalFunc allows customization of global error marshaling
	ErrorMarshalFunc = func(err error) interface{} {
		return err
	}
)
