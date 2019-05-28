// Copyright 2019 ThoughtWorks, Inc.

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

package parser

import (
	"fmt"
	"sync"
)

type specFileCollection struct {
	mutex     sync.Mutex
	index     int
	specFiles []string
}

func NewSpecFileCollection(s []string) *specFileCollection {
	return &specFileCollection{specFiles: s}
}

func (s *specFileCollection) Next() (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.index < len(s.specFiles) {
		specFile := s.specFiles[s.index]
		s.index++
		return specFile, nil
	}
	return "", fmt.Errorf("no files in collection")
}
