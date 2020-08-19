/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package parser

import (
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/gauge"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestParsingSpecHeading(c *C) {
	parser := new(SpecParser)

	specText := newSpecBuilder().specHeading("Spec Heading").String()
	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)
	c.Assert(tokens[0].Kind, Equals, gauge.SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec Heading")
}

func (s *MySuite) TestParsingASingleStep(c *C) {
	parser := new(SpecParser)
	tokens, err := parser.GenerateTokens("* test step \"arg\" ", "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)
	c.Assert(tokens[0].Kind, Equals, gauge.StepKind)
}

func (s *MySuite) TestParsingMultipleSpecHeading(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec Heading").specHeading("Another Spec Heading").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)
	c.Assert(tokens[0].Kind, Equals, gauge.SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec Heading")
	c.Assert(tokens[1].Kind, Equals, gauge.SpecKind)
	c.Assert(tokens[1].Value, Equals, "Another Spec Heading")
}

func (s *MySuite) TestParsingThrowErrorForEmptySpecHeading(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("").text("dsfdsf").String()

	_, res, err := parser.Parse(specText, gauge.NewConceptDictionary(), "foo.spec")
	c.Assert(err, IsNil)
	c.Assert(len(res.ParseErrors) > 0, Equals, true)
	c.Assert(res.ParseErrors[0].Error(), Equals, "foo.spec:1 Spec heading should have at least one character => ''")
}

func (s *MySuite) TestParsingScenarioHeading(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec Heading").scenarioHeading("First scenario").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)
	c.Assert(tokens[1].Kind, Equals, gauge.ScenarioKind)
	c.Assert(tokens[1].Value, Equals, "First scenario")
}

func (s *MySuite) TestParsingThrowErrorForEmptyScenarioHeading(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec Heading").scenarioHeading("").String()

	_, errs := parser.GenerateTokens(specText, "foo.spec")

	c.Assert(len(errs) > 0, Equals, true)
	c.Assert(errs[0].Error(), Equals, "foo.spec:2 Scenario heading should have at least one character => ''")
}

func (s *MySuite) TestParsingScenarioWithoutSpecHeading(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().scenarioHeading("Scenario Heading").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)
	c.Assert(tokens[0].Kind, Equals, gauge.ScenarioKind)
}

func (s *MySuite) TestParsingComments(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec Heading").text("Hello i am a comment ").text("### A h3 comment").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[1].Kind, Equals, gauge.CommentKind)
	c.Assert(tokens[1].Value, Equals, "Hello i am a comment")

	c.Assert(tokens[2].Kind, Equals, gauge.CommentKind)
	c.Assert(tokens[2].Value, Equals, "### A h3 comment")
}

func (s *MySuite) TestParsingSpecHeadingWithUnderlineOneChar(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().text("Spec heading with underline ").text("=").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)

	c.Assert(tokens[0].Kind, Equals, gauge.SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec heading with underline")

}

func (s *MySuite) TestParsingSpecHeadingWithUnderlineMultipleChar(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().text("Spec heading with underline ").text("=====").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)

	c.Assert(tokens[0].Kind, Equals, gauge.SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec heading with underline")

}

func (s *MySuite) TestParsingCommentWithUnderlineAndInvalidCharacters(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().text("A comment that will be with invalid underline").text("===89s").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)

	c.Assert(tokens[0].Kind, Equals, gauge.CommentKind)
	c.Assert(tokens[0].Value, Equals, "A comment that will be with invalid underline")

	c.Assert(tokens[1].Kind, Equals, gauge.CommentKind)
	c.Assert(tokens[1].Value, Equals, "===89s")
}

func (s *MySuite) TestParsingScenarioHeadingWithUnderline(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().text("Spec heading with underline ").text("=").text("Scenario heading with underline").text("-").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)

	c.Assert(tokens[0].Kind, Equals, gauge.SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec heading with underline")

	c.Assert(tokens[1].Kind, Equals, gauge.ScenarioKind)
	c.Assert(tokens[1].Value, Equals, "Scenario heading with underline")

}

func (s *MySuite) TestParsingScenarioHeadingWithUnderlineMultipleChar(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().text("Spec heading with underline ").text("=").text("Scenario heading with underline").text("----").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)

	c.Assert(tokens[0].Kind, Equals, gauge.SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec heading with underline")

	c.Assert(tokens[1].Kind, Equals, gauge.ScenarioKind)
	c.Assert(tokens[1].Value, Equals, "Scenario heading with underline")

}

