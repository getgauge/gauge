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
	. "gopkg.in/check.v1"
	"github.com/getgauge/gauge/util"
)

func (s *MySuite) TestParsingFileSpecialType(c *C) {
	resolver := newSpecialTypeResolver()
	resolver.predefinedResolvers["file"] = func(value string) (*StepArg, error) {
		return &StepArg{Value: "dummy", ArgType: Static}, nil
	}

	stepArg, _ := resolver.resolve("file:foo")
	c.Assert(stepArg.Value, Equals, "dummy")
	c.Assert(stepArg.ArgType, Equals, Static)
	c.Assert(stepArg.Name, Equals, "file:foo")
}

func (s *MySuite) TestParsingFileAsSpecialParamWithWindowsPathAsValue(c *C) {
	resolver := newSpecialTypeResolver()
	resolver.predefinedResolvers["file"] = func(value string) (*StepArg, error) {
		return &StepArg{Value: "hello", ArgType: SpecialString}, nil
	}

	stepArg, _ := resolver.resolve(`file:C:\Users\abc`)
	c.Assert(stepArg.Value, Equals, "hello")
	c.Assert(stepArg.ArgType, Equals, SpecialString)
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
	c.Assert(len(table.columns), Equals, 0)
}

func (s *MySuite) TestParsingUnknownSpecialType(c *C) {
	resolver := newSpecialTypeResolver()

	_, err := resolver.getStepArg("unknown", "foo", "unknown:foo")
	c.Assert(err.Error(), Equals, "Resolver not found for special param <unknown:foo>")
}

func (s *MySuite) TestPopulatingConceptLookup(c *C) {
	parser := new(SpecParser)
	conceptDictionary := new(ConceptDictionary)

	specText := SpecBuilder().specHeading("A spec heading").
		tableHeader("id", "name", "phone").
		tableHeader("123", "foo", "888").
		scenarioHeading("First scenario").
		step("create user <id> <name> and <phone>").
		String()

	conceptText := SpecBuilder().
		specHeading("create user <user-id> <user-name> and <user-phone>").
		step("assign id <user-id> and name <user-name>").
		step("assign number <user-phone>").String()

	concepts, _ := new(ConceptParser).Parse(conceptText)

	conceptDictionary.Add(concepts, "file.cpt")
	spec, _ := parser.Parse(specText, conceptDictionary)
	concept := spec.Scenarios[0].Steps[0]

	dataTableLookup := new(ArgLookup).FromDataTableRow(&spec.DataTable.Table, 0)
	PopulateConceptDynamicParams(concept, dataTableLookup)

	c.Assert(concept.getArg("user-id").Value, Equals, "123")
	c.Assert(concept.getArg("user-name").Value, Equals, "foo")
	c.Assert(concept.getArg("user-phone").Value, Equals, "888")

}

func (s *MySuite) TestPopulatingNestedConceptLookup(c *C) {
	parser := new(SpecParser)
	conceptDictionary := new(ConceptDictionary)

	specText := SpecBuilder().specHeading("A spec heading").
		tableHeader("id", "name", "phone").
		tableHeader("123", "prateek", "8800").
		scenarioHeading("First scenario").
		step("create user <id> <name> and <phone>").
		step("create user \"456\" \"foo\" and \"9900\"").
		String()

	conceptText := SpecBuilder().
		specHeading("create user <user-id> <user-name> and <user-phone>").
		step("assign id <user-id> and name <user-name>").
		specHeading("assign id <userid> and name <username>").
		step("add id <userid>").
		step("add name <username>").String()

	concepts, _ := new(ConceptParser).Parse(conceptText)

	conceptDictionary.Add(concepts, "file.cpt")
	spec, _ := parser.Parse(specText, conceptDictionary)
	concept1 := spec.Scenarios[0].Steps[0]

	dataTableLookup := new(ArgLookup).FromDataTableRow(&spec.DataTable.Table, 0)
	PopulateConceptDynamicParams(concept1, dataTableLookup)

	c.Assert(concept1.getArg("user-id").Value, Equals, "123")
	c.Assert(concept1.getArg("user-name").Value, Equals, "prateek")
	c.Assert(concept1.getArg("user-phone").Value, Equals, "8800")

	nestedConcept := concept1.ConceptSteps[0]
	c.Assert(nestedConcept.getArg("userid").Value, Equals, "123")
	c.Assert(nestedConcept.getArg("username").Value, Equals, "prateek")

	concept2 := spec.Scenarios[0].Steps[1]
	c.Assert(concept2.getArg("user-id").Value, Equals, "456")
	c.Assert(concept2.getArg("user-name").Value, Equals, "foo")
	c.Assert(concept2.getArg("user-phone").Value, Equals, "9900")

	nestedConcept2 := concept2.ConceptSteps[0]
	c.Assert(nestedConcept2.getArg("userid").Value, Equals, "456")
	c.Assert(nestedConcept2.getArg("username").Value, Equals, "foo")

}

