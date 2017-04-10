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

package execution

import (
	"fmt"
	"path/filepath"

	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/validation"
	. "gopkg.in/check.v1"
)

type specBuilder struct {
	lines []string
}

func SpecBuilder() *specBuilder {
	return &specBuilder{lines: make([]string, 0)}
}

func (specBuilder *specBuilder) addPrefix(prefix string, line string) string {
	return fmt.Sprintf("%s%s\n", prefix, line)
}

func (specBuilder *specBuilder) String() string {
	var result string
	for _, line := range specBuilder.lines {
		result = fmt.Sprintf("%s%s", result, line)
	}
	return result
}

func (specBuilder *specBuilder) specHeading(heading string) *specBuilder {
	line := specBuilder.addPrefix("#", heading)
	specBuilder.lines = append(specBuilder.lines, line)
	return specBuilder
}

func (specBuilder *specBuilder) scenarioHeading(heading string) *specBuilder {
	line := specBuilder.addPrefix("##", heading)
	specBuilder.lines = append(specBuilder.lines, line)
	return specBuilder
}

func (specBuilder *specBuilder) step(stepText string) *specBuilder {
	line := specBuilder.addPrefix("* ", stepText)
	specBuilder.lines = append(specBuilder.lines, line)
	return specBuilder
}

func (specBuilder *specBuilder) tags(tags ...string) *specBuilder {
	tagText := ""
	for i, tag := range tags {
		tagText = fmt.Sprintf("%s%s", tagText, tag)
		if i != len(tags)-1 {
			tagText = fmt.Sprintf("%s,", tagText)
		}
	}
	line := specBuilder.addPrefix("tags: ", tagText)
	specBuilder.lines = append(specBuilder.lines, line)
	return specBuilder
}

func (specBuilder *specBuilder) tableHeader(cells ...string) *specBuilder {
	return specBuilder.tableRow(cells...)
}
func (specBuilder *specBuilder) tableRow(cells ...string) *specBuilder {
	rowInMarkdown := "|"
	for _, cell := range cells {
		rowInMarkdown = fmt.Sprintf("%s%s|", rowInMarkdown, cell)
	}
	specBuilder.lines = append(specBuilder.lines, fmt.Sprintf("%s\n", rowInMarkdown))
	return specBuilder
}

func (specBuilder *specBuilder) text(comment string) *specBuilder {
	specBuilder.lines = append(specBuilder.lines, fmt.Sprintf("%s\n", comment))
	return specBuilder
}

func (s *MySuite) TestResolveConceptToProtoConceptItem(c *C) {
	conceptDictionary := gauge.NewConceptDictionary()

	specText := SpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("create user \"456\" \"foo\" and \"9900\"").
		String()
	path, _ := filepath.Abs(filepath.Join("testdata", "concept.cpt"))
	parser.AddConcepts(path, conceptDictionary)

	spec, _ := new(parser.SpecParser).Parse(specText, conceptDictionary, "")

	specExecutor := newSpecExecutor(spec, nil, nil, nil, 0)
	specExecutor.errMap = getValidationErrorMap()
	protoConcept := specExecutor.resolveToProtoConceptItem(*spec.Scenarios[0].Steps[0]).GetConcept()

	checkConceptParameterValuesInOrder(c, protoConcept, "456", "foo", "9900")
	firstNestedStep := protoConcept.GetSteps()[0].GetConcept().GetSteps()[0].GetStep()
	params := getParameters(firstNestedStep.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, gauge_messages.Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "456")

	secondNestedStep := protoConcept.GetSteps()[0].GetConcept().GetSteps()[1].GetStep()
	params = getParameters(secondNestedStep.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, gauge_messages.Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "foo")

	secondStep := protoConcept.GetSteps()[1].GetStep()
	params = getParameters(secondStep.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, gauge_messages.Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "9900")

}

func (s *MySuite) TestResolveNestedConceptToProtoConceptItem(c *C) {
	conceptDictionary := gauge.NewConceptDictionary()

	specText := SpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("create user \"456\" \"foo\" and \"9900\"").
		String()

	path, _ := filepath.Abs(filepath.Join("testdata", "concept.cpt"))
	parser.AddConcepts(path, conceptDictionary)
	specParser := new(parser.SpecParser)
	spec, _ := specParser.Parse(specText, conceptDictionary, "")

	specExecutor := newSpecExecutor(spec, nil, nil, nil, 0)
	specExecutor.errMap = getValidationErrorMap()
	protoConcept := specExecutor.resolveToProtoConceptItem(*spec.Scenarios[0].Steps[0]).GetConcept()
	checkConceptParameterValuesInOrder(c, protoConcept, "456", "foo", "9900")

	c.Assert(protoConcept.GetSteps()[0].GetItemType(), Equals, gauge_messages.ProtoItem_Concept)

	nestedConcept := protoConcept.GetSteps()[0].GetConcept()
	checkConceptParameterValuesInOrder(c, nestedConcept, "456", "foo")

	firstNestedStep := nestedConcept.GetSteps()[0].GetStep()
	params := getParameters(firstNestedStep.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, gauge_messages.Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "456")

	secondNestedStep := nestedConcept.GetSteps()[1].GetStep()
	params = getParameters(secondNestedStep.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, gauge_messages.Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "foo")

	c.Assert(protoConcept.GetSteps()[1].GetItemType(), Equals, gauge_messages.ProtoItem_Step)
	secondStepInConcept := protoConcept.GetSteps()[1].GetStep()
	params = getParameters(secondStepInConcept.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, gauge_messages.Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "9900")

}

