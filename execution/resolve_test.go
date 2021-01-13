/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package execution

import (
	"path/filepath"
	"testing"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/gauge"
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
	_, _, err := parser.AddConcepts([]string{path}, conceptDictionary)
	if err != nil {
		c.Error(err)
	}
	spec, _, err := new(parser.SpecParser).Parse(specText, conceptDictionary, "")
	if err != nil {
		c.Error(err)
	}

	specExecutor := newSpecExecutor(spec, nil, nil, nil, 0)
	specExecutor.errMap = getValidationErrorMap()
	lookup, err := specExecutor.dataTableLookup()
	c.Assert(err, IsNil)
	cItem, err := resolveToProtoConceptItem(*spec.Scenarios[0].Steps[0], lookup, specExecutor.setSkipInfo)
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
	_, _, err := parser.AddConcepts([]string{path}, conceptDictionary)
	if err != nil {
		c.Error(err)
	}
	specParser := new(parser.SpecParser)
	spec, _, _ := specParser.Parse(specText, conceptDictionary, "")

	specExecutor := newSpecExecutor(spec, nil, nil, nil, 0)
	specExecutor.errMap = getValidationErrorMap()
	lookup, err := specExecutor.dataTableLookup()
	c.Assert(err, IsNil)
	cItem, err := resolveToProtoConceptItem(*spec.Scenarios[0].Steps[0], lookup, specExecutor.setSkipInfo)
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
	_, _, err := parser.AddConcepts([]string{path}, conceptDictionary)
	if err != nil {
		t.Error(err)
	}
	specParser := new(parser.SpecParser)
	spec, _, _ := specParser.Parse(specText, conceptDictionary, "")

	specExecutor := newSpecExecutor(spec, nil, nil, nil, 0)
	specExecutor.errMap = getValidationErrorMap()
	lookup, err := specExecutor.dataTableLookup()
	if err != nil {
		t.Errorf("Expected no error. Got : %s", err.Error())
	}
	cItem, err := resolveToProtoConceptItem(*spec.Scenarios[0].Steps[0], lookup, specExecutor.setSkipInfo)
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
	_, _, err := parser.AddConcepts([]string{path}, conceptDictionary)
	if err != nil {
		c.Error(err)
	}
	specParser := new(parser.SpecParser)
	spec, _, _ := specParser.Parse(specText, conceptDictionary, "")

	specExecutor := newSpecExecutor(spec, nil, nil, nil, 0)

	specExecutor.errMap = gauge.NewBuildErrors()
	lookup, err := specExecutor.dataTableLookup()
	c.Assert(err, IsNil)
	cItem, err := resolveToProtoConceptItem(*spec.Scenarios[0].Steps[0], lookup, specExecutor.setSkipInfo)
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
