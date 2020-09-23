package structmerge

import (
	"errors"
)

var (
	errEmptySources = errors.New("sources cannot be empty")
)

// Source holds source
type Source struct {
	name    string
	orderID int
	data    interface{}
}

// NewSource creates new instance of the source
func NewSource(name string, data interface{}) *Source {
	return &Source{
		name: name,
		data: data,
	}
}

// WithOrderID sets order id
func (s *Source) WithOrderID(orderID int) *Source {
	s.orderID = orderID

	return s
}

// Name returns source name
func (s *Source) Name() string {
	return s.name
}

// Data returns source data
func (s *Source) Data() interface{} {
	return s.data
}

// Sources holds sources
type Sources struct {
	sources []*Source
}

// NewSources creates new instance
func NewSources() *Sources {
	return &Sources{
		sources: make([]*Source, 0),
	}
}

// Add adds source to the sources with the given name
func (s *Sources) Add(name string, data interface{}) *Sources {
	source := NewSource(name, data).WithOrderID(len(s.sources))
	s.sources = append(s.sources, source)

	return s
}

// Sources returns sources list
func (s Sources) Sources() []*Source {
	return s.sources
}

// check checks if sources length is proper
func (s *Sources) check() error {
	if len(s.sources) == 0 {
		return errEmptySources
	}

	return nil
}
