package structmerge

import (
	"reflect"
	"testing"
)

type (
	Config struct {
		MainStruct  RootStruct
		OtherStruct NodeStruct
	}

	OtherConfig struct {
		MainStruct  RootStruct
		OtherStruct NodeStruct
	}

	BadConfig struct {
		ParamStr    *string
		ParamInt    *int
		MainStruct  *RootStruct
		OtherStruct *NodeStruct
	}

	BadConfigAlternate struct {
		ParamStr    string
		ParamInt    int
		MainStruct  RootStruct
		OtherStruct NodeStruct
	}

	RootStruct struct {
		ParamStr  *string
		ParamInt  *int
		ParamBool *bool
	}

	NodeStruct struct {
		NodeParamStr *string
	}

	Settable struct {
		ParamInt int
		ParamStr string
	}
)

func dataConfig(sourceName string) *Config {
	paramStr := sourceName + ":paramStr"
	paramInt := 10
	paramBool := true
	nodeParamStr := sourceName + "nodeParamStr"

	return &Config{
		MainStruct: RootStruct{
			ParamStr:  &paramStr,
			ParamInt:  &paramInt,
			ParamBool: &paramBool,
		},
		OtherStruct: NodeStruct{
			NodeParamStr: &nodeParamStr,
		},
	}
}

func dataOtherConfig(sourceName string) *OtherConfig {
	paramStr := sourceName + ":paramStr"
	paramInt := 10
	paramBool := true
	nodeParamStr := sourceName + "nodeParamStr"

	return &OtherConfig{
		MainStruct: RootStruct{
			ParamStr:  &paramStr,
			ParamInt:  &paramInt,
			ParamBool: &paramBool,
		},
		OtherStruct: NodeStruct{
			NodeParamStr: &nodeParamStr,
		},
	}
}

func dataBadConfig(sourceName string) *BadConfig {
	paramStr := sourceName + ":paramStr"
	paramInt := 10
	paramBool := true
	nodeParamStr := sourceName + "nodeParamStr"

	return &BadConfig{
		ParamStr: &paramStr,
		ParamInt: &paramInt,
		MainStruct: &RootStruct{
			ParamStr:  &paramStr,
			ParamInt:  &paramInt,
			ParamBool: &paramBool,
		},
		OtherStruct: &NodeStruct{
			NodeParamStr: &nodeParamStr,
		},
	}
}

func dataBadConfigAlternate(sourceName string) *BadConfigAlternate {
	paramStr := sourceName + ":paramStr"
	paramInt := 10
	paramBool := true
	nodeParamStr := sourceName + "nodeParamStr"

	return &BadConfigAlternate{
		ParamStr: paramStr,
		ParamInt: paramInt,
		MainStruct: RootStruct{
			ParamStr:  &paramStr,
			ParamInt:  &paramInt,
			ParamBool: &paramBool,
		},
		OtherStruct: NodeStruct{
			NodeParamStr: &nodeParamStr,
		},
	}
}

