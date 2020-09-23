package rest

import (
	"bytes"
	"golang.org/x/net/context"
	"io"
	"reflect"
	"testing"
)

func Test_calculateChecksum(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "test 1 - good pattern",
			args: args{r: bytes.NewReader([]byte("The sums are computed as described in FIPS-180-4"))},
			want: "5d403c7674225a0b81e45b008656545479333dc30ff5a18a34188419bd557e13",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateChecksum(tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateChecksum() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isChecksumValid(t *testing.T) {
	type args struct {
		r            io.Reader
		formChecksum string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test 1 - checksum OK",
			args: args{
				r:            bytes.NewReader([]byte("The sums are computed as described in FIPS-180-4")),
				formChecksum: "5d403c7674225a0b81e45b008656545479333dc30ff5a18a34188419bd557e13",
			},
			want: true,
		},
		{
			name: "test 2 - OK, but no formChecksum found",
			args: args{
				r:            bytes.NewReader([]byte("The sums are computed as described in FIPS-180-4")),
				formChecksum: "",
			},
			want: true,
		},
		{
			name: "test 3 - wrong checksum",
			args: args{
				r:            bytes.NewReader([]byte("The sums are computed as described in FIPS-180-4")),
				formChecksum: "25a0b81e45b008656545479333dc30ff5a18a34188419bd557e13",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isChecksumValid(context.Background(), tt.args.r, tt.args.formChecksum); got != tt.want {
				t.Errorf("isChecksumValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
