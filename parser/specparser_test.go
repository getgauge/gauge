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
	"testing"
)

// Hook up gocheck into the "go test" runner.
func Test(t *testing.T) { TestingT(t) }

type MySuite struct{}

var _ = Suite(&MySuite{})

func (s *MySuite) TestParsingSpecHeading(c *C) {
	parser := new(SpecParser)

	specText := SpecBuilder().specHeading("Spec Heading").String()
	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)
	c.Assert(tokens[0].Kind, Equals, SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec Heading")
}

func (s *MySuite) TestParsingASingleStep(c *C) {
	parser := new(SpecParser)
	tokens, err := parser.GenerateTokens("* test step \"arg\" ")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)
	c.Assert(tokens[0].Kind, Equals, StepKind)
}

func (s *MySuite) TestParsingMultipleSpecHeading(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec Heading").specHeading("Another Spec Heading").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)
	c.Assert(tokens[0].Kind, Equals, SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec Heading")
	c.Assert(tokens[1].Kind, Equals, SpecKind)
	c.Assert(tokens[1].Value, Equals, "Another Spec Heading")
}

func (s *MySuite) TestParsingThrowErrorForEmptySpecHeading(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("").specHeading("Another Spec Heading").String()

	_, err := parser.GenerateTokens(specText)

	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "line no: 1, Spec heading should have at least one character")
}

func (s *MySuite) TestParsingScenarioHeading(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec Heading").scenarioHeading("First scenario").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)
	c.Assert(tokens[1].Kind, Equals, ScenarioKind)
	c.Assert(tokens[1].Value, Equals, "First scenario")
}

func (s *MySuite) TestParsingThrowErrorForEmptyScenarioHeading(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec Heading").scenarioHeading("").String()

	_, err := parser.GenerateTokens(specText)

	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "line no: 2, Scenario heading should have at least one character")
}

func (s *MySuite) TestParsingScenarioWithoutSpecHeading(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().scenarioHeading("Scenario Heading").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)
	c.Assert(tokens[0].Kind, Equals, ScenarioKind)
}

func (s *MySuite) TestParsingComments(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec Heading").text("Hello i am a comment ").text("### A h3 comment").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[1].Kind, Equals, CommentKind)
	c.Assert(tokens[1].Value, Equals, "Hello i am a comment")

	c.Assert(tokens[2].Kind, Equals, CommentKind)
	c.Assert(tokens[2].Value, Equals, "### A h3 comment")
}

func (s *MySuite) TestParsingSpecHeadingWithUnderlineOneChar(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().text("Spec heading with underline ").text("=").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)

	c.Assert(tokens[0].Kind, Equals, SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec heading with underline")

}

func (s *MySuite) TestParsingSpecHeadingWithUnderlineMultipleChar(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().text("Spec heading with underline ").text("=====").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)

	c.Assert(tokens[0].Kind, Equals, SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec heading with underline")

}

func (s *MySuite) TestParsingCommentWithUnderlineAndInvalidCharacters(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().text("A comment that will be with invalid underline").text("===89s").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)

	c.Assert(tokens[0].Kind, Equals, CommentKind)
	c.Assert(tokens[0].Value, Equals, "A comment that will be with invalid underline")

	c.Assert(tokens[1].Kind, Equals, CommentKind)
	c.Assert(tokens[1].Value, Equals, "===89s")
}

func (s *MySuite) TestParsingScenarioHeadingWithUnderline(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().text("Spec heading with underline ").text("=").text("Scenario heading with underline").text("-").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)

	c.Assert(tokens[0].Kind, Equals, SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec heading with underline")

	c.Assert(tokens[1].Kind, Equals, ScenarioKind)
	c.Assert(tokens[1].Value, Equals, "Scenario heading with underline")

}

func (s *MySuite) TestParsingScenarioHeadingWithUnderlineMultipleChar(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().text("Spec heading with underline ").text("=").text("Scenario heading with underline").text("----").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)

	c.Assert(tokens[0].Kind, Equals, SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec heading with underline")

	c.Assert(tokens[1].Kind, Equals, ScenarioKind)
	c.Assert(tokens[1].Value, Equals, "Scenario heading with underline")

}

