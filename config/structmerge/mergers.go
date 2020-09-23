package structmerge

import (
	"fmt"
	"reflect"

	"github.com/pkg/errors"
)

const (
	// FieldPathSeparator is used to separate field names
	FieldPathSeparator = "."
)

var (
	errNilData         = errors.New("data cannot be nil")
	errNotPtr          = errors.New("data is not a pointer")
	errNotStruct       = errors.New("data is not a struct")
	errInvalidType     = errors.New("invalid type")
	errUnsupportedType = errors.New("unsupported type")
	errReallyWrong     = errors.New("something goes really wrong") // used for things which should not happen :P
)

// Merge is a structure used for merger implementation(s)
type Merge struct {
	mergedMeta []MergedMeta
}

// newDefaultMerger creates instance of the default merger
func newDefaultMerger() *Merge {
	return &Merge{
		mergedMeta: make([]MergedMeta, 0),
	}
}

// Merge merges all sources into given data. All types of data and sources must be the same.
func (m *Merge) Merge(data interface{}, s *Sources) ([]MergedMeta, error) {
	// all types of data & sources must be the same
	if err := m.checkDataTypes(data, s); err != nil {
		return nil, err
	}

	if err := m.mergeSources(data, s.Sources()); err != nil {
		return nil, err
	}

	return m.mergedMeta, nil
}

// merge merges all sources into data
func (m *Merge) mergeSources(data interface{}, s []*Source) error {
	travVals := newTraversedValues()
	travVals.data = reflect.ValueOf(data).Elem()
	// add sources in reversed order (the merge order)
	for i := len(s) - 1; i >= 0; i-- {
		travVals.sources = append(travVals.sources, &traversedSource{
			name: s[i].Name(),
			data: reflect.ValueOf(s[i].Data()).Elem(),
		})
	}

	return m.merge(travVals)
}

// mergeSources traverses through data and sources and sets value from source
func (m *Merge) merge(travVals *traversedValues) error {
	switch travVals.data.Kind() {
	case reflect.Struct:
		for i := 0; i < travVals.data.NumField(); i++ {
			newTravVals := newTraversedValues()
			newTravVals.data = travVals.data.Field(i)
			for _, s := range travVals.sources {
				newTravVals.sources = append(newTravVals.sources, &traversedSource{
					name: s.name,
					data: s.data.Field(i),
				})
			}
			newTravVals.joinFieldPath(travVals.fieldPath, travVals.data.Type().Field(i).Name)

			err := m.merge(newTravVals)
			if err != nil {
				return err
			}
		}
	case reflect.Ptr:
		switch travVals.data.Type().Elem().Kind() {
		case reflect.Struct, reflect.Map, reflect.Slice, reflect.Array, reflect.Interface, reflect.Ptr:
			return errors.Wrapf(errUnsupportedType, "merge: `%s`", travVals.data.Type().Elem().Kind())
		case reflect.Invalid:
			return errReallyWrong
		default:
			selectedValue, selectedSource, err := m.selectValue(travVals)
			if err != nil {
				return err
			}

			travVals.data.Set(selectedValue)

			m.mergedMeta = append(m.mergedMeta, MergedMeta{
				FieldPath: travVals.fieldPath,
				Source:    selectedSource,
				Value:     fmt.Sprintf("%+v", selectedValue.Elem()),
			})
		}
	case reflect.Invalid:
		return errReallyWrong
	default:
		return errors.Wrapf(errUnsupportedType, "merge: `%s`", travVals.data.Kind())
	}

	return nil
}

// selectValues selects value from sources
func (m *Merge) selectValue(travVals *traversedValues) (reflect.Value, string, error) {
	var sourceName string
	var value reflect.Value
	var valueFromSource bool

	// select value from sources without last one
	// last source is treated as the default one
	for i := 0; i < len(travVals.sources)-1; i++ {
		if !m.isSettable(travVals.sources[i].data) {
			continue
		}

		sourceName = travVals.sources[i].name
		value = travVals.sources[i].data
		valueFromSource = true
	}

	// select value from last source
	if !valueFromSource {
		sourceName = travVals.sources[len(travVals.sources)-1].name
		value = travVals.sources[len(travVals.sources)-1].data
	}

	// selected value must be a proper value (cannot be invalid or nil)
	// protection to avoid unnecessary checks in your code
	if value.Kind() == reflect.Invalid {
		return reflect.Value{}, sourceName, fmt.Errorf("value from last source `%s` is invalid", sourceName)
	} else if value.IsNil() {
		return reflect.Value{}, sourceName, fmt.Errorf("value from last source `%s` cannot be nil", sourceName)
	}

	return value, sourceName, nil
}

// isSettable checks if the value can be set
func (m *Merge) isSettable(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.String:
		return v.Len() != 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() != 0
	case reflect.Float32, reflect.Float64:
		return v.Float() != 0
	case reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return !v.IsNil()
	}

	return true
}

// checkDataTypes checks data type and kind againts sources data types. All must be the same.
func (m *Merge) checkDataTypes(data interface{}, s *Sources) error {
	dataType, err := m.getDataType(data)
	if err != nil {
		return err
	}

	for _, v := range s.Sources() {
		sourceDataType, err := m.getDataType(v.data)
		if err != nil {
			return err
		}

		if sourceDataType != dataType {
			return errInvalidType
		}
	}

	return nil
}

// getDataType gets data dynamic type. Given data cannot be nil and must be a pointer to a struct.
func (m *Merge) getDataType(data interface{}) (reflect.Type, error) {
	if data == nil {
		return nil, errNilData
	}

	t := reflect.TypeOf(data)
	if t.Kind() != reflect.Ptr {
		return nil, errNotPtr
	} else if t.Elem().Kind() != reflect.Struct {
		return nil, errNotStruct
	}

	return t, nil
}

type traversedSource struct {
	name string
	data reflect.Value
}

type traversedValues struct {
	data      reflect.Value
	sources   []*traversedSource
	fieldPath string
}

func newTraversedValues() *traversedValues {
	return &traversedValues{
		sources: make([]*traversedSource, 0),
	}
}

// joinFieldPath creates field path from field names
func (v *traversedValues) joinFieldPath(path, partPath string) error {
	if partPath == "" {
		return errReallyWrong
	}

	if path == "" {
		v.fieldPath = partPath
	} else {
		v.fieldPath = path + FieldPathSeparator + partPath
	}

	return nil
}
