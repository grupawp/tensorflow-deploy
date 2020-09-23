package plaintext

import (
	"testing"

	"golang.org/x/net/context"
)

func Test_skipLine(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name           string
		args           args
		wantResultTrue bool
	}{

		{
			name: "skip, empty line ",
			args: args{
				line: "",
			},
			wantResultTrue: true,
		},
		{
			name: "shouldn't skip, line is valid, at the beginning are spaces",
			args: args{
				line: "    tfs-instance-name",
			},
			wantResultTrue: false,
		},
		{
			name: "shouldn't skip,  line is valid",
			args: args{
				line: "tfs-instance-name",
			},
			wantResultTrue: false,
		},
		{
			name: "should skip, at the beginning is ;",
			args: args{
				line: ";    tfs-instance-name",
			},
			wantResultTrue: true,
		},
		{
			name: "should skip, at the beginning is #",
			args: args{
				line: "#    tfs-instance-name",
			},
			wantResultTrue: true,
		},
	}
	for _, tt := range tests {
		pt, _ := NewPlaintext(context.Background(), "")
		t.Run(tt.name, func(t *testing.T) {
			if result := pt.skipLine(tt.args.line); (result != true) == tt.wantResultTrue {
				t.Errorf("skipLine() result = %t name test %s", result, tt.name)
			}
		})
	}

}

func Test_extractInstanceName(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "empty line",
			args: args{
				line: "",
			},
			wantErr: true,
		},
		{
			name: "invalid instanceName",
			args: args{
				line: "instanceName",
			},
			wantErr: true,
		},
		{
			name: "invalid instanceName",
			args: args{
				line: "instanceName instanceName",
			},
			wantErr: true,
		},
		{
			name: "valid servableIDInstanceName",
			args: args{
				line: "tfs-instance-name",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pt, _ := NewPlaintext(context.Background(), "")
			if _, err := pt.extractInstanceName(tt.args.line); (err != nil) != tt.wantErr {
				t.Errorf("extractInstanceName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

}

func Test_isValidInstanceName(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name           string
		args           args
		wantResultTrue bool
	}{
		{
			name: "valid instanceName",
			args: args{
				line: "tfs-moo-ppp",
			},
			wantResultTrue: true,
		},
		{
			name: "valid instanceName",
			args: args{
				line: "tfs-moomoomoomoomoomoomoo-ppp",
			},
			wantResultTrue: true,
		},
		{
			name: "invalid prefix",
			args: args{
				line: "tfx-moomoomoomoomoomoomoo-ppp",
			},
			wantResultTrue: false,
		},
		{
			name: "invalid char",
			args: args{
				line: "tfs-`-ppp",
			},
			wantResultTrue: false,
		},
		{
			name: "invalid char",
			args: args{
				line: "tfs-(-ppp",
			},
			wantResultTrue: false,
		},
	}
	for _, tt := range tests {
		pt, _ := NewPlaintext(context.Background(), "")
		t.Run(tt.name, func(t *testing.T) {
			if result := pt.isValidInstanceName(tt.args.line); (result != true) == tt.wantResultTrue {
				t.Errorf("isValidInstanceName() result = %t name test %s", result, tt.name)
			}
		})
	}

}

func Test_isInstanceValid(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name           string
		args           args
		wantResultTrue bool
	}{
		{
			name: "valid address 1",
			args: args{
				line: "192.168.1.1:8500",
			},
			wantResultTrue: true,
		},

		{
			name: "valid address 2",
			args: args{
				line: "192.168.1.1:65001",
			},
			wantResultTrue: true,
		},
		{
			name: "valid address 3",
			args: args{
				line: "192.168.1.1:1",
			},
			wantResultTrue: true,
		},
		{
			name: "invalid address 4",
			args: args{
				line: "192.168.1.1:-1",
			},
			wantResultTrue: false,
		},
		{
			name: "invalid address 5",
			args: args{
				line: "192.168.1.1:00023",
			},
			wantResultTrue: false,
		},
	}
	for _, tt := range tests {
		pt, _ := NewPlaintext(context.Background(), "")
		t.Run(tt.name, func(t *testing.T) {
			if result := pt.isValidInstance(tt.args.line); (result != true) == tt.wantResultTrue {
				t.Errorf("isValidInstance() result = %t name test %s", result, tt.name)
			}
		})
	}

}

func Test_parseLine(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "valid line",
			args: args{
				line: "tfs-mta-dev 10.10.10.21:8500 	  10.10.2.21:8500",
			},
			wantErr: false,
		},
		{
			name: "valid line",
			args: args{
				line: "tfs-mta-dev 10.10.10.21:8500 	10.10.2.21:8500 	10.10.3.21:8500 10.10.2.21:8500 10.10.4.21:8500 10.10.5.21:8500 10.10.6.21:8500 10.10.7.21:8500 10.10.8.21:8500 ",
			},
			wantErr: false,
		},
		{
			name: "invalid address",
			args: args{
				line: "tfs-mta-dev 10.10.10.21:8500 :	10.10.2.21:8500 	10.10.3.21:8500 10.10.2.21:8500 10.10.4.21:8500 10.10.5.21:8500 10.10.6.21:8500 10.10.7.21:8500 10.10.8.21:8500 ",
			},
			wantErr: true,
		},
		{
			name: "invalid char",
			args: args{
				line: "tfs-`mta-dev 10.10.10.21:8500 	10.10.2.21:8500  ",
			},
			wantErr: true,
		},
		{
			name: "invalid char",
			args: args{
				line: "tfs-`mta-dev 10.10.10.21:8500 	10.10.2.21:8500 ",
			},
			wantErr: true,
		},
		{
			name: "invalid prefix",
			args: args{
				line: "tfo-mta-dev 10.10.10.21:8500 	10.10.2.21:8500 ",
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			args: args{
				line: "tfs-mta-dev 10.10.10.21:08500 10.10.2.21:00001",
			},
			wantErr: true,
		},
	}
	for k, tt := range tests {
		ctx := context.Background()
		pt, _ := NewPlaintext(ctx, "")
		t.Run(tt.name, func(t *testing.T) {
			if _, err := pt.readInstances(ctx, k, tt.args.line); (err != nil) != tt.wantErr {
				t.Errorf("parseLine() error = %v, wantErr %v, name %s ", err, tt.wantErr, tt.name)
			}
		})
	}

}