func (s *MySuite) TestParsingHeadingWithUnderlineAndHash(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading with hash ").text("=====").scenarioHeading("Scenario heading with hash").text("----").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 4)

	c.Assert(tokens[0].Kind, Equals, gauge.SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec heading with hash")

	c.Assert(tokens[1].Kind, Equals, gauge.CommentKind)
	c.Assert(tokens[1].Value, Equals, "=====")

	c.Assert(tokens[2].Kind, Equals, gauge.ScenarioKind)
	c.Assert(tokens[2].Value, Equals, "Scenario heading with hash")

	c.Assert(tokens[3].Kind, Equals, gauge.CommentKind)
	c.Assert(tokens[3].Value, Equals, "----")

}

func (s *MySuite) TestParseSpecTags(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading with hash ").tags("tag1", "tag2").scenarioHeading("Scenario Heading").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[1].Kind, Equals, gauge.TagKind)
	c.Assert(len(tokens[1].Args), Equals, 2)
	c.Assert(tokens[1].Args[0], Equals, "tag1")
	c.Assert(tokens[1].Args[1], Equals, "tag2")
	c.Assert(tokens[1].LineText(), Equals, "tags: tag1,tag2")
	c.Assert(tokens[1].Value, Equals, "tag1,tag2")
}

func (s *MySuite) TestParseSpecTagsWithSpace(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading with hash ").text(" tags :tag1,tag2").scenarioHeading("Scenario Heading").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[1].Kind, Equals, gauge.TagKind)
	c.Assert(len(tokens[1].Args), Equals, 2)
	c.Assert(tokens[1].Args[0], Equals, "tag1")
	c.Assert(tokens[1].Args[1], Equals, "tag2")
	c.Assert(tokens[1].LineText(), Equals, " tags :tag1,tag2")
	c.Assert(tokens[1].Value, Equals, "tag1,tag2")
}

func (s *MySuite) TestParseEmptyTags(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading with hash ").tags("tag1", "", "tag2", "").scenarioHeading("Scenario Heading").String()
	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[1].Kind, Equals, gauge.TagKind)
	c.Assert(len(tokens[1].Args), Equals, 2)
	c.Assert(tokens[1].Args[0], Equals, "tag1")
	c.Assert(tokens[1].Args[1], Equals, "tag2")
	c.Assert(tokens[1].LineText(), Equals, "tags: tag1,,tag2,")
	c.Assert(tokens[1].Value, Equals, "tag1,,tag2,")
}

func (s *MySuite) TestParseScenarioTags(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading with hash ").scenarioHeading("Scenario Heading").tags("tag1", "tag2").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[2].Kind, Equals, gauge.TagKind)
	c.Assert(len(tokens[2].Args), Equals, 2)
	c.Assert(tokens[2].Args[0], Equals, "tag1")
	c.Assert(tokens[2].Args[1], Equals, "tag2")
	c.Assert(tokens[2].LineText(), Equals, "tags: tag1,tag2")
	c.Assert(tokens[2].Value, Equals, "tag1,tag2")
}

func (s *MySuite) TestParseScenarioWithTagsInMultipleLines(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading with hash ").scenarioHeading("Scenario Heading").tags("tag1", "\ntag2").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 4)

	c.Assert(tokens[2].Kind, Equals, gauge.TagKind)
	c.Assert(len(tokens[2].Args), Equals, 1)
	c.Assert(tokens[2].Args[0], Equals, "tag1")
	c.Assert(tokens[2].LineText(), Equals, "tags: tag1,")
	c.Assert(tokens[2].Value, Equals, "tag1,")
	c.Assert(tokens[3].Args[0], Equals, "tag2")
	c.Assert(tokens[3].LineText(), Equals, "tag2")
	c.Assert(tokens[3].Value, Equals, "tag2")
}

func (s *MySuite) TestParseSpecTagsBeforeSpecHeading(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().tags("tag1 ").specHeading("Spec heading with hash ").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)

	c.Assert(tokens[0].Kind, Equals, gauge.TagKind)
	c.Assert(len(tokens[0].Args), Equals, 1)
	c.Assert(tokens[0].Args[0], Equals, "tag1")
	c.Assert(tokens[0].LineText(), Equals, "tags: tag1 ")
	c.Assert(tokens[0].Value, Equals, "tag1")
}

