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
	"github.com/getgauge/gauge/gauge_messages"
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
	lookup.AddArgValue("param1", &StepArg{Value: "value1", ArgType: Static})
	step := &Step{Lookup: *lookup}

	c.Assert(step.GetArg("param1").Value, Equals, "value1")
}

func (s *MySuite) TestGetArgForConceptStep(c *C) {
	lookup := new(ArgLookup)
	lookup.AddArgName("param1")
	lookup.AddArgValue("param1", &StepArg{Value: "value1", ArgType: Static})
	concept := &Step{Lookup: *lookup, IsConcept: true}
	stepLookup := new(ArgLookup)
	stepLookup.AddArgName("param1")
	stepLookup.AddArgValue("param1", &StepArg{Value: "param1", ArgType: Dynamic})
	step := &Step{Parent: concept, Lookup: *stepLookup}

	c.Assert(step.GetArg("param1").Value, Equals, "value1")
}

func (s *MySuite) TestPopulateFragmentsForStepWithParameters(c *C) {
	arg1 := &StepArg{Value: "first", ArgType: Static}
	arg2 := &StepArg{Value: "second", ArgType: Dynamic, Name: "second"}
	argTable := new(Table)
	headers := []string{"header1", "header2"}
	row1 := []string{"row1", "row2"}
	argTable.AddHeaders(headers)
	argTable.AddRowValues(row1)
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
	originalStep := &Step{LineNo: 12,
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
		LineNo:         12,
		Value:          "foo {}",
		LineText:       "foo <bar>",
		Args:           argsInStep,
		IsConcept:      true,
		Lookup:         *argLookup,
		Fragments:      fragments,
		ConceptSteps:   conceptSteps,
		HasInlineTable: false}

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
	isRefactored := originalStep.Rename(*originalStep, *newStep, false, orderMap, &IsConcept)

	c.Assert(isRefactored, Equals, true)
	c.Assert(originalStep.Value, Equals, "step from {} {}")
	c.Assert(originalStep.Args[0].Name, Equals, "arg2")
	c.Assert(originalStep.Args[1].Name, Equals, "arg1")
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
	c.Assert(step.Args[0].Value, Equals, "first/first.csv")
	c.Assert(step.Args[1].ArgType, Equals, Dynamic)
	c.Assert(step.Args[1].Value, Equals, "second/second.txt")
}

func (s *MySuite) TestReplaceArgs(c *C) {
	arg1 := &StepArg{Name: "first", ArgType: Static, Value: "text from file"}
	arg2 := &StepArg{Name: "second", ArgType: Static, Value: "text from file"}

	stepArgs := []*StepArg{arg1, arg2}
	step := &Step{Value: "step with {} and {}", Args: stepArgs}

	step.ReplaceArgsWithDynamic(stepArgs)
	c.Assert(step.Args[0].ArgType, Equals, Dynamic)
	c.Assert(step.Args[0].Value, Equals, "text from file")
	c.Assert(step.Args[1].ArgType, Equals, Dynamic)
	c.Assert(step.Args[1].Value, Equals, "text from file")
}

func (s *MySuite) TestGetDynamicParameters(c *C) {
	fragment1 := &gauge_messages.Fragment{
		FragmentType: gauge_messages.Fragment_Text,
		Text:         "print ",
	}

	fragment2 := &gauge_messages.Fragment{
		FragmentType: gauge_messages.Fragment_Text,
		Parameter: &gauge_messages.Parameter{
			ParameterType: gauge_messages.Parameter_Dynamic,
			Name:          "name",
		},
	}

	fragment3 := &gauge_messages.Fragment{
		FragmentType: gauge_messages.Fragment_Text,
		Parameter: &gauge_messages.Parameter{
			ParameterType: gauge_messages.Parameter_Dynamic,
			Name:          "id",
		},
	}

	fragment4 := &gauge_messages.Fragment{
		FragmentType: gauge_messages.Fragment_Text,
		Parameter: &gauge_messages.Parameter{
			ParameterType: gauge_messages.Parameter_Static,
			Value:         "abc",
		},
	}

	fragments := []*gauge_messages.Fragment{fragment1, fragment2, fragment3, fragment4}

	step := &Step{
		Fragments: fragments,
	}

	dynamicParams := step.GetDynamicParamas()
	c.Assert(2, Equals, len(dynamicParams))
	c.Assert(dynamicParams[0], Equals, "name")
	c.Assert(dynamicParams[1], Equals, "id")
}
