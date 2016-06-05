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

	spec, _ := new(parser.SpecParser).Parse(specText, conceptDictionary)

	specExecutor := newSpecExecutor(spec, nil, nil, indexRange{start: 0, end: 0}, nil, 0)
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
	spec, _ := specParser.Parse(specText, conceptDictionary)

	specExecutor := newSpecExecutor(spec, nil, nil, indexRange{start: 0, end: 0}, nil, 0)
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
	spec, _ := specParser.Parse(specText, conceptDictionary)

	specExecutor := newSpecExecutor(spec, nil, nil, indexRange{start: 0, end: 0}, nil, 0)

	// For first row
	specExecutor.currentTableRow = 0
	specExecutor.errMap = getValidationErrorMap()
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

func (s *MySuite) TestToGetDataTableRowsRangeFromInputFlag(c *C) {
	rowsRange, err := getDataTableRowsRange("5-6", 7)
	c.Assert(err, Equals, nil)
	c.Assert(rowsRange.start, Equals, 4)
	c.Assert(rowsRange.end, Equals, 5)
}

func (s *MySuite) TestToGetDataTableRow(c *C) {
	rowsRange, err := getDataTableRowsRange("5", 7)
	c.Assert(err, Equals, nil)
	c.Assert(rowsRange.start, Equals, 4)
	c.Assert(rowsRange.end, Equals, 4)
}

func (s *MySuite) TestToGetDataTableRowFromInvalidInput(c *C) {
	_, err := getDataTableRowsRange("a", 7)
	c.Assert(err.Error(), Equals, "Table rows range validation failed.")
	_, err = getDataTableRowsRange("a-5", 7)
	c.Assert(err.Error(), Equals, "Table rows range validation failed.")
	_, err = getDataTableRowsRange("a-qwerty", 7)
	c.Assert(err.Error(), Equals, "Table rows range validation failed.")
	_, err = getDataTableRowsRange("aas-helloo", 7)
	c.Assert(err.Error(), Equals, "Table rows range validation failed.")
	_, err = getDataTableRowsRange("apoorva", 7)
	c.Assert(err.Error(), Equals, "Table rows range validation failed.")
	_, err = getDataTableRowsRange("8-9", 7)
	c.Assert(err.Error(), Equals, "Table rows range validation failed.")
	_, err = getDataTableRowsRange("12-9", 7)
	c.Assert(err.Error(), Equals, "Table rows range validation failed.")
	_, err = getDataTableRowsRange("4:5", 6)
	c.Assert(err.Error(), Equals, "Table rows range validation failed.")
	_, err = getDataTableRowsRange("4-5-8", 6)
	c.Assert(err.Error(), Equals, "Table rows range validation failed.")
	_, err = getDataTableRowsRange("4", 3)
	c.Assert(err.Error(), Equals, "Table rows range validation failed.")
	_, err = getDataTableRowsRange("0", 3)
	c.Assert(err.Error(), Equals, "Table rows range validation failed.")
	_, err = getDataTableRowsRange("", 3)
	c.Assert(err.Error(), Equals, "Table rows range validation failed.")
}

func (s *MySuite) TestCreateSkippedSpecResult(c *C) {
	spec := &gauge.Specification{Heading: &gauge.Heading{LineNo: 0, Value: "SPEC_HEADING"}, FileName: "FILE"}

	se := newSpecExecutor(spec, nil, nil, indexRange{start: 0, end: 0}, nil, 0)
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

	se := newSpecExecutor(anySpec(), nil, nil, indexRange{start: 0, end: 0}, nil, 0)
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

func anySpec() *gauge.Specification{
	specText := SpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("create user \"456\" \"foo\" and \"9900\"").
		String()

	spec, _ := new(parser.SpecParser).Parse(specText, gauge.NewConceptDictionary())
	spec.FileName = "FILE"
	return spec
}

func (s *MySuite) TestSpecIsSkippedIfDataRangeIsInvalid(c *C) {
	errMap := &validation.ValidationErrMaps{
		SpecErrs:     make(map[*gauge.Specification][]*validation.StepValidationError),
		ScenarioErrs: make(map[*gauge.Scenario][]*validation.StepValidationError),
		StepErrs:     make(map[*gauge.Step]*validation.StepValidationError),
	}
	se := newSpecExecutor(anySpec(), nil, nil, indexRange{start: -1, end: -1},errMap , 0)

	result:= se.execute()

	c.Assert(result.Skipped, Equals, true)
}