func (s *MySuite) TestParsingSimpleDataTable(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading").text("|name|id|").text("|---|---|").text("|john|123|").text("|james|007|").String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 5)

	c.Assert(tokens[1].Kind, Equals, gauge.TableHeader)
	c.Assert(len(tokens[1].Args), Equals, 2)
	c.Assert(tokens[1].Args[0], Equals, "name")
	c.Assert(tokens[1].Args[1], Equals, "id")

	c.Assert(tokens[2].Kind, Equals, gauge.TableRow)
	c.Assert(len(tokens[2].Args), Equals, 2)
	c.Assert(tokens[2].Args[0], Equals, "---")
	c.Assert(tokens[2].Args[1], Equals, "---")

	c.Assert(tokens[3].Kind, Equals, gauge.TableRow)
	c.Assert(len(tokens[3].Args), Equals, 2)
	c.Assert(tokens[3].Args[0], Equals, "john")
	c.Assert(tokens[3].Args[1], Equals, "123")

	c.Assert(tokens[4].Kind, Equals, gauge.TableRow)
	c.Assert(len(tokens[4].Args), Equals, 2)
	c.Assert(tokens[4].Args[0], Equals, "james")
	c.Assert(tokens[4].Args[1], Equals, "007")

}
func (s *MySuite) TestParsingMultipleDataTable(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading").text("|name|id|").text("|john|123|").text("|james|007|").step("Example step").text("|user|role|").text("|root | admin|").String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 7)

	c.Assert(tokens[1].Kind, Equals, gauge.TableHeader)
	c.Assert(len(tokens[1].Args), Equals, 2)
	c.Assert(tokens[1].Args[0], Equals, "name")
	c.Assert(tokens[1].Args[1], Equals, "id")

	c.Assert(tokens[2].Kind, Equals, gauge.TableRow)
	c.Assert(len(tokens[2].Args), Equals, 2)
	c.Assert(tokens[2].Args[0], Equals, "john")
	c.Assert(tokens[2].Args[1], Equals, "123")

	c.Assert(tokens[3].Kind, Equals, gauge.TableRow)
	c.Assert(len(tokens[3].Args), Equals, 2)
	c.Assert(tokens[3].Args[0], Equals, "james")
	c.Assert(tokens[3].Args[1], Equals, "007")

	c.Assert(tokens[5].Kind, Equals, gauge.TableHeader)
	c.Assert(len(tokens[5].Args), Equals, 2)
	c.Assert(tokens[5].Args[0], Equals, "user")
	c.Assert(tokens[5].Args[1], Equals, "role")

	c.Assert(tokens[6].Kind, Equals, gauge.TableRow)
	c.Assert(len(tokens[6].Args), Equals, 2)
	c.Assert(tokens[6].Args[0], Equals, "root")
	c.Assert(tokens[6].Args[1], Equals, "admin")
}

func (s *MySuite) TestParsingDataTableWithEmptyHeaderSeparatorRow(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading").text("|name|id|").text("|||").text("|john|123|").String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 4)

	c.Assert(tokens[1].Kind, Equals, gauge.TableHeader)
	c.Assert(len(tokens[1].Args), Equals, 2)
	c.Assert(tokens[1].Args[0], Equals, "name")
	c.Assert(tokens[1].Args[1], Equals, "id")

	c.Assert(tokens[2].Kind, Equals, gauge.TableRow)
	c.Assert(len(tokens[2].Args), Equals, 2)
	c.Assert(tokens[2].Args[0], Equals, "")
	c.Assert(tokens[2].Args[1], Equals, "")

	c.Assert(tokens[3].Kind, Equals, gauge.TableRow)
	c.Assert(len(tokens[3].Args), Equals, 2)
	c.Assert(tokens[3].Args[0], Equals, "john")
	c.Assert(tokens[3].Args[1], Equals, "123")

}

func (s *MySuite) TestParsingDataTableRowEscapingPipe(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading").text("| name|id | address| phone|").text("| escape \\| pipe |second|third|").String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[1].Kind, Equals, gauge.TableHeader)
	c.Assert(len(tokens[1].Args), Equals, 4)
	c.Assert(tokens[1].Args[0], Equals, "name")
	c.Assert(tokens[1].Args[1], Equals, "id")
	c.Assert(tokens[1].Args[2], Equals, "address")
	c.Assert(tokens[1].Args[3], Equals, "phone")

	c.Assert(tokens[2].Kind, Equals, gauge.TableRow)
	c.Assert(len(tokens[2].Args), Equals, 3)
	c.Assert(tokens[2].Args[0], Equals, "escape | pipe")
	c.Assert(tokens[2].Args[1], Equals, "second")
	c.Assert(tokens[2].Args[2], Equals, "third")

}

