package storage

import "testing"

var (
	directoryRegexp = []string{`^variables/$`, `^variables/variables\.data-[0-9]{5}-of-[0-9]{5}$`, `^variables/variables\.index$`, `^saved_model\..*$`, `^README\.md$`}
)

func Test_modelStorage_isDirectoryLayoutValid(t *testing.T) {
	type args struct {
		directoryLayout []string
		directoryRegexp []string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test 1 - directory structure is valid",
			args: args{
				directoryLayout: []string{"variables/", "variables/variables.data-12345-of-12345", "variables/variables.index", "saved_model.pb", "README.md"},
				directoryRegexp: directoryRegexp,
			},
			want: true,
		},
		{
			name: "test 2 - variables.data is invalid",
			args: args{
				directoryLayout: []string{"variables/", "variables/variables.data-abcd1-of-zzzzz", "variables/variables.index", "saved_model.pb", "README.md"},
				directoryRegexp: directoryRegexp,
			},
			want: false,
		},
		{
			name: "test 3 - README.md file is missing",
			args: args{
				directoryLayout: []string{"variables/", "variables/variables.data-12345-of-12345", "variables/variables.index", "saved_model.pb"},
				directoryRegexp: directoryRegexp,
			},
			want: false,
		},
		{
			name: "test 4 - wrong variables.data file name",
			args: args{
				directoryLayout: []string{"variables/", "variables/variables.data-12-of-12345", "variables/variables.index", "saved_model.pb", "README.md"},
				directoryRegexp: directoryRegexp,
			},
			want: false,
		},
		{
			name: "test 5 - wrong vvvvvvariables file name",
			args: args{
				directoryLayout: []string{"vvvvvvariables/", "variables/variables.data-12345-of-12345", "variables/variables.index", "saved_model.pb", "README.md"},
				directoryRegexp: directoryRegexp,
			},
			want: false,
		},
		{
			name: "test 6 - wrong variables/variables.index-bak file name, all regexps use ^$ pattern",
			args: args{
				directoryLayout: []string{"variables/", "variables/variables.data-12345-of-12345", "variables/variables.index-bak", "saved_model.pb", "README.md"},
				directoryRegexp: directoryRegexp,
			},
			want: false,
		},
		{
			name: "test 7 - directory structure is valid",
			args: args{
				directoryLayout: []string{"assets", "variables/", "variables/variables.data-12345-of-12345", "variables/variables.index", "saved_model.pb", "README.md"},
				directoryRegexp: directoryRegexp,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDirectoryLayoutValid(tt.args.directoryLayout, tt.args.directoryRegexp); got != tt.want {
				t.Errorf("modelStorage.isDirectoryLayoutValid() = %v, want %v", got, tt.want)
			}
		})
	}
}
