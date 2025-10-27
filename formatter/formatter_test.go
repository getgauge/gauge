/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package formatter

import (
	"testing"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/config"
	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

// Setup method to make sure each test starts with a known configuration
func (s *MySuite) SetUpTest(c *C) {
	config.SetSkipEmptyLineInsertions(false)
}

func (s *MySuite) TestFormatSpecification(c *C) {
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&parser.Token{Kind: gauge.TagKind, Args: []string{"test_tag1", "test_tag2"}, LineNo: 3},
		&parser.Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 4, Lines: []string{"Example step"}},
		&parser.Token{Kind: gauge.StepKind, Value: "Step with inline table", LineNo: 5, Lines: []string{"Step with inline table "}},
		&parser.Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"<1>", "foo"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}},
	}

	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")

	formatted := FormatSpecification(spec)

	c.Assert(formatted, Equals,
		`# Spec Heading
## Scenario Heading

tags: test_tag1, test_tag2

* Example step
* Step with inline table

   |id |name|
   |---|----|
   |<1>|foo |
   |2  |bar |
`)
}

func (s *MySuite) TestFormatSpecificationSkipEmptyLineInsertions(c *C) {
	config.SetSkipEmptyLineInsertions(true)
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&parser.Token{Kind: gauge.TagKind, Args: []string{"test_tag1", "test_tag2"}, LineNo: 3},
		&parser.Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 4, Lines: []string{"Example step"}},
		&parser.Token{Kind: gauge.StepKind, Value: "Step with inline table", LineNo: 5, Lines: []string{"Step with inline table "}},
		&parser.Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"<1>", "foo"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}},
	}

	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")

	formatted := FormatSpecification(spec)

	c.Assert(formatted, Equals,
		`# Spec Heading
## Scenario Heading
tags: test_tag1, test_tag2
* Example step
* Step with inline table
   |id |name|
   |---|----|
   |<1>|foo |
   |2  |bar |
`)
}



func (s *MySuite) TestFormatTable(c *C) {
	cell1 := gauge.TableCell{Value: "john", CellType: gauge.Static}
	cell2 := gauge.TableCell{Value: "doe", CellType: gauge.Static}

	headers := []string{"name1", "name2"}
	cols := [][]gauge.TableCell{{cell1}, {cell2}}

	table := gauge.NewTable(headers, cols, 10)

	got := FormatTable(table)
	want := `
   |name1|name2|
   |-----|-----|
   |john |doe  |
`

	c.Assert(got, Equals, want)
}

func (s *MySuite) TestFormatTableWithUmlautChars(c *C) {
	// umlaut characters are unicode and can take up twice the space of regular chars
	cell1 := gauge.TableCell{Value: "Büsingen", CellType: gauge.Static}
	cell2 := gauge.TableCell{Value: "Hauptstraße", CellType: gauge.Static}

	headers := []string{"col1", "col2"}
	cols := [][]gauge.TableCell{{cell1}, {cell2}}

	table := gauge.NewTable(headers, cols, 10)

	got := FormatTable(table)
	want := `
   |col1    |col2       |
   |--------|-----------|
   |Büsingen|Hauptstraße|
`

	c.Assert(got, Equals, want)
}

func (s *MySuite) TestFormatConcepts(c *C) {
	dictionary := gauge.NewConceptDictionary()
	step1 := &gauge.Step{Value: "sdsf", LineText: "sdsf", IsConcept: true, LineNo: 1, PreComments: []*gauge.Comment{&gauge.Comment{Value: "COMMENT", LineNo: 1}}}
	step2 := &gauge.Step{Value: "dsfdsfdsf", LineText: "dsfdsfdsf", IsConcept: true, LineNo: 2, Items: []gauge.Item{&gauge.Step{Value: "sfd", LineText: "sfd", IsConcept: false}, &gauge.Step{Value: "sdfsdf" + "T", LineText: "sdfsdf" + "T", IsConcept: false}}}

	dictionary.ConceptsMap[step1.Value] = &gauge.Concept{ConceptStep: step1, FileName: "file.cpt"}
	dictionary.ConceptsMap[step2.Value] = &gauge.Concept{ConceptStep: step2, FileName: "file.cpt"}

	formatted := FormatConcepts(dictionary)
	c.Assert(formatted["file.cpt"], Equals, `COMMENT
# sdsf
# dsfdsfdsf
* sdfsdfT
`)
}

