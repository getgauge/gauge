/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package gauge

import "fmt"

type Table struct {
	headerIndexMap map[string]int
	Columns        [][]TableCell
	Headers        []string
	LineNo         int
}

type DataTable struct {
	Table      *Table
	Value      string
	LineNo     int
	IsExternal bool
}

type TableCell struct {
	Value    string
	CellType ArgType
}

func NewTable(headers []string, cols [][]TableCell, lineNo int) *Table {
	headerIndx := make(map[string]int)
	for i, h := range headers {
		headerIndx[h] = i
	}

	return &Table{
		headerIndexMap: headerIndx,
		Columns:        cols,
		Headers:        headers,
		LineNo:         lineNo,
	}
}

func (table *Table) IsInitialized() bool {
	return table != nil && table.headerIndexMap != nil
}

func (cell *TableCell) GetValue() string {
	value := cell.Value
	if cell.CellType == Dynamic || cell.CellType == SpecialString {
		value = fmt.Sprintf("<%s>", value)
	}
	return value
}

func (dataTable *DataTable) IsInitialized() bool {
	return dataTable.Table != nil && dataTable.Table.headerIndexMap != nil
}

func (table *Table) String() string {
	return fmt.Sprintf("%v\n%v", table.Headers, table.Columns)
}

func (table *Table) GetDynamicArgs() []string {
	args := make([]string, 0)
	for _, row := range table.Columns {
		for _, column := range row {
			if column.CellType == Dynamic {
				args = append(args, column.Value)
			}
		}
	}
	return args
}

func (table *Table) Get(header string) ([]TableCell, error) {
	if !table.headerExists(header) {
		return nil, fmt.Errorf("Table column %s not found", header)
	}
	return table.Columns[table.headerIndexMap[header]], nil
}

func (table *Table) headerExists(header string) bool {
	_, ok := table.headerIndexMap[header]
	return ok
}

func (table *Table) AddHeaders(columnNames []string) {
	table.headerIndexMap = make(map[string]int)
	table.Headers = make([]string, len(columnNames))
	table.Columns = make([][]TableCell, len(columnNames))
	for i, column := range columnNames {
		table.Headers[i] = column
		table.headerIndexMap[column] = i
		table.Columns[i] = make([]TableCell, 0)
	}
}

func (table *Table) AddRowValues(tableCells []TableCell) {
	table.addRows(tableCells)
}

func (table *Table) CreateTableCells(rowValues []string) []TableCell {
	tableCells := make([]TableCell, 0)
	for _, value := range rowValues {
		tableCells = append(tableCells, GetTableCell(value))
	}
	return tableCells
}

func (table *Table) toHeaderSizeRow(rows []TableCell) []TableCell {
	finalCells := make([]TableCell, 0)
	for i := range table.Headers {
		var cell TableCell
		if len(rows)-1 >= i {
			cell = rows[i]
		} else {
			cell = GetDefaultTableCell()
		}
		finalCells = append(finalCells, cell)
	}
	return finalCells
}

func (table *Table) addRows(rows []TableCell) {
	for i, value := range table.toHeaderSizeRow(rows) {
		table.Columns[i] = append(table.Columns[i], value)
	}
}

func (table *Table) Rows() [][]string {
	if !table.IsInitialized() {
		return nil
	}

	tableRows := make([][]string, 0)
	if len(table.Columns) == 0 {
		return tableRows
	}
	for i := 0; i < len(table.Columns[0]); i++ {
		row := make([]string, 0)
		for _, header := range table.Headers {
			tableCells, _ := table.Get(header)
			tableCell := tableCells[i]
			value := tableCell.GetValue()
			row = append(row, value)
		}
		tableRows = append(tableRows, row)
	}
	return tableRows
}

func (table *Table) GetRowCount() int {
	if table.IsInitialized() {
		return len(table.Columns[0])
	}
	return 0
}

func (table *Table) Kind() TokenKind {
	return TableKind
}

func (externalTable *DataTable) Kind() TokenKind {
	return DataTableKind
}

func GetTableCell(value string) TableCell {
	return TableCell{Value: value, CellType: Static}
}

func GetDefaultTableCell() TableCell {
	return TableCell{Value: "", CellType: Static}
}
