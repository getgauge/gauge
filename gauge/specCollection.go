// Copyright 2015 ThoughtWorks, Inc.

// This file is part of Gauge.

// Gauge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Gauge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.

package gauge

import "sync"

type SpecCollection struct {
	mutex sync.Mutex
	index int
	specs []*Specification
}

func NewSpecCollection(s []*Specification) *SpecCollection {
	return &SpecCollection{specs: s}
}

func (s *SpecCollection) Add(spec *Specification) {
	s.specs = append(s.specs, spec)
}

func (s *SpecCollection) Specs() []*Specification {
	return s.specs
}

func (s *SpecCollection) HasNext() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.index < len(s.specs)
}

func (s *SpecCollection) Next() *Specification {
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
	for _, spec := range s.specs {
		specNames = append(specNames, spec.FileName)
	}
	return specNames
}