func (s *MySuite) TestFormatSpecificationWithTags(c *C) {
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "My Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.TagKind, Args: []string{"tag1", "tag2"}, LineNo: 2},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 3},
		&parser.Token{Kind: gauge.TagKind, Args: []string{"tag3", "tag4"}, LineNo: 4},
		&parser.Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 5, Lines: []string{"Example step"}},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading1", LineNo: 6},
		&parser.Token{Kind: gauge.TagKind, Args: []string{"tag3", "tag4"}, LineNo: 7},
		&parser.Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 8, Lines: []string{"Example step"}},
	}

	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	formatted := FormatSpecification(spec)
	c.Assert(formatted, Equals,
		`# My Spec Heading

tags: tag1, tag2

## Scenario Heading

tags: tag3, tag4

* Example step
## Scenario Heading1

tags: tag3, tag4

* Example step
`)

}

func (s *MySuite) TestFormatSpecificationWithTagsInMutipleLines(c *C) {
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "My Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.TagKind, Args: []string{"tag1", "tag2"}, LineNo: 2},
		&parser.Token{Kind: gauge.TagKind, Args: []string{"tag3", "tag4"}, LineNo: 3},
		&parser.Token{Kind: gauge.TagKind, Args: []string{"tag5"}, LineNo: 4},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 5},
		&parser.Token{Kind: gauge.TagKind, Args: []string{"tag3", "tag4"}, LineNo: 6},
		&parser.Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 7, Lines: []string{"Example step"}},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading1", LineNo: 8},
		&parser.Token{Kind: gauge.TagKind, Args: []string{"tag3", "tag4"}, LineNo: 9},
		&parser.Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 10, Lines: []string{"Example step"}},
	}

	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	formatted := FormatSpecification(spec)
	c.Assert(formatted, Equals,
		`# My Spec Heading

tags: tag1, tag2,`+string(" \n      ")+`tag3, tag4,`+string(" \n      ")+`tag5

## Scenario Heading

tags: tag3, tag4

* Example step
## Scenario Heading1

tags: tag3, tag4

* Example step
`)
}

func (s *MySuite) TestFormatSpecificationWithTeardownSteps(c *C) {
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "My Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.TagKind, Args: []string{"tag1", "tag2"}, LineNo: 2},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 3},
		&parser.Token{Kind: gauge.TagKind, Args: []string{"tag3", "tag4"}, LineNo: 4},
		&parser.Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 5, Lines: []string{"Example step"}},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading1", LineNo: 6},
		&parser.Token{Kind: gauge.TagKind, Args: []string{"tag3", "tag4"}, LineNo: 7},
		&parser.Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 8, Lines: []string{"Example step"}},
		&parser.Token{Kind: gauge.TearDownKind, Value: "____", LineNo: 9},
		&parser.Token{Kind: gauge.StepKind, Value: "Example step1", LineNo: 10, Lines: []string{"Example step1"}},
		&parser.Token{Kind: gauge.StepKind, Value: "Example step2", LineNo: 11, Lines: []string{"Example step2"}},
	}

	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	formatted := FormatSpecification(spec)
	c.Assert(formatted, Equals,
		`# My Spec Heading

tags: tag1, tag2

## Scenario Heading

tags: tag3, tag4

* Example step
## Scenario Heading1

tags: tag3, tag4

* Example step
____
* Example step1
* Example step2
`)

}

func (s *MySuite) TestFormatStep(c *C) {
	step := &gauge.Step{Value: "my step with {}, {}, {} and {}", Args: []*gauge.StepArg{&gauge.StepArg{Value: "static \"foo\"", ArgType: gauge.Static},
		&gauge.StepArg{Name: "dynamic", Value: "\"foo\"", ArgType: gauge.Dynamic},
		&gauge.StepArg{Name: "file:user\".txt", ArgType: gauge.SpecialString},
		&gauge.StepArg{Name: "table :hell\".csv", ArgType: gauge.SpecialTable}}}
	formatted := FormatStep(step)
	c.Assert(formatted, Equals, `* my step with "static \"foo\"", <dynamic>, <file:user\".txt> and <table :hell\".csv>
`)
}

