/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package gauge

import (
	"github.com/getgauge/gauge-proto/go/gauge_messages"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestPopulateFragmentsForSimpleStep(c *C) {
	step := &Step{Value: "This is a simple step"}

	step.PopulateFragments()

	c.Assert(len(step.Fragments), Equals, 1)
	fragment := step.Fragments[0]
	c.Assert(fragment.GetText(), Equals, "This is a simple step")
	c.Assert(fragment.GetFragmentType(), Equals, gauge_messages.Fragment_Text)
}

func (s *MySuite) TestGetArgForStep(c *C) {
	lookup := new(ArgLookup)
	lookup.AddArgName("param1")
	err := lookup.AddArgValue("param1", &StepArg{Value: "value1", ArgType: Static})
	c.Assert(err, IsNil)

	step := &Step{Lookup: *lookup}
	stepArg, err := step.GetArg("param1")
	c.Assert(err, IsNil)
	c.Assert(stepArg.Value, Equals, "value1")
}

func (s *MySuite) TestGetArgForConceptStep(c *C) {
	lookup := new(ArgLookup)
	lookup.AddArgName("param1")
	err := lookup.AddArgValue("param1", &StepArg{Value: "value1", ArgType: Static})
	c.Assert(err, IsNil)

	concept := &Step{Lookup: *lookup, IsConcept: true}
	stepLookup := new(ArgLookup)
	stepLookup.AddArgName("param1")
	err = stepLookup.AddArgValue("param1", &StepArg{Value: "param1", ArgType: Dynamic})
	c.Assert(err, IsNil)

	step := &Step{Parent: concept, Lookup: *stepLookup}

	stepArg, err := step.GetArg("param1")
	c.Assert(err, IsNil)
	c.Assert(stepArg.Value, Equals, "value1")
}

func (s *MySuite) TestPopulateFragmentsForStepWithParameters(c *C) {
	arg1 := &StepArg{Value: "first", ArgType: Static}
	arg2 := &StepArg{Value: "second", ArgType: Dynamic, Name: "second"}
	argTable := new(Table)
	headers := []string{"header1", "header2"}
	row1 := []string{"row1", "row2"}
	argTable.AddHeaders(headers)
	argTable.AddRowValues(argTable.CreateTableCells(row1))
	arg3 := &StepArg{ArgType: SpecialString, Value: "text from file", Name: "file:foo.txt"}
	arg4 := &StepArg{Table: *argTable, ArgType: TableArg}
	stepArgs := []*StepArg{arg1, arg2, arg3, arg4}
	step := &Step{Value: "{} step with {} and {}, {}", Args: stepArgs}

	step.PopulateFragments()

	c.Assert(len(step.Fragments), Equals, 7)
	fragment1 := step.Fragments[0]
	c.Assert(fragment1.GetFragmentType(), Equals, gauge_messages.Fragment_Parameter)
	c.Assert(fragment1.GetParameter().GetValue(), Equals, "first")
	c.Assert(fragment1.GetParameter().GetParameterType(), Equals, gauge_messages.Parameter_Static)

	fragment2 := step.Fragments[1]
	c.Assert(fragment2.GetText(), Equals, " step with ")
	c.Assert(fragment2.GetFragmentType(), Equals, gauge_messages.Fragment_Text)

	fragment3 := step.Fragments[2]
	c.Assert(fragment3.GetFragmentType(), Equals, gauge_messages.Fragment_Parameter)
	c.Assert(fragment3.GetParameter().GetValue(), Equals, "second")
	c.Assert(fragment3.GetParameter().GetParameterType(), Equals, gauge_messages.Parameter_Dynamic)

	fragment4 := step.Fragments[3]
	c.Assert(fragment4.GetText(), Equals, " and ")
	c.Assert(fragment4.GetFragmentType(), Equals, gauge_messages.Fragment_Text)

	fragment5 := step.Fragments[4]
	c.Assert(fragment5.GetFragmentType(), Equals, gauge_messages.Fragment_Parameter)
	c.Assert(fragment5.GetParameter().GetValue(), Equals, "text from file")
	c.Assert(fragment5.GetParameter().GetParameterType(), Equals, gauge_messages.Parameter_Special_String)
	c.Assert(fragment5.GetParameter().GetName(), Equals, "file:foo.txt")

	fragment6 := step.Fragments[5]
	c.Assert(fragment6.GetText(), Equals, ", ")
	c.Assert(fragment6.GetFragmentType(), Equals, gauge_messages.Fragment_Text)

	fragment7 := step.Fragments[6]
	c.Assert(fragment7.GetFragmentType(), Equals, gauge_messages.Fragment_Parameter)
	c.Assert(fragment7.GetParameter().GetParameterType(), Equals, gauge_messages.Parameter_Table)
	protoTable := fragment7.GetParameter().GetTable()
	c.Assert(protoTable.GetHeaders().GetCells(), DeepEquals, headers)
	c.Assert(len(protoTable.GetRows()), Equals, 1)
	c.Assert(protoTable.GetRows()[0].GetCells(), DeepEquals, row1)
}

func (s *MySuite) TestUpdatePropertiesFromAnotherStep(c *C) {
	argsInStep := []*StepArg{&StepArg{Name: "arg1", Value: "arg value", ArgType: Dynamic}}
	fragments := []*gauge_messages.Fragment{&gauge_messages.Fragment{Text: "foo"}}
	originalStep := &Step{
		LineNo:         0,
		Value:          "foo {}",
		LineText:       "foo <bar>",
		Args:           argsInStep,
		IsConcept:      false,
		Fragments:      fragments,
		HasInlineTable: false}

	destinationStep := new(Step)
	destinationStep.CopyFrom(originalStep)

	c.Assert(destinationStep, DeepEquals, originalStep)
}

func (s *MySuite) TestUpdatePropertiesFromAnotherConcept(c *C) {
	argsInStep := []*StepArg{&StepArg{Name: "arg1", Value: "arg value", ArgType: Dynamic}}
	argLookup := new(ArgLookup)
	argLookup.AddArgName("name")
	argLookup.AddArgName("id")
	fragments := []*gauge_messages.Fragment{&gauge_messages.Fragment{Text: "foo"}}
	conceptSteps := []*Step{&Step{Value: "step 1"}}
	originalConcept := &Step{
		LineNo:         0,
		Value:          "foo {}",
		LineText:       "foo <bar>",
		Args:           argsInStep,
		IsConcept:      true,
		Lookup:         *argLookup,
		Fragments:      fragments,
		ConceptSteps:   conceptSteps,
		HasInlineTable: false,
	}

	destinationConcept := new(Step)
	destinationConcept.CopyFrom(originalConcept)

	c.Assert(destinationConcept, DeepEquals, originalConcept)
}

func (s *MySuite) TestRenameStep(c *C) {
	argsInStep := []*StepArg{&StepArg{Name: "arg1", Value: "value", ArgType: Static}, &StepArg{Name: "arg2", Value: "value1", ArgType: Static}}
	originalStep := &Step{
		LineNo:         12,
		Value:          "step with {}",
		Args:           argsInStep,
		IsConcept:      false,
		HasInlineTable: false}

	argsInStep = []*StepArg{&StepArg{Name: "arg2", Value: "value1", ArgType: Static}, &StepArg{Name: "arg1", Value: "value", ArgType: Static}}
	newStep := &Step{
		LineNo:         12,
		Value:          "step from {} {}",
		Args:           argsInStep,
		IsConcept:      false,
		HasInlineTable: false}
	orderMap := make(map[int]int)
	orderMap[0] = 1
	orderMap[1] = 0
	IsConcept := false
	diff, isRefactored := originalStep.Rename(originalStep, newStep, false, orderMap, &IsConcept)

	c.Assert(isRefactored, Equals, true)
	c.Assert(originalStep.Value, Equals, "step from {} {}")
	c.Assert(originalStep.Args[0].Name, Equals, "arg2")
	c.Assert(originalStep.Args[1].Name, Equals, "arg1")
	c.Assert(diff.OldStep.Value, Equals, "step with {}")
	c.Assert(diff.NewStep.Value, Equals, "step from {} {}")
}

func (s *MySuite) TestRenameConcept(c *C) {
	originalStep := &Step{
		LineNo:         3,
		Value:          "concept with text file",
		IsConcept:      true,
		HasInlineTable: false}

	argsInStep := []*StepArg{&StepArg{Name: "file:foo.txt", Value: "text from file", ArgType: SpecialString}}
	newStep := &Step{
		LineNo:         3,
		Value:          "concept with text file {}",
		Args:           argsInStep,
		IsConcept:      true,
		HasInlineTable: false}
	orderMap := make(map[int]int)
	orderMap[0] = -1
	IsConcept := true
	diff, isRefactored := originalStep.Rename(originalStep, newStep, false, orderMap, &IsConcept)
	c.Assert(isRefactored, Equals, true)
	c.Assert(originalStep.Value, Equals, "concept with text file {}")
	c.Assert(originalStep.Args[0].Name, Equals, "arg0")
	c.Assert(newStep.Args[0].Name, Equals, "file:foo.txt")
	c.Assert(diff.OldStep.Value, Equals, "concept with text file")
	c.Assert(diff.NewStep.Value, Equals, "concept with text file {}")
}

func (s *MySuite) TestGetLineTextForStep(c *C) {
	step := &Step{LineText: "foo"}

	c.Assert(step.GetLineText(), Equals, "foo")
}

func (s *MySuite) TestGetLineTextForStepWithTable(c *C) {
	step := &Step{
		LineText:       "foo",
		HasInlineTable: true}

	c.Assert(step.GetLineText(), Equals, "foo <table>")
}

func (s *MySuite) TestReplaceArgsWithDynamicForSpecialParam(c *C) {
	arg1 := &StepArg{Name: "table:first/first.csv", ArgType: SpecialString, Value: "text from file"}
	arg2 := &StepArg{Name: "file:second/second.txt", ArgType: SpecialTable, Value: "text from file"}

	stepArgs := []*StepArg{arg1, arg2}
	step := &Step{Value: "step with {} and {}", Args: stepArgs}

	step.ReplaceArgsWithDynamic(stepArgs)
	c.Assert(step.Args[0].ArgType, Equals, Dynamic)
	c.Assert(step.Args[0].Name, Equals, "first/first.csv")
	c.Assert(step.Args[1].ArgType, Equals, Dynamic)
	c.Assert(step.Args[1].Name, Equals, "second/second.txt")
}

func (s *MySuite) TestReplaceArgs(c *C) {
	arg1 := &StepArg{Name: "first", ArgType: Static, Value: "text from file"}
	arg2 := &StepArg{Name: "second", ArgType: Static, Value: "text from file"}

	stepArgs := []*StepArg{arg1, arg2}
	step := &Step{Value: "step with {} and {}", Args: stepArgs}

	step.ReplaceArgsWithDynamic(stepArgs)
	c.Assert(step.Args[0].ArgType, Equals, Dynamic)
	c.Assert(step.Args[0].Name, Equals, "text from file")
	c.Assert(step.Args[1].ArgType, Equals, Dynamic)
	c.Assert(step.Args[1].Name, Equals, "text from file")
}

func (s *MySuite) TestUsageDynamicArgs(c *C) {

	dArgs := []string{"first", "second"}

	sArg := &StepArg{ArgType: Dynamic, Value: "first"}

	step := &Step{Value: "step with {}", Args: []*StepArg{sArg}}

	usesDynamicArgs := step.UsesDynamicArgs(dArgs...)

	c.Assert(usesDynamicArgs, Equals, true)

}

func (s *MySuite) TestStepDoesNotUsageDynamicArgs(c *C) {

	dArgs := []string{"first", "second"}

	sArg := &StepArg{ArgType: Dynamic, Value: "third"}

	step := &Step{Value: "step with {}", Args: []*StepArg{sArg}}

	usesDynamicArgs := step.UsesDynamicArgs(dArgs...)

	c.Assert(usesDynamicArgs, Equals, false)

}

func (s *MySuite) TestInlineTableUsageDynamicArgs(c *C) {
	headers := []string{"header"}
	cells := [][]TableCell{
		{
			{
				CellType: Dynamic,
				Value:    "first",
			},
		},
	}
	table := NewTable(headers, cells, 1)

	dArgs := []string{"first", "second"}

	sArg := &StepArg{Name: "hello", ArgType: TableArg, Table: *table}

	step := &Step{Value: "step with {}", Args: []*StepArg{sArg}}

	usesDynamicArgs := step.UsesDynamicArgs(dArgs...)

	c.Assert(usesDynamicArgs, Equals, true)

}

func (s *MySuite) TestLastArgs(c *C) {
	headers := []string{"header"}
	cells := [][]TableCell{
		{
			{
				CellType: Dynamic,
				Value:    "first",
			},
		},
	}

	table := NewTable(headers, cells, 1)

	dArg := &StepArg{Name: "hello", ArgType: TableArg, Table: *table}

	step := &Step{Value: "step with {}", Args: []*StepArg{dArg}}

	la := step.GetLastArg()

	c.Assert(la, DeepEquals, dArg)

}