func (s *MySuite) TestParsingDataTableThrowsErrorWithEmptyHeader(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading").text("| name|id |||").text("| escape \\| pipe |second|third|second|").String()

	_, errs := parser.GenerateTokens(specText, "foo.spec")
	c.Assert(len(errs) > 0, Equals, true)
	c.Assert(errs[0].Error(), Equals, "foo.spec:2 Table header should not be blank => '| name|id |||'")
}

func (s *MySuite) TestParsingDataTableThrowsErrorWithSameColumnHeader(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading").text("| name|id|name|").text("|1|2|3|").String()

	_, errs := parser.GenerateTokens(specText, "foo.spec")
	c.Assert(len(errs) > 0, Equals, true)
	c.Assert(errs[0].Error(), Equals, "foo.spec:2 Table header cannot have repeated column values => '| name|id|name|'")
}

func (s *MySuite) TestParsingDataTableWithSeparatorAsHeader(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading").text("|---|--|-|").text("|---|--|-|").text("|---|--|-|").text("| escape \\| pipe |second|third|").String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 5)

	c.Assert(tokens[1].Kind, Equals, gauge.TableHeader)
	c.Assert(len(tokens[1].Args), Equals, 3)
	c.Assert(tokens[1].Args[0], Equals, "---")
	c.Assert(tokens[1].Args[1], Equals, "--")
	c.Assert(tokens[1].Args[2], Equals, "-")

	c.Assert(tokens[2].Kind, Equals, gauge.TableRow)
	c.Assert(len(tokens[2].Args), Equals, 3)
	c.Assert(tokens[2].Args[0], Equals, "---")
	c.Assert(tokens[2].Args[1], Equals, "--")
	c.Assert(tokens[2].Args[2], Equals, "-")

}

func (s *MySuite) TestParsingSpecWithMultipleLines(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("A spec heading").
		text("Hello, i am a comment").
		text(" ").
		step("Context step with \"param\" and <file:foo>").
		text("|a|b|c|").
		text("|--||").
		text("|a1|a2|a3|").
		tags("one", "two").
		scenarioHeading("First flow").
		tags("tag1", "tag2").
		step("first with \"fpp\" and <bar>").
		text("").
		text("Comment in scenario").
		step("<table:file.csv> and <another> with \"foo\"").
		scenarioHeading("First flow").
		step("another").String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 15)

	c.Assert(tokens[0].Kind, Equals, gauge.SpecKind)
	c.Assert(tokens[1].Kind, Equals, gauge.CommentKind)
	c.Assert(tokens[2].Kind, Equals, gauge.CommentKind)

	c.Assert(tokens[3].Kind, Equals, gauge.StepKind)
	c.Assert(tokens[3].Value, Equals, "Context step with {static} and {special}")

	c.Assert(tokens[4].Kind, Equals, gauge.TableHeader)
	c.Assert(tokens[5].Kind, Equals, gauge.TableRow)
	c.Assert(tokens[6].Kind, Equals, gauge.TableRow)
	c.Assert(tokens[7].Kind, Equals, gauge.TagKind)
	c.Assert(tokens[8].Kind, Equals, gauge.ScenarioKind)
	c.Assert(tokens[9].Kind, Equals, gauge.TagKind)

	c.Assert(tokens[10].Kind, Equals, gauge.StepKind)
	c.Assert(tokens[10].Value, Equals, "first with {static} and {dynamic}")

	c.Assert(tokens[11].Kind, Equals, gauge.CommentKind)

	c.Assert(tokens[12].Kind, Equals, gauge.StepKind)
	c.Assert(tokens[12].Value, Equals, "{special} and {dynamic} with {static}")

	c.Assert(tokens[13].Kind, Equals, gauge.ScenarioKind)

	c.Assert(tokens[14].Kind, Equals, gauge.StepKind)
	c.Assert(tokens[14].Value, Equals, "another")

}

func (s *MySuite) TestParsingSimpleScenarioDataTable(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading").
		scenarioHeading("Scenario Heading").
		text("|name|id|").
		text("|---|---|").
		text("|john|123|").
		text("|james|007|").String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 6)

	c.Assert(tokens[2].Kind, Equals, gauge.TableHeader)
	c.Assert(len(tokens[2].Args), Equals, 2)
	c.Assert(tokens[2].Args[0], Equals, "name")
	c.Assert(tokens[2].Args[1], Equals, "id")

	c.Assert(tokens[3].Kind, Equals, gauge.TableRow)
	c.Assert(len(tokens[3].Args), Equals, 2)
	c.Assert(tokens[3].Args[0], Equals, "---")
	c.Assert(tokens[3].Args[1], Equals, "---")

	c.Assert(tokens[4].Kind, Equals, gauge.TableRow)
	c.Assert(len(tokens[4].Args), Equals, 2)
	c.Assert(tokens[4].Args[0], Equals, "john")
	c.Assert(tokens[4].Args[1], Equals, "123")

	c.Assert(tokens[5].Kind, Equals, gauge.TableRow)
	c.Assert(len(tokens[5].Args), Equals, 2)
	c.Assert(tokens[5].Args[0], Equals, "james")
	c.Assert(tokens[5].Args[1], Equals, "007")

}