func (s *MySuite) TestFormatStepsWithResolveArgs(c *C) {
	step := &gauge.Step{Value: "my step with {}, {}", Args: []*gauge.StepArg{&gauge.StepArg{Value: "static \"foo\"", ArgType: gauge.Static},
		&gauge.StepArg{Name: "dynamic", Value: "\"foo\"", ArgType: gauge.Dynamic}},
		Fragments: []*gauge_messages.Fragment{
			&gauge_messages.Fragment{Text: "my step with "},
			&gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Parameter, Parameter: &gauge_messages.Parameter{Value: "static \"foo\"", ParameterType: gauge_messages.Parameter_Static}},
			&gauge_messages.Fragment{Text: ", "},
			&gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Parameter, Parameter: &gauge_messages.Parameter{Value: "\"foo\"", ParameterType: gauge_messages.Parameter_Dynamic}}}}
	formatted := FormatStepWithResolvedArgs(step)
	c.Assert(formatted, Equals, `* my step with "static "foo"", ""foo""
`)
}

func (s *MySuite) TestFormatStepsWithResolveArgsWithConcept(c *C) {
	step := &gauge.Step{Value: "my step with {}, {}", Args: []*gauge.StepArg{&gauge.StepArg{Value: "static \"foo\"", ArgType: gauge.Dynamic},
		&gauge.StepArg{Name: "dynamic", Value: "\"foo\"", ArgType: gauge.Static}},
		Fragments: []*gauge_messages.Fragment{
			&gauge_messages.Fragment{Text: "my step with "},
			&gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Parameter, Parameter: &gauge_messages.Parameter{Value: "static \"foo\"", ParameterType: gauge_messages.Parameter_Dynamic}},
			&gauge_messages.Fragment{Text: ", "},
			&gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Parameter, Parameter: &gauge_messages.Parameter{Value: "\"foo\"", ParameterType: gauge_messages.Parameter_Static}}}}
	formatted := FormatStepWithResolvedArgs(step)
	c.Assert(formatted, Equals, `* my step with "static "foo"", ""foo""
`)
}

func (s *MySuite) TestFormatStepsWithResolveArgsWithSpecialArguments(c *C) {
	step := &gauge.Step{Value: "my step with {}, {}", Args: []*gauge.StepArg{&gauge.StepArg{Value: "static \"foo\"", ArgType: gauge.SpecialString},
		&gauge.StepArg{Name: "dynamic", Value: "\"foo\"", ArgType: gauge.SpecialTable}},
		Fragments: []*gauge_messages.Fragment{
			&gauge_messages.Fragment{Text: "my step with "},
			&gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Parameter, Parameter: &gauge_messages.Parameter{Value: "static \"foo\"", ParameterType: gauge_messages.Parameter_Special_String}},
			&gauge_messages.Fragment{Text: ", "},
			&gauge_messages.Fragment{FragmentType: gauge_messages.Fragment_Parameter, Parameter: &gauge_messages.Parameter{Value: "\"foo\"", ParameterType: gauge_messages.Parameter_Special_Table}}}}
	formatted := FormatStepWithResolvedArgs(step)
	c.Assert(formatted, Equals, `* my step with "static "foo"", ""foo""
`)
}

func (s *MySuite) TestFormattingWithTableAsAComment(c *C) {
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "My Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.CommentKind, Value: "\n", LineNo: 2},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 3},
		&parser.Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}, Lines: []string{" |id|name|"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"1", "foo"}, Lines: []string{" |1|foo|"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}, Lines: []string{"|2|bar|"}},
		&parser.Token{Kind: gauge.CommentKind, Value: "\n", LineNo: 7},
		&parser.Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}, Lines: []string{" |id|name1|"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"1", "foo"}, Lines: []string{" |1|foo|"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}, Lines: []string{"|2|bar|"}},
		&parser.Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 11, Lines: []string{"Example step"}},
	}

	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	formatted := FormatSpecification(spec)
	c.Assert(formatted, Equals,
		`# My Spec Heading

## Scenario Heading
   |id|name|
   |--|----|
   |1 |foo |
   |2 |bar |

 |id|name1|
 |1|foo|
|2|bar|
* Example step
`)
}