func (s *MySuite) TestResolveToProtoConceptItemWithDataTable(c *C) {
	conceptDictionary := gauge.NewConceptDictionary()

	specText := SpecBuilder().specHeading("A spec heading").
		tableHeader("id", "name", "phone").
		tableHeader("123", "foo", "8800").
		tableHeader("666", "bar", "9900").
		scenarioHeading("First scenario").
		step("create user <id> <name> and <phone>").
		String()

	path, _ := filepath.Abs(filepath.Join("testdata", "concept.cpt"))
	parser.AddConcepts(path, conceptDictionary)
	specParser := new(parser.SpecParser)
	spec, _ := specParser.Parse(specText, conceptDictionary, "")

	specExecutor := newSpecExecutor(spec, nil, nil, nil, 0)

	// For first row
	specExecutor.currentTableRow = 0
	specExecutor.errMap = gauge.NewBuildErrors()
	protoConcept := specExecutor.resolveToProtoConceptItem(*spec.Scenarios[0].Steps[0]).GetConcept()
	checkConceptParameterValuesInOrder(c, protoConcept, "123", "foo", "8800")

	c.Assert(protoConcept.GetSteps()[0].GetItemType(), Equals, gauge_messages.ProtoItem_Concept)
	nestedConcept := protoConcept.GetSteps()[0].GetConcept()
	checkConceptParameterValuesInOrder(c, nestedConcept, "123", "foo")
	firstNestedStep := nestedConcept.GetSteps()[0].GetStep()
	params := getParameters(firstNestedStep.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, gauge_messages.Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "123")

	secondNestedStep := nestedConcept.GetSteps()[1].GetStep()
	params = getParameters(secondNestedStep.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, gauge_messages.Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "foo")

	c.Assert(protoConcept.GetSteps()[1].GetItemType(), Equals, gauge_messages.ProtoItem_Step)
	secondStepInConcept := protoConcept.GetSteps()[1].GetStep()
	params = getParameters(secondStepInConcept.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, gauge_messages.Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "8800")

	// For second row
	specExecutor.currentTableRow = 1
	protoConcept = specExecutor.resolveToProtoConceptItem(*spec.Scenarios[0].Steps[0]).GetConcept()
	c.Assert(protoConcept.GetSteps()[0].GetItemType(), Equals, gauge_messages.ProtoItem_Concept)
	checkConceptParameterValuesInOrder(c, protoConcept, "666", "bar", "9900")

	nestedConcept = protoConcept.GetSteps()[0].GetConcept()
	checkConceptParameterValuesInOrder(c, nestedConcept, "666", "bar")
	firstNestedStep = nestedConcept.GetSteps()[0].GetStep()
	params = getParameters(firstNestedStep.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, gauge_messages.Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "666")

	secondNestedStep = nestedConcept.GetSteps()[1].GetStep()
	params = getParameters(secondNestedStep.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, gauge_messages.Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "bar")

	c.Assert(protoConcept.GetSteps()[1].GetItemType(), Equals, gauge_messages.ProtoItem_Step)
	secondStepInConcept = protoConcept.GetSteps()[1].GetStep()
	params = getParameters(secondStepInConcept.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, gauge_messages.Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "9900")
}

func checkConceptParameterValuesInOrder(c *C, concept *gauge_messages.ProtoConcept, paramValues ...string) {
	params := getParameters(concept.GetConceptStep().Fragments)
	c.Assert(len(params), Equals, len(paramValues))
	for i, param := range params {
		c.Assert(param.GetValue(), Equals, paramValues[i])
	}

}

type tableRow struct {
	name           string
	input          string // input by user for data table rows
	output         []int  // data table indexes to be executed
	tableRowsCount int    // total rows in given data table
}

