package parser

import (
	"sync"
)

type SpecFileCollection struct {
	mutex     *sync.Mutex
	index     int
	specFiles []string
}

func NewSpecFileCollection(s []string) *SpecFileCollection {
	return &SpecFileCollection{specFiles: s, index: 0, mutex: &sync.Mutex{}}
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