func (s *MySuite) TestFormatSpecificationWithTableContainingDynamicParameters(c *C) {
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.TableHeader, Args: []string{"id", "foo"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"1", "f"}},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&parser.Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 3, Lines: []string{"Example step"}},
		&parser.Token{Kind: gauge.StepKind, Value: "Step with inline table", LineNo: 3, Lines: []string{"Step with inline table "}},
		&parser.Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"1", "<foo>"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}},
	}

	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")

	formatted := FormatSpecification(spec)

	c.Assert(formatted, Equals,
		`# Spec Heading
   |id|foo|
   |--|---|
   |1 |f  |
## Scenario Heading
* Example step
* Step with inline table

   |id|name |
   |--|-----|
   |1 |<foo>|
   |2 |bar  |
`)
}

func (s *MySuite) TestFormatShouldRetainNewlines(c *C) {
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "My Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.CommentKind, Value: "\n", LineNo: 2},
		&parser.Token{Kind: gauge.CommentKind, Value: "\n", LineNo: 3},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 4},
		&parser.Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}, Lines: []string{" |id|name|"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"1", "foo"}, Lines: []string{" |1|foo|"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}, Lines: []string{"|2|bar|"}},
		&parser.Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 8, Lines: []string{"Example step"}},
		&parser.Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"1", "<foo>"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}},
	}

	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	formatted := FormatSpecification(spec)
	c.Assert(formatted, Equals,
		`# My Spec Heading


## Scenario Heading
   |id|name|
   |--|----|
   |1 |foo |
   |2 |bar |
* Example step

   |id|name |
   |--|-----|
   |1 |<foo>|
   |2 |bar  |
`)
}

func (s *MySuite) TestFormatShouldRetainNewlinesBetweenSteps(c *C) {
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "My Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 4},
		&parser.Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 6, Lines: []string{"Example step"}, Suffix: "\n\n"},
		&parser.Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 9, Lines: []string{"Example step"}, Suffix: "\n\n"},
	}

	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	formatted := FormatSpecification(spec)
	c.Assert(formatted, Equals,
		`# My Spec Heading
## Scenario Heading
* Example step


* Example step


`)
}

func (s *MySuite) TestFormatShouldStripDuplicateNewlinesBeforeInlineTable(c *C) {
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "My Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.CommentKind, Value: "\n", LineNo: 2},
		&parser.Token{Kind: gauge.CommentKind, Value: "\n", LineNo: 3},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 4},
		&parser.Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}, Lines: []string{" |id|name|"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"1", "foo"}, Lines: []string{" |1|foo|"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}, Lines: []string{"|2|bar|"}},
		&parser.Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 8, Lines: []string{"Example step\n\n"}},
		&parser.Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"1", "<foo>"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}},
	}

	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	formatted := FormatSpecification(spec)
	c.Assert(formatted, Equals,
		`# My Spec Heading


## Scenario Heading
   |id|name|
   |--|----|
   |1 |foo |
   |2 |bar |
* Example step

   |id|name |
   |--|-----|
   |1 |<foo>|
   |2 |bar  |
`)
}

func (s *MySuite) TestFormatShouldStripDuplicateNewlinesBeforeInlineTableInTeardown(c *C) {
	tokens := []*parser.Token{
		&parser.Token{Kind: gauge.SpecKind, Value: "My Spec Heading", LineNo: 1},
		&parser.Token{Kind: gauge.CommentKind, Value: "\n", LineNo: 2},
		&parser.Token{Kind: gauge.CommentKind, Value: "\n", LineNo: 3},
		&parser.Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 4},
		&parser.Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}, Lines: []string{" |id|name|"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"1", "foo"}, Lines: []string{" |1|foo|"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}, Lines: []string{"|2|bar|"}},
		&parser.Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 8, Lines: []string{"Example step\n\n"}},
		&parser.Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"1", "<foo>"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}},
		&parser.Token{Kind: gauge.CommentKind, Value: "\n", LineNo: 10},
		&parser.Token{Kind: gauge.TearDownKind, Value: "____", LineNo: 9},
		&parser.Token{Kind: gauge.CommentKind, Value: "\n", LineNo: 10},
		&parser.Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 8, Lines: []string{"Example step\n\n\n"}},
		&parser.Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"1", "<foo>"}},
		&parser.Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}},
	}

	spec, _, _ := new(parser.SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary(), "")
	formatted := FormatSpecification(spec)
	c.Assert(formatted, Equals,
		`# My Spec Heading


## Scenario Heading
   |id|name|
   |--|----|
   |1 |foo |
   |2 |bar |
* Example step

   |id|name |
   |--|-----|
   |1 |<foo>|
   |2 |bar  |

____

* Example step

   |id|name |
   |--|-----|
   |1 |<foo>|
   |2 |bar  |
`)
}

