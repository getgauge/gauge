/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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
