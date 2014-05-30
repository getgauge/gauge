package main

import (
	. "launchpad.net/gocheck"
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestParsingSpecHeading(c *C) {
	parser := new(specParser)

	specText := SpecBuilder().specHeading("Spec Heading").String()
	tokens, err := parser.generateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)
	c.Assert(tokens[0].kind, Equals, specKind)
	c.Assert(tokens[0].value, Equals, "Spec Heading")
}

func (s *MySuite) TestParsingMultipleSpecHeading(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().specHeading("Spec Heading").specHeading("Another Spec Heading").String()

	tokens, err := parser.generateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)
	c.Assert(tokens[0].kind, Equals, specKind)
	c.Assert(tokens[0].value, Equals, "Spec Heading")
	c.Assert(tokens[1].kind, Equals, specKind)
	c.Assert(tokens[1].value, Equals, "Another Spec Heading")
}

func (s *MySuite) TestParsingThrowErrorForEmptySpecHeading(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().specHeading("").specHeading("Another Spec Heading").String()

	_, err := parser.generateTokens(specText)

	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "line no: 1, Spec heading should have at least one character")
}

func (s *MySuite) TestParsingScenarioHeading(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().specHeading("Spec Heading").scenarioHeading("First scenario").String()

	tokens, err := parser.generateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)
	c.Assert(tokens[1].kind, Equals, scenarioKind)
	c.Assert(tokens[1].value, Equals, "First scenario")
}

func (s *MySuite) TestParsingThrowErrorForEmptyScenarioHeading(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().specHeading("Spec Heading").scenarioHeading("").String()

	_, err := parser.generateTokens(specText)

	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "line no: 2, Scenario heading should have at least one character")
}

func (s *MySuite) TestParsingScenarioWithoutSpecHeading(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().scenarioHeading("Scenario Heading").String()

	tokens, err := parser.generateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)
	c.Assert(tokens[0].kind, Equals, scenarioKind)
}

func (s *MySuite) TestParsingComments(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().specHeading("Spec Heading").text("Hello i am a comment ").text("### A h3 comment").String()

	tokens, err := parser.generateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[1].kind, Equals, commentKind)
	c.Assert(tokens[1].value, Equals, "Hello i am a comment")

	c.Assert(tokens[2].kind, Equals, commentKind)
	c.Assert(tokens[2].value, Equals, "### A h3 comment")
}

func (s *MySuite) TestParsingSpecHeadingWithUnderlineOneChar(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().text("Spec heading with underline ").text("=").String()

	tokens, err := parser.generateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)

	c.Assert(tokens[0].kind, Equals, specKind)
	c.Assert(tokens[0].value, Equals, "Spec heading with underline")

}

func (s *MySuite) TestParsingSpecHeadingWithUnderlineMultipleChar(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().text("Spec heading with underline ").text("=====").String()

	tokens, err := parser.generateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)

	c.Assert(tokens[0].kind, Equals, specKind)
	c.Assert(tokens[0].value, Equals, "Spec heading with underline")

}

func (s *MySuite) TestParsingCommentWithUnderlineAndInvalidCharacters(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().text("A comment that will be with invalid underline").text("===89s").String()

	tokens, err := parser.generateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)

	c.Assert(tokens[0].kind, Equals, commentKind)
	c.Assert(tokens[0].value, Equals, "A comment that will be with invalid underline")

	c.Assert(tokens[1].kind, Equals, commentKind)
	c.Assert(tokens[1].value, Equals, "===89s")
}

func (s *MySuite) TestParsingScenarioHeadingWithUnderline(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().text("Spec heading with underline ").text("=").text("Scenario heading with underline").text("-").String()

	tokens, err := parser.generateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)

	c.Assert(tokens[0].kind, Equals, specKind)
	c.Assert(tokens[0].value, Equals, "Spec heading with underline")

	c.Assert(tokens[1].kind, Equals, scenarioKind)
	c.Assert(tokens[1].value, Equals, "Scenario heading with underline")

}