func (s *MySuite) TestFormatShouldNotAddExtraNewLinesBeforeDataTable(c *C) {
	spec, _, _ := new(parser.SpecParser).Parse(`# Specification Heading

     |Word  |Vowel Count|
     |------|-----------|
     |Gauge |3          |
     |Mingle|2          |
     |Snap  |1          |
     |GoCD  |1          |
     |Rhythm|0          |
`, gauge.NewConceptDictionary(), "")
	formatted := FormatSpecification(spec)
	c.Assert(formatted, Equals,
		`# Specification Heading

   |Word  |Vowel Count|
   |------|-----------|
   |Gauge |3          |
   |Mingle|2          |
   |Snap  |1          |
   |GoCD  |1          |
   |Rhythm|0          |
`)
}


func TestFormatStepWithMultilineSpecialStringXML(t *testing.T) {
	xmlContent := `<user>
<name>Alice</name>
<age>30</age>
<isStudent>false</isStudent>
<hobbies>
    <hobby>reading</hobby>
    <hobby>hiking</hobby>
    <hobby>gaming</hobby>
</hobbies>
</user>`

	step := &gauge.Step{
		Value: "Step with multiline {}",
		Args: []*gauge.StepArg{
			{
				Name:    "xmlContent",
				ArgType: gauge.MultilineString,
				Value:   xmlContent,
			},
		},
	}

	got := FormatStep(step)

	want := `* Step with multiline
"""
<user>
<name>Alice</name>
<age>30</age>
<isStudent>false</isStudent>
<hobbies>
    <hobby>reading</hobby>
    <hobby>hiking</hobby>
    <hobby>gaming</hobby>
</hobbies>
</user>
"""
`

	if got != want {
		t.Errorf("unexpected formatted step.\nGot:\n%q\nWant:\n%q", got, want)
	}
}

func TestFormatStepWithMultilineSpecialStringJSON(t *testing.T) {
	jsonContent := `{
"name": "Alice",
"age": 30,
"isStudent": false,
"hobbies": ["reading", "hiking", "gaming"]
}`

	step := &gauge.Step{
		Value: "Step with multiline {}",
		Args: []*gauge.StepArg{
			{
				Name:    "jsonContent",
				ArgType: gauge.MultilineString,
				Value:   jsonContent,
			},
		},
	}

	got := FormatStep(step)

	want := `* Step with multiline
"""
{
"name": "Alice",
"age": 30,
"isStudent": false,
"hobbies": ["reading", "hiking", "gaming"]
}
"""
`

	if got != want {
		t.Errorf("unexpected formatted step.\nGot:\n%q\nWant:\n%q", got, want)
	}
}

func TestFormatStepWithMultilineSpecialStringWithEscapedNewlines(t *testing.T) {
	contentWithEscapedNewlines := "line 1\nline 2\nline 3"

	step := &gauge.Step{
		Value: "Step with escaped newlines {}",
		Args: []*gauge.StepArg{
			{
				Name:    "multilineContent",
				ArgType: gauge.MultilineString,
				Value:   contentWithEscapedNewlines,
			},
		},
	}

	got := FormatStep(step)

	want := `* Step with escaped newlines
"""
line 1
line 2
line 3
"""
`

	if got != want {
		t.Errorf("unexpected formatted step.\nGot:\n%q\nWant:\n%q", got, want)
	}
}

