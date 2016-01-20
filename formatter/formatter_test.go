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

package formatter

import (
	"testing"

	"github.com/getgauge/gauge/parser"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestFormatSpecification(c *C) {
	tokens := []*parser.Token{
		&parser.Token{Kind: parser.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&parser.Token{Kind: parser.StepKind, Value: "Example step", LineNo: 3, LineText: "Example step"},
		&parser.Token{Kind: parser.StepKind, Value: "Step with inline table", LineNo: 3, LineText: "Step with inline table "},
		&parser.Token{Kind: parser.TableHeader, Args: []string{"id", "name"}},
		&parser.Token{Kind: parser.TableRow, Args: []string{"<1>", "foo"}},
		&parser.Token{Kind: parser.TableRow, Args: []string{"2", "bar"}},
	}

	spec, _ := new(parser.SpecParser).CreateSpecification(tokens, new(parser.ConceptDictionary))

	formatted := FormatSpecification(spec)

	c.Assert(formatted, Equals,
		`Spec Heading
============
Scenario Heading
----------------
* Example step
* Step with inline table`+" "+`
     |id |name|
     |---|----|
     |<1>|foo |
     |2  |bar |
`)
}

func (s *MySuite) TestFormatConcepts(c *C) {
	dictionary := parser.NewConceptDictionary()
	step1 := &parser.Step{Value: "sdsf", LineText: "sdsf", IsConcept: true, LineNo: 1, PreComments: []*parser.Comment{&parser.Comment{Value: "COMMENT", LineNo: 1}}}
	step2 := &parser.Step{Value: "dsfdsfdsf", LineText: "dsfdsfdsf", IsConcept: true, LineNo: 2, Items: []parser.Item{&parser.Step{Value: "sfd", LineText: "sfd", IsConcept: false}, &parser.Step{Value: "sdfsdf" + "T", LineText: "sdfsdf" + "T", IsConcept: false}}}

	dictionary.ConceptsMap[step1.Value] = &parser.Concept{ConceptStep: step1, FileName: "file.cpt"}
	dictionary.ConceptsMap[step2.Value] = &parser.Concept{ConceptStep: step2, FileName: "file.cpt"}

	formatted := FormatConcepts(dictionary)
	c.Assert(formatted["file.cpt"], Equals, `COMMENT
# sdsf
# dsfdsfdsf
* sdfsdfT
`)
}

func (s *MySuite) TestFormatSpecificationWithTags(c *C) {
	tokens := []*parser.Token{
		&parser.Token{Kind: parser.SpecKind, Value: "My Spec Heading", LineNo: 1},
		&parser.Token{Kind: parser.TagKind, Args: []string{"tag1", "tag2"}, LineNo: 2},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading", LineNo: 3},
		&parser.Token{Kind: parser.TagKind, Args: []string{"tag3", "tag4"}, LineNo: 4},
		&parser.Token{Kind: parser.StepKind, Value: "Example step", LineNo: 5, LineText: "Example step"},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading1", LineNo: 6},
		&parser.Token{Kind: parser.TagKind, Args: []string{"tag3", "tag4"}, LineNo: 7},
		&parser.Token{Kind: parser.StepKind, Value: "Example step", LineNo: 8, LineText: "Example step"},
	}

	spec, _ := new(parser.SpecParser).CreateSpecification(tokens, new(parser.ConceptDictionary))
	formatted := FormatSpecification(spec)
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

func (s *MySuite) TestFormatSpecificationWithTeardownSteps(c *C) {
	tokens := []*parser.Token{
		&parser.Token{Kind: parser.SpecKind, Value: "My Spec Heading", LineNo: 1},
		&parser.Token{Kind: parser.TagKind, Args: []string{"tag1", "tag2"}, LineNo: 2},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading", LineNo: 3},
		&parser.Token{Kind: parser.TagKind, Args: []string{"tag3", "tag4"}, LineNo: 4},
		&parser.Token{Kind: parser.StepKind, Value: "Example step", LineNo: 5, LineText: "Example step"},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading1", LineNo: 6},
		&parser.Token{Kind: parser.TagKind, Args: []string{"tag3", "tag4"}, LineNo: 7},
		&parser.Token{Kind: parser.StepKind, Value: "Example step", LineNo: 8, LineText: "Example step"},
		&parser.Token{Kind: parser.TearDownKind, Value: "____", LineNo: 9},
		&parser.Token{Kind: parser.StepKind, Value: "Example step1", LineNo: 10, LineText: "Example step1"},
		&parser.Token{Kind: parser.StepKind, Value: "Example step2", LineNo: 11, LineText: "Example step2"},
	}

	spec, _ := new(parser.SpecParser).CreateSpecification(tokens, new(parser.ConceptDictionary))
	formatted := FormatSpecification(spec)
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
____
* Example step1
* Example step2
`)

}

func (s *MySuite) TestFormatStep(c *C) {
	step := &parser.Step{Value: "my step with {}, {}, {} and {}", Args: []*parser.StepArg{&parser.StepArg{Value: "static \"foo\"", ArgType: parser.Static},
		&parser.StepArg{Value: "dynamic \"foo\"", ArgType: parser.Dynamic},
		&parser.StepArg{Name: "file:user\".txt", ArgType: parser.SpecialString},
		&parser.StepArg{Name: "table :hell\".csv", ArgType: parser.SpecialTable}}}
	formatted := FormatStep(step)
	c.Assert(formatted, Equals, `* my step with "static \"foo\"", <dynamic \"foo\">, <file:user\".txt> and <table :hell\".csv>
`)
}

func (s *MySuite) TestFormattingWithTableAsAComment(c *C) {
	tokens := []*parser.Token{
		&parser.Token{Kind: parser.SpecKind, Value: "My Spec Heading", LineNo: 1},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading", LineNo: 3},
		&parser.Token{Kind: parser.TableHeader, Args: []string{"id", "name"}, LineText: " |id|name|"},
		&parser.Token{Kind: parser.TableRow, Args: []string{"1", "foo"}, LineText: " |1|foo|"},
		&parser.Token{Kind: parser.TableRow, Args: []string{"2", "bar"}, LineText: "|2|bar|"},
		&parser.Token{Kind: parser.StepKind, Value: "Example step", LineNo: 5, LineText: "Example step"},
	}

	spec, _ := new(parser.SpecParser).CreateSpecification(tokens, new(parser.ConceptDictionary))
	formatted := FormatSpecification(spec)
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

func (s *MySuite) TestFormatSpecificationWithTableContainingDynamicParameters(c *C) {
	tokens := []*parser.Token{
		&parser.Token{Kind: parser.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: parser.TableHeader, Args: []string{"id", "foo"}},
		&parser.Token{Kind: parser.TableRow, Args: []string{"1", "f"}},
		&parser.Token{Kind: parser.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&parser.Token{Kind: parser.StepKind, Value: "Example step", LineNo: 3, LineText: "Example step"},
		&parser.Token{Kind: parser.StepKind, Value: "Step with inline table", LineNo: 3, LineText: "Step with inline table "},
		&parser.Token{Kind: parser.TableHeader, Args: []string{"id", "name"}},
		&parser.Token{Kind: parser.TableRow, Args: []string{"1", "<foo>"}},
		&parser.Token{Kind: parser.TableRow, Args: []string{"2", "bar"}},
	}

	spec, _ := new(parser.SpecParser).CreateSpecification(tokens, new(parser.ConceptDictionary))

	formatted := FormatSpecification(spec)

	c.Assert(formatted, Equals,
		`Spec Heading
============
     |id|foo|
     |--|---|
     |1 |f  |
Scenario Heading
----------------
* Example step
* Step with inline table `+`
     |id|name |
     |--|-----|
     |1 |<foo>|
     |2 |bar  |
`)
}
