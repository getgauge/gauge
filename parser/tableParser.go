/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package parser

import (
	"encoding/csv"
	"os"
	"strings"

	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/gauge"
)

func convertCsvToTable(csvContents string) (*gauge.Table, error) {
	r := csv.NewReader(strings.NewReader(csvContents))
	var de = os.Getenv(env.CsvDelimiter)
	if de != "" {
		r.Comma = []rune(os.Getenv(env.CsvDelimiter))[0]
	}
	r.Comment = '#'
	lines, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	table := new(gauge.Table)
	for i, line := range lines {
		if i == 0 {
			table.AddHeaders(line)
		} else {
			table.AddRowValues(table.CreateTableCells(line))
		}
	}
	return table, nil
}
