package rz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

const (
	cReset    = 0
	cBold     = 1
	cRed      = 31
	cGreen    = 32
	cYellow   = 33
	cBlue     = 34
	cMagenta  = 35
	cCyan     = 36
	cGray     = 37
	cDarkGray = 90
)

// FormatterConsole prettify output for human cosumption
func FormatterConsole() LogFormatter {
	return func(ev *Event) ([]byte, error) {
		var event map[string]interface{}
		var ret = new(bytes.Buffer)

		d := json.NewDecoder(bytes.NewReader(ev.buf))
		d.UseNumber()
		err := d.Decode(&event)
		if err != nil {
			return ret.Bytes(), err
		}

		lvlColor := cReset
		level := "????"
		if l, ok := event[DefaultLevelFieldName].(string); ok {
			lvlColor = levelColor(l)
			level = strings.ToUpper(l)[0:4]
		}

		message := ""
		if m, ok := event[DefaultMessageFieldName].(string); ok {
			message = m
		}

		timestamp := ""
		if t, ok := event[DefaultTimestampFieldName].(string); ok {
			timestamp = t
		}

		ret.WriteString(fmt.Sprintf("%-20s |%-4s|",
			timestamp,
			colorize(level, lvlColor),
		))
		if message != "" {
			ret.WriteString(" " + message)
		}

		fields := make([]string, 0, len(event))
		for field := range event {
			switch field {
			case DefaultTimestampFieldName, DefaultMessageFieldName, DefaultLevelFieldName:
				continue
			}

			fields = append(fields, field)
		}

		sort.Strings(fields)
		for _, field := range fields {
			if needsQuote(field) {
				field = strconv.Quote(field)
			}
			fmt.Fprintf(ret, " %s=", colorize(field, lvlColor))

			switch value := event[field].(type) {
			case string:
				if len(value) == 0 {
					ret.WriteString("\"\"")
				} else if needsQuote(value) {
					ret.WriteString(strconv.Quote(value))
				} else {
					ret.WriteString(value)
				}
			default:
				b, err := json.Marshal(value)
				if err != nil {
					return ret.Bytes(), err
				}
				fmt.Fprint(ret, string(b))
			}
		}

		ret.WriteByte('\n')

		return ret.Bytes(), nil
	}
}

func colorize(s interface{}, color int) string {
	return fmt.Sprintf("\x1b[%dm%v\x1b[0m", color, s)
}

func levelColor(level string) int {
	switch level {
	case "debug":
		return cMagenta
	case "info":
		return cCyan
	case "warning":
		return cYellow
	case "error", "fatal", "panic":
		return cRed
	default:
		return cReset
	}
}

func needsQuote(s string) bool {
	for i := range s {
		if s[i] < 0x20 || s[i] > 0x7e || s[i] == ' ' || s[i] == '\\' || s[i] == '"' {
			return true
		}
	}
	return false
}
