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

package main

import (
	"fmt"
	"github.com/getgauge/gauge/gauge_messages"
	"strings"
)

type table struct {
	headerIndexMap map[string]int
	columns        [][]tableCell
	headers        []string
	lineNo         int
}

type tableCell struct {
	value    string
	cellType argType
}

func (table *table) isInitialized() bool {
	return table.headerIndexMap != nil
}

func (table *table) get(header string) []tableCell {
	if !table.headerExists(header) {
		panic(fmt.Sprintf("Table column %s not found", header))
	}
	return table.columns[table.headerIndexMap[header]]
}

func (table *table) headerExists(header string) bool {
	_, ok := table.headerIndexMap[header]
	return ok
}

func (table *table) addHeaders(columnNames []string) {
	table.headerIndexMap = make(map[string]int)
	table.headers = make([]string, len(columnNames))
	table.columns = make([][]tableCell, len(columnNames))
	for i, column := range columnNames {
		trimmedHeader := strings.TrimSpace(column)
		table.headers[i] = trimmedHeader
		table.headerIndexMap[trimmedHeader] = i
		table.columns[i] = make([]tableCell, 0)
	}
}

func (table *table) addRowValues(rowValues []string) {
	tableCells := table.createTableCells(rowValues)
	table.addRows(tableCells)
}

func (table *table) createTableCells(rowValues []string) []tableCell {
	tableCells := make([]tableCell, 0)
	for _, value := range rowValues {
		tableCells = append(tableCells, getTableCell(strings.TrimSpace(value)))
	}
	return tableCells
}

func (table *table) toHeaderSizeRow(rows []tableCell) []tableCell {
	finalCells := make([]tableCell, 0)
	for i, _ := range table.headers {
		var cell tableCell
		if len(rows)-1 >= i {
			cell = rows[i]
		} else {
			cell = getDefaultTableCell()
		}
		finalCells = append(finalCells, cell)
	}
	return finalCells
}

func (table *table) addRows(rows []tableCell) {
	for i, value := range table.toHeaderSizeRow(rows) {
		table.columns[i] = append(table.columns[i], value)
	}
}

func (table *table) getRows() [][]string {
	if !table.isInitialized() {
		return nil
	}

	tableRows := make([][]string, 0)
	for i := 0; i < len(table.columns[0]); i++ {
		row := make([]string, 0)
		for _, header := range table.headers {
			tableCell := table.get(header)[i]
			value := tableCell.value
			row = append(row, value)
		}
		tableRows = append(tableRows, row)
	}
	return tableRows
}

func (table *table) getRowCount() int {
	if table.isInitialized() {
		return len(table.columns[0])
	} else {
		return 0
	}
}

func (table *table) kind() tokenKind {
	return tableKind
}

func getTableCell(value string) tableCell {
	return tableCell{value: value, cellType: static}
}

func getDefaultTableCell() tableCell {
	return tableCell{value: "", cellType: static}
}

func tableFrom(protoTable *gauge_messages.ProtoTable) *table {
	table := &table{}
	table.addHeaders(protoTable.GetHeaders().GetCells())
	for _, row := range protoTable.GetRows() {
		table.addRowValues(row.GetCells())
	}
	return table
}