func (s *MySuite) TestParsingHeadingWithUnderlineAndHash(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading with hash ").text("=====").scenarioHeading("Scenario heading with hash").text("----").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 4)

	c.Assert(tokens[0].Kind, Equals, SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec heading with hash")

	c.Assert(tokens[1].Kind, Equals, CommentKind)
	c.Assert(tokens[1].Value, Equals, "=====")

	c.Assert(tokens[2].Kind, Equals, ScenarioKind)
	c.Assert(tokens[2].Value, Equals, "Scenario heading with hash")

	c.Assert(tokens[3].Kind, Equals, CommentKind)
	c.Assert(tokens[3].Value, Equals, "----")

}

func (s *MySuite) TestParseSpecTags(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading with hash ").tags("tag1", "tag2").scenarioHeading("Scenario Heading").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[1].Kind, Equals, TagKind)
	c.Assert(len(tokens[1].Args), Equals, 2)
	c.Assert(tokens[1].Args[0], Equals, "tag1")
	c.Assert(tokens[1].Args[1], Equals, "tag2")
	c.Assert(tokens[1].LineText, Equals, "tags: tag1,tag2")
	c.Assert(tokens[1].Value, Equals, "tag1,tag2")
}

func (s *MySuite) TestParseSpecTagsWithSpace(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading with hash ").text(" tags :tag1,tag2").scenarioHeading("Scenario Heading").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[1].Kind, Equals, TagKind)
	c.Assert(len(tokens[1].Args), Equals, 2)
	c.Assert(tokens[1].Args[0], Equals, "tag1")
	c.Assert(tokens[1].Args[1], Equals, "tag2")
	c.Assert(tokens[1].LineText, Equals, " tags :tag1,tag2")
	c.Assert(tokens[1].Value, Equals, "tag1,tag2")
}

func (s *MySuite) TestParseEmptyTags(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading with hash ").tags("tag1", "", "tag2", "").scenarioHeading("Scenario Heading").String()
	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[1].Kind, Equals, TagKind)
	c.Assert(len(tokens[1].Args), Equals, 2)
	c.Assert(tokens[1].Args[0], Equals, "tag1")
	c.Assert(tokens[1].Args[1], Equals, "tag2")
	c.Assert(tokens[1].LineText, Equals, "tags: tag1,,tag2,")
	c.Assert(tokens[1].Value, Equals, "tag1,,tag2,")
}

func (s *MySuite) TestParseScenarioTags(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading with hash ").scenarioHeading("Scenario Heading").tags("tag1", "tag2").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[2].Kind, Equals, TagKind)
	c.Assert(len(tokens[2].Args), Equals, 2)
	c.Assert(tokens[2].Args[0], Equals, "tag1")
	c.Assert(tokens[2].Args[1], Equals, "tag2")
	c.Assert(tokens[2].LineText, Equals, "tags: tag1,tag2")
	c.Assert(tokens[2].Value, Equals, "tag1,tag2")
}

func (s *MySuite) TestParseSpecTagsBeforeSpecHeading(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().tags("tag1 ").specHeading("Spec heading with hash ").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)

	c.Assert(tokens[0].Kind, Equals, TagKind)
	c.Assert(len(tokens[0].Args), Equals, 1)
	c.Assert(tokens[0].Args[0], Equals, "tag1")
	c.Assert(tokens[0].LineText, Equals, "tags: tag1 ")
	c.Assert(tokens[0].Value, Equals, "tag1")
}

func (s *MySuite) TestParsingSimpleDataTable(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading").text("|name|id|").text("|---|---|").text("|john|123|").text("|james|007|").String()

	tokens, err := parser.GenerateTokens(specText)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 5)

	c.Assert(tokens[1].Kind, Equals, TableHeader)
	c.Assert(len(tokens[1].Args), Equals, 2)
	c.Assert(tokens[1].Args[0], Equals, "name")
	c.Assert(tokens[1].Args[1], Equals, "id")

	c.Assert(tokens[2].Kind, Equals, TableRow)
	c.Assert(len(tokens[2].Args), Equals, 2)
	c.Assert(tokens[2].Args[0], Equals, "---")
	c.Assert(tokens[2].Args[1], Equals, "---")

	c.Assert(tokens[3].Kind, Equals, TableRow)
	c.Assert(len(tokens[3].Args), Equals, 2)
	c.Assert(tokens[3].Args[0], Equals, "john")
	c.Assert(tokens[3].Args[1], Equals, "123")

	c.Assert(tokens[4].Kind, Equals, TableRow)
	c.Assert(len(tokens[4].Args), Equals, 2)
	c.Assert(tokens[4].Args[0], Equals, "james")
	c.Assert(tokens[4].Args[1], Equals, "007")

}

