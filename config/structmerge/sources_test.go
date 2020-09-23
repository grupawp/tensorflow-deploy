package structmerge

import (
	"reflect"
	"testing"
)

func TestNewSources(t *testing.T) {
	tests := []struct {
		name string
		want *Sources
	}{
		{
			name: "test",
			want: &Sources{
				sources: make([]*Source, 0),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSources(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSources() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSources_Add(t *testing.T) {
	type fields struct {
		sources []*Source
	}
	type args struct {
		name string
		data interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Sources
	}{
		{
			name: "test",
			fields: fields{
				sources: make([]*Source, 0),
			},
			args: args{
				name: "some-source",
				data: &params{},
			},
			want: NewSources().Add("some-source", &params{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sources{
				sources: tt.fields.sources,
			}
			if got := s.Add(tt.args.name, tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sources.Add() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSources_Sources(t *testing.T) {
	type fields struct {
		sources []*Source
	}
	tests := []struct {
		name   string
		fields fields
		want   []*Source
	}{
		{
			name: "test",
			fields: fields{
				sources: make([]*Source, 0),
			},
			want: make([]*Source, 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Sources{
				sources: tt.fields.sources,
			}
			if got := s.Sources(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Sources.Sources() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSources_check(t *testing.T) {
	type fields struct {
		sources []*Source
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "test - check passed",
			fields: fields{
				sources: NewSources().Add("some-source", &params{}).Sources(),
			},
			wantErr: false,
		},
		{
			name: "test - check not passed",
			fields: fields{
				sources: make([]*Source, 0),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Sources{
				sources: tt.fields.sources,
			}
			if err := s.check(); (err != nil) != tt.wantErr {
				t.Errorf("Sources.check() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewSource(t *testing.T) {
	type args struct {
		name string
		data interface{}
	}
	tests := []struct {
		name string
		args args
		want *Source
	}{
		{
			name: "test",
			args: args{
				name: "some-source",
				data: &params{},
			},
			want: &Source{
				name: "some-source",
				data: &params{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSource(tt.args.name, tt.args.data); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSource_WithOrderID(t *testing.T) {
	type fields struct {
		name    string
		orderID int
		data    interface{}
	}
	type args struct {
		orderID int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *Source
	}{
		{
			name: "test",
			fields: fields{
				name:    "some-source",
				orderID: 0,
				data:    &params{},
			},
			args: args{
				orderID: 0,
			},
			want: NewSource("some-source", &params{}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Source{
				name:    tt.fields.name,
				orderID: tt.fields.orderID,
				data:    tt.fields.data,
			}
			if got := s.WithOrderID(tt.args.orderID); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Source.WithOrderID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSource_Name(t *testing.T) {
	type fields struct {
		name    string
		orderID int
		data    interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "test",
			fields: fields{
				name:    "some-source",
				orderID: 0,
				data:    &params{},
			},
			want: "some-source",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Source{
				name:    tt.fields.name,
				orderID: tt.fields.orderID,
				data:    tt.fields.data,
			}
			if got := s.Name(); got != tt.want {
				t.Errorf("Source.Name() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSource_Data(t *testing.T) {
	type fields struct {
		name    string
		orderID int
		data    interface{}
	}
	tests := []struct {
		name   string
		fields fields
		want   interface{}
	}{
		{
			name: "test",
			fields: fields{
				name:    "some-source",
				orderID: 0,
				data:    &params{},
			},
			want: &params{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Source{
				name:    tt.fields.name,
				orderID: tt.fields.orderID,
				data:    tt.fields.data,
			}
			if got := s.Data(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Source.Data() = %v, want %v", got, tt.want)
			}
		})
	}
}
