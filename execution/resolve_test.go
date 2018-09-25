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
	"path/filepath"
	"testing"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/parser"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestResolveConceptToProtoConceptItem(c *C) {
	conceptDictionary := gauge.NewConceptDictionary()

	specText := newSpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("create user \"456\" \"foo\" and \"9900\"").
		String()
	path, _ := filepath.Abs(filepath.Join("testdata", "concept.cpt"))
	parser.AddConcepts([]string{path}, conceptDictionary)

	spec, _, _ := new(parser.SpecParser).Parse(specText, conceptDictionary, "")

	specExecutor := newSpecExecutor(spec, nil, nil, nil, 0)
	specExecutor.errMap = getValidationErrorMap()
	cItem, err := resolveToProtoConceptItem(*spec.Scenarios[0].Steps[0], specExecutor.dataTableLookup, specExecutor.setSkipInfo)
	c.Assert(err, IsNil)
	protoConcept := cItem.GetConcept()

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

	specText := newSpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("create user \"456\" \"foo\" and \"9900\"").
		String()

	path, _ := filepath.Abs(filepath.Join("testdata", "concept.cpt"))
	parser.AddConcepts([]string{path}, conceptDictionary)
	specParser := new(parser.SpecParser)
	spec, _, _ := specParser.Parse(specText, conceptDictionary, "")

	specExecutor := newSpecExecutor(spec, nil, nil, nil, 0)
	specExecutor.errMap = getValidationErrorMap()
	cItem, err := resolveToProtoConceptItem(*spec.Scenarios[0].Steps[0], specExecutor.dataTableLookup, specExecutor.setSkipInfo)
	c.Assert(err, IsNil)
	protoConcept := cItem.GetConcept()
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

func TestResolveNestedConceptAndTableParamToProtoConceptItem(t *testing.T) {
	conceptDictionary := gauge.NewConceptDictionary()

	specText := newSpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("create user \"456\"").
		String()
	want := "456"
	path, _ := filepath.Abs(filepath.Join("testdata", "conceptTable.cpt"))
	parser.AddConcepts([]string{path}, conceptDictionary)
	specParser := new(parser.SpecParser)
	spec, _, _ := specParser.Parse(specText, conceptDictionary, "")

	specExecutor := newSpecExecutor(spec, nil, nil, nil, 0)
	specExecutor.errMap = getValidationErrorMap()
	cItem, err := resolveToProtoConceptItem(*spec.Scenarios[0].Steps[0], specExecutor.dataTableLookup, specExecutor.setSkipInfo)
	if err != nil {
		t.Errorf("Expected no error. Got : %s", err.Error())
	}
	protoConcept := cItem.GetConcept()
	got := getParameters(protoConcept.GetSteps()[0].GetStep().GetFragments())[0].GetTable().GetRows()[1].Cells[0]

	if want != got {
		t.Errorf("Did not resolve dynamic param in table for concept. Got %s, want: %s", got, want)
	}
}

func (s *MySuite) TestResolveToProtoConceptItemWithDataTable(c *C) {
	conceptDictionary := gauge.NewConceptDictionary()

	specText := newSpecBuilder().specHeading("A spec heading").
		tableHeader("id", "name", "phone").
		tableHeader("123", "foo", "8800").
		tableHeader("666", "bar", "9900").
		scenarioHeading("First scenario").
		step("create user <id> <name> and <phone>").
		String()

	path, _ := filepath.Abs(filepath.Join("testdata", "concept.cpt"))
	parser.AddConcepts([]string{path}, conceptDictionary)
	specParser := new(parser.SpecParser)
	spec, _, _ := specParser.Parse(specText, conceptDictionary, "")

	specExecutor := newSpecExecutor(spec, nil, nil, nil, 0)

	specExecutor.errMap = gauge.NewBuildErrors()
	cItem, err := resolveToProtoConceptItem(*spec.Scenarios[0].Steps[0], specExecutor.dataTableLookup, specExecutor.setSkipInfo)
	c.Assert(err, IsNil)
	protoConcept := cItem.GetConcept()
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
}

func checkConceptParameterValuesInOrder(c *C, concept *gauge_messages.ProtoConcept, paramValues ...string) {
	params := getParameters(concept.GetConceptStep().Fragments)
	c.Assert(len(params), Equals, len(paramValues))
	for i, param := range params {
		c.Assert(param.GetValue(), Equals, paramValues[i])
	}
}