func (s *MySuite) TestParsingScenarioHeadingWithUnderlineMultipleChar(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().text("Spec heading with underline ").text("=").text("Scenario heading with underline").text("----").String()

	tokens, err := parser.generateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)

	c.Assert(tokens[0].kind, Equals, specKind)
	c.Assert(tokens[0].value, Equals, "Spec heading with underline")

	c.Assert(tokens[1].kind, Equals, scenarioKind)
	c.Assert(tokens[1].value, Equals, "Scenario heading with underline")

}

func (s *MySuite) TestParsingHeadingWithUnderlineAndHash(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().specHeading("Spec heading with hash ").text("=====").scenarioHeading("Scenario heading with hash").text("----").String()

	tokens, err := parser.generateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 4)

	c.Assert(tokens[0].kind, Equals, specKind)
	c.Assert(tokens[0].value, Equals, "Spec heading with hash")

	c.Assert(tokens[1].kind, Equals, commentKind)
	c.Assert(tokens[1].value, Equals, "=====")

	c.Assert(tokens[2].kind, Equals, scenarioKind)
	c.Assert(tokens[2].value, Equals, "Scenario heading with hash")

	c.Assert(tokens[3].kind, Equals, commentKind)
	c.Assert(tokens[3].value, Equals, "----")

}

func (s *MySuite) TestParseSpecTags(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().specHeading("Spec heading with hash ").tags("tag1", "tag2").scenarioHeading("Scenario Heading").String()

	tokens, err := parser.generateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[1].kind, Equals, tagKind)
	c.Assert(len(tokens[1].args), Equals, 2)
	c.Assert(tokens[1].args[0], Equals, "tag1")
	c.Assert(tokens[1].args[1], Equals, "tag2")
	c.Assert(tokens[1].lineText, Equals, "tags: tag1,tag2")
	c.Assert(tokens[1].value, Equals, "tag1,tag2")
}

func (s *MySuite) TestParseSpecTagsWithSpace(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().specHeading("Spec heading with hash ").text(" tags :tag1,tag2").scenarioHeading("Scenario Heading").String()

	tokens, err := parser.generateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[1].kind, Equals, tagKind)
	c.Assert(len(tokens[1].args), Equals, 2)
	c.Assert(tokens[1].args[0], Equals, "tag1")
	c.Assert(tokens[1].args[1], Equals, "tag2")
	c.Assert(tokens[1].lineText, Equals, " tags :tag1,tag2")
	c.Assert(tokens[1].value, Equals, "tag1,tag2")
}

func (s *MySuite) TestParseEmptyTags(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().specHeading("Spec heading with hash ").tags("tag1", "", "tag2", "").scenarioHeading("Scenario Heading").String()
	tokens, err := parser.generateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[1].kind, Equals, tagKind)
	c.Assert(len(tokens[1].args), Equals, 2)
	c.Assert(tokens[1].args[0], Equals, "tag1")
	c.Assert(tokens[1].args[1], Equals, "tag2")
	c.Assert(tokens[1].lineText, Equals, "tags: tag1,,tag2,")
	c.Assert(tokens[1].value, Equals, "tag1,,tag2,")
}

func (s *MySuite) TestParseScenarioTags(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().specHeading("Spec heading with hash ").scenarioHeading("Scenario Heading").tags("tag1", "tag2").String()

	tokens, err := parser.generateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[2].kind, Equals, tagKind)
	c.Assert(len(tokens[2].args), Equals, 2)
	c.Assert(tokens[2].args[0], Equals, "tag1")
	c.Assert(tokens[2].args[1], Equals, "tag2")
	c.Assert(tokens[2].lineText, Equals, "tags: tag1,tag2")
	c.Assert(tokens[2].value, Equals, "tag1,tag2")
}

