package logging

import (
	"context"
	"net/http"
	"os"

	"github.com/bloom42/rz-go/v2"
	"github.com/bloom42/rz-go/v2/rzhttp"
	"github.com/grupawp/tensorflow-deploy/exterr"
	"github.com/segmentio/ksuid"
)

// logger as global
var logger = newLogger()

// newLogger creates logger with default settings
func newLogger() rz.Logger {
	l := rz.New(rz.CallerSkipFrameCount(4))
	l = l.With(rz.Level(rz.InfoLevel))

	return l
}

// SetDebugLevel sets logger with debug level
func SetDebugLevel() {
	logger = logger.With(rz.Level(rz.DebugLevel))
}

// SetFatalLevel sets logger with fatal level.
func SetFatalLevel() {
	logger = logger.With(rz.Level(rz.FatalLevel))
}

// SetDisabled disable logger.
// Be carefull fatal and panic dont stop application.
func SetDisabled() {
	logger = logger.With(rz.Level(rz.Disabled))
}

// Debug logs a new message with debug level.
func Debug(ctx context.Context, message string) {
	logger.Debug(message)
}

// Info logs a new message with info level.
func Info(ctx context.Context, message string) {
	logger.Info(message)
}

// Warn logs a new message with warn level.
func Warn(ctx context.Context, message string) {
	logger.Warn(message)
}

// Error logs a message and code with error level.
func Error(ctx context.Context, message, code string) {
	fields := fields(ctx, code)
	logger.Error(message, fields...)
}

// ErrorWithStack logs a message of error and all previous errors.
func ErrorWithStack(ctx context.Context, err error) {
	requestID, ok := ctx.Value(rzhttp.RequestIDCtxKey).(string)
	if !ok {
		return
	}

	if newErr, ok := err.(*exterr.Error); ok {
		for _, item := range newErr.DumpStackWithID(requestID) {
			logger.Error(item)
		}
		return
	}
	logger.Error(err.Error())
}

// ErrorWithStackWithoutRequestID logs a message of error without requestID
func ErrorWithStackWithoutRequestID(ctx context.Context, err error) {
	dumpStackLogger(ctx, err)
}

// FatalErrorWithStack logs a message of error and all previous errors and call Fatal .
func FatalErrorWithStack(ctx context.Context, err error, code string) {
	dumpStackLogger(ctx, err)
	Fatal(ctx, err.Error(), code)
}

// Fatal logs a new message and code with fatal level.
func Fatal(ctx context.Context, message, code string) {
	fields := fields(ctx, code)
	logger.Fatal(message, fields...)
	os.Exit(1)
}

// Panic logs a new message and code with panic level.
func Panic(ctx context.Context, message, code string) {
	fields := fields(ctx, code)
	logger.Panic(message, fields...)
}

// HTTPCtxValuesMiddleware ->
func HTTPCtxValuesMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// requestID
		requestID := GenerateRequestID()
		w.Header().Set("X-Request-ID", requestID)
		ctx = context.WithValue(ctx, rzhttp.RequestIDCtxKey, requestID)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// HTTPRequestMiddleware ->
func HTTPRequestMiddleware() func(next http.Handler) http.Handler {
	return rzhttp.Handler(logger)
}

func fields(ctx context.Context, code string) []rz.Field {
	fields := make([]rz.Field, 0)
	// debug fields
	fields = append(fields, rz.Caller(true))
	// required fields
	fields = append(fields, rz.String("code", code))
	// context fields
	fields = fieldsFromCtx(ctx, fields)

	return fields
}

// fieldsFromCtx gets hard-coded fields from context
func fieldsFromCtx(ctx context.Context, fields []rz.Field) []rz.Field {
	// requestID
	if requestID, ok := ctx.Value(rzhttp.RequestIDCtxKey).(string); ok {
		fields = append(fields, rz.String("request_id", requestID))
	}

	return fields
}

func dumpStackLogger(ctx context.Context, err error) {
	if newErr, ok := err.(*exterr.Error); ok {
		for _, item := range newErr.DumpStack() {
			logger.Error(item)
		}
	}
}

// GenerateRequestID generates unique request_id
func GenerateRequestID() string {
	id := ksuid.New()

	return id.String()
}
