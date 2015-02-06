// Copyright 2014 ThoughtWorks, Inc.

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
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestFormatSpecification(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 2},
		&token{kind: stepKind, value: "Example step", lineNo: 3, lineText: "Example step"},
		&token{kind: stepKind, value: "Step with inline table", lineNo: 3, lineText: "Step with inline table "},
		&token{kind: tableHeader, args: []string{"id", "name"}},
		&token{kind: tableRow, args: []string{"1", "foo"}},
		&token{kind: tableRow, args: []string{"2", "bar"}},
	}

	spec, _ := new(specParser).createSpecification(tokens, new(conceptDictionary))

	formatted := formatSpecification(spec)

	c.Assert(formatted, Equals,
		`Spec Heading
============
Scenario Heading
----------------
* Example step
* Step with inline table 
     |id|name|
     |--|----|
     |1 |foo |
     |2 |bar |
`)
}

func (s *MySuite) TestFormatConcepts(c *C) {
	dictionary := new(conceptDictionary)
	step1 := &step{value: "sdsf", lineText: "sdsf", isConcept: true, lineNo: 1, preComments: []*comment{&comment{value: "COMMENT", lineNo: 1}}}
	step2 := &step{value: "dsfdsfdsf", lineText: "dsfdsfdsf", isConcept: true, lineNo: 2, items: []item{&step{value: "sfd", lineText: "sfd", isConcept: false}, &step{value: "sdfsdf" + "T", lineText: "sdfsdf" + "T", isConcept: false}}}
	dictionary.add([]*step{step1, step2}, "file.cpt")

	formatted := formatConcepts(dictionary)
	c.Assert(formatted["file.cpt"], Equals, `COMMENT
# sdsf
# dsfdsfdsf
* sdfsdfT
`)
}

func (s *MySuite) TestFormatSpecificationWithTags(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "My Spec Heading", lineNo: 1},
		&token{kind: tagKind, args: []string{"tag1", "tag2"}, lineNo: 2},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 3},
		&token{kind: tagKind, args: []string{"tag3", "tag4"}, lineNo: 4},
		&token{kind: stepKind, value: "Example step", lineNo: 5, lineText: "Example step"},
		&token{kind: scenarioKind, value: "Scenario Heading1", lineNo: 6},
		&token{kind: tagKind, args: []string{"tag3", "tag4"}, lineNo: 7},
		&token{kind: stepKind, value: "Example step", lineNo: 8, lineText: "Example step"},
	}

	spec, _ := new(specParser).createSpecification(tokens, new(conceptDictionary))
	formatted := formatSpecification(spec)
	c.Assert(formatted, Equals,
		`My Spec Heading
===============
tags: tag1, tag2
Scenario Heading
----------------
tags: tag3, tag4
* Example step
Scenario Heading1
-----------------
tags: tag3, tag4
* Example step
`)

}

func (s *MySuite) TestFormatStep(c *C) {
	step := &step{value: "my step with {}, {}, {} and {}", args: []*stepArg{&stepArg{value: "static \"foo\"", argType: static},
		&stepArg{value: "dynamic \"foo\"", argType: dynamic},
		&stepArg{name: "file:user\".txt", argType: specialString},
		&stepArg{name: "table :hell\".csv", argType: specialTable}}}
	formatted := formatStep(step)
	c.Assert(formatted, Equals, `* my step with "static \"foo\"", <dynamic \"foo\">, <file:user\".txt> and <table :hell\".csv>
`)
}

func (s *MySuite) TestFormattingWithTableAsAComment(c *C) {
	tokens := []*token{
		&token{kind: specKind, value: "My Spec Heading", lineNo: 1},
		&token{kind: scenarioKind, value: "Scenario Heading", lineNo: 3},
		&token{kind: tableHeader, args: []string{"id", "name"}, lineText: " |id|name|"},
		&token{kind: tableRow, args: []string{"1", "foo"}, lineText: " |1|foo|"},
		&token{kind: tableRow, args: []string{"2", "bar"}, lineText: "|2|bar|"},
		&token{kind: stepKind, value: "Example step", lineNo: 5, lineText: "Example step"},
	}

	spec, _ := new(specParser).createSpecification(tokens, new(conceptDictionary))
	formatted := formatSpecification(spec)
	c.Assert(formatted, Equals,
		`My Spec Heading
===============
Scenario Heading
----------------
 |id|name|
 |1|foo|
|2|bar|
* Example step
`)
}