func (s *MySuite) TestParseSpecTagsBeforeSpecHeading(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().tags("tag1 ").specHeading("Spec heading with hash ").String()

	tokens, err := parser.generateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)

	c.Assert(tokens[0].kind, Equals, tagKind)
	c.Assert(len(tokens[0].args), Equals, 1)
	c.Assert(tokens[0].args[0], Equals, "tag1")
	c.Assert(tokens[0].lineText, Equals, "tags: tag1 ")
	c.Assert(tokens[0].value, Equals, "tag1")
}

func (s *MySuite) TestParsingSimpleDataTable(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().specHeading("Spec heading").text("|name|id|").text("|---|---|").text("|john|123|").text("|james|007|").String()

	tokens, err := parser.generateTokens(specText)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 4)

	c.Assert(tokens[1].kind, Equals, tableHeader)
	c.Assert(len(tokens[1].args), Equals, 2)
	c.Assert(tokens[1].args[0], Equals, "name")
	c.Assert(tokens[1].args[1], Equals, "id")

	c.Assert(tokens[2].kind, Equals, tableRow)
	c.Assert(len(tokens[2].args), Equals, 2)
	c.Assert(tokens[2].args[0], Equals, "john")
	c.Assert(tokens[2].args[1], Equals, "123")

	c.Assert(tokens[3].kind, Equals, tableRow)
	c.Assert(len(tokens[3].args), Equals, 2)
	c.Assert(tokens[3].args[0], Equals, "james")
	c.Assert(tokens[3].args[1], Equals, "007")

}

func (s *MySuite) TestParsingMultipleDataTable(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().specHeading("Spec heading").text("|name|id|").text("|---|---|").text("|john|123|").text("|james|007|").step("Example step").text("|user|role|").text("|root | admin|").String()

	tokens, err := parser.generateTokens(specText)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 7)

	c.Assert(tokens[1].kind, Equals, tableHeader)
	c.Assert(len(tokens[1].args), Equals, 2)
	c.Assert(tokens[1].args[0], Equals, "name")
	c.Assert(tokens[1].args[1], Equals, "id")

	c.Assert(tokens[2].kind, Equals, tableRow)
	c.Assert(len(tokens[2].args), Equals, 2)
	c.Assert(tokens[2].args[0], Equals, "john")
	c.Assert(tokens[2].args[1], Equals, "123")

	c.Assert(tokens[3].kind, Equals, tableRow)
	c.Assert(len(tokens[3].args), Equals, 2)
	c.Assert(tokens[3].args[0], Equals, "james")
	c.Assert(tokens[3].args[1], Equals, "007")

	c.Assert(tokens[5].kind, Equals, tableHeader)
	c.Assert(len(tokens[5].args), Equals, 2)
	c.Assert(tokens[5].args[0], Equals, "user")
	c.Assert(tokens[5].args[1], Equals, "role")

	c.Assert(tokens[6].kind, Equals, tableRow)
	c.Assert(len(tokens[6].args), Equals, 2)
	c.Assert(tokens[6].args[0], Equals, "root")
	c.Assert(tokens[6].args[1], Equals, "admin")
}

func (s *MySuite) TestParsingDataTableWithEmptyHeaderSeparatorRow(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().specHeading("Spec heading").text("|name|id|").text("|||").text("|john|123|").String()

	tokens, err := parser.generateTokens(specText)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[1].kind, Equals, tableHeader)
	c.Assert(len(tokens[1].args), Equals, 2)
	c.Assert(tokens[1].args[0], Equals, "name")
	c.Assert(tokens[1].args[1], Equals, "id")

	c.Assert(tokens[2].kind, Equals, tableRow)
	c.Assert(len(tokens[2].args), Equals, 2)
	c.Assert(tokens[2].args[0], Equals, "john")
	c.Assert(tokens[2].args[1], Equals, "123")

}