func (s *MySuite) TestParsingMultipleDataTable(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading").text("|name|id|").text("|john|123|").text("|james|007|").step("Example step").text("|user|role|").text("|root | admin|").String()

	tokens, err := parser.GenerateTokens(specText)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 7)

	c.Assert(tokens[1].Kind, Equals, TableHeader)
	c.Assert(len(tokens[1].Args), Equals, 2)
	c.Assert(tokens[1].Args[0], Equals, "name")
	c.Assert(tokens[1].Args[1], Equals, "id")

	c.Assert(tokens[2].Kind, Equals, TableRow)
	c.Assert(len(tokens[2].Args), Equals, 2)
	c.Assert(tokens[2].Args[0], Equals, "john")
	c.Assert(tokens[2].Args[1], Equals, "123")

	c.Assert(tokens[3].Kind, Equals, TableRow)
	c.Assert(len(tokens[3].Args), Equals, 2)
	c.Assert(tokens[3].Args[0], Equals, "james")
	c.Assert(tokens[3].Args[1], Equals, "007")

	c.Assert(tokens[5].Kind, Equals, TableHeader)
	c.Assert(len(tokens[5].Args), Equals, 2)
	c.Assert(tokens[5].Args[0], Equals, "user")
	c.Assert(tokens[5].Args[1], Equals, "role")

	c.Assert(tokens[6].Kind, Equals, TableRow)
	c.Assert(len(tokens[6].Args), Equals, 2)
	c.Assert(tokens[6].Args[0], Equals, "root")
	c.Assert(tokens[6].Args[1], Equals, "admin")
}

func (s *MySuite) TestParsingDataTableWithEmptyHeaderSeparatorRow(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading").text("|name|id|").text("|||").text("|john|123|").String()

	tokens, err := parser.GenerateTokens(specText)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 4)

	c.Assert(tokens[1].Kind, Equals, TableHeader)
	c.Assert(len(tokens[1].Args), Equals, 2)
	c.Assert(tokens[1].Args[0], Equals, "name")
	c.Assert(tokens[1].Args[1], Equals, "id")

	c.Assert(tokens[2].Kind, Equals, TableRow)
	c.Assert(len(tokens[2].Args), Equals, 2)
	c.Assert(tokens[2].Args[0], Equals, "")
	c.Assert(tokens[2].Args[1], Equals, "")

	c.Assert(tokens[3].Kind, Equals, TableRow)
	c.Assert(len(tokens[3].Args), Equals, 2)
	c.Assert(tokens[3].Args[0], Equals, "john")
	c.Assert(tokens[3].Args[1], Equals, "123")

}

func (s *MySuite) TestParsingDataTableRowEscapingPipe(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading").text("| name|id | address| phone|").text("| escape \\| pipe |second|third|").String()

	tokens, err := parser.GenerateTokens(specText)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[1].Kind, Equals, TableHeader)
	c.Assert(len(tokens[1].Args), Equals, 4)
	c.Assert(tokens[1].Args[0], Equals, "name")
	c.Assert(tokens[1].Args[1], Equals, "id")
	c.Assert(tokens[1].Args[2], Equals, "address")
	c.Assert(tokens[1].Args[3], Equals, "phone")

	c.Assert(tokens[2].Kind, Equals, TableRow)
	c.Assert(len(tokens[2].Args), Equals, 3)
	c.Assert(tokens[2].Args[0], Equals, "escape | pipe")
	c.Assert(tokens[2].Args[1], Equals, "second")
	c.Assert(tokens[2].Args[2], Equals, "third")

}

func (s *MySuite) TestParsingDataTableThrowsErrorWithEmptyHeader(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading").text("| name|id |||").text("| escape \\| pipe |second|third|second|").String()

	_, err := parser.GenerateTokens(specText)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "line no: 2, Table header should not be blank")
}

func (s *MySuite) TestParsingDataTableThrowsErrorWithSameColumnHeader(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading").text("| name|id|name|").text("|1|2|3|").String()

	_, err := parser.GenerateTokens(specText)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "line no: 2, Table header cannot have repeated column values")
}

func (s *MySuite) TestParsingDataTableWithSeparatorAsHeader(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading").text("|---|--|-|").text("|---|--|-|").text("|---|--|-|").text("| escape \\| pipe |second|third|").String()

	tokens, err := parser.GenerateTokens(specText)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 5)

	c.Assert(tokens[1].Kind, Equals, TableHeader)
	c.Assert(len(tokens[1].Args), Equals, 3)
	c.Assert(tokens[1].Args[0], Equals, "---")
	c.Assert(tokens[1].Args[1], Equals, "--")
	c.Assert(tokens[1].Args[2], Equals, "-")

	c.Assert(tokens[2].Kind, Equals, TableRow)
	c.Assert(len(tokens[2].Args), Equals, 3)
	c.Assert(tokens[2].Args[0], Equals, "---")
	c.Assert(tokens[2].Args[1], Equals, "--")
	c.Assert(tokens[2].Args[2], Equals, "-")

}

