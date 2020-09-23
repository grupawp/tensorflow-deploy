package structmerge

// Merger ...
type Merger interface {
	Merge(destination interface{}, s *Sources) ([]MergedMeta, error)
}

// MergedMeta holds metadata of merged source
type MergedMeta struct {
	FieldPath string
	Source    string
	Value     string // WARNING! For security reasons, don't print or log all values (e.g. passwords can be a sensitive data)
}

// StructMerge holds sources and merged metadata
type StructMerge struct {
	sources    *Sources
	MergedMeta []MergedMeta
}

// NewStructMerge creates new instance
func NewStructMerge() *StructMerge {
	return &StructMerge{
		sources:    NewSources(),
		MergedMeta: make([]MergedMeta, 0),
	}
}

// WithSource adds data with given name to the sources
func (sm *StructMerge) WithSource(name string, data interface{}) *StructMerge {
	sm.sources.Add(name, data)

	return sm
}

// WithSources sets sources
func (sm *StructMerge) WithSources(s *Sources) *StructMerge {
	sm.sources = s

	return sm
}

// Merge merges structures into given data using default implementation of merger
func (sm *StructMerge) Merge(data interface{}) error {
	return sm.MergeWithMerger(data, newDefaultMerger())
}

// MergeWithMerger merges structures into given data with custom merger
func (sm *StructMerge) MergeWithMerger(data interface{}, m Merger) error {
	if err := sm.sources.check(); err != nil {
		return err
	}

	mergedMeta, err := m.Merge(data, sm.sources)
	if err != nil {
		return err
	}
	sm.MergedMeta = mergedMeta

	return nil
}
