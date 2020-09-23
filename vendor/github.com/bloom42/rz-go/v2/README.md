<p align="center">
  <h3 align="center">rz</h3>
  <p align="center">RipZap - The fastest structured, leveled JSON logger for Go ⚡️. Dependency free.</p>
</p>

--------

[Make logging great again](https://kerkour.com/post/logging/)

[![GoDoc](https://godoc.org/github.com/bloom42/rz-go?status.svg)](https://godoc.org/github.com/bloom42/rz-go)
[![Build Status](https://travis-ci.org/bloom42/rz-go.svg?branch=master)](https://travis-ci.org/bloom42/rz-go)
[![GitHub release](https://img.shields.io/github/release/bloom42/rz-go.svg)](https://github.com/bloom42/rz-go/releases)
<!-- [![Coverage](http://gocover.io/_badge/github.com/bloom42/rz-go)](http://gocover.io/github.com/bloom42/rz-go) -->

![Console logging](docs/example_screenshot.png)

The rz package provides a fast and simple logger dedicated to JSON output avoiding allocations and reflection..

Uber's [zap](https://godoc.org/go.uber.org/zap) and rs's [zerolog](https://godoc.org/github.com/rs/zerolog)
libraries pioneered this approach.

ripzap is an update of zerolog taking this concept to the next level with a **simpler** to use and **safer**
API and even better [performance](#benchmarks).

To keep the code base and the API simple, ripzap focuses on efficient structured logging only.
Pretty logging on the console is made possible using the provided (but inefficient)
[`Formatter`s](https://godoc.org/github.com/bloom42/rz-go#LogFormatter).


1. [Quickstart](#quickstart)
2. [Configuration](#configuration)
3. [Field types](#field-types)
4. [HTTP Handler](#http-handler)
5. [Examples](#examples)
6. [Benchmarks](#benchmarks)
7. [Versions](#versions)
8. [Contributing](#contributing)
9. [License](#license)

-------------------

## Quickstart

```go
package main

import (
	"os"

	"github.com/bloom42/rz-go/v2"
	"github.com/bloom42/rz-go/v2/log"
)

func main() {

	env := os.Getenv("GO_ENV")
	hostname, _ := os.Hostname()

	// update global logger's context fields
	log.SetLogger(log.With(rz.Fields(
		rz.String("hostname", hostname), rz.String("environment", env),
	)))

	if env == "production" {
		log.SetLogger(log.With(rz.Level(rz.InfoLevel)))
	}

	log.Info("info from logger", rz.String("hello", "world"))
	// {"level":"info","hostname":"","environment":"","hello":"world","timestamp":"2019-02-07T09:30:07Z","message":"info from logger"}
}
```


## Configuration

### Logger
```go
// Writer update logger's writer.
func Writer(writer io.Writer) LoggerOption {}
// Level update logger's level.
func Level(lvl LogLevel) LoggerOption {}
// Sampler update logger's sampler.
func Sampler(sampler LogSampler) LoggerOption {}
// AddHook appends hook to logger's hook
func AddHook(hook LogHook) LoggerOption {}
// Hooks replaces logger's hooks
func Hooks(hooks ...LogHook) LoggerOption {}
// With replaces logger's context fields
func With(fields func(*Event)) LoggerOption {}
// Stack enable/disable stack in error messages.
func Stack(enableStack bool) LoggerOption {}
// Timestamp enable/disable timestamp logging in error messages.
func Timestamp(enableTimestamp bool) LoggerOption {}
// Caller enable/disable caller field in message messages.
func Caller(enableCaller bool) LoggerOption {}
// Formatter update logger's formatter.
func Formatter(formatter LogFormatter) LoggerOption {}
// TimestampFieldName update logger's timestampFieldName.
func TimestampFieldName(timestampFieldName string) LoggerOption {}
// LevelFieldName update logger's levelFieldName.
func LevelFieldName(levelFieldName string) LoggerOption {}
// MessageFieldName update logger's messageFieldName.
func MessageFieldName(messageFieldName string) LoggerOption {}
// ErrorFieldName update logger's errorFieldName.
func ErrorFieldName(errorFieldName string) LoggerOption {}
// CallerFieldName update logger's callerFieldName.
func CallerFieldName(callerFieldName string) LoggerOption {}
// CallerSkipFrameCount update logger's callerSkipFrameCount.
func CallerSkipFrameCount(callerSkipFrameCount int) LoggerOption {}
// ErrorStackFieldName update logger's errorStackFieldName.
func ErrorStackFieldName(errorStackFieldName string) LoggerOption {}
// TimeFieldFormat update logger's timeFieldFormat.
func TimeFieldFormat(timeFieldFormat string) LoggerOption {}
// TimestampFunc update logger's timestampFunc.
func TimestampFunc(timestampFunc func() time.Time) LoggerOption {}
```

### Global
```go
var (
	// DurationFieldUnit defines the unit for time.Duration type fields added
	// using the Duration method.
	DurationFieldUnit = time.Millisecond

	// DurationFieldInteger renders Duration fields as integer instead of float if
	// set to true.
	DurationFieldInteger = false

	// ErrorHandler is called whenever rz fails to write an event on its
	// output. If not set, an error is printed on the stderr. This handler must
	// be thread safe and non-blocking.
	ErrorHandler func(err error)
)
```


## Field Types

### Standard Types

* `String`
* `Bool`
* `Int`, `Int8`, `Int16`, `Int32`, `Int64`
* `Uint`, `Uint8`, `Uint16`, `Uint32`, `Uint64`
* `Float32`, `Float64`

### Advanced Fields

* `Err`: Takes an `error` and render it as a string using the `logger.errorFieldName` field name.
* `Error`: Adds a field with a `error`.
* `Timestamp`: Insert a timestamp field with `logger.timestampFieldName` field name and formatted using `logger.timeFieldFormat`.
* `Time`: Adds a field with the time formated with the `logger.timeFieldFormat`.
* `Duration`: Adds a field with a `time.Duration`.
* `Dict`: Adds a sub-key/value as a field of the event.
* `Interface`: Uses reflection to marshal the type.


## HTTP Handler

See the [bloom42/rz-go/rzhttp](https://godoc.org/github.com/bloom42/rz-go/rzhttp) package or the
[example here](https://github.com/bloom42/rz-go/tree/master/examples/http).


## Examples

See the [examples](https://github.com/bloom42/rz-go/tree/master/examples) folder.


## Benchmarks

```
$ make benchmarks
cd benchmarks && ./run.sh
goos: linux
goarch: amd64
pkg: github.com/bloom42/rz-go/benchmarks
BenchmarkDisabledWithoutFields/sirupsen/logrus-4         	100000000	        17.2 ns/op	      16 B/op	       1 allocs/op
BenchmarkDisabledWithoutFields/uber-go/zap-4             	30000000	        37.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkDisabledWithoutFields/rs/zerolog-4              	1000000000	         2.35 ns/op	       0 B/op	       0 allocs/op
BenchmarkDisabledWithoutFields/bloom42/rz-go-4           	2000000000	         1.67 ns/op	       0 B/op	       0 allocs/op
BenchmarkWithoutFields/sirupsen/logrus-4                 	  300000	      4521 ns/op	    1126 B/op	      24 allocs/op
BenchmarkWithoutFields/uber-go/zap-4                     	 5000000	       322 ns/op	       0 B/op	       0 allocs/op
BenchmarkWithoutFields/rs/zerolog-4                      	10000000	       118 ns/op	       0 B/op	       0 allocs/op
BenchmarkWithoutFields/bloom42/rz-go-4                   	10000000	       113 ns/op	       0 B/op	       0 allocs/op
Benchmark10Context/sirupsen/logrus-4                     	  100000	     18139 ns/op	    3246 B/op	      50 allocs/op
Benchmark10Context/uber-go/zap-4                         	 5000000	       331 ns/op	       0 B/op	       0 allocs/op
Benchmark10Context/rs/zerolog-4                          	10000000	       133 ns/op	       0 B/op	       0 allocs/op
Benchmark10Context/bloom42/rz-go-4                       	10000000	       124 ns/op	       0 B/op	       0 allocs/op
Benchmark10Fields/sirupsen/logrus-4                      	  100000	     21705 ns/op	    4025 B/op	      54 allocs/op
Benchmark10Fields/uber-go/zap-4                          	  500000	      2400 ns/op	      80 B/op	       1 allocs/op
Benchmark10Fields/rs/zerolog-4                           	 1000000	      1914 ns/op	     640 B/op	       6 allocs/op
Benchmark10Fields/bloom42/rz-go-4                        	 1000000	      1794 ns/op	     512 B/op	       3 allocs/op
Benchmark10Fields10Context/sirupsen/logrus-4             	  100000	     23873 ns/op	    4550 B/op	      53 allocs/op
Benchmark10Fields10Context/uber-go/zap-4                 	  500000	      2381 ns/op	      80 B/op	       1 allocs/op
Benchmark10Fields10Context/rs/zerolog-4                  	 1000000	      1914 ns/op	     640 B/op	       6 allocs/op
Benchmark10Fields10Context/bloom42/rz-go-4               	 1000000	      1807 ns/op	     512 B/op	       3 allocs/op
PASS
ok  	github.com/bloom42/rz-go/benchmarks	36.783s
```

## Versions

For v2 (current) see the [master branch](https://github.com/bloom42/rz-go).

For v1 see the [v1 branch](https://github.com/bloom42/rz-go/tree/v1).


## Contributing

See [https://opensource.bloom.sh/contributing](https://opensource.bloom.sh/contributing)


## License

See `LICENSE.txt` and [https://opensource.bloom.sh/licensing](https://opensource.bloom.sh/licensing)

From an original work by [rs](https://github.com/rs): [zerolog](https://github.com/rs/zerolog) - commit aa55558e4cb2e8f05cd079342d430f77e946d00a