func (s *MySuite) TestPopulatingNestedConceptsWithStaticParametersLookup(c *C) {
	parser := new(SpecParser)
	conceptDictionary := new(ConceptDictionary)

	specText := SpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("create user \"456\" \"foo\" and \"prateek\"").
		String()

	conceptText := SpecBuilder().
		specHeading("assign id <userid> and name <username>").
		step("add id \"some-id\"").
		step("second nested \"s-value\"").
		specHeading("create user <user-id> <user-name> and <user-phone>").
		step("assign id <user-id> and name \"static-name\"").
		specHeading("second nested <baz>").
		step("add id <baz>").String()

	concepts, _ := new(ConceptParser).Parse(conceptText)

	conceptDictionary.Add(concepts, "file.cpt")
	spec, _ := parser.Parse(specText, conceptDictionary)
	concept1 := spec.Scenarios[0].Steps[0]

	dataTableLookup := new(ArgLookup).FromDataTableRow(&spec.DataTable.Table, 0)
	PopulateConceptDynamicParams(concept1, dataTableLookup)

	c.Assert(concept1.getArg("user-id").Value, Equals, "456")
	c.Assert(concept1.getArg("user-name").Value, Equals, "foo")
	c.Assert(concept1.getArg("user-phone").Value, Equals, "prateek")

	nestedConcept := concept1.ConceptSteps[0]
	c.Assert(nestedConcept.getArg("userid").Value, Equals, "456")
	c.Assert(nestedConcept.getArg("username").Value, Equals, "static-name")

	c.Assert(nestedConcept.ConceptSteps[0].Args[0].ArgType, Equals, Static)
	c.Assert(nestedConcept.ConceptSteps[0].Args[0].Value, Equals, "some-id")

	secondLevelNestedConcept := nestedConcept.ConceptSteps[1]
	c.Assert(secondLevelNestedConcept.getArg("baz").Value, Equals, "s-value")
	c.Assert(secondLevelNestedConcept.getArg("baz").ArgType, Equals, Static)
}

func (s *MySuite) TestEachConceptUsageIsUpdatedWithRespectiveParams(c *C) {
	parser := new(SpecParser)
	conceptDictionary := new(ConceptDictionary)

	specText := SpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("Concept").
		String()

	conceptText := SpecBuilder().
		specHeading("Concept").
		step("Heading \"a\"").
		step("Heading \"b\"").
		specHeading("Heading <h>").
		step("Say <h> to <h>").String()

	concepts, _ := new(ConceptParser).Parse(conceptText)

	conceptDictionary.Add(concepts, "file.cpt")
	spec, _ := parser.Parse(specText, conceptDictionary)
	concept1 := spec.Scenarios[0].Steps[0]

	nestedConcept := concept1.ConceptSteps[0]
	nestedConcept1 := concept1.ConceptSteps[1]

	c.Assert(nestedConcept.getArg("h").Value, Equals, "a")
	c.Assert(nestedConcept1.getArg("h").Value, Equals, "b")
}

func (s *MySuite) TestEachConceptUsageIsUpdatedWithRespectiveParamsIncludingDynamicParams(c *C) {
	parser := new(SpecParser)
	conceptDictionary := new(ConceptDictionary)

	specText := SpecBuilder().specHeading("A spec heading").
		scenarioHeading("First scenario").
		step("Concept \"abc\"").
		String()

	conceptText := SpecBuilder().
		specHeading("Concept <message>").
		step("Heading \"a\" and \"a\"").
		step("Heading \"b\" and \"b\"").
		step("Heading <message> and \"a\"").
		step("Heading \"c\" and \"a\"").
		specHeading("Heading <h> and <i>").
		step("Say <h> to <i>").String()

	concepts, _ := new(ConceptParser).Parse(conceptText)

	conceptDictionary.Add(concepts, "file.cpt")
	spec, _ := parser.Parse(specText, conceptDictionary)
	concept1 := spec.Scenarios[0].Steps[0]

	nestedConcept := concept1.ConceptSteps[0]
	c.Assert(nestedConcept.getArg("h").Value, Equals, "a")
	c.Assert(nestedConcept.getArg("i").Value, Equals, "a")

	nestedConcept1 := concept1.ConceptSteps[1]
	c.Assert(nestedConcept1.getArg("h").Value, Equals, "b")
	c.Assert(nestedConcept1.getArg("i").Value, Equals, "b")

	nestedConcept2 := concept1.ConceptSteps[2]
	c.Assert(nestedConcept2.getArg("h").Value, Equals, "abc")
	c.Assert(nestedConcept2.getArg("i").Value, Equals, "a")

	nestedConcept3 := concept1.ConceptSteps[3]
	c.Assert(nestedConcept3.getArg("h").Value, Equals, "c")
	c.Assert(nestedConcept3.getArg("i").Value, Equals, "a")
}
