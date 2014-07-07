package main

import (
	"code.google.com/p/goprotobuf/proto"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/getgauge/common"
	"regexp"
	"strings"
)

type resolverFn func(string) (*stepArg, error)
type specialTypeResolver struct {
	predefinedResolvers map[string]resolverFn
}

type paramResolver struct {
}

func (paramResolver *paramResolver) getResolvedParams(stepArgs []*stepArg, lookup *argLookup, dataTableLookup *argLookup) []*Parameter {
	parameters := make([]*Parameter, 0)
	for _, arg := range stepArgs {
		parameter := new(Parameter)
		parameter.Name = proto.String(arg.name)
		if arg.argType == static {
			parameter.ParameterType = Parameter_Static.Enum()
			parameter.Value = proto.String(arg.value)
		} else if arg.argType == dynamic {
			resolvedArg := lookup.getArg(arg.value)
			//In case a special table used in a concept, you will get a dynamic table value which has to be resolved from the concept lookup
			parameter.Name = proto.String(resolvedArg.name)
			if resolvedArg.table.isInitialized() {
				parameter.ParameterType = Parameter_Special_Table.Enum()
				parameter.Table = paramResolver.createProtoStepTable(&resolvedArg.table, lookup, dataTableLookup)
			} else {
				parameter.ParameterType = Parameter_Dynamic.Enum()
				parameter.Value = proto.String(resolvedArg.value)
			}
		} else if arg.argType == specialString {
			parameter.ParameterType = Parameter_Special_String.Enum()
			parameter.Value = proto.String(arg.value)
		} else if arg.argType == specialTable {
			parameter.ParameterType = Parameter_Special_Table.Enum()
			parameter.Table = paramResolver.createProtoStepTable(&arg.table, lookup, dataTableLookup)
		} else {
			parameter.ParameterType = Parameter_Table.Enum()
			parameter.Table = paramResolver.createProtoStepTable(&arg.table, lookup, dataTableLookup)

		}
		parameters = append(parameters, parameter)
	}

	return parameters

}

func (resolver *paramResolver) createProtoStepTable(table *table, lookup *argLookup, dataTableLookup *argLookup) *ProtoTable {
	protoTable := new(ProtoTable)
	protoTable.Headers = &ProtoTableRow{Cells: table.headers}
	tableRows := make([]*ProtoTableRow, 0)
	for i := 0; i < len(table.columns[0]); i++ {
		row := make([]string, 0)
		for _, header := range table.headers {
			tableCell := table.get(header)[i]
			value := tableCell.value
			if tableCell.cellType == dynamic {
				if lookup.containsArg(tableCell.value) {
					value = lookup.getArg(tableCell.value).value
				} else {
					//if concept has a table with dynamic cell, arglookup won't have the table value, so fetch from datatable itself
					value = dataTableLookup.getArg(tableCell.value).value
				}
			}
			row = append(row, value)
		}
		tableRows = append(tableRows, &ProtoTableRow{Cells: row})
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
