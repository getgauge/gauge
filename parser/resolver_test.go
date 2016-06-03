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
	"path/filepath"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/util"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestParsingFileSpecialType(c *C) {
	resolver := newSpecialTypeResolver()
	resolver.predefinedResolvers["file"] = func(value string) (*gauge.StepArg, error) {
		return &gauge.StepArg{Value: "dummy", ArgType: gauge.Static}, nil
	}

	stepArg, _ := resolver.resolve("file:foo")
	c.Assert(stepArg.Value, Equals, "dummy")
	c.Assert(stepArg.ArgType, Equals, gauge.Static)
	c.Assert(stepArg.Name, Equals, "file:foo")
}

func (s *MySuite) TestParsingFileAsSpecialParamWithWindowsPathAsValue(c *C) {
	resolver := newSpecialTypeResolver()
	resolver.predefinedResolvers["file"] = func(value string) (*gauge.StepArg, error) {
		return &gauge.StepArg{Value: "hello", ArgType: gauge.SpecialString}, nil
	}

	stepArg, _ := resolver.resolve(`file:C:\Users\abc`)
	c.Assert(stepArg.Value, Equals, "hello")
	c.Assert(stepArg.ArgType, Equals, gauge.SpecialString)
	if util.IsWindows() {
		c.Assert(stepArg.Name, Equals, `file:C:\\Users\\abc`)
	} else {
		c.Assert(stepArg.Name, Equals, `file:C:\Users\abc`)
	}
}

func (s *MySuite) TestParsingInvalidSpecialType(c *C) {
	resolver := newSpecialTypeResolver()

	_, err := resolver.resolve("unknown:foo")
	c.Assert(err.Error(), Equals, "Resolver not found for special param <unknown:foo>")
}

func (s *MySuite) TestConvertCsvToTable(c *C) {
	table, _ := convertCsvToTable("id,name \n1,foo\n2,bar")

	idColumn := table.Get("id")
	c.Assert(idColumn[0].Value, Equals, "1")
	c.Assert(idColumn[1].Value, Equals, "2")

	nameColumn := table.Get("name")
	c.Assert(nameColumn[0].Value, Equals, "foo")
	c.Assert(nameColumn[1].Value, Equals, "bar")
}

func (s *MySuite) TestConvertEmptyCsvToTable(c *C) {
	table, _ := convertCsvToTable("")
	c.Assert(len(table.Columns), Equals, 0)
}

func (s *MySuite) TestParsingUnknownSpecialType(c *C) {
	resolver := newSpecialTypeResolver()

	_, err := resolver.getStepArg("unknown", "foo", "unknown:foo")
	c.Assert(err.Error(), Equals, "Resolver not found for special param <unknown:foo>")
}

func (s *MySuite) TestPopulatingConceptLookup(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("A spec heading").
		tableHeader("id", "name", "phone").
		tableHeader("123", "foo", "888").
		scenarioHeading("First scenario").
		step("create user <id> <name> and <phone>").
		String()

	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "dynamic_param_concept.cpt"))
	AddConcepts(path, conceptDictionary)
	spec, _ := parser.Parse(specText, conceptDictionary, "")
	concept := spec.Scenarios[0].Steps[0]

	dataTableLookup := new(gauge.ArgLookup).FromDataTableRow(&spec.DataTable.Table, 0)
	PopulateConceptDynamicParams(concept, dataTableLookup)

	c.Assert(concept.GetArg("user-id").Value, Equals, "123")
	c.Assert(concept.GetArg("user-name").Value, Equals, "foo")
	c.Assert(concept.GetArg("user-phone").Value, Equals, "888")

}

func (s *MySuite) TestPopulatingNestedConceptLookup(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("A spec heading").
		tableHeader("id", "name", "phone").
		tableHeader("123", "prateek", "8800").
		scenarioHeading("First scenario").
		step("create user <id> <name> and <phone>").
		step("create user \"456\" \"foo\" and \"9900\"").
		String()

	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "dynamic_param_concept.cpt"))
	AddConcepts(path, conceptDictionary)
	spec, _ := parser.Parse(specText, conceptDictionary, "")
	concept1 := spec.Scenarios[0].Steps[0]

	dataTableLookup := new(gauge.ArgLookup).FromDataTableRow(&spec.DataTable.Table, 0)
	PopulateConceptDynamicParams(concept1, dataTableLookup)

	c.Assert(concept1.GetArg("user-id").Value, Equals, "123")
	c.Assert(concept1.GetArg("user-name").Value, Equals, "prateek")
	c.Assert(concept1.GetArg("user-phone").Value, Equals, "8800")

	nestedConcept := concept1.ConceptSteps[0]
	c.Assert(nestedConcept.GetArg("userid").Value, Equals, "123")
	c.Assert(nestedConcept.GetArg("username").Value, Equals, "prateek")

	concept2 := spec.Scenarios[0].Steps[1]
	c.Assert(concept2.GetArg("user-id").Value, Equals, "456")
	c.Assert(concept2.GetArg("user-name").Value, Equals, "foo")
	c.Assert(concept2.GetArg("user-phone").Value, Equals, "9900")

	nestedConcept2 := concept2.ConceptSteps[0]
	c.Assert(nestedConcept2.GetArg("userid").Value, Equals, "456")
	c.Assert(nestedConcept2.GetArg("username").Value, Equals, "foo")

}

func (s *MySuite) TestPopulatingNestedConceptsWithStaticParametersLookup(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("create user \"456\" \"foo\" and \"123456\"").
		String()

	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "static_param_concept.cpt"))
	AddConcepts(path, conceptDictionary)

	spec, _ := parser.Parse(specText, conceptDictionary, "")
	concept1 := spec.Scenarios[0].Steps[0]

	dataTableLookup := new(gauge.ArgLookup).FromDataTableRow(&spec.DataTable.Table, 0)
	PopulateConceptDynamicParams(concept1, dataTableLookup)

	c.Assert(concept1.GetArg("user-id").Value, Equals, "456")
	c.Assert(concept1.GetArg("user-name").Value, Equals, "foo")
	c.Assert(concept1.GetArg("user-phone").Value, Equals, "123456")

	nestedConcept := concept1.ConceptSteps[0]
	c.Assert(nestedConcept.GetArg("userid").Value, Equals, "456")
	c.Assert(nestedConcept.GetArg("username").Value, Equals, "static-value")
}

func (s *MySuite) TestEachConceptUsageIsUpdatedWithRespectiveParams(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("create user \"sdf\" \"name\" and \"1234\"").
		String()

	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "static_param_concept.cpt"))
	AddConcepts(path, conceptDictionary)
	spec, _ := parser.Parse(specText, conceptDictionary, "")
	concept1 := spec.Scenarios[0].Steps[0]

	nestedConcept := concept1.ConceptSteps[0]
	nestedConcept1 := concept1.ConceptSteps[1]

	c.Assert(nestedConcept.GetArg("username").Value, Equals, "static-value")
	c.Assert(nestedConcept1.GetArg("username").Value, Equals, "static-value1")
	c.Assert(nestedConcept.GetArg("userid").Value, Equals, "sdf")
	c.Assert(nestedConcept1.GetArg("userid").Value, Equals, "sdf")
}
