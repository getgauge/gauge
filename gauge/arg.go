package gauge

import "fmt"

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

func (lookup *ArgLookup) AddArgValue(param string, stepArg *StepArg) {
	paramIndex, ok := lookup.ParamIndexMap[param]
	if !ok {
		panic(fmt.Sprintf("Accessing an invalid parameter (%s)", param))
	}
	lookup.paramValue[paramIndex].stepArg = stepArg
}

func (lookup *ArgLookup) ContainsArg(param string) bool {
	_, ok := lookup.ParamIndexMap[param]
	return ok
}

func (lookup *ArgLookup) GetArg(param string) *StepArg {
	paramIndex, ok := lookup.ParamIndexMap[param]
	if !ok {
		panic(fmt.Sprintf("Accessing an invalid parameter (%s)", param))
	}
	return lookup.paramValue[paramIndex].stepArg
}

func (lookup *ArgLookup) GetCopy() *ArgLookup {
	lookupCopy := new(ArgLookup)
	for key, _ := range lookup.ParamIndexMap {
		lookupCopy.AddArgName(key)
		arg := lookup.GetArg(key)
		if arg != nil {
			lookupCopy.AddArgValue(key, &StepArg{Value: arg.Value, ArgType: arg.ArgType, Table: arg.Table, Name: arg.Name})
		}
	}
	return lookupCopy
}

func (lookup *ArgLookup) FromDataTableRow(datatable *Table, index int) *ArgLookup {
	dataTableLookup := new(ArgLookup)
	if !datatable.IsInitialized() {
		return dataTableLookup
	}
	for _, header := range datatable.Headers {
		dataTableLookup.AddArgName(header)
		dataTableLookup.AddArgValue(header, &StepArg{Value: datatable.Get(header)[index].Value, ArgType: Static})
	}
	return dataTableLookup
}

//create an empty lookup with only args to resolve dynamic params for steps
func (lookup *ArgLookup) FromDataTable(datatable *Table) *ArgLookup {
	dataTableLookup := new(ArgLookup)
	if !datatable.IsInitialized() {
		return dataTableLookup
	}
	for _, header := range datatable.Headers {
		dataTableLookup.AddArgName(header)
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
