/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package config

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type Formatter interface {
	Format([]Property) (string, error)
}

type JsonFormatter struct {
}

func (f JsonFormatter) Format(p []Property) (string, error) {
	sort.Sort(byPropertyKey(p))
	bytes, err := json.MarshalIndent(p, "", "\t")
	return string(bytes), err
}

type TextFormatter struct {
	Headers []string
}

func (f TextFormatter) Format(p []Property) (string, error) {
	sort.Sort(byPropertyKey(p))
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
	if len(f.Headers) == 0 {
		f.Headers = []string{"Key", "Value"}
	}
	s = append([]string{fmt.Sprintf(format, f.Headers[0], f.Headers[1]), strings.Repeat("-", max)}, s...)
	return strings.Join(s, "\n"), nil
}

type byPropertyKey []Property

func (a byPropertyKey) Len() int           { return len(a) }
func (a byPropertyKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPropertyKey) Less(i, j int) bool { return a[i].Key < a[j].Key }
