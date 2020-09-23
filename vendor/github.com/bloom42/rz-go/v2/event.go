package rz

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"sync"
	"time"
)

var eventPool = &sync.Pool{
	New: func() interface{} {
		return &Event{
			buf: make([]byte, 0, 500),
		}
	},
}

// Event represents a log event. It is instanced by one of the level method.
type Event struct {
	buf                  []byte
	w                    LevelWriter
	level                LogLevel
	done                 func(msg string)
	stack                bool      // enable error stack trace
	caller               bool      // enable caller field
	timestamp            bool      // enable timestamp
	ch                   []LogHook // hooks from context
	timestampFieldName   string
	levelFieldName       string
	messageFieldName     string
	errorFieldName       string
	callerFieldName      string
	errorStackFieldName  string
	timeFieldFormat      string
	callerSkipFrameCount int
	formatter            LogFormatter
	timestampFunc        func() time.Time
	encoder              Encoder
}

func putEvent(e *Event) {
	// Proper usage of a sync.Pool requires each entry to have approximately
	// the same memory cost. To obtain this property when the stored type
	// contains a variably-sized buffer, we add a hard limit on the maximum buffer
	// to place back in the pool.
	//
	// See https://golang.org/issue/23199
	const maxSize = 1 << 16 // 64KiB
	if cap(e.buf) > maxSize {
		return
	}
	eventPool.Put(e)
}

// LogObjectMarshaler provides a strongly-typed and encoding-agnostic interface
// to be implemented by types used with Event/Context's Object methods.
type LogObjectMarshaler interface {
	MarshalRzObject(*Event)
}

// LogArrayMarshaler provides a strongly-typed and encoding-agnostic interface
// to be implemented by types used with Event/Context's Array methods.
type logArrayMarshaler interface {
	MarshalRzArray(*array)
}

func newEvent(w LevelWriter, level LogLevel) *Event {
	e := eventPool.Get().(*Event)
	e.buf = e.buf[:0]
	e.ch = nil
	e.buf = enc.AppendBeginMarker(e.buf)
	e.w = w
	e.level = level
	return e
}

// Enabled return false if the *Event is going to be filtered out by
// log level or sampling.
func (e *Event) Enabled() bool {
	return e.level != Disabled
}

// Append the given fields to the event
func (e *Event) Append(fields ...Field) {
	for i := range fields {
		fields[i](e)
	}
}

// Fields returns the fields from the event.
// Note that this call is very expensive and should be used sparingly.
func (e *Event) Fields() (map[string]interface{}, error) {
	var fields map[string]interface{}

	r := io.MultiReader(bytes.NewReader(e.buf), bytes.NewReader([]byte{'}', '\n'}))

	d := json.NewDecoder(r)
	err := d.Decode(&fields)
	if err != nil {
		return nil, err
	}

	return fields, nil
}

// Discard disables the event
func (e *Event) discard() {
	e.level = Disabled
}

// Fields is a helper function to use a map to set fields using type assertion.
func (e *Event) fields(fields map[string]interface{}) {
	e.buf = e.appendFields(e.buf, fields)
}

// Dict adds the field key with a dict to the event context.
// Use rz.Dict() to create the dictionary.
func (e *Event) dict(key string, dict *Event) {
	dict.buf = enc.AppendEndMarker(dict.buf)
	e.buf = append(enc.AppendKey(e.buf, key), dict.buf...)
	putEvent(dict)
}

func newDict() *Event {
	return newEvent(nil, 0)
}

// Array adds the field key with an array to the event context.
// Use Event.Arr() to create the array or pass a type that
// implement the LogArrayMarshaler interface.
func (e *Event) array(key string, arr logArrayMarshaler) {
	e.buf = enc.AppendKey(e.buf, key)
	var a *array
	if aa, ok := arr.(*array); ok {
		a = aa
	} else {
		a = e.arr()
		arr.MarshalRzArray(a)
	}
	e.buf = a.write(e.buf)
}

