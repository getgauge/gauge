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

import . "gopkg.in/check.v1"

func (s *MySuite) TestParsingFileSpecialType(c *C) {
	resolver := newSpecialTypeResolver()
	resolver.predefinedResolvers["file"] = func(value string) (*stepArg, error) {
		return &stepArg{value: "dummy", argType: static}, nil
	}

	stepArg, _ := resolver.resolve("file:foo")
	c.Assert(stepArg.value, Equals, "dummy")
	c.Assert(stepArg.argType, Equals, static)
	c.Assert(stepArg.name, Equals, "file:foo")
}

func (s *MySuite) TestParsingInvalidSpecialType(c *C) {
	resolver := newSpecialTypeResolver()

	_, err := resolver.resolve("unknown:foo")
	c.Assert(err.Error(), Equals, "Resolver not found for special param <unknown:foo>")
}

func (s *MySuite) TestConvertCsvToTable(c *C) {
	table, _ := convertCsvToTable("id,name \n1,foo\n2,bar")

	idColumn := table.get("id")
	c.Assert(idColumn[0].value, Equals, "1")
	c.Assert(idColumn[1].value, Equals, "2")

	nameColumn := table.get("name")
	c.Assert(nameColumn[0].value, Equals, "foo")
	c.Assert(nameColumn[1].value, Equals, "bar")
}

func (s *MySuite) TestParsingUnknownSpecialType(c *C) {
	resolver := newSpecialTypeResolver()

	_, err := resolver.getStepArg("unknown", "foo", "unknown:foo")
	c.Assert(err.Error(), Equals, "Resolver not found for special param <unknown:foo>")
}

func (s *MySuite) TestPopulatingConceptLookup(c *C) {
	parser := new(specParser)
	conceptDictionary := new(conceptDictionary)

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

	concepts, _ := new(conceptParser).parse(conceptText)

	conceptDictionary.add(concepts, "file.cpt")
	spec, _ := parser.parse(specText, conceptDictionary)
	concept := spec.scenarios[0].steps[0]

	dataTableLookup := new(argLookup).fromDataTableRow(&spec.dataTable.table, 0)
	populateConceptDynamicParams(concept, dataTableLookup)

	c.Assert(concept.getArg("user-id").value, Equals, "123")
	c.Assert(concept.getArg("user-name").value, Equals, "foo")
	c.Assert(concept.getArg("user-phone").value, Equals, "888")

}

func (s *MySuite) TestPopulatingNestedConceptLookup(c *C) {
	parser := new(specParser)
	conceptDictionary := new(conceptDictionary)

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

	concepts, _ := new(conceptParser).parse(conceptText)

	conceptDictionary.add(concepts, "file.cpt")
	spec, _ := parser.parse(specText, conceptDictionary)
	concept1 := spec.scenarios[0].steps[0]

	dataTableLookup := new(argLookup).fromDataTableRow(&spec.dataTable.table, 0)
	populateConceptDynamicParams(concept1, dataTableLookup)

	c.Assert(concept1.getArg("user-id").value, Equals, "123")
	c.Assert(concept1.getArg("user-name").value, Equals, "prateek")
	c.Assert(concept1.getArg("user-phone").value, Equals, "8800")

	nestedConcept := concept1.conceptSteps[0]
	c.Assert(nestedConcept.getArg("userid").value, Equals, "123")
	c.Assert(nestedConcept.getArg("username").value, Equals, "prateek")

	concept2 := spec.scenarios[0].steps[1]
	c.Assert(concept2.getArg("user-id").value, Equals, "456")
	c.Assert(concept2.getArg("user-name").value, Equals, "foo")
	c.Assert(concept2.getArg("user-phone").value, Equals, "9900")

	nestedConcept2 := concept2.conceptSteps[0]
	c.Assert(nestedConcept2.getArg("userid").value, Equals, "456")
	c.Assert(nestedConcept2.getArg("username").value, Equals, "foo")

}
