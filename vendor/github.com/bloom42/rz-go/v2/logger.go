package rz

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/bloom42/rz-go/v2/internal/json"
)

// A Logger represents an active logging object that generates lines
// of JSON output to an io.Writer. Each logging operation makes a single
// call to the Writer's Write method. There is no guaranty on access
// serialization to the Writer. If your Writer is not thread safe,
// you may consider a sync wrapper.
type Logger struct {
	writer               LevelWriter
	stack                bool
	caller               bool
	timestamp            bool
	level                LogLevel
	sampler              LogSampler
	context              []byte
	hooks                []LogHook
	timestampFieldName   string
	levelFieldName       string
	messageFieldName     string
	errorFieldName       string
	callerFieldName      string
	callerSkipFrameCount int
	errorStackFieldName  string
	timeFieldFormat      string
	formatter            LogFormatter
	timestampFunc        func() time.Time
	contextMutext        *sync.Mutex
	encoder              Encoder
}

// New creates a root logger with given options. If the output writer implements
// the LevelWriter interface, the WriteLevel method will be called instead of the Write
// one. Default writer is os.Stdout
//
// Each logging operation makes a single call to the Writer's Write method. There is no
// guaranty on access serialization to the Writer. If your Writer is not thread safe,
// you may consider using sync wrapper.
func New(options ...LoggerOption) Logger {
	logger := Logger{
		writer:               levelWriterAdapter{os.Stdout},
		level:                DebugLevel,
		timestamp:            true,
		timestampFieldName:   DefaultTimestampFieldName,
		levelFieldName:       DefaultLevelFieldName,
		messageFieldName:     DefaultMessageFieldName,
		errorFieldName:       DefaultErrorFieldName,
		callerFieldName:      DefaultCallerFieldName,
		callerSkipFrameCount: DefaultCallerSkipFrameCount,
		errorStackFieldName:  DefaultErrorStackFieldName,
		timeFieldFormat:      DefaultTimeFieldFormat,
		timestampFunc:        DefaultTimestampFunc,
		contextMutext:        &sync.Mutex{},
		encoder:              json.Encoder{},
	}
	return logger.With(options...)
}

// Nop returns a disabled logger for which all operation are no-op.
func Nop() Logger {
	return New(Writer(nil), Level(Disabled))
}

// With create a new copy of the logger and apply all the options to the new logger
func (l Logger) With(options ...LoggerOption) Logger {
	oldContext := l.context
	l.context = make([]byte, 0, 500)
	l.contextMutext = &sync.Mutex{}
	if oldContext != nil {
		l.context = append(l.context, oldContext...)
	}
	for _, option := range options {
		option(&l)
	}
	return l
}

// GetLevel returns the current log level.
func (l *Logger) GetLevel() LogLevel {
	return l.level
}

// Debug logs a new message with debug level.
func (l *Logger) Debug(message string, fields ...Field) {
	l.logEvent(DebugLevel, message, nil, fields)
}

// Info logs a new message with info level.
func (l *Logger) Info(message string, fields ...Field) {
	l.logEvent(InfoLevel, message, nil, fields)
}

// Warn logs a new message with warn level.
func (l *Logger) Warn(message string, fields ...Field) {
	l.logEvent(WarnLevel, message, nil, fields)
}

// Error logs a message with error level.
func (l *Logger) Error(message string, fields ...Field) {
	l.logEvent(ErrorLevel, message, nil, fields)
}

// Fatal logs a new message with fatal level. The os.Exit(1) function
// is then called, which terminates the program immediately.
func (l *Logger) Fatal(message string, fields ...Field) {
	l.logEvent(FatalLevel, message, func(msg string) { os.Exit(1) }, fields)
}

// Panic logs a new message with panic level. The panic() function
// is then called, which stops the ordinary flow of a goroutine.
func (l *Logger) Panic(message string, fields ...Field) {
	l.logEvent(PanicLevel, message, func(msg string) { panic(msg) }, fields)
}

// Log logs a new message with no level. Setting GlobalLevel to Disabled
// will still disable events produced by this method.
func (l *Logger) Log(message string, fields ...Field) {
	l.logEvent(NoLevel, message, nil, fields)
}

// NewDict creates an Event to be used with the Dict method.
// Call usual field methods like Str, Int etc to add fields to this
// event and give it as argument the *Event.Dict method.
func (l *Logger) NewDict(fields ...Field) *Event {
	e := newEvent(nil, 0)
	copyInternalLoggerFieldsToEvent(l, e)
	e.Append(fields...)
	return e
}