var tableRowTests = []*tableRow{
	{"Valid single row number", "2", []int{1}, 5},
	{"Valid row numbers list", "2,3,4", []int{1, 2, 3}, 4},
	{"Valid table rows range", "2-5", []int{1, 2, 3, 4}, 5},
	{"Empty table rows range", "", []int{0, 1, 2, 3}, 4},
	{"Table rows list with spaces", "2, 4 ", []int{1, 3}, 4},
	{"Row count is zero with empty input", "", []int{}, 0},
	{"Row count is non-zero with empty input", "", []int{0, 1}, 2},
	{"Row count is non-zero with non-empty input", "2", []int{1}, 2},
}

func (s *MySuite) TestToGetDataTableRowsRangeFromInputFlag(c *C) {
	for _, test := range tableRowTests {
		TableRows = test.input
		got := getDataTableRows(test.tableRowsCount)
		want := test.output
		c.Assert(got, DeepEquals, want, Commentf(test.name))
	}
}

func (s *MySuite) TestCreateSkippedSpecResult(c *C) {
	spec := &gauge.Specification{Heading: &gauge.Heading{LineNo: 0, Value: "SPEC_HEADING"}, FileName: "FILE"}

	se := newSpecExecutor(spec, nil, nil, nil, 0)
	se.errMap = getValidationErrorMap()
	se.specResult = &result.SpecResult{}
	se.skipSpecForError(fmt.Errorf("ERROR"))

	c.Assert(se.specResult.IsFailed, Equals, false)
	c.Assert(se.specResult.Skipped, Equals, true)

	// c.Assert(len(se.errMap.SpecErrs[spec]), Equals, 1)
	// c.Assert(se.errMap.SpecErrs[spec][0].message, Equals, "ERROR")
	// c.Assert(se.errMap.SpecErrs[spec][0].fileName, Equals, "FILE")
	// c.Assert(se.errMap.SpecErrs[spec][0].step.LineNo, Equals, 0)
	// c.Assert(se.errMap.SpecErrs[spec][0].step.LineText, Equals, "SPEC_HEADING")
}

func (s *MySuite) TestCreateSkippedSpecResultWithScenarios(c *C) {
	se := newSpecExecutor(anySpec(), nil, nil, nil, 0)
	se.errMap = getValidationErrorMap()
	se.specResult = &result.SpecResult{ProtoSpec: &gauge_messages.ProtoSpec{}}
	se.skipSpecForError(fmt.Errorf("ERROR"))

	c.Assert(se.specResult.IsFailed, Equals, false)
	c.Assert(se.specResult.Skipped, Equals, true)

	// c.Assert(len(specExecutor.errMap.SpecErrs[spec]), Equals, 1)
	// c.Assert(specExecutor.errMap.SpecErrs[spec][0].message, Equals, "ERROR")
	// c.Assert(specExecutor.errMap.SpecErrs[spec][0].fileName, Equals, "FILE")
	// c.Assert(specExecutor.errMap.SpecErrs[spec][0].step.LineNo, Equals, 1)
	// c.Assert(specExecutor.errMap.SpecErrs[spec][0].step.LineText, Equals, "A spec heading")
	// c.Assert(len(specExecutor.errMap.ScenarioErrs[spec.Scenarios[0]]), Equals, 1)
	// c.Assert(specExecutor.errMap.ScenarioErrs[spec.Scenarios[0]][0].message, Equals, "ERROR")
	// c.Assert(specExecutor.errMap.ScenarioErrs[spec.Scenarios[0]][0].fileName, Equals, "FILE")
	// c.Assert(specExecutor.errMap.ScenarioErrs[spec.Scenarios[0]][0].step.LineNo, Equals, 1)
	// c.Assert(specExecutor.errMap.ScenarioErrs[spec.Scenarios[0]][0].step.LineText, Equals, "A spec heading")
}

func (s *MySuite) TestSkipSpecWithDataTableScenarios(c *C) {
	stepText := "Unimplememted step"

	specText := SpecBuilder().specHeading("A spec heading").
		tableHeader("id", "name", "phone").
		tableRow("123", "foo", "8800").
		tableRow("666", "bar", "9900").
		scenarioHeading("First scenario").
		step(stepText).
		step("create user <id> <name> and <phone>").
		String()

	spec, _ := new(parser.SpecParser).Parse(specText, gauge.NewConceptDictionary(), "")

	errMap := &gauge.BuildErrors{
		SpecErrs:     make(map[*gauge.Specification][]error),
		ScenarioErrs: make(map[*gauge.Scenario][]error),
		StepErrs:     make(map[*gauge.Step]error),
	}

	errMap.SpecErrs[spec] = []error{validation.NewSpecValidationError("Step implementation not found", spec.FileName)}
	se := newSpecExecutor(spec, nil, nil, errMap, 0)
	specInfo := &gauge_messages.SpecInfo{Name: se.specification.Heading.Value,
		FileName: se.specification.FileName,
		IsFailed: false, Tags: getTagValue(se.specification.Tags)}
	se.currentExecutionInfo = &gauge_messages.ExecutionInfo{CurrentSpec: specInfo}
	se.specResult = gauge.NewSpecResult(se.specification)
	resolvedSpecItems := se.resolveItems(se.specification.GetSpecItems())
	se.specResult.AddSpecItems(resolvedSpecItems)

	se.skipSpec()

	c.Assert(se.specResult.ProtoSpec.GetIsTableDriven(), Equals, true)
	c.Assert(len(se.specResult.ProtoSpec.GetItems()), Equals, 3)

}

