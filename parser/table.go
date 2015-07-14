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
	"fmt"
	"github.com/getgauge/gauge/gauge_messages"
	"strings"
)

type Table struct {
	headerIndexMap map[string]int
	columns        [][]TableCell
	Headers        []string
	LineNo         int
}

type DataTable struct {
	Table      Table
	Value      string
	LineNo     int
	IsExternal bool
}

type TableCell struct {
	Value    string
	CellType ArgType
}

func (table *Table) IsInitialized() bool {
	return table.headerIndexMap != nil
}

func (cell *TableCell) GetValue() string {
	value := cell.Value
	if cell.CellType == Dynamic {
		value = fmt.Sprintf("<%s>", value)
	}
	return value
}

func (dataTable *DataTable) IsInitialized() bool {
	return dataTable.Table.headerIndexMap != nil
}

func (table *Table) String() string {
	return fmt.Sprintf("%v\n%v", table.Headers, table.columns)
}

func (table *Table) GetDynamicArgs() []string {
	args := make([]string, 0)
	for _, row := range table.columns {
		for _, column := range row {
			if column.CellType == Dynamic {
				args = append(args, column.Value)
			}
		}
	}
	return args
}

func (table *Table) Get(header string) []TableCell {
	if !table.headerExists(header) {
		panic(fmt.Sprintf("Table column %s not found", header))
	}
	return table.columns[table.headerIndexMap[header]]
}

func (table *Table) headerExists(header string) bool {
	_, ok := table.headerIndexMap[header]
	return ok
}

func (table *Table) addHeaders(columnNames []string) {
	table.headerIndexMap = make(map[string]int)
	table.Headers = make([]string, len(columnNames))
	table.columns = make([][]TableCell, len(columnNames))
	for i, column := range columnNames {
		trimmedHeader := strings.TrimSpace(column)
		table.Headers[i] = trimmedHeader
		table.headerIndexMap[trimmedHeader] = i
		table.columns[i] = make([]TableCell, 0)
	}
}

func (table *Table) addRowValues(rowValues []string) {
	tableCells := table.createTableCells(rowValues)
	table.addRows(tableCells)
}

func (table *Table) createTableCells(rowValues []string) []TableCell {
	tableCells := make([]TableCell, 0)
	for _, value := range rowValues {
		tableCells = append(tableCells, getTableCell(strings.TrimSpace(value)))
	}
	return tableCells
}

func (table *Table) toHeaderSizeRow(rows []TableCell) []TableCell {
	finalCells := make([]TableCell, 0)
	for i, _ := range table.Headers {
		var cell TableCell
		if len(rows)-1 >= i {
			cell = rows[i]
		} else {
			cell = getDefaultTableCell()
		}
		finalCells = append(finalCells, cell)
	}
	return finalCells
}

func (table *Table) addRows(rows []TableCell) {
	for i, value := range table.toHeaderSizeRow(rows) {
		table.columns[i] = append(table.columns[i], value)
	}
}

func (table *Table) Rows() [][]string {
	if !table.IsInitialized() {
		return nil
	}

	tableRows := make([][]string, 0)
	for i := 0; i < len(table.columns[0]); i++ {
		row := make([]string, 0)
		for _, header := range table.Headers {
			tableCell := table.Get(header)[i]
			value := tableCell.GetValue()
			row = append(row, value)
		}
		tableRows = append(tableRows, row)
	}
	return tableRows
}

func (table *Table) GetRowCount() int {
	if table.IsInitialized() {
		return len(table.columns[0])
	} else {
		return 0
	}
}

func (table *Table) Kind() TokenKind {
	return TableKind
}

func (externalTable *DataTable) Kind() TokenKind {
	return DataTableKind
}

func getTableCell(value string) TableCell {
	return TableCell{Value: value, CellType: Static}
}

func getDefaultTableCell() TableCell {
	return TableCell{Value: "", CellType: Static}
}

func TableFrom(protoTable *gauge_messages.ProtoTable) *Table {
	table := &Table{}
	table.addHeaders(protoTable.GetHeaders().GetCells())
	for _, row := range protoTable.GetRows() {
		table.addRowValues(row.GetCells())
	}
	return table
}

func convertCsvToTable(csvContents string) (*Table, error) {
	r := csv.NewReader(strings.NewReader(csvContents))
	lines, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	table := new(Table)
	for i, line := range lines {
		if i == 0 {
			table.addHeaders(line)
		} else {
			table.addRowValues(line)
		}
	}
	return table, nil
}