// Write implements the io.Writer interface. This is useful to set as a writer
// for the standard library log.
//
//    log := rz.New()
//    stdlog.SetFlags(0)
//    stdlog.SetOutput(log)
//    stdlog.Print("hello world")
func (l Logger) Write(p []byte) (n int, err error) {
	n = len(p)
	if n > 0 && p[n-1] == '\n' {
		// Trim CR added by stdlog.
		p = p[0 : n-1]
	}
	l.Log(string(p), nil)
	return
}

func (l *Logger) logEvent(level LogLevel, message string, done func(string), fields []Field) {
	enabled := l.should(level)
	if !enabled {
		return
	}
	e := newEvent(l.writer, level)
	e.ch = l.hooks
	copyInternalLoggerFieldsToEvent(l, e)
	if level != NoLevel {
		e.string(e.levelFieldName, level.String())
	}
	if l.context != nil && len(l.context) > 0 {
		e.buf = enc.AppendObjectData(e.buf, l.context)
	}

	for i := range fields {
		fields[i](e)
	}

	writeEvent(e, message, done)
}

func writeEvent(e *Event, msg string, done func(string)) {
	// run hooks
	if len(e.ch) > 0 {
		e.ch[0].Run(e, e.level, msg)
		if len(e.ch) > 1 {
			for _, hook := range e.ch[1:] {
				hook.Run(e, e.level, msg)
			}
		}
	}

	if done != nil {
		defer done(msg)
	}

	// if hooks didn't disabled our event, continue
	if e.level != Disabled {
		var err error

		if e.timestamp {
			e.buf = enc.AppendTime(enc.AppendKey(e.buf, e.timestampFieldName), e.timestampFunc(), e.timeFieldFormat)
		}

		if msg != "" {
			e.buf = enc.AppendString(enc.AppendKey(e.buf, e.messageFieldName), msg)
		}
		if e.caller {
			_, file, line, ok := runtime.Caller(e.callerSkipFrameCount)
			if ok {
				e.buf = enc.AppendString(enc.AppendKey(e.buf, e.callerFieldName), file+":"+strconv.Itoa(line))
			}
		}

		// end json payload
		e.buf = enc.AppendEndMarker(e.buf)
		e.buf = enc.AppendLineBreak(e.buf)
		if e.formatter != nil {
			e.buf, err = e.formatter(e)
		}
		if e.w != nil {
			_, err = e.w.WriteLevel(e.level, e.buf)
		}

		putEvent(e)

		if err != nil {
			if ErrorHandler != nil {
				ErrorHandler(err)
			} else {
				fmt.Fprintf(os.Stderr, "rz: could not write event: %v\n", err)
			}
		}
	}

}

// should returns true if the log event should be logged.
func (l *Logger) should(lvl LogLevel) bool {
	if lvl < l.level {
		return false
	}
	if l.sampler != nil {
		return l.sampler.Sample(lvl)
	}
	return true
}

// Append the fields to the internal logger's context.
// It does not create a noew copy of the logger and rely on a mutex to enable thread safety,
// so `With(Fields(fields...))` often is preferable.
func (l *Logger) Append(fields ...Field) {
	e := newEvent(l.writer, l.level)
	e.buf = nil
	copyInternalLoggerFieldsToEvent(l, e)
	for i := range fields {
		fields[i](e)
	}
	l.contextMutext.Lock()
	if e.stack != l.stack {
		l.stack = e.stack
	}
	if e.caller != l.caller {
		l.caller = e.caller
	}
	if e.timestamp != l.timestamp {
		l.timestamp = e.timestamp
	}
	if e.buf != nil {
		l.context = enc.AppendObjectData(l.context, e.buf)
	}
	l.contextMutext.Unlock()
}

func copyInternalLoggerFieldsToEvent(l *Logger, e *Event) {
	e.stack = l.stack
	e.caller = l.caller
	e.timestamp = l.timestamp
	e.timestampFieldName = l.timestampFieldName
	e.levelFieldName = l.levelFieldName
	e.messageFieldName = l.messageFieldName
	e.errorFieldName = l.errorFieldName
	e.callerFieldName = l.callerFieldName
	e.timeFieldFormat = l.timeFieldFormat
	e.errorStackFieldName = l.errorStackFieldName
	e.callerSkipFrameCount = l.callerSkipFrameCount
	e.formatter = l.formatter
	e.timestampFunc = l.timestampFunc
	e.encoder = l.encoder
}