func (e *Event) appendObject(obj LogObjectMarshaler) {
	e.buf = enc.AppendBeginMarker(e.buf)
	obj.MarshalRzObject(e)
	e.buf = enc.AppendEndMarker(e.buf)
}

// Object marshals an object that implement the LogObjectMarshaler interface.
func (e *Event) object(key string, obj LogObjectMarshaler) {
	e.buf = enc.AppendKey(e.buf, key)
	e.appendObject(obj)
}

// embedObject marshals an object that implement the LogObjectMarshaler interface.
func (e *Event) embedObject(obj LogObjectMarshaler) {
	obj.MarshalRzObject(e)
}

// String adds the field key with val as a string to the *Event context.
func (e *Event) string(key, val string) {
	e.buf = enc.AppendString(enc.AppendKey(e.buf, key), val)
}

// Strings adds the field key with vals as a []string to the *Event context.
func (e *Event) strings(key string, vals []string) {
	e.buf = enc.AppendStrings(enc.AppendKey(e.buf, key), vals)
}

// Bytes adds the field key with val as a string to the *Event context.
//
// Runes outside of normal ASCII ranges will be hex-encoded in the resulting
// JSON.
func (e *Event) bytes(key string, val []byte) {
	e.buf = enc.AppendBytes(enc.AppendKey(e.buf, key), val)
}

// Hex adds the field key with val as a hex string to the *Event context.
func (e *Event) hex(key string, val []byte) {
	e.buf = enc.AppendHex(enc.AppendKey(e.buf, key), val)
}

// RawJSON adds already encoded JSON to the log line under key.
//
// No sanity check is performed on b; it must not contain carriage returns and
// be valid JSON.
func (e *Event) rawJSON(key string, b []byte) {
	e.buf = appendJSON(enc.AppendKey(e.buf, key), b)
}

// Error adds the field key with serialized err to the *Event context.
// If err is nil, no field is added.
func (e *Event) error(key string, err error) {
	switch m := ErrorMarshalFunc(err).(type) {
	case nil:
	case LogObjectMarshaler:
		e.object(key, m)
	case error:
		e.string(key, m.Error())
	case string:
		e.string(key, m)
	default:
		e.iinterface(key, m)
	}
}

// Errors adds the field key with errs as an array of serialized errors to the
// *Event context.
func (e *Event) errors(key string, errs []error) {
	arr := e.arr()
	for _, err := range errs {
		switch m := ErrorMarshalFunc(err).(type) {
		case LogObjectMarshaler:
			arr = arr.Object(m)
		case error:
			arr = arr.Err(m)
		case string:
			arr = arr.Str(m)
		default:
			arr = arr.Interface(m)
		}
	}

	e.array(key, arr)
}

// Err adds the field "error" with serialized err to the *Event context.
// If err is nil, no field is added.
// To customize the key name, uze rz.ErrorFieldName.
////
// If Stack() has been called before and rz.ErrorStackMarshaler is defined,
// the err is passed to ErrorStackMarshaler and the result is appended to the
// rz.ErrorStackFieldName.
func (e *Event) err(err error) {
	if e.stack && ErrorStackMarshaler != nil {
		switch m := ErrorStackMarshaler(err).(type) {
		case nil:
		case LogObjectMarshaler:
			e.object(e.errorStackFieldName, m)
		case error:
			e.string(e.errorStackFieldName, m.Error())
		case string:
			e.string(e.errorStackFieldName, m)
		default:
			e.iinterface(e.errorStackFieldName, m)
		}
	}
	e.error(e.errorFieldName, err)
}

// Stack enables stack trace printing for the error passed to Err().
//
// ErrorStackMarshaler must be set for this method to do something.
func (e *Event) enableStack(enable bool) {
	e.stack = enable
}

// Bool adds the field key with val as a bool to the *Event context.
func (e *Event) bool(key string, b bool) {
	e.buf = enc.AppendBool(enc.AppendKey(e.buf, key), b)
}

