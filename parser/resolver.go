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
	"fmt"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"
	"regexp"
	"strings"
)

type invalidSpecialParamError struct {
	message string
}

type resolverFn func(string) (*StepArg, error)
type specialTypeResolver struct {
	predefinedResolvers map[string]resolverFn
}

type paramResolver struct {
}

func (invalidSpecialParamError invalidSpecialParamError) Error() string {
	return invalidSpecialParamError.message
}

func (paramResolver *paramResolver) getResolvedParams(step *Step, parent *Step, dataTableLookup *ArgLookup) []*gauge_messages.Parameter {
	parameters := make([]*gauge_messages.Parameter, 0)
	for _, arg := range step.args {
		parameter := new(gauge_messages.Parameter)
		parameter.Name = proto.String(arg.Name)
		if arg.ArgType == Static {
			parameter.ParameterType = gauge_messages.Parameter_Static.Enum()
			parameter.Value = proto.String(arg.Value)
		} else if arg.ArgType == Dynamic {
			var resolvedArg *StepArg
			if parent != nil {
				resolvedArg = parent.getArg(arg.Value)
			} else {
				resolvedArg = dataTableLookup.getArg(arg.Value)
			}
			//In case a special table used in a concept, you will get a dynamic table value which has to be resolved from the concept lookup
			parameter.Name = proto.String(resolvedArg.Name)
			if resolvedArg.Table.isInitialized() {
				parameter.ParameterType = gauge_messages.Parameter_Special_Table.Enum()
				parameter.Table = paramResolver.createProtoStepTable(&resolvedArg.Table, dataTableLookup)
			} else {
				parameter.ParameterType = gauge_messages.Parameter_Dynamic.Enum()
				parameter.Value = proto.String(resolvedArg.Value)
			}
		} else if arg.ArgType == SpecialString {
			parameter.ParameterType = gauge_messages.Parameter_Special_String.Enum()
			parameter.Value = proto.String(arg.Value)
		} else if arg.ArgType == SpecialTable {
			parameter.ParameterType = gauge_messages.Parameter_Special_Table.Enum()
			parameter.Table = paramResolver.createProtoStepTable(&arg.Table, dataTableLookup)
		} else {
			parameter.ParameterType = gauge_messages.Parameter_Table.Enum()
			parameter.Table = paramResolver.createProtoStepTable(&arg.Table, dataTableLookup)

		}
		parameters = append(parameters, parameter)
	}

	return parameters

}

func (resolver *paramResolver) createProtoStepTable(table *Table, dataTableLookup *ArgLookup) *gauge_messages.ProtoTable {
	protoTable := new(gauge_messages.ProtoTable)
	protoTable.Headers = &gauge_messages.ProtoTableRow{Cells: table.Headers}
	tableRows := make([]*gauge_messages.ProtoTableRow, 0)
	for i := 0; i < len(table.columns[0]); i++ {
		row := make([]string, 0)
		for _, header := range table.Headers {
			tableCell := table.Get(header)[i]
			value := tableCell.value
			if tableCell.cellType == Dynamic {
				//if concept has a table with dynamic cell, fetch from datatable
				value = dataTableLookup.getArg(tableCell.value).Value
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
		"file": func(filePath string) (*StepArg, error) {
			fileContent, err := common.ReadFileContents(filePath)
			if err != nil {
				return nil, err
			}
			return &StepArg{Value: fileContent, ArgType: SpecialString}, nil
		},
		"table": func(filePath string) (*StepArg, error) {
			csv, err := common.ReadFileContents(filePath)
			if err != nil {
				return nil, err
			}
			csvTable, err := convertCsvToTable(csv)
			if err != nil {
				return nil, err
			}
			return &StepArg{Table: *csvTable, ArgType: SpecialTable}, nil
		},
	}
}

func (resolver *specialTypeResolver) resolve(arg string) (*StepArg, error) {
	//	fmt.Println(arg)
	regEx := regexp.MustCompile("(.*):(.*)")
	match := regEx.FindAllStringSubmatch(arg, -1)
	specialType := strings.TrimSpace(match[0][1])
	value := strings.TrimSpace(match[0][2])
	stepArg, err := resolver.getStepArg(specialType, value, arg)
	if err == nil {
		stepArg.Name = arg
	}
	return stepArg, err
}

func (resolver *specialTypeResolver) getStepArg(specialType string, value string, arg string) (*StepArg, error) {
	resolveFunc, found := resolver.predefinedResolvers[specialType]
	if found {
		return resolveFunc(value)
	}
	return nil, invalidSpecialParamError{message: fmt.Sprintf("Resolver not found for special param <%s>", arg)}
}

// Creating a copy of the lookup and populating table values
func populateConceptDynamicParams(concept *Step, dataTableLookup *ArgLookup) {
	//If it is a top level concept
	if concept.parent == nil {
		lookup := concept.lookup.getCopy()
		for key, _ := range lookup.paramIndexMap {
			conceptLookupArg := lookup.getArg(key)
			if conceptLookupArg.ArgType == Dynamic {
				resolvedArg := dataTableLookup.getArg(conceptLookupArg.Value)
				lookup.addArgValue(key, resolvedArg)
			}
		}
		concept.lookup = *lookup
	}

	//Updating values inside the concept step as well
	newArgs := make([]*StepArg, 0)
	for _, arg := range concept.args {
		if arg.ArgType == Dynamic {
			if concept.parent != nil {
				newArgs = append(newArgs, concept.parent.getArg(arg.Value))
			} else {
				newArgs = append(newArgs, dataTableLookup.getArg(arg.Value))
			}
		} else {
			newArgs = append(newArgs, arg)
		}
	}
	concept.args = newArgs
	concept.populateFragments()
}
