package rz

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"time"
)

// FormatterLogfmt prettify output for human consumption, using the logfmt format.
func FormatterLogfmt() LogFormatter {
	return func(ev *Event) ([]byte, error) {
		var event map[string]interface{}
		var ret = new(bytes.Buffer)

		d := json.NewDecoder(bytes.NewReader(ev.buf))
		d.UseNumber()
		err := d.Decode(&event)
		if err != nil {
			return ret.Bytes(), err
		}

		fields := make([]string, 0, len(event))
		for field := range event {
			fields = append(fields, field)
		}

		sort.Strings(fields)
		for _, field := range fields {
			if needsQuote(field) {
				field = strconv.Quote(field)
			}
			fmt.Fprintf(ret, " %s=", field)

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
					fmt.Fprint(ret, strconv.Quote(string(b)))
				}
			}

		}

		return ret.Bytes(), nil
	}
}
