// Copyright 2019 ThoughtWorks, Inc.

/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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
