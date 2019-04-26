package parser

import (
	"sync"

	"github.com/getgauge/gauge/gauge"
)

type SpecFileCollection struct {
	mutex             sync.Mutex
	index             int
	specFiles         []string
	conceptDictionary *gauge.ConceptDictionary
}

func NewSpecFileCollection(s []string, cptDict *gauge.ConceptDictionary) *SpecFileCollection {
	return &SpecFileCollection{specFiles: s, conceptDictionary: cptDict}
}

func (s *SpecFileCollection) HasNext() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.index < len(s.specFiles)
}

func (s *SpecFileCollection) Next() string {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	specFile := s.specFiles[s.index]
	s.index++
	return specFile
}