func (s *MySuite) TestParsingDataTableRowEscapingPipe(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().specHeading("Spec heading").text("| name|id | address| phone|").text("| escape \\| pipe |second|third|").String()

	tokens, err := parser.generateTokens(specText)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[1].kind, Equals, tableHeader)
	c.Assert(len(tokens[1].args), Equals, 4)
	c.Assert(tokens[1].args[0], Equals, "name")
	c.Assert(tokens[1].args[1], Equals, "id")
	c.Assert(tokens[1].args[2], Equals, "address")
	c.Assert(tokens[1].args[3], Equals, "phone")

	c.Assert(tokens[2].kind, Equals, tableRow)
	c.Assert(len(tokens[2].args), Equals, 3)
	c.Assert(tokens[2].args[0], Equals, "escape | pipe")
	c.Assert(tokens[2].args[1], Equals, "second")
	c.Assert(tokens[2].args[2], Equals, "third")

}

func (s *MySuite) TestParsingDataTableThrowsErrorWithEmptyHeader(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().specHeading("Spec heading").text("| name|id |||").text("| escape \\| pipe |second|third|second|").String()

	_, err := parser.generateTokens(specText)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "line no: 2, Table header should not be blank")
}

func (s *MySuite) TestParsingDataTableThrowsErrorWithSameColumnHeader(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().specHeading("Spec heading").text("| name|id|name|").text("|1|2|3|").String()

	_, err := parser.generateTokens(specText)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "line no: 2, Table header cannot have repeated column values")
}

func (s *MySuite) TestParsingDataTableWithSeparatorAsHeader(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().specHeading("Spec heading").text("|---|--|-|").text("|---|--|-|").text("|---|--|-|").text("| escape \\| pipe |second|third|").String()

	tokens, err := parser.generateTokens(specText)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 4)

	c.Assert(tokens[1].kind, Equals, tableHeader)
	c.Assert(len(tokens[1].args), Equals, 3)
	c.Assert(tokens[1].args[0], Equals, "---")
	c.Assert(tokens[1].args[1], Equals, "--")
	c.Assert(tokens[1].args[2], Equals, "-")

	c.Assert(tokens[2].kind, Equals, tableRow)
	c.Assert(len(tokens[2].args), Equals, 3)
	c.Assert(tokens[2].args[0], Equals, "---")
	c.Assert(tokens[2].args[1], Equals, "--")
	c.Assert(tokens[2].args[2], Equals, "-")

}

func (s *MySuite) TestParsingSpecWithMultipleLines(c *C) {
	parser := new(specParser)
	specText := SpecBuilder().specHeading("A spec heading").
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
		text("Comment in scenario").
		step("<table:file.csv> and <another> with \"foo\"").
		scenarioHeading("First flow").
		step("another").String()

	tokens, err := parser.generateTokens(specText)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 14)

	c.Assert(tokens[0].kind, Equals, specKind)
	c.Assert(tokens[1].kind, Equals, commentKind)
	c.Assert(tokens[2].kind, Equals, commentKind)

	c.Assert(tokens[3].kind, Equals, stepKind)
	c.Assert(tokens[3].value, Equals, "Context step with {static} and {special}")

	c.Assert(tokens[4].kind, Equals, tableHeader)
	c.Assert(tokens[5].kind, Equals, tableRow)
	c.Assert(tokens[6].kind, Equals, tagKind)
	c.Assert(tokens[7].kind, Equals, scenarioKind)
	c.Assert(tokens[8].kind, Equals, tagKind)

	c.Assert(tokens[9].kind, Equals, stepKind)
	c.Assert(tokens[9].value, Equals, "first with {static} and {dynamic}")

	c.Assert(tokens[10].kind, Equals, commentKind)

	c.Assert(tokens[11].kind, Equals, stepKind)
	c.Assert(tokens[11].value, Equals, "{special} and {dynamic} with {static}")

	c.Assert(tokens[12].kind, Equals, scenarioKind)

	c.Assert(tokens[13].kind, Equals, stepKind)
	c.Assert(tokens[13].value, Equals, "another")

}
