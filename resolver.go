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
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"
	"regexp"
	"strings"
)

type resolverFn func(string) (*stepArg, error)
type specialTypeResolver struct {
	predefinedResolvers map[string]resolverFn
}

type paramResolver struct {
}

func (paramResolver *paramResolver) getResolvedParams(step *step, parent *step, dataTableLookup *argLookup) []*gauge_messages.Parameter {
	parameters := make([]*gauge_messages.Parameter, 0)
	for _, arg := range step.args {
		parameter := new(gauge_messages.Parameter)
		parameter.Name = proto.String(arg.name)
		if arg.argType == static {
			parameter.ParameterType = gauge_messages.Parameter_Static.Enum()
			parameter.Value = proto.String(arg.value)
		} else if arg.argType == dynamic {
			var resolvedArg *stepArg
			if parent != nil {
				resolvedArg = parent.getArg(arg.value)
			} else {
				resolvedArg = dataTableLookup.getArg(arg.value)
			}
			//In case a special table used in a concept, you will get a dynamic table value which has to be resolved from the concept lookup
			parameter.Name = proto.String(resolvedArg.name)
			if resolvedArg.table.isInitialized() {
				parameter.ParameterType = gauge_messages.Parameter_Special_Table.Enum()
				parameter.Table = paramResolver.createProtoStepTable(&resolvedArg.table, dataTableLookup)
			} else {
				parameter.ParameterType = gauge_messages.Parameter_Dynamic.Enum()
				parameter.Value = proto.String(resolvedArg.value)
			}
		} else if arg.argType == specialString {
			parameter.ParameterType = gauge_messages.Parameter_Special_String.Enum()
			parameter.Value = proto.String(arg.value)
		} else if arg.argType == specialTable {
			parameter.ParameterType = gauge_messages.Parameter_Special_Table.Enum()
			parameter.Table = paramResolver.createProtoStepTable(&arg.table, dataTableLookup)
		} else {
			parameter.ParameterType = gauge_messages.Parameter_Table.Enum()
			parameter.Table = paramResolver.createProtoStepTable(&arg.table, dataTableLookup)

		}
		parameters = append(parameters, parameter)
	}

	return parameters

}

func (resolver *paramResolver) createProtoStepTable(table *table, dataTableLookup *argLookup) *gauge_messages.ProtoTable {
	protoTable := new(gauge_messages.ProtoTable)
	protoTable.Headers = &gauge_messages.ProtoTableRow{Cells: table.headers}
	tableRows := make([]*gauge_messages.ProtoTableRow, 0)
	for i := 0; i < len(table.columns[0]); i++ {
		row := make([]string, 0)
		for _, header := range table.headers {
			tableCell := table.get(header)[i]
			value := tableCell.value
			if tableCell.cellType == dynamic {
				//if concept has a table with dynamic cell, fetch from datatable
				value = dataTableLookup.getArg(tableCell.value).value
			}
			row = append(row, value)
		}
		tableRows = append(tableRows, &gauge_messages.ProtoTableRow{Cells: row})
	}
	protoTable.Rows = tableRows
	return protoTable
}

func newSpecialTypeResolver() *specialTypeResolver {
	resolver := new(specialTypeResolver)
	resolver.predefinedResolvers = initializePredefinedResolvers()
	return resolver
}

func initializePredefinedResolvers() map[string]resolverFn {
	return map[string]resolverFn{
		"file": func(filePath string) (*stepArg, error) {
			fileContent, err := common.ReadFileContents(filePath)
			if err != nil {
				return nil, err
			}
			return &stepArg{value: fileContent, argType: specialString}, nil
		},
		"table": func(filePath string) (*stepArg, error) {
			csv, err := common.ReadFileContents(filePath)
			if err != nil {
				return nil, err
			}
			csvTable, err := convertCsvToTable(csv)
			if err != nil {
				return nil, err
			}
			return &stepArg{table: *csvTable, argType: specialTable}, nil
		},
	}
}

func convertCsvToTable(csvContents string) (*table, error) {
	r := csv.NewReader(strings.NewReader(csvContents))
	lines, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	table := new(table)
	for i, line := range lines {
		if i == 0 {
			table.addHeaders(line)
		} else {
			table.addRowValues(line)
		}
	}
	return table, nil
}

func (resolver *specialTypeResolver) resolve(arg string) (*stepArg, error) {
	regEx := regexp.MustCompile("(.*):(.*)")
	match := regEx.FindAllStringSubmatch(arg, -1)
	specialType := strings.TrimSpace(match[0][1])
	value := strings.TrimSpace(match[0][2])
	stepArg, err := resolver.getStepArg(specialType, value, arg)
	if err == nil {
		stepArg.name = arg
	}
	return stepArg, err
}

func (resolver *specialTypeResolver) getStepArg(specialType string, value string, arg string) (*stepArg, error) {
	resolveFunc, found := resolver.predefinedResolvers[specialType]
	if found {
		return resolveFunc(value)
	}
	return nil, errors.New(fmt.Sprintf("Resolver not found for special param <%s>", arg))
}
