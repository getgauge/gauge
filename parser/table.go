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

package parser

import (
	"encoding/csv"
	"strings"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
)

func TableFrom(protoTable *gauge_messages.ProtoTable) *gauge.Table {
	table := &gauge.Table{}
	table.AddHeaders(protoTable.GetHeaders().GetCells())
	for _, row := range protoTable.GetRows() {
		table.AddRowValues(row.GetCells())
	}
	return table
}

func convertCsvToTable(csvContents string) (*gauge.Table, error) {
	r := csv.NewReader(strings.NewReader(csvContents))
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
			table.AddRowValues(line)
		}
	}
	return table, nil
}
