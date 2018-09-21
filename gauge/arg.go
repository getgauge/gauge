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

package gauge

import (
	"fmt"
)

type ArgType string

const (
	Static               ArgType = "static"
	Dynamic              ArgType = "dynamic"
	TableArg             ArgType = "table"
	SpecialString        ArgType = "special_string"
	SpecialTable         ArgType = "special_table"
	ParameterPlaceholder         = "{}"
)

type ArgLookup struct {
	//helps to access the index of an arg at O(1)
	ParamIndexMap map[string]int
	paramValue    []paramNameValue
}

func (lookup ArgLookup) String() string {
	return fmt.Sprintln(lookup.paramValue)
}

func (lookup *ArgLookup) AddArgName(argName string) {
	if lookup.ParamIndexMap == nil {
		lookup.ParamIndexMap = make(map[string]int)
		lookup.paramValue = make([]paramNameValue, 0)
	}
	lookup.ParamIndexMap[argName] = len(lookup.paramValue)
	lookup.paramValue = append(lookup.paramValue, paramNameValue{name: argName})
}

func (lookup *ArgLookup) AddArgValue(param string, stepArg *StepArg) error {
	paramIndex, ok := lookup.ParamIndexMap[param]
	if !ok {
		return fmt.Errorf("Accessing an invalid parameter (%s)", param)
	}
	lookup.paramValue[paramIndex].stepArg = stepArg
	return nil
}

func (lookup *ArgLookup) ContainsArg(param string) bool {
	_, ok := lookup.ParamIndexMap[param]
	return ok
}

func (lookup *ArgLookup) GetArg(param string) (*StepArg, error) {
	paramIndex, ok := lookup.ParamIndexMap[param]
	if !ok {
		return nil, fmt.Errorf("Accessing an invalid parameter (%s)", param)
	}
	return lookup.paramValue[paramIndex].stepArg, nil
}

func (lookup *ArgLookup) GetCopy() (*ArgLookup, error) {
	lookupCopy := new(ArgLookup)
	var err error
	for key, _ := range lookup.ParamIndexMap {
		lookupCopy.AddArgName(key)
		var arg *StepArg
		arg, err = lookup.GetArg(key)
		if arg != nil {
			err = lookupCopy.AddArgValue(key, &StepArg{Value: arg.Value, ArgType: arg.ArgType, Table: arg.Table, Name: arg.Name})
		}
	}
	return lookupCopy, err
}

func (lookup *ArgLookup) FromDataTableRow(datatable *Table, index int) (*ArgLookup, error) {
	dataTableLookup := new(ArgLookup)
	var err error
	if !datatable.IsInitialized() {
		return dataTableLookup, err
	}
	for _, header := range datatable.Headers {
		dataTableLookup.AddArgName(header)
		tableCells, _ := datatable.Get(header)
		err = dataTableLookup.AddArgValue(header, &StepArg{Value: tableCells[index].Value, ArgType: Static})
	}
	return dataTableLookup, err
}

//FromDataTables creates an empty lookup with only args to resolve dynamic params for steps from given list of tables
func (lookup *ArgLookup) FromDataTables(tables ...*Table) *ArgLookup {
	dataTableLookup := new(ArgLookup)
	for _, table := range tables {
		if table.IsInitialized() {
			for _, header := range table.Headers {
				dataTableLookup.AddArgName(header)
			}
		}
	}
	return dataTableLookup
}

type paramNameValue struct {
	name    string
	stepArg *StepArg
}

func (paramNameValue paramNameValue) String() string {
	return fmt.Sprintf("ParamName: %s, stepArg: %s", paramNameValue.name, paramNameValue.stepArg)
}

type StepArg struct {
	Name    string
	Value   string
	ArgType ArgType
	Table   Table
}

func (stepArg *StepArg) String() string {
	return fmt.Sprintf("{Name: %s,value %s,argType %s,table %v}", stepArg.Name, stepArg.Value, string(stepArg.ArgType), stepArg.Table)
}

func (stepArg *StepArg) ArgValue() string {
	switch stepArg.ArgType {
	case Static, Dynamic:
		return stepArg.Value
	case TableArg:
		return "table"
	case SpecialString, SpecialTable:
		return stepArg.Name
	}
	return ""
}
