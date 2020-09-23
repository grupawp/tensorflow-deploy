package rz

import (
	"net"
	"sync"
	"time"
)

var arrayPool = &sync.Pool{
	New: func() interface{} {
		return &array{
			buf: make([]byte, 0, 500),
		}
	},
}

// Array is used to prepopulate an array of items
// which can be re-used to add to log messages.
type array struct {
	buf             []byte
	timeFieldFormat string
}

func putArray(a *array) {
	// Proper usage of a sync.Pool requires each entry to have approximately
	// the same memory cost. To obtain this property when the stored type
	// contains a variably-sized buffer, we add a hard limit on the maximum buffer
	// to place back in the pool.
	//
	// See https://golang.org/issue/23199
	const maxSize = 1 << 16 // 64KiB
	if cap(a.buf) > maxSize {
		return
	}
	arrayPool.Put(a)
}

// Arr creates an array to be added to an Event or Context.
func (e *Event) arr() *array {
	a := arrayPool.Get().(*array)
	a.buf = a.buf[:0]
	a.timeFieldFormat = e.timeFieldFormat
	return a
}

// MarshalRzArray method here is no-op - since data is
// already in the needed format.
func (*array) MarshalRzArray(*array) {
}

func (a *array) write(dst []byte) []byte {
	dst = enc.AppendArrayStart(dst)
	if len(a.buf) > 0 {
		dst = append(append(dst, a.buf...))
	}
	dst = enc.AppendArrayEnd(dst)
	putArray(a)
	return dst
}

// Object marshals an object that implement the LogObjectMarshaler
// interface and append append it to the array.
func (a *array) Object(obj LogObjectMarshaler) *array {
	e := newDict()
	e.timeFieldFormat = a.timeFieldFormat
	obj.MarshalRzObject(e)
	e.buf = enc.AppendEndMarker(e.buf)
	a.buf = append(enc.AppendArrayDelim(a.buf), e.buf...)
	putEvent(e)
	return a
}

// Str append append the val as a string to the array.
func (a *array) Str(val string) *array {
	a.buf = enc.AppendString(enc.AppendArrayDelim(a.buf), val)
	return a
}

// Bytes append append the val as a string to the array.
func (a *array) Bytes(val []byte) *array {
	a.buf = enc.AppendBytes(enc.AppendArrayDelim(a.buf), val)
	return a
}

// Hex append append the val as a hex string to the array.
func (a *array) Hex(val []byte) *array {
	a.buf = enc.AppendHex(enc.AppendArrayDelim(a.buf), val)
	return a
}

// Err serializes and appends the err to the array.
func (a *array) Err(err error) *array {
	marshaled := ErrorMarshalFunc(err)
	switch m := marshaled.(type) {
	case LogObjectMarshaler:
		e := newEvent(nil, 0)
		e.buf = e.buf[:0]
		e.appendObject(m)
		a.buf = append(enc.AppendArrayDelim(a.buf), e.buf...)
		putEvent(e)
	case error:
		a.buf = enc.AppendString(enc.AppendArrayDelim(a.buf), m.Error())
	case string:
		a.buf = enc.AppendString(enc.AppendArrayDelim(a.buf), m)
	default:
		a.buf = enc.AppendInterface(enc.AppendArrayDelim(a.buf), m)
	}

	return a
}

// Bool append append the val as a bool to the array.
func (a *array) Bool(b bool) *array {
	a.buf = enc.AppendBool(enc.AppendArrayDelim(a.buf), b)
	return a
}

// Int append append i as a int to the array.
func (a *array) Int(i int) *array {
	a.buf = enc.AppendInt(enc.AppendArrayDelim(a.buf), i)
	return a
}

// Int8 append append i as a int8 to the array.
func (a *array) Int8(i int8) *array {
	a.buf = enc.AppendInt8(enc.AppendArrayDelim(a.buf), i)
	return a
}

// Int16 append append i as a int16 to the array.
func (a *array) Int16(i int16) *array {
	a.buf = enc.AppendInt16(enc.AppendArrayDelim(a.buf), i)
	return a
}

// Int32 append append i as a int32 to the array.
func (a *array) Int32(i int32) *array {
	a.buf = enc.AppendInt32(enc.AppendArrayDelim(a.buf), i)
	return a
}

// Int64 append append i as a int64 to the array.
func (a *array) Int64(i int64) *array {
	a.buf = enc.AppendInt64(enc.AppendArrayDelim(a.buf), i)
	return a
}

// Uint append append i as a uint to the array.
func (a *array) Uint(i uint) *array {
	a.buf = enc.AppendUint(enc.AppendArrayDelim(a.buf), i)
	return a
}

// Uint8 append append i as a uint8 to the array.
func (a *array) Uint8(i uint8) *array {
	a.buf = enc.AppendUint8(enc.AppendArrayDelim(a.buf), i)
	return a
}

// Uint16 append append i as a uint16 to the array.
func (a *array) Uint16(i uint16) *array {
	a.buf = enc.AppendUint16(enc.AppendArrayDelim(a.buf), i)
	return a
}

// Uint32 append append i as a uint32 to the array.
func (a *array) Uint32(i uint32) *array {
	a.buf = enc.AppendUint32(enc.AppendArrayDelim(a.buf), i)
	return a
}

// Uint64 append append i as a uint64 to the array.
func (a *array) Uint64(i uint64) *array {
	a.buf = enc.AppendUint64(enc.AppendArrayDelim(a.buf), i)
	return a
}

// Float32 append append f as a float32 to the array.
func (a *array) Float32(f float32) *array {
	a.buf = enc.AppendFloat32(enc.AppendArrayDelim(a.buf), f)
	return a
}

// Float64 append append f as a float64 to the array.
func (a *array) Float64(f float64) *array {
	a.buf = enc.AppendFloat64(enc.AppendArrayDelim(a.buf), f)
	return a
}

// Time append append t formated as string using rz.TimeFieldFormat.
func (a *array) Time(t time.Time) *array {
	a.buf = enc.AppendTime(enc.AppendArrayDelim(a.buf), t, a.timeFieldFormat)
	return a
}

// Dur append append d to the array.
func (a *array) Dur(d time.Duration) *array {
	a.buf = enc.AppendDuration(enc.AppendArrayDelim(a.buf), d, DurationFieldUnit, DurationFieldInteger)
	return a
}

// Interface append append i marshaled using reflection.
func (a *array) Interface(i interface{}) *array {
	if obj, ok := i.(LogObjectMarshaler); ok {
		return a.Object(obj)
	}
	a.buf = enc.AppendInterface(enc.AppendArrayDelim(a.buf), i)
	return a
}

// IPAddr adds IPv4 or IPv6 address to the array
func (a *array) IPAddr(ip net.IP) *array {
	a.buf = enc.AppendIPAddr(enc.AppendArrayDelim(a.buf), ip)
	return a
}

// IPPrefix adds IPv4 or IPv6 Prefix (IP + mask) to the array
func (a *array) IPPrefix(pfx net.IPNet) *array {
	a.buf = enc.AppendIPPrefix(enc.AppendArrayDelim(a.buf), pfx)
	return a
}

// MACAddr adds a MAC (Ethernet) address to the array
func (a *array) MACAddr(ha net.HardwareAddr) *array {
	a.buf = enc.AppendMACAddr(enc.AppendArrayDelim(a.buf), ha)
	return a
}