func (s *MySuite) TestParsingExternalScenarioDataTable(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading").
		scenarioHeading("Scenario Heading").
		text("table:data/foo.csv").String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)
	c.Assert(tokens[2].Kind, Equals, gauge.DataTableKind)
}

func (s *MySuite) TestParsingStepWIthNewlineAndTableParam(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().
		step("step1").
		text("").
		tableHeader("foo|bar").
		tableRow("somerow|another").String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[0].Kind, Equals, gauge.StepKind)

	c.Assert(tokens[1].Kind, Equals, gauge.TableHeader)
	c.Assert(tokens[2].Kind, Equals, gauge.TableRow)
}

func (s *MySuite) TestParsingMultilineStep(c *C) {
	env.AllowMultiLineStep = func() bool { return true }
	parser := new(SpecParser)
	specText := newSpecBuilder().
		step("step1").
		text("second line").String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)

	c.Assert(tokens[0].Kind, Equals, gauge.StepKind)
	c.Assert(tokens[0].Value, Equals, "step1 second line")
}

func (s *MySuite) TestParsingMultilineStepWithParams(c *C) {
	env.AllowMultiLineStep = func() bool { return true }
	parser := new(SpecParser)
	specText := newSpecBuilder().
		step("step1").
		text("second line \"foo\"").
		text("third line <bar>").String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)

	c.Assert(tokens[0].Kind, Equals, gauge.StepKind)
	c.Assert(tokens[0].Value, Equals, "step1 second line {static} third line {dynamic}")
	c.Assert(len(tokens[0].Args), Equals, 2)
}

func (s *MySuite) TestParsingMultilineStepWithTableParam(c *C) {
	env.AllowMultiLineStep = func() bool { return true }
	parser := new(SpecParser)
	specText := newSpecBuilder().
		step("step1").
		text("second line").
		text("").
		tableHeader("foo|bar").
		tableRow("somerow|another").String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[0].Kind, Equals, gauge.StepKind)
	c.Assert(tokens[1].Kind, Equals, gauge.TableHeader)
	c.Assert(tokens[2].Kind, Equals, gauge.TableRow)
}

func (s *MySuite) TestParsingMultilineStepScenarioNext(c *C) {
	env.AllowMultiLineStep = func() bool { return true }
	parser := new(SpecParser)
	specText := newSpecBuilder().
		step("step1").
		text("Scenario1").
		text("---------").String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)

	c.Assert(tokens[0].Kind, Equals, gauge.StepKind)
	c.Assert(tokens[0].Value, Equals, "step1 Scenario1")
	c.Assert(tokens[1].Kind, Equals, gauge.CommentKind)
}

func (s *MySuite) TestParsingMultilineStepWithSpecNext(c *C) {
	env.AllowMultiLineStep = func() bool { return true }
	parser := new(SpecParser)
	specText := newSpecBuilder().
		step("step1").
		text("Concept1").
		text("========").String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)

	c.Assert(tokens[0].Kind, Equals, gauge.StepKind)
	c.Assert(tokens[0].Value, Equals, "step1 Concept1")
	c.Assert(tokens[1].Kind, Equals, gauge.CommentKind)

}

func (s *MySuite) TestParsingSpecWithTearDownSteps(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("A spec heading").
		text("Hello, i am a comment").
		scenarioHeading("First flow").
		step("another").
		text("_____").
		step("step1").
		step("step2").String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 7)

	c.Assert(tokens[0].Kind, Equals, gauge.SpecKind)
	c.Assert(tokens[1].Kind, Equals, gauge.CommentKind)

	c.Assert(tokens[2].Kind, Equals, gauge.ScenarioKind)
	c.Assert(tokens[3].Kind, Equals, gauge.StepKind)
	c.Assert(tokens[3].Value, Equals, "another")
	c.Assert(tokens[4].Kind, Equals, gauge.TearDownKind)

	c.Assert(tokens[5].Kind, Equals, gauge.StepKind)
	c.Assert(tokens[5].Value, Equals, "step1")
	c.Assert(tokens[6].Kind, Equals, gauge.StepKind)
	c.Assert(tokens[6].Value, Equals, "step2")
}