func anySpec() *gauge.Specification {

	specText := SpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("create user \"456\" \"foo\" and \"9900\"").
		String()

	spec, _ := new(parser.SpecParser).Parse(specText, gauge.NewConceptDictionary(), "")
	spec.FileName = "FILE"
	return spec
}

func (s *MySuite) TestSpecIsSkippedIfDataRangeIsInvalid(c *C) {
	errMap := &gauge.BuildErrors{
		SpecErrs:     make(map[*gauge.Specification][]error),
		ScenarioErrs: make(map[*gauge.Scenario][]error),
		StepErrs:     make(map[*gauge.Step]error),
	}
	spec := anySpec()
	errMap.SpecErrs[spec] = []error{validation.NewSpecValidationError("Table row number out of range", spec.FileName)}
	se := newSpecExecutor(spec, nil, nil, errMap, 0)

	specResult := se.execute(true, true, true)
	c.Assert(specResult.Skipped, Equals, true)
}

func (s *MySuite) TestDataTableRowsAreSkippedForUnimplemetedStep(c *C) {
	stepText := "Unimplememted step"

	specText := SpecBuilder().specHeading("A spec heading").
		tableHeader("id", "name", "phone").
		tableRow("123", "foo", "8800").
		tableRow("666", "bar", "9900").
		scenarioHeading("First scenario").
		step(stepText).
		step("create user <id> <name> and <phone>").
		String()

	spec, _ := new(parser.SpecParser).Parse(specText, gauge.NewConceptDictionary(), "")

	errMap := &gauge.BuildErrors{
		SpecErrs:     make(map[*gauge.Specification][]error),
		ScenarioErrs: make(map[*gauge.Scenario][]error),
		StepErrs:     make(map[*gauge.Step]error),
	}

	errMap.SpecErrs[spec] = []error{validation.NewSpecValidationError("Step implementation not found", spec.FileName)}
	se := newSpecExecutor(spec, nil, nil, errMap, 0)

	specResult := se.execute(true, true, true)
	c.Assert(specResult.ProtoSpec.GetIsTableDriven(), Equals, true)
	c.Assert(specResult.Skipped, Equals, true)
}

func (s *MySuite) TestConvertParseErrorToGaugeMessagesError(c *C) {
	e := parser.ParseError{Message: "Message", LineNo: 5, FileName: "filename"}
	se := newSpecExecutor(nil, nil, nil, nil, 0)

	errs := se.convertErrors([]error{e})

	expected := gauge_messages.Error{
		Type:       gauge_messages.Error_PARSE_ERROR,
		Message:    "filename:5 Message => ''",
		LineNumber: 5,
		Filename:   "filename",
	}

	c.Assert(len(errs), DeepEquals, 1)
	c.Assert(*(errs[0]), DeepEquals, expected)
}

func (s *MySuite) TestConvertSpecValidationErrorToGaugeMessagesError(c *C) {
	e := validation.NewSpecValidationError("Message", "filename")
	se := newSpecExecutor(nil, nil, nil, nil, 0)

	errs := se.convertErrors([]error{e})

	expected := gauge_messages.Error{
		Type:    gauge_messages.Error_VALIDATION_ERROR,
		Message: "filename Message",
	}

	c.Assert(len(errs), DeepEquals, 1)
	c.Assert(*(errs[0]), DeepEquals, expected)
}

func (s *MySuite) TestConvertStepValidationErrorToGaugeMessagesError(c *C) {
	e := validation.NewStepValidationError(&gauge.Step{LineText: "step", LineNo: 3}, "Step Message", "filename", nil)
	se := newSpecExecutor(nil, nil, nil, nil, 0)

	errs := se.convertErrors([]error{e})

	expected := gauge_messages.Error{
		Type:    gauge_messages.Error_VALIDATION_ERROR,
		Message: "filename:3 Step Message => 'step'",
	}

	c.Assert(len(errs), DeepEquals, 1)
	c.Assert(*(errs[0]), DeepEquals, expected)
}
