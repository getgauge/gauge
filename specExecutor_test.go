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
	"github.com/getgauge/gauge/gauge_messages"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestResolveConceptToProtoConceptItem(c *C) {
	parser := new(specParser)
	conceptDictionary := new(conceptDictionary)

	specText := SpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("create user \"456\" \"foo\" and \"9900\"").
		String()

	conceptText := SpecBuilder().
		specHeading("create user <user-id> <user-name> and <user-phone>").
		step("assign id <user-id> and name <user-name>").
		step("assign phone <user-phone>").String()

	concepts, _ := new(conceptParser).parse(conceptText)
	conceptDictionary.add(concepts, "file.cpt")
	spec, _ := parser.parse(specText, conceptDictionary)

	specExecutor := newSpecExecutor(spec, nil, nil, nil, indexRange{start: 0, end: 0})
	protoConcept := specExecutor.resolveToProtoConceptItem(*spec.scenarios[0].steps[0]).GetConcept()

	checkConceptParameterValuesInOrder(c, protoConcept, "456", "foo", "9900")

	firstStep := protoConcept.GetSteps()[0].GetStep()
	params := getParameters(firstStep.GetFragments())
	c.Assert(2, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, gauge_messages.Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "456")

	secondStep := protoConcept.GetSteps()[1].GetStep()
	params = getParameters(secondStep.GetFragments())
	c.Assert(1, Equals, len(params))
	c.Assert(params[0].GetParameterType(), Equals, gauge_messages.Parameter_Dynamic)
	c.Assert(params[0].GetValue(), Equals, "9900")

}

func (s *MySuite) TestResolveNestedConceptToProtoConceptItem(c *C) {
	parser := new(specParser)
	conceptDictionary := new(conceptDictionary)

	specText := SpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("create user \"456\" \"foo\" and \"9900\"").
		String()

	conceptText := SpecBuilder().
		specHeading("create user <user-id> <user-name> and <user-phone>").
		step("assign id <user-id> and name <user-name>").
		step("assign phone <user-phone>").
		specHeading("assign id <userid> and name <username>").
		step("add id <userid>").
		step("add name <username>").String()

	concepts, _ := new(conceptParser).parse(conceptText)
	conceptDictionary.add(concepts, "file.cpt")
	spec, _ := parser.parse(specText, conceptDictionary)

	specExecutor := newSpecExecutor(spec, nil, nil, nil, indexRange{start: 0, end: 0})
	protoConcept := specExecutor.resolveToProtoConceptItem(*spec.scenarios[0].steps[0]).GetConcept()
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
	parser := new(specParser)
	conceptDictionary := new(conceptDictionary)

	specText := SpecBuilder().specHeading("A spec heading").
		tableHeader("id", "name", "phone").
		tableHeader("123", "foo", "8800").
		tableHeader("666", "bar", "9900").
		scenarioHeading("First scenario").
		step("create user <id> <name> and <phone>").
		String()

	conceptText := SpecBuilder().
		specHeading("create user <user-id> <user-name> and <user-phone>").
		step("assign id <user-id> and name <user-name>").
		step("assign phone <user-phone>").
		specHeading("assign id <userid> and name <username>").
		step("add id <userid>").
		step("add name <username>").String()

	concepts, _ := new(conceptParser).parse(conceptText)
	conceptDictionary.add(concepts, "file.cpt")
	spec, _ := parser.parse(specText, conceptDictionary)

	specExecutor := newSpecExecutor(spec, nil, nil, nil, indexRange{start: 0, end: 0})

	// For first row
	specExecutor.currentTableRow = 0
	protoConcept := specExecutor.resolveToProtoConceptItem(*spec.scenarios[0].steps[0]).GetConcept()
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
	protoConcept = specExecutor.resolveToProtoConceptItem(*spec.scenarios[0].steps[0]).GetConcept()
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