func Test_newDefaultMerge(t *testing.T) {
	tests := []struct {
		name string
		want *Merge
	}{
		{
			name: "test",
			want: &Merge{
				mergedMeta: make([]MergedMeta, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newDefaultMerger(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newDefaultMerge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMerge_Merge(t *testing.T) {
	wantedMergeMeta, _ := newDefaultMerger().Merge(&Config{}, NewSources().Add("defaults", dataConfig("defaults")).Add("cli", dataConfig("cli")))

	type fields struct {
		mergedMeta []MergedMeta
	}
	type args struct {
		data interface{}
		s    *Sources
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []MergedMeta
		wantErr bool
	}{
		{
			name: "test - passed",
			fields: fields{
				mergedMeta: make([]MergedMeta, 0),
			},
			args: args{
				data: &Config{},
				s:    NewSources().Add("defaults", dataConfig("defaults")).Add("cli", dataConfig("cli")),
			},
			want:    wantedMergeMeta,
			wantErr: false,
		},
		{
			name:   "test - invalid type between data and sources",
			fields: fields{},
			args: args{
				data: &RootStruct{},
				s:    NewSources().Add("defaults", dataConfig("defaults")).Add("cli", dataOtherConfig("cli")),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "test - invalid type on sources",
			fields: fields{},
			args: args{
				data: &Config{},
				s:    NewSources().Add("defaults", dataConfig("defaults")).Add("cli", dataOtherConfig("cli")),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "test - bad struct",
			fields: fields{},
			args: args{
				data: &BadConfig{},
				s:    NewSources().Add("defaults", dataBadConfig("defaults")).Add("cli", dataBadConfig("cli")),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "test - bad struct alternate",
			fields: fields{},
			args: args{
				data: &BadConfigAlternate{},
				s:    NewSources().Add("defaults", dataBadConfigAlternate("defaults")).Add("cli", dataBadConfigAlternate("cli")),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "test - nil struct",
			fields: fields{},
			args: args{
				data: &Config{},
				s:    NewSources().Add("defaults", &Config{}).Add("cli", &Config{}),
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Merge{
				mergedMeta: tt.fields.mergedMeta,
			}
			got, err := m.Merge(tt.args.data, tt.args.s)
			if (err != nil) != tt.wantErr {
				t.Errorf("Merge.Merge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Merge.Merge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMerge_isSettable(t *testing.T) {
	rv := make(map[string]reflect.Value, 0)
	rv["string"] = reflect.ValueOf("string")
	rv["int"] = reflect.ValueOf(int(10))
	rv["uint"] = reflect.ValueOf(uint(10))
	rv["float"] = reflect.ValueOf(float64(10.3))
	rv["bool"] = reflect.ValueOf(bool(false))
	rv["stringptr"] = reflect.ValueOf(new(string))

	type fields struct {
		mergedMeta []MergedMeta
	}
	type args struct {
		v reflect.Value
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{name: "test - string", args: args{v: rv["string"]}, want: true},
		{name: "test - int", args: args{v: rv["int"]}, want: true},
		{name: "test - uint", args: args{v: rv["uint"]}, want: true},
		{name: "test - float", args: args{v: rv["float"]}, want: true},
		{name: "test - bool", args: args{v: rv["bool"]}, want: true},
		{name: "test - string ptr", args: args{v: rv["stringptr"]}, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Merge{
				mergedMeta: tt.fields.mergedMeta,
			}
			if got := m.isSettable(tt.args.v); got != tt.want {
				t.Errorf("Merge.isSettable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMerge_checkDataTypes(t *testing.T) {
	type fields struct {
		mergedMeta []MergedMeta
	}
	type args struct {
		data interface{}
		s    *Sources
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test - invalid type between data and sources",
			fields: fields{
				mergedMeta: make([]MergedMeta, 0),
			},
			args: args{
				data: &RootStruct{},
				s:    NewSources().Add("defaults", dataConfig("defaults")).Add("cli", dataConfig("cli")),
			},
			wantErr: true,
		},
		{
			name: "test - invalid type on sources",
			fields: fields{
				mergedMeta: make([]MergedMeta, 0),
			},
			args: args{
				data: &Config{},
				s:    NewSources().Add("defaults", dataConfig("defaults")).Add("cli", dataOtherConfig("cli")),
			},
			wantErr: true,
		},
		{
			name: "test - bad source",
			fields: fields{
				mergedMeta: make([]MergedMeta, 0),
			},
			args: args{
				data: &Config{},
				s:    NewSources().Add("defaults", dataConfig("defaults")).Add("cli", nil),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Merge{
				mergedMeta: tt.fields.mergedMeta,
			}
			if err := m.checkDataTypes(tt.args.data, tt.args.s); (err != nil) != tt.wantErr {
				t.Errorf("Merge.checkDataTypes() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMerge_getDataType(t *testing.T) {
	type fields struct {
		mergedMeta []MergedMeta
	}
	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    reflect.Type
		wantErr bool
	}{
		{
			name: "test - nil data",
			fields: fields{
				mergedMeta: make([]MergedMeta, 0),
			},
			args: args{
				data: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test - data not a ptr",
			fields: fields{
				mergedMeta: make([]MergedMeta, 0),
			},
			args: args{
				data: Config{},
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test - data not a struct",
			fields: fields{
				mergedMeta: make([]MergedMeta, 0),
			},
			args: args{
				data: dataConfig("defaults").MainStruct.ParamStr,
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Merge{
				mergedMeta: tt.fields.mergedMeta,
			}
			got, err := m.getDataType(tt.args.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("Merge.getDataType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Merge.getDataType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newTraversedValues(t *testing.T) {
	tests := []struct {
		name string
		want *traversedValues
	}{
		{
			name: "test",
			want: &traversedValues{
				sources: make([]*traversedSource, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newTraversedValues(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newTraversedValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_traversedValues_joinFieldPath(t *testing.T) {
	type fields struct {
		data      reflect.Value
		sources   []*traversedSource
		fieldPath string
	}
	type args struct {
		path     string
		partPath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "test - empty partPath",
			fields:  fields{},
			args:    args{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &traversedValues{
				data:      tt.fields.data,
				sources:   tt.fields.sources,
				fieldPath: tt.fields.fieldPath,
			}
			if err := v.joinFieldPath(tt.args.path, tt.args.partPath); (err != nil) != tt.wantErr {
				t.Errorf("traversedValues.joinFieldPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
