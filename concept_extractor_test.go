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

func (s *MySuite) TestExtractConceptWithoutParameters(c *C) {
	STEP := "* step that takes a table \"arg\""
	result := getTextForConcept(STEP)

	c.Assert(result.stepTexts, Equals, STEP+"\n")
	c.Assert(result.heading, Equals, CONCEPT_HEADING_TEMPLATE)
	c.Assert(result.conceptText, Equals, "* "+CONCEPT_HEADING_TEMPLATE)
	c.Assert(result.hasParam, Equals, false)
}

func (s *MySuite) TestExtractConcept(c *C) {
	STEP := "* step that takes a table \"arg\""
	result := getTextForConcept(STEP + "\n" + STEP)

	c.Assert(result.stepTexts, Equals, "* step that takes a table <arg>\n* step that takes a table <arg>\n")
	c.Assert(result.heading, Equals, CONCEPT_HEADING_TEMPLATE+" <arg>")
	c.Assert(result.conceptText, Equals, "* "+CONCEPT_HEADING_TEMPLATE+" \"arg\"")
	c.Assert(result.hasParam, Equals, true)
}

func (s *MySuite) TestExtractConceptWithTableAsArg(c *C) {
	STEP := `* Step that takes a table
	|Product|Description                  |
	|-------|-----------------------------|
	|Gauge  |BDD style testing with ease  |
	|Mingle |Agile project management     |
	|Snap   |Hosted continuous integration|
	|Gocd   |Continuous delivery platform |
	* Step that takes a table
	|Product|Description                  |
	|-------|-----------------------------|
	|Gauge  |BDD style testing with ease  |
	|Mingle |Agile project management     |
	|Snap   |Hosted continuous integration|
	|Gocd   |Continuous delivery platform |
	`
	result := getTextForConcept(STEP)

	c.Assert(result.stepTexts, Equals, "* Step that takes a table <table>\n* Step that takes a table <table>\n")
	c.Assert(result.heading, Equals, CONCEPT_HEADING_TEMPLATE+" <table>")
	c.Assert(result.conceptText, Equals, "* "+CONCEPT_HEADING_TEMPLATE+`
     |Product|Description                  |
     |-------|-----------------------------|
     |Gauge  |BDD style testing with ease  |
     |Mingle |Agile project management     |
     |Snap   |Hosted continuous integration|
     |Gocd   |Continuous delivery platform |
`)
	c.Assert(result.hasParam, Equals, true)
}

func (s *MySuite) TestExtractConceptWithTablesAsArg(c *C) {
	STEP := `* Step that takes a table
	|Product|Description                  |
	|-------|-----------------------------|
	|Gauge  |BDD style testing with ease  |
	|Mingle |Agile project management     |
	|Snap   |Hosted continuous integration|
	|Gocd   |Continuous delivery platform |
	* Step that takes a table
	|Product|Description                  |
	|-------|-----------------------------|
	|Gauge  |BDD style testing with ease  |
	* Step that takes a table
	|Product|Description                  |
	|-------|-----------------------------|
	|Gauge  |BDD style testing with ease  |
	|Mingle |Agile project management     |
	|Snap   |Hosted continuous integration|
	|Gocd   |Continuous delivery platform |
	`
	result := getTextForConcept(STEP)

	c.Assert(result.stepTexts, Equals, "* Step that takes a table <table>\n* Step that takes a table \n     |Product|Description                |\n     |-------|---------------------------|\n     |Gauge  |BDD style testing with ease|\n* Step that takes a table <table>\n")
	c.Assert(result.heading, Equals, CONCEPT_HEADING_TEMPLATE+" <table>")
	c.Assert(result.conceptText, Equals, "* "+CONCEPT_HEADING_TEMPLATE+`
     |Product|Description                  |
     |-------|-----------------------------|
     |Gauge  |BDD style testing with ease  |
     |Mingle |Agile project management     |
     |Snap   |Hosted continuous integration|
     |Gocd   |Continuous delivery platform |
`)
	c.Assert(result.hasParam, Equals, true)
}