// Bools adds the field key with val as a []bool to the *Event context.
func (e *Event) bools(key string, b []bool) {
	e.buf = enc.AppendBools(enc.AppendKey(e.buf, key), b)
}

// Int adds the field key with i as a int to the *Event context.
func (e *Event) int(key string, i int) {
	e.buf = enc.AppendInt(enc.AppendKey(e.buf, key), i)
}

// Ints adds the field key with i as a []int to the *Event context.
func (e *Event) ints(key string, i []int) {
	e.buf = enc.AppendInts(enc.AppendKey(e.buf, key), i)
}

// Int8 adds the field key with i as a int8 to the *Event context.
func (e *Event) int8(key string, i int8) {
	e.buf = enc.AppendInt8(enc.AppendKey(e.buf, key), i)
}

// Ints8 adds the field key with i as a []int8 to the *Event context.
func (e *Event) ints8(key string, i []int8) {
	e.buf = enc.AppendInts8(enc.AppendKey(e.buf, key), i)
}

// Int16 adds the field key with i as a int16 to the *Event context.
func (e *Event) int16(key string, i int16) {
	e.buf = enc.AppendInt16(enc.AppendKey(e.buf, key), i)
}

// Ints16 adds the field key with i as a []int16 to the *Event context.
func (e *Event) ints16(key string, i []int16) {
	e.buf = enc.AppendInts16(enc.AppendKey(e.buf, key), i)
}

// Int32 adds the field key with i as a int32 to the *Event context.
func (e *Event) int32(key string, i int32) {
	e.buf = enc.AppendInt32(enc.AppendKey(e.buf, key), i)
}

// Ints32 adds the field key with i as a []int32 to the *Event context.
func (e *Event) ints32(key string, i []int32) {
	e.buf = enc.AppendInts32(enc.AppendKey(e.buf, key), i)
}

// Int64 adds the field key with i as a int64 to the *Event context.
func (e *Event) int64(key string, i int64) {
	e.buf = enc.AppendInt64(enc.AppendKey(e.buf, key), i)
}

// Ints64 adds the field key with i as a []int64 to the *Event context.
func (e *Event) ints64(key string, i []int64) {
	e.buf = enc.AppendInts64(enc.AppendKey(e.buf, key), i)
}

// Uint adds the field key with i as a uint to the *Event context.
func (e *Event) uint(key string, i uint) {
	e.buf = enc.AppendUint(enc.AppendKey(e.buf, key), i)
}

// Uints adds the field key with i as a []int to the *Event context.
func (e *Event) uints(key string, i []uint) {
	e.buf = enc.AppendUints(enc.AppendKey(e.buf, key), i)
}

// Uint8 adds the field key with i as a uint8 to the *Event context.
func (e *Event) uint8(key string, i uint8) {
	e.buf = enc.AppendUint8(enc.AppendKey(e.buf, key), i)
}

// Uints8 adds the field key with i as a []int8 to the *Event context.
func (e *Event) uints8(key string, i []uint8) {
	e.buf = enc.AppendUints8(enc.AppendKey(e.buf, key), i)
}

// Uint16 adds the field key with i as a uint16 to the *Event context.
func (e *Event) uint16(key string, i uint16) {
	e.buf = enc.AppendUint16(enc.AppendKey(e.buf, key), i)
}

// Uints16 adds the field key with i as a []int16 to the *Event context.
func (e *Event) uints16(key string, i []uint16) {
	e.buf = enc.AppendUints16(enc.AppendKey(e.buf, key), i)
}

// Uint32 adds the field key with i as a uint32 to the *Event context.
func (e *Event) uint32(key string, i uint32) {
	e.buf = enc.AppendUint32(enc.AppendKey(e.buf, key), i)
}

// Uints32 adds the field key with i as a []int32 to the *Event context.
func (e *Event) uints32(key string, i []uint32) {
	e.buf = enc.AppendUints32(enc.AppendKey(e.buf, key), i)
}

// Uint64 adds the field key with i as a uint64 to the *Event context.
func (e *Event) uint64(key string, i uint64) {
	e.buf = enc.AppendUint64(enc.AppendKey(e.buf, key), i)
}