func TestFormatStepWithSingleLineSpecialString(t *testing.T) {
	singleLineContent := "simple string"

	step := &gauge.Step{
		Value: "Step with single line {}",
		Args: []*gauge.StepArg{
			{
				Name:    "singleLineContent",
				ArgType: gauge.Static,
				Value:   singleLineContent,
			},
		},
	}

	got := FormatStep(step)

	want := `* Step with single line "simple string"
`

	if got != want {
		t.Errorf("unexpected formatted step.\nGot:\n%q\nWant:\n%q", got, want)
	}
}
func TestFormatStepWithMultilineSpecialStringHTML(t *testing.T) {
	htmlContent := `<html>
<head>
    <title>Test Page</title>
    <style>
        body { font-family: Arial, sans-serif; }
        .container { margin: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Welcome to the Test Page</h1>
        <p>This is a paragraph with <strong>bold</strong> and <em>italic</em> text.</p>
        <ul>
            <li>Item 1</li>
            <li>Item 2</li>
            <li>Item 3</li>
        </ul>
    </div>
</body>
</html>`

	step := &gauge.Step{
		Value: "Render HTML page {}",
		Args: []*gauge.StepArg{
			{
				Name:    "htmlContent",
				ArgType: gauge.MultilineString,
				Value:   htmlContent,
			},
		},
	}

	got := FormatStep(step)

	want := `* Render HTML page
"""
<html>
<head>
    <title>Test Page</title>
    <style>
        body { font-family: Arial, sans-serif; }
        .container { margin: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Welcome to the Test Page</h1>
        <p>This is a paragraph with <strong>bold</strong> and <em>italic</em> text.</p>
        <ul>
            <li>Item 1</li>
            <li>Item 2</li>
            <li>Item 3</li>
        </ul>
    </div>
</body>
</html>
"""
`

	if got != want {
		t.Errorf("unexpected formatted step.\nGot:\n%q\nWant:\n%q", got, want)
	}
}

func TestFormatStepWithMultilineSpecialStringHTMLFragment(t *testing.T) {
	htmlFragment := `<div class="user-profile">
    <h2>User Information</h2>
    <p><strong>Name:</strong> John Doe</p>
    <p><strong>Email:</strong> john.doe@example.com</p>
    <p><strong>Status:</strong> <span class="active">Active</span></p>
</div>`

	step := &gauge.Step{
		Value: "Display user profile {}",
		Args: []*gauge.StepArg{
			{
				Name:    "profileHtml",
				ArgType: gauge.MultilineString,
				Value:   htmlFragment,
			},
		},
	}

	got := FormatStep(step)

	want := `* Display user profile
"""
<div class="user-profile">
    <h2>User Information</h2>
    <p><strong>Name:</strong> John Doe</p>
    <p><strong>Email:</strong> john.doe@example.com</p>
    <p><strong>Status:</strong> <span class="active">Active</span></p>
</div>
"""
`

	if got != want {
		t.Errorf("unexpected formatted step.\nGot:\n%q\nWant:\n%q", got, want)
	}
}

func TestFormatStepWithMultilineSpecialStringMarkdown(t *testing.T) {
	markdownContent := `# Project Documentation

## Introduction
This is a sample markdown document with various elements.

### Features
- Feature 1: Does something amazing
- Feature 2: Provides excellent performance
- Feature 3: Easy to use

### Code Block
We can show code examples here.

> Note: This is a blockquote emphasizing important information.`

	step := &gauge.Step{
		Value: "Generate documentation {}",
		Args: []*gauge.StepArg{
			{
				Name:    "markdownContent",
				ArgType: gauge.MultilineString,
				Value:   markdownContent,
			},
		},
	}

	got := FormatStep(step)

	want := `* Generate documentation
"""
# Project Documentation

## Introduction
This is a sample markdown document with various elements.

### Features
- Feature 1: Does something amazing
- Feature 2: Provides excellent performance
- Feature 3: Easy to use

### Code Block
We can show code examples here.

> Note: This is a blockquote emphasizing important information.
"""
`

	if got != want {
		t.Errorf("unexpected formatted step.\nGot:\n%q\nWant:\n%q", got, want)
	}
}