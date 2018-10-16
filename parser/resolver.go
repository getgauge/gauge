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
	"regexp"
	"strings"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/util"
)

type invalidSpecialParamError struct {
	message string
}

type resolverFn func(string) (*gauge.StepArg, error)
type specialTypeResolver struct {
	predefinedResolvers map[string]resolverFn
}

func (invalidSpecialParamError invalidSpecialParamError) Error() string {
	return invalidSpecialParamError.message
}

//Resolve takes a step, a lookup and updates the target after reconciling the dynamic paramters from the given lookup
func Resolve(step *gauge.Step, parent *gauge.Step, lookup *gauge.ArgLookup, target *gauge_messages.ProtoStep) error {
	stepParameters, err := getResolvedParams(step, parent, lookup)
	if err != nil {
		return err
	}
	paramIndex := 0
	for fragmentIndex, fragment := range target.Fragments {
		if fragment.GetFragmentType() == gauge_messages.Fragment_Parameter {
			target.Fragments[fragmentIndex].Parameter = stepParameters[paramIndex]
			paramIndex++
		}
	}
	return nil
}

// getResolvedParams based on the arg type(static, dynamic, table, special_string, special_table) resolves the parameter of a step.
func getResolvedParams(step *gauge.Step, parent *gauge.Step, lookup *gauge.ArgLookup) ([]*gauge_messages.Parameter, error) {
	parameters := make([]*gauge_messages.Parameter, 0)
	for _, arg := range step.Args {
		parameter := new(gauge_messages.Parameter)
		parameter.Name = arg.Name
		if arg.ArgType == gauge.Static {
			parameter.ParameterType = gauge_messages.Parameter_Static
			parameter.Value = arg.Value
		} else if arg.ArgType == gauge.Dynamic {
			var resolvedArg *gauge.StepArg
			var err error
			if parent != nil {
				resolvedArg, err = parent.GetArg(arg.Value)
			} else {
				resolvedArg, err = lookup.GetArg(arg.Value)
			}
			if err != nil {
				return nil, err
			}
			//In case a special table used in a concept, you will get a dynamic table value which has to be resolved from the concept lookup
			parameter.Name = resolvedArg.Name
			if resolvedArg.Table.IsInitialized() {
				parameter.ParameterType = gauge_messages.Parameter_Special_Table
				table, err := createProtoStepTable(&resolvedArg.Table, lookup)
				if err != nil {
					return nil, err
				}
				parameter.Table = table
			} else {
				parameter.ParameterType = gauge_messages.Parameter_Dynamic
				parameter.Value = resolvedArg.Value
			}
		} else if arg.ArgType == gauge.SpecialString {
			parameter.ParameterType = gauge_messages.Parameter_Special_String
			parameter.Value = arg.Value
		} else if arg.ArgType == gauge.SpecialTable {
			parameter.ParameterType = gauge_messages.Parameter_Special_Table
			table, err := createProtoStepTable(&arg.Table, lookup)
			if err != nil {
				return nil, err
			}
			parameter.Table = table
		} else {
			parameter.ParameterType = gauge_messages.Parameter_Table
			table, err := createProtoStepTable(&arg.Table, lookup)
			if err != nil {
				return nil, err
			}
			parameter.Table = table
		}
		parameters = append(parameters, parameter)
	}

	return parameters, nil
}

func createProtoStepTable(table *gauge.Table, lookup *gauge.ArgLookup) (*gauge_messages.ProtoTable, error) {
	protoTable := new(gauge_messages.ProtoTable)
	protoTable.Headers = &gauge_messages.ProtoTableRow{Cells: table.Headers}
	tableRows := make([]*gauge_messages.ProtoTableRow, 0)
	if len(table.Columns) == 0 {
		protoTable.Rows = tableRows
		return protoTable, nil
	}
	for i := 0; i < len(table.Columns[0]); i++ {
		row := make([]string, 0)
		for _, header := range table.Headers {
			tableCells, _ := table.Get(header)
			value := tableCells[i].Value
			if tableCells[i].CellType == gauge.Dynamic {
				//if concept has a table with dynamic cell, fetch from datatable
				arg, err := lookup.GetArg(tableCells[i].Value)
				if err != nil {
					return nil, err
				}
				value = arg.Value
			} else if tableCells[i].CellType == gauge.SpecialString {
				resolvedArg, _ := newSpecialTypeResolver().resolve(value)
				value = resolvedArg.Value
			}
			row = append(row, value)
		}
		tableRows = append(tableRows, &gauge_messages.ProtoTableRow{Cells: row})
	}
	protoTable.Rows = tableRows
	return protoTable, nil
}

func newSpecialTypeResolver() *specialTypeResolver {
	resolver := new(specialTypeResolver)
	resolver.predefinedResolvers = initializePredefinedResolvers()
	return resolver
}

func initializePredefinedResolvers() map[string]resolverFn {
	return map[string]resolverFn{
		"file": func(filePath string) (*gauge.StepArg, error) {
			fileContent, err := util.GetFileContents(filePath)
			if err != nil {
				return nil, err
			}
			return &gauge.StepArg{Value: fileContent, ArgType: gauge.SpecialString}, nil
		},
		"table": func(filePath string) (*gauge.StepArg, error) {
			csv, err := util.GetFileContents(filePath)
			if err != nil {
				return nil, err
			}
			csvTable, err := convertCsvToTable(csv)
			if err != nil {
				return nil, err
			}
			return &gauge.StepArg{Table: *csvTable, ArgType: gauge.SpecialTable}, nil
		},
	}
}

func (resolver *specialTypeResolver) resolve(arg string) (*gauge.StepArg, error) {
	if util.IsWindows() {
		arg = GetUnescapedString(arg)
	}
	regEx := regexp.MustCompile("(.*?):(.*)")
	match := regEx.FindAllStringSubmatch(arg, -1)
	specialType := strings.TrimSpace(match[0][1])
	value := strings.TrimSpace(match[0][2])
	stepArg, err := resolver.getStepArg(specialType, value, arg)
	if err == nil {
		stepArg.Name = arg
	}
	return stepArg, err
}

func (resolver *specialTypeResolver) getStepArg(specialType string, value string, arg string) (*gauge.StepArg, error) {
	resolveFunc, found := resolver.predefinedResolvers[specialType]
	if found {
		return resolveFunc(value)
	}
	return nil, invalidSpecialParamError{message: fmt.Sprintf("Resolver not found for special param <%s>", arg)}
}

// PopulateConceptDynamicParams creates a copy of the lookup and populates table values
func PopulateConceptDynamicParams(concept *gauge.Step, dataTableLookup *gauge.ArgLookup) error {
	//If it is a top level concept
	lookup, err := concept.Lookup.GetCopy()
	if err != nil {
		return err
	}
	for key := range lookup.ParamIndexMap {
		conceptLookupArg, err := lookup.GetArg(key)
		if err != nil {
			return err
		}
		if conceptLookupArg.ArgType == gauge.Dynamic {
			resolvedArg, err := dataTableLookup.GetArg(conceptLookupArg.Value)
			if err != nil {
				return err
			}
			if err = lookup.AddArgValue(key, resolvedArg); err != nil {
				return err
			}
		}
	}
	concept.Lookup = *lookup

	//Updating values inside the concept step as well
	newArgs := make([]*gauge.StepArg, 0)
	for _, arg := range concept.Args {
		if arg.ArgType == gauge.Dynamic {
			if concept.Parent != nil {
				cArg, err := concept.Parent.GetArg(arg.Value)
				if err != nil {
					return err
				}
				newArgs = append(newArgs, cArg)
			} else {
				dArg, err := dataTableLookup.GetArg(arg.Value)
				if err != nil {
					return err
				}
				newArgs = append(newArgs, dArg)
			}
		} else {
			newArgs = append(newArgs, arg)
		}
	}
	concept.Args = newArgs
	concept.PopulateFragments()
	return nil
}

// GetResolvedDataTablerows resolves any dynamic parameters in a table cell
func GetResolvedDataTablerows(table gauge.Table) {
	for i, cells := range table.Columns {
		for j, cell := range cells {
			if cell.CellType == gauge.SpecialString {
				resolvedArg, _ := newSpecialTypeResolver().resolve(cell.Value)
				table.Columns[i][j].Value = resolvedArg.Value
			}
		}
	}
}
