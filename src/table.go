package main

type table struct {
	headerIndexMap map[string]int
	columns        [][]string
	headers        []string
}

func (table *table) isInitialized() bool {
	return table.headerIndexMap != nil
}

func (table *table) get(columnName string) []string {
	return table.columns[table.headerIndexMap[columnName]]
}

func (table *table) addHeaders(columns []string) {
	table.headerIndexMap = make(map[string]int)
	table.headers = make([]string, len(columns))
	table.columns = make([][]string, len(columns))
	for i, column := range columns {
		table.headers[i] = columns[i]
		table.headerIndexMap[column] = i
		table.columns[i] = make([]string, 0)
	}
}

func (table *table) addRowValues(rowValues []string) {
	for i, value := range rowValues {
		table.columns[i] = append(table.columns[i], value)
	}
	if len(table.headers) > len(rowValues) {
		for i := len(rowValues); i < len(table.headers); i++ {
			table.columns[i] = append(table.columns[i], "")
		}
	}
}


func (table *table) getRowCount() int {
	if table.isInitialized() {
		return len(table.columns[0])
	} else {
		return 0
	}

}