// Uints64 adds the field key with i as a []int64 to the *Event context.
func (e *Event) uints64(key string, i []uint64) {
	e.buf = enc.AppendUints64(enc.AppendKey(e.buf, key), i)
}

// Float32 adds the field key with f as a float32 to the *Event context.
func (e *Event) float32(key string, f float32) {
	e.buf = enc.AppendFloat32(enc.AppendKey(e.buf, key), f)
}

// Floats32 adds the field key with f as a []float32 to the *Event context.
func (e *Event) floats32(key string, f []float32) {
	e.buf = enc.AppendFloats32(enc.AppendKey(e.buf, key), f)
}

// Float64 adds the field key with f as a float64 to the *Event context.
func (e *Event) float64(key string, f float64) {
	e.buf = enc.AppendFloat64(enc.AppendKey(e.buf, key), f)
}

// Floats64 adds the field key with f as a []float64 to the *Event context.
func (e *Event) floats64(key string, f []float64) {
	e.buf = enc.AppendFloats64(enc.AppendKey(e.buf, key), f)
}

// Timestamp adds the current local time as UNIX timestamp to the *Event context with the
// logger.TimestampFieldName key.
// func (e *Event) Timestamp() {
// 	e.timestamp = false
// 	e.buf = enc.AppendTime(enc.AppendKey(e.buf, e.timestampFieldName), e.timestampFunc(), e.timeFieldFormat)
// 	return e
// }
func (e *Event) enableTimestamp(enable bool) {
	e.timestamp = enable
}

// Time adds the field key with t formated as string using rz.TimeFieldFormat.
func (e *Event) time(key string, t time.Time) {
	e.buf = enc.AppendTime(enc.AppendKey(e.buf, key), t, e.timeFieldFormat)
}

// Times adds the field key with t formated as string using rz.TimeFieldFormat.
func (e *Event) times(key string, t []time.Time) {
	e.buf = enc.AppendTimes(enc.AppendKey(e.buf, key), t, e.timeFieldFormat)
}

// Duration adds the field key with duration d stored as rz.DurationFieldUnit.
// If rz.DurationFieldInteger is true, durations are rendered as integer
// instead of float.
func (e *Event) duration(key string, d time.Duration) {
	e.buf = enc.AppendDuration(enc.AppendKey(e.buf, key), d, DurationFieldUnit, DurationFieldInteger)
}

// Durations adds the field key with duration d stored as rz.DurationFieldUnit.
// If rz.DurationFieldInteger is true, durations are rendered as integer
// instead of float.
func (e *Event) durations(key string, d []time.Duration) {
	e.buf = enc.AppendDurations(enc.AppendKey(e.buf, key), d, DurationFieldUnit, DurationFieldInteger)
}

// Interface adds the field key with i marshaled using reflection.
func (e *Event) iinterface(key string, i interface{}) {
	if obj, ok := i.(LogObjectMarshaler); ok {
		e.object(key, obj)
	}
	e.buf = enc.AppendInterface(enc.AppendKey(e.buf, key), i)
}

// enableCaller adds the file:line of the caller with the rz.CallerFieldName key.
func (e *Event) enableCaller(enable bool) {
	e.caller = enable
}

// ip adds IPv4 or IPv6 Address to the event
func (e *Event) ip(key string, ip net.IP) {
	e.buf = enc.AppendIPAddr(enc.AppendKey(e.buf, key), ip)
}

// ipNet adds IPv4 or IPv6 Prefix (address and mask) to the event
func (e *Event) ipNet(key string, pfx net.IPNet) {
	e.buf = enc.AppendIPPrefix(enc.AppendKey(e.buf, key), pfx)
}

// hardwareAddr adds MAC address to the event
func (e *Event) hardwareAddr(key string, ha net.HardwareAddr) {
	e.buf = enc.AppendMACAddr(enc.AppendKey(e.buf, key), ha)
}
