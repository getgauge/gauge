/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package gauge

import (
	"sync"
)

type SpecCollection struct {
	mutex sync.Mutex
	index int
	specs [][]*Specification
}

func NewSpecCollection(s []*Specification, groupDataTableSpecs bool) *SpecCollection {
	if !groupDataTableSpecs {
		var specs [][]*Specification
		for _, spec := range s {
			specs = append(specs, []*Specification{spec})
		}
		return &SpecCollection{specs: specs}
	}
	return &SpecCollection{specs: combineDataTableSpecs(s)}
}

func combineDataTableSpecs(s []*Specification) (specs [][]*Specification) {
	combinedSpecs := make(map[string][]*Specification)
	for _, spec := range s {
		combinedSpecs[spec.FileName] = append(combinedSpecs[spec.FileName], spec)
	}
	for _, spec := range s {
		if _, ok := combinedSpecs[spec.FileName]; ok {
			specs = append(specs, combinedSpecs[spec.FileName])
			delete(combinedSpecs, spec.FileName)
		}
	}
	return
}

func (s *SpecCollection) Add(spec *Specification) {
	s.specs = append(s.specs, []*Specification{spec})
}

func (s *SpecCollection) Specs() (specs []*Specification) {
	for _, subSpecs := range s.specs {
		specs = append(specs, subSpecs...)
	}
	return specs
}

func (s *SpecCollection) HasNext() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.index < len(s.specs)
}

func (s *SpecCollection) Next() []*Specification {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	spec := s.specs[s.index]
	s.index++
	return spec
}

func (s *SpecCollection) Size() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	length := len(s.specs)
	return length
}

func (s *SpecCollection) SpecNames() []string {
	specNames := make([]string, 0)
	for _, specs := range s.specs {
		for _, subSpec := range specs {
			specNames = append(specNames, subSpec.FileName)
		}
	}
	return specNames
}
