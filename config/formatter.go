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

package config

import (
	"encoding/json"
	"fmt"
	"strings"
)

type formatter interface {
	format([]property) (string, error)
}

type jsonFormatter struct {
}

func (f jsonFormatter) format(p []property) (string, error) {
	bytes, err := json.MarshalIndent(p, "", "\t")
	return string(bytes), err
}

type textFormatter struct {
}

func (f textFormatter) format(p []property) (string, error) {
	format := "%-30s\t%-35s"
	var s []string
	max := 0
	for _, v := range p {
		text := fmt.Sprintf(format, v.Key, v.Value)
		s = append(s, text)
		if max < len(text) {
			max = len(text)
		}
	}
	s = append([]string{fmt.Sprintf(format, "Key", "Value"), strings.Repeat("-", max)}, s...)
	return strings.Join(s, "\n"), nil
}
