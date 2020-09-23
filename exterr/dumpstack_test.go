package exterr

import (
	"strings"
	"testing"

	goError "errors"
)

const (
	errMsg1 = "example error1"
	errMsg2 = "example error2"
	errMsg3 = "example error3"
)

func prepareErrors(messages []string) error {
	var result error
	for k, v := range messages {
		if k == 0 {
			result = NewErrorWithErr(goError.New(v))
			continue
		}
		result = WrapWithErr(result, NewErrorWithErr(goError.New(v)))
	}
	return result
}

func Test_DumpStack(t *testing.T) {
	type args struct {
		errorMessages []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "DumpStack OK",
			args: args{
				errorMessages: []string{errMsg1, errMsg2, errMsg3},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			preparedErrors := prepareErrors(tt.args.errorMessages)
			newErr, ok := preparedErrors.(*Error)
			if !ok {
				t.Errorf("preparedErrors are empty, test: %s", tt.name)
			}

			resultDumpStack := newErr.DumpStack()
			if len(resultDumpStack) != len(tt.args.errorMessages) {
				t.Errorf("size result (%d) != size data entries (%d), test name: %s", len(resultDumpStack), len(tt.args.errorMessages), tt.name)
				return
			}

			for k, v := range resultDumpStack {
				if !strings.Contains(v, tt.args.errorMessages[k]) {
					t.Errorf("result is invalid for data entries, test name: %s, item: %s", tt.name, tt.args.errorMessages[k])
				}
			}
		})

	}

}

func Test_DumpStackWithID(t *testing.T) {
	type args struct {
		errorMessages []string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "DumpStackWithID OK",
			args: args{
				errorMessages: []string{errMsg1, errMsg2, errMsg3},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := "DumpStackWithID"
			preparedErrors := prepareErrors(tt.args.errorMessages)
			newErr, ok := preparedErrors.(*Error)
			if !ok {
				t.Errorf("preparedErrors are empty, test: %s", tt.name)
			}

			resultDumpStack := newErr.DumpStackWithID(id)
			if len(resultDumpStack) != len(tt.args.errorMessages) {
				t.Errorf("size result (%d) != size data entries (%d), test name: %s", len(resultDumpStack), len(tt.args.errorMessages), tt.name)
				return
			}

			for k, v := range resultDumpStack {
				if !strings.Contains(v, tt.args.errorMessages[k]) {
					t.Errorf("result is invalid for data entries, test name: %s, item: %s", tt.name, tt.args.errorMessages[k])
				}
			}
		})

	}

}
