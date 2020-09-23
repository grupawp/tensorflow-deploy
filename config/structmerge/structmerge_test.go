package structmerge

import (
	"reflect"
	"testing"
)

type (
	params struct{}
)

func TestNewStructMerge(t *testing.T) {
	tests := []struct {
		name string
		want *StructMerge
	}{
		{
			name: "test",
			want: &StructMerge{
				sources:    NewSources(),
				MergedMeta: make([]MergedMeta, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewStructMerge(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewStructMerge() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStructMerge_WithSource(t *testing.T) {
	type fields struct {
		sources    *Sources
		mergedMeta []MergedMeta
	}
	type args struct {
		name string
		data interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *StructMerge
	}{
		{
			name: "test",
			fields: fields{
				sources:    NewSources(),
				mergedMeta: make([]MergedMeta, 0),
			},
			args: args{
				name: "some-source",
				data: &params{},
			},
			want: NewStructMerge().WithSource("some-source", &params{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := &StructMerge{
				sources:    tt.fields.sources,
				MergedMeta: tt.fields.mergedMeta,
			}
			if got := sm.WithSource(tt.args.name, tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StructMerge.WithSource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStructMerge_WithSources(t *testing.T) {
	type fields struct {
		sources    *Sources
		mergedMeta []MergedMeta
	}
	type args struct {
		s *Sources
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *StructMerge
	}{
		{
			name: "test",
			fields: fields{
				sources:    NewSources(),
				mergedMeta: make([]MergedMeta, 0),
			},
			args: args{
				s: NewSources().Add("defaults", &params{}).Add("cli", &params{}),
			},
			want: NewStructMerge().WithSource("defaults", &params{}).WithSource("cli", &params{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := &StructMerge{
				sources:    tt.fields.sources,
				MergedMeta: tt.fields.mergedMeta,
			}
			if got := sm.WithSources(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StructMerge.WithSources() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStructMerge_Merge(t *testing.T) {
	type fields struct {
		sources    *Sources
		mergedMeta []MergedMeta
	}
	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test - no sources given",
			fields: fields{
				sources:    NewSources(),
				mergedMeta: make([]MergedMeta, 0),
			},
			args: args{
				data: &params{},
			},
			wantErr: true,
		},
		{
			name: "test - data not a ptr",
			fields: fields{
				sources:    NewSources().Add("defaults", params{}),
				mergedMeta: make([]MergedMeta, 0),
			},
			args: args{
				data: params{},
			},
			wantErr: true,
		},
		{
			name: "test - sources given",
			fields: fields{
				sources:    NewSources().Add("defaults", &params{}),
				mergedMeta: make([]MergedMeta, 0),
			},
			args: args{
				data: &params{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := &StructMerge{
				sources:    tt.fields.sources,
				MergedMeta: tt.fields.mergedMeta,
			}
			if err := sm.Merge(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("StructMerge.Merge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStructMerge_MergeWithMerger(t *testing.T) {
	type fields struct {
		sources    *Sources
		mergedMeta []MergedMeta
	}
	type args struct {
		data interface{}
		m    Merger
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "test - proper implementation of merger",
			fields: fields{
				sources:    NewSources().Add("defaults", &params{}),
				mergedMeta: make([]MergedMeta, 0),
			},
			args: args{
				data: &params{},
				m:    newDefaultMerger(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sm := &StructMerge{
				sources:    tt.fields.sources,
				MergedMeta: tt.fields.mergedMeta,
			}
			if err := sm.MergeWithMerger(tt.args.data, tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("StructMerge.MergeWithMerger() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