func (s *MySuite) TestParsingSpecWithMultipleLines(c *C) {
	parser := new(SpecParser)
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

	tokens, err := parser.GenerateTokens(specText)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 15)

	c.Assert(tokens[0].Kind, Equals, SpecKind)
	c.Assert(tokens[1].Kind, Equals, CommentKind)
	c.Assert(tokens[2].Kind, Equals, CommentKind)

	c.Assert(tokens[3].Kind, Equals, StepKind)
	c.Assert(tokens[3].Value, Equals, "Context step with {static} and {special}")

	c.Assert(tokens[4].Kind, Equals, TableHeader)
	c.Assert(tokens[5].Kind, Equals, TableRow)
	c.Assert(tokens[6].Kind, Equals, TableRow)
	c.Assert(tokens[7].Kind, Equals, TagKind)
	c.Assert(tokens[8].Kind, Equals, ScenarioKind)
	c.Assert(tokens[9].Kind, Equals, TagKind)

	c.Assert(tokens[10].Kind, Equals, StepKind)
	c.Assert(tokens[10].Value, Equals, "first with {static} and {dynamic}")

	c.Assert(tokens[11].Kind, Equals, CommentKind)

	c.Assert(tokens[12].Kind, Equals, StepKind)
	c.Assert(tokens[12].Value, Equals, "{special} and {dynamic} with {static}")

	c.Assert(tokens[13].Kind, Equals, ScenarioKind)

	c.Assert(tokens[14].Kind, Equals, StepKind)
	c.Assert(tokens[14].Value, Equals, "another")

}

func (s *MySuite) TestParsingConceptInSpec(c *C) {
	parser := new(SpecParser)
	conceptDictionary := new(ConceptDictionary)
	specText := SpecBuilder().specHeading("A spec heading").
		scenarioHeading("First flow").
		step("concept step").
		step("another step").String()
	step1 := &Step{Value: "step 1"}
	step2 := &Step{Value: "step 2"}
	concept1 := &Step{Value: "concept step", ConceptSteps: []*Step{step1, step2}, IsConcept: true}
	err := conceptDictionary.Add([]*Step{concept1}, "file.cpt")
	tokens, err := parser.GenerateTokens(specText)
	c.Assert(err, IsNil)
	spec, parseResult := parser.CreateSpecification(tokens, conceptDictionary)

	c.Assert(parseResult.Ok, Equals, true)
	firstStepInSpec := spec.Scenarios[0].Steps[0]
	secondStepInSpec := spec.Scenarios[0].Steps[1]
	c.Assert(firstStepInSpec.ConceptSteps[0].Parent, Equals, firstStepInSpec)
	c.Assert(firstStepInSpec.ConceptSteps[1].Parent, Equals, firstStepInSpec)
	c.Assert(firstStepInSpec.Parent, IsNil)
	c.Assert(secondStepInSpec.Parent, IsNil)
}

func (s *MySuite) TestTableFromInvalidFile(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading").text("table: inputinvalid.csv").String()

	_, err := parser.GenerateTokens(specText)
	c.Assert(err.Message, Equals, "Could not resolve table from table: inputinvalid.csv")
}

func (s *MySuite) TestTableInputFromInvalidFileAndDataTableNotInitialized(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading").text("table: inputinvalid.csv").String()

	_, parseRes := parser.parse(specText, new(ConceptDictionary))
	c.Assert(parseRes.ParseError.Message, Equals, "Could not resolve table from table: inputinvalid.csv")
	c.Assert(parseRes.Ok, Equals, false)
}

func (s *MySuite) TestTableInputFromFile(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading").text("Table: inputinvalid.csv").String()

	_, parseRes := parser.parse(specText, new(ConceptDictionary))
	c.Assert(parseRes.ParseError.Message, Equals, "Could not resolve table from Table: inputinvalid.csv")
	c.Assert(parseRes.Ok, Equals, false)
}

func (s *MySuite) TestTableInputFromFileIfPathNotSpecified(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading").text("Table: ").String()

	_, parseRes := parser.parse(specText, new(ConceptDictionary))
	c.Assert(parseRes.ParseError.Message, Equals, "Table location not specified")
	c.Assert(parseRes.Ok, Equals, false)
}
