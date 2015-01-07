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
`)

}
