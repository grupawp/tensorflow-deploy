package rz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"
)

// FormatterCLI prettify output suitable for command-line interfaces.
func FormatterCLI() LogFormatter {
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
		level := ""
		if l, ok := event[ev.levelFieldName].(string); ok {
			lvlColor = levelColor(l)
			level = l
		}

		message := ""
		if m, ok := event[ev.messageFieldName].(string); ok {
			message = m
		}

		if level != "" {
			ret.WriteString(colorize(levelSymbol(level), lvlColor))
		}
		ret.WriteString(message)

		fields := make([]string, 0, len(event))
		for field := range event {
			switch field {
			case ev.timestampFieldName, ev.messageFieldName, ev.levelFieldName:
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
			case time.Time:
				ret.WriteString(value.Format(time.RFC3339))
			default:
				b, err := json.Marshal(value)
				if err != nil {
					fmt.Fprintf(ret, "[error: %v]", err)
				} else {
					fmt.Fprint(ret, string(b))
				}
			}

		}

		return ret.Bytes(), nil
	}
}

func levelSymbol(level string) string {
	switch level {
	case "info":
		return "✔ "
	case "warning":
		return "⚠ "
	case "error", "fatal":
		return "✘ "
	default:
		return "• "
	}
}
