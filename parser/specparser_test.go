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
	"path/filepath"
	"testing"

	"github.com/getgauge/gauge/gauge"

	. "gopkg.in/check.v1"
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
	c.Assert(tokens[0].Kind, Equals, gauge.SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec Heading")
}

func (s *MySuite) TestParsingASingleStep(c *C) {
	parser := new(SpecParser)
	tokens, err := parser.GenerateTokens("* test step \"arg\" ")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)
	c.Assert(tokens[0].Kind, Equals, gauge.StepKind)
}

func (s *MySuite) TestParsingMultipleSpecHeading(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec Heading").specHeading("Another Spec Heading").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)
	c.Assert(tokens[0].Kind, Equals, gauge.SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec Heading")
	c.Assert(tokens[1].Kind, Equals, gauge.SpecKind)
	c.Assert(tokens[1].Value, Equals, "Another Spec Heading")
}

func (s *MySuite) TestParsingThrowErrorForEmptySpecHeading(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("").specHeading("Another Spec Heading").String()

	_, err := parser.GenerateTokens(specText)

	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "1: Spec heading should have at least one character => ''")
}

func (s *MySuite) TestParsingScenarioHeading(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec Heading").scenarioHeading("First scenario").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)
	c.Assert(tokens[1].Kind, Equals, gauge.ScenarioKind)
	c.Assert(tokens[1].Value, Equals, "First scenario")
}

func (s *MySuite) TestParsingThrowErrorForEmptyScenarioHeading(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec Heading").scenarioHeading("").String()

	_, err := parser.GenerateTokens(specText)

	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "2: Scenario heading should have at least one character => ''")
}

func (s *MySuite) TestParsingScenarioWithoutSpecHeading(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().scenarioHeading("Scenario Heading").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)
	c.Assert(tokens[0].Kind, Equals, gauge.ScenarioKind)
}

func (s *MySuite) TestParsingComments(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec Heading").text("Hello i am a comment ").text("### A h3 comment").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[1].Kind, Equals, gauge.CommentKind)
	c.Assert(tokens[1].Value, Equals, "Hello i am a comment")

	c.Assert(tokens[2].Kind, Equals, gauge.CommentKind)
	c.Assert(tokens[2].Value, Equals, "### A h3 comment")
}

func (s *MySuite) TestParsingSpecHeadingWithUnderlineOneChar(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().text("Spec heading with underline ").text("=").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)

	c.Assert(tokens[0].Kind, Equals, gauge.SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec heading with underline")

}

func (s *MySuite) TestParsingSpecHeadingWithUnderlineMultipleChar(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().text("Spec heading with underline ").text("=====").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)

	c.Assert(tokens[0].Kind, Equals, gauge.SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec heading with underline")

}

func (s *MySuite) TestParsingCommentWithUnderlineAndInvalidCharacters(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().text("A comment that will be with invalid underline").text("===89s").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)

	c.Assert(tokens[0].Kind, Equals, gauge.CommentKind)
	c.Assert(tokens[0].Value, Equals, "A comment that will be with invalid underline")

	c.Assert(tokens[1].Kind, Equals, gauge.CommentKind)
	c.Assert(tokens[1].Value, Equals, "===89s")
}

func (s *MySuite) TestParsingScenarioHeadingWithUnderline(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().text("Spec heading with underline ").text("=").text("Scenario heading with underline").text("-").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)

	c.Assert(tokens[0].Kind, Equals, gauge.SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec heading with underline")

	c.Assert(tokens[1].Kind, Equals, gauge.ScenarioKind)
	c.Assert(tokens[1].Value, Equals, "Scenario heading with underline")

}

func (s *MySuite) TestParsingScenarioHeadingWithUnderlineMultipleChar(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().text("Spec heading with underline ").text("=").text("Scenario heading with underline").text("----").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 2)

	c.Assert(tokens[0].Kind, Equals, gauge.SpecKind)
	c.Assert(tokens[0].Value, Equals, "Spec heading with underline")

	c.Assert(tokens[1].Kind, Equals, gauge.ScenarioKind)
	c.Assert(tokens[1].Value, Equals, "Scenario heading with underline")

}

func (s *MySuite) TestParsingHeadingWithUnderlineAndHash(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading with hash ").text("=====").scenarioHeading("Scenario heading with hash").text("----").String()

	tokens, err := parser.GenerateTokens(specText)

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
	specText := SpecBuilder().specHeading("Spec heading with hash ").tags("tag1", "tag2").scenarioHeading("Scenario Heading").String()

	tokens, err := parser.GenerateTokens(specText)

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[1].Kind, Equals, gauge.TagKind)
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

	c.Assert(tokens[1].Kind, Equals, gauge.TagKind)
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

	c.Assert(tokens[1].Kind, Equals, gauge.TagKind)
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

	c.Assert(tokens[2].Kind, Equals, gauge.TagKind)
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

	c.Assert(tokens[0].Kind, Equals, gauge.TagKind)
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
	specText := SpecBuilder().specHeading("Spec heading").text("|name|id|").text("|john|123|").text("|james|007|").step("Example step").text("|user|role|").text("|root | admin|").String()

	tokens, err := parser.GenerateTokens(specText)
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
	specText := SpecBuilder().specHeading("Spec heading").text("|name|id|").text("|||").text("|john|123|").String()

	tokens, err := parser.GenerateTokens(specText)
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
	specText := SpecBuilder().specHeading("Spec heading").text("| name|id | address| phone|").text("| escape \\| pipe |second|third|").String()

	tokens, err := parser.GenerateTokens(specText)
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
	specText := SpecBuilder().specHeading("Spec heading").text("| name|id |||").text("| escape \\| pipe |second|third|second|").String()

	_, err := parser.GenerateTokens(specText)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "2: Table header should not be blank => '| name|id |||'")
}

func (s *MySuite) TestParsingDataTableThrowsErrorWithSameColumnHeader(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading").text("| name|id|name|").text("|1|2|3|").String()

	_, err := parser.GenerateTokens(specText)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "2: Table header cannot have repeated column values => '| name|id|name|'")
}

func (s *MySuite) TestParsingDataTableWithSeparatorAsHeader(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading").text("|---|--|-|").text("|---|--|-|").text("|---|--|-|").text("| escape \\| pipe |second|third|").String()

	tokens, err := parser.GenerateTokens(specText)
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

	c.Assert(tokens[0].Kind, Equals, gauge.SpecKind)
	c.Assert(tokens[1].Kind, Equals, gauge.CommentKind)
	c.Assert(tokens[2].Kind, Equals, gauge.NewLineKind)

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

func (s *MySuite) TestParsingMultilineStep(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().
		step("step1").
		text("").
		tableHeader("foo|bar").
		tableRow("somerow|another").String()

	tokens, err := parser.GenerateTokens(specText)
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 4)

	c.Assert(tokens[0].Kind, Equals, gauge.StepKind)
	c.Assert(tokens[1].Kind, Equals, gauge.NewLineKind)

	c.Assert(tokens[2].Kind, Equals, gauge.TableHeader)
	c.Assert(tokens[3].Kind, Equals, gauge.TableRow)
}

func (s *MySuite) TestParsingSpecWithTearDownSteps(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("A spec heading").
		text("Hello, i am a comment").
		scenarioHeading("First flow").
		step("another").
		text("_____").
		step("step1").
		step("step2").String()

	tokens, err := parser.GenerateTokens(specText)
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

func (s *MySuite) TestParsingConceptInSpec(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("A spec heading").
		scenarioHeading("First flow").
		step("test concept step 1").
		step("another step").String()
	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "concept.cpt"))
	AddConcepts(path, conceptDictionary)
	tokens, err := parser.GenerateTokens(specText)
	c.Assert(err, IsNil)
	spec, parseResult := parser.CreateSpecification(tokens, conceptDictionary)

	c.Assert(parseResult.Ok, Equals, true)
	firstStepInSpec := spec.Scenarios[0].Steps[0]
	secondStepInSpec := spec.Scenarios[0].Steps[1]
	c.Assert(firstStepInSpec.ConceptSteps[0].Parent, Equals, firstStepInSpec)
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

	_, parseRes := parser.Parse(specText, gauge.NewConceptDictionary())
	c.Assert(parseRes.ParseErrors[0].Message, Equals, "Could not resolve table from table: inputinvalid.csv")
	c.Assert(parseRes.Ok, Equals, false)
}

func (s *MySuite) TestTableInputFromFile(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading").text("Table: inputinvalid.csv").String()

	_, parseRes := parser.Parse(specText, gauge.NewConceptDictionary())
	c.Assert(parseRes.ParseErrors[0].Message, Equals, "Could not resolve table from Table: inputinvalid.csv")
	c.Assert(parseRes.Ok, Equals, false)
}

func (s *MySuite) TestTableInputFromFileIfPathNotSpecified(c *C) {
	parser := new(SpecParser)
	specText := SpecBuilder().specHeading("Spec heading").text("Table: ").String()

	_, parseRes := parser.Parse(specText, gauge.NewConceptDictionary())
	c.Assert(parseRes.ParseErrors[0].Message, Equals, "Table location not specified")
	c.Assert(parseRes.Ok, Equals, false)
}

func (s *MySuite) TestToSplitTagNames(c *C) {
	allTags := splitAndTrimTags("tag1 , tag2,   tag3")
	c.Assert(allTags[0], Equals, "tag1")
	c.Assert(allTags[1], Equals, "tag2")
	c.Assert(allTags[2], Equals, "tag3")
}

func (s *MySuite) TestThrowsErrorForMultipleSpecHeading(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 3},
		&Token{Kind: gauge.SpecKind, Value: "Another Heading", LineNo: 4},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, false)

	c.Assert(result.ParseErrors[0].Message, Equals, "Parse error: Multiple spec headings found in same file")
	c.Assert(result.ParseErrors[0].LineNo, Equals, 4)
}

func (s *MySuite) TestThrowsErrorForScenarioWithoutSpecHeading(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 1},
		&Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 2},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, false)

	c.Assert(result.ParseErrors[0].Message, Equals, "Parse error: Scenario should be defined after the spec heading")
	c.Assert(result.ParseErrors[0].LineNo, Equals, 1)
}

func (s *MySuite) TestThrowsErrorForDuplicateScenariosWithinTheSameSpec(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 3},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 4},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, false)

	c.Assert(result.ParseErrors[0].Message, Equals, "Parse error: Duplicate scenario definition 'Scenario Heading' found in the same specification")
	c.Assert(result.ParseErrors[0].LineNo, Equals, 4)
}

func (s *MySuite) TestSpecWithHeadingAndSimpleSteps(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 3},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(len(spec.Items), Equals, 1)
	c.Assert(spec.Items[0], Equals, spec.Scenarios[0])
	scenarioItems := (spec.Items[0]).(*gauge.Scenario).Items
	c.Assert(scenarioItems[0], Equals, spec.Scenarios[0].Steps[0])

	c.Assert(result.Ok, Equals, true)
	c.Assert(spec.Heading.LineNo, Equals, 1)
	c.Assert(spec.Heading.Value, Equals, "Spec Heading")

	c.Assert(len(spec.Scenarios), Equals, 1)
	c.Assert(spec.Scenarios[0].Heading.LineNo, Equals, 2)
	c.Assert(spec.Scenarios[0].Heading.Value, Equals, "Scenario Heading")
	c.Assert(len(spec.Scenarios[0].Steps), Equals, 1)
	c.Assert(spec.Scenarios[0].Steps[0].Value, Equals, "Example step")
}

func (s *MySuite) TestStepsAndComments(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.CommentKind, Value: "A comment with some text and **bold** characters", LineNo: 2},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 3},
		&Token{Kind: gauge.CommentKind, Value: "Another comment", LineNo: 4},
		&Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 5},
		&Token{Kind: gauge.CommentKind, Value: "Third comment", LineNo: 6},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())
	c.Assert(len(spec.Items), Equals, 2)
	c.Assert(spec.Items[0], Equals, spec.Comments[0])
	c.Assert(spec.Items[1], Equals, spec.Scenarios[0])

	scenarioItems := (spec.Items[1]).(*gauge.Scenario).Items
	c.Assert(3, Equals, len(scenarioItems))
	c.Assert(scenarioItems[0], Equals, spec.Scenarios[0].Comments[0])
	c.Assert(scenarioItems[1], Equals, spec.Scenarios[0].Steps[0])
	c.Assert(scenarioItems[2], Equals, spec.Scenarios[0].Comments[1])

	c.Assert(result.Ok, Equals, true)
	c.Assert(spec.Heading.Value, Equals, "Spec Heading")

	c.Assert(len(spec.Comments), Equals, 1)
	c.Assert(spec.Comments[0].LineNo, Equals, 2)
	c.Assert(spec.Comments[0].Value, Equals, "A comment with some text and **bold** characters")

	c.Assert(len(spec.Scenarios), Equals, 1)
	scenario := spec.Scenarios[0]

	c.Assert(2, Equals, len(scenario.Comments))
	c.Assert(scenario.Comments[0].LineNo, Equals, 4)
	c.Assert(scenario.Comments[0].Value, Equals, "Another comment")

	c.Assert(scenario.Comments[1].LineNo, Equals, 6)
	c.Assert(scenario.Comments[1].Value, Equals, "Third comment")

	c.Assert(scenario.Heading.Value, Equals, "Scenario Heading")
	c.Assert(len(scenario.Steps), Equals, 1)
}

func (s *MySuite) TestStepsWithParam(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.TableHeader, Args: []string{"id"}, LineNo: 2},
		&Token{Kind: gauge.TableRow, Args: []string{"1"}, LineNo: 3},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 4},
		&Token{Kind: gauge.StepKind, Value: "enter {static} with {dynamic}", LineNo: 5, Args: []string{"user \\n foo", "id"}},
		&Token{Kind: gauge.StepKind, Value: "sample \\{static\\}", LineNo: 6},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())
	c.Assert(result.Ok, Equals, true)
	step := spec.Scenarios[0].Steps[0]
	c.Assert(step.Value, Equals, "enter {} with {}")
	c.Assert(step.LineNo, Equals, 5)
	c.Assert(len(step.Args), Equals, 2)
	c.Assert(step.Args[0].Value, Equals, "user \\n foo")
	c.Assert(step.Args[0].ArgType, Equals, gauge.Static)
	c.Assert(step.Args[1].Value, Equals, "id")
	c.Assert(step.Args[1].ArgType, Equals, gauge.Dynamic)
	c.Assert(step.Args[1].Name, Equals, "id")

	escapedStep := spec.Scenarios[0].Steps[1]
	c.Assert(escapedStep.Value, Equals, "sample \\{static\\}")
	c.Assert(len(escapedStep.Args), Equals, 0)
}

func (s *MySuite) TestStepsWithKeywords(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "sample {static} and {dynamic}", LineNo: 3, Args: []string{"name"}},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result, NotNil)
	c.Assert(result.Ok, Equals, false)
	c.Assert(result.ParseErrors[0].Message, Equals, "Step text should not have '{static}' or '{dynamic}' or '{special}'")
	c.Assert(result.ParseErrors[0].LineNo, Equals, 3)
}

func (s *MySuite) TestContextWithKeywords(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.StepKind, Value: "sample {static} and {dynamic}", LineNo: 3, Args: []string{"name"}},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result, NotNil)
	c.Assert(result.Ok, Equals, false)
	c.Assert(result.ParseErrors[0].Message, Equals, "Step text should not have '{static}' or '{dynamic}' or '{special}'")
	c.Assert(result.ParseErrors[0].LineNo, Equals, 3)
}

func (s *MySuite) TestSpecWithDataTable(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading"},
		&Token{Kind: gauge.CommentKind, Value: "Comment before data table"},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}},
		&Token{Kind: gauge.TableRow, Args: []string{"1", "foo"}},
		&Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}},
		&Token{Kind: gauge.CommentKind, Value: "Comment before data table"},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(len(spec.Items), Equals, 3)
	c.Assert(spec.Items[0], Equals, spec.Comments[0])
	c.Assert(spec.Items[1], DeepEquals, &spec.DataTable)
	c.Assert(spec.Items[2], Equals, spec.Comments[1])

	c.Assert(result.Ok, Equals, true)
	c.Assert(spec.DataTable, NotNil)
	c.Assert(len(spec.DataTable.Table.Get("id")), Equals, 2)
	c.Assert(len(spec.DataTable.Table.Get("name")), Equals, 2)
	c.Assert(spec.DataTable.Table.Get("id")[0].Value, Equals, "1")
	c.Assert(spec.DataTable.Table.Get("id")[0].CellType, Equals, gauge.Static)
	c.Assert(spec.DataTable.Table.Get("id")[1].Value, Equals, "2")
	c.Assert(spec.DataTable.Table.Get("id")[1].CellType, Equals, gauge.Static)
	c.Assert(spec.DataTable.Table.Get("name")[0].Value, Equals, "foo")
	c.Assert(spec.DataTable.Table.Get("name")[0].CellType, Equals, gauge.Static)
	c.Assert(spec.DataTable.Table.Get("name")[1].Value, Equals, "bar")
	c.Assert(spec.DataTable.Table.Get("name")[1].CellType, Equals, gauge.Static)
}

func (s *MySuite) TestStepWithInlineTable(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "Step with inline table", LineNo: 3},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}},
		&Token{Kind: gauge.TableRow, Args: []string{"1", "foo"}},
		&Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, true)
	step := spec.Scenarios[0].Steps[0]

	c.Assert(step.Args[0].ArgType, Equals, gauge.TableArg)
	inlineTable := step.Args[0].Table
	c.Assert(inlineTable, NotNil)

	c.Assert(step.Value, Equals, "Step with inline table {}")
	c.Assert(step.HasInlineTable, Equals, true)
	c.Assert(len(inlineTable.Get("id")), Equals, 2)
	c.Assert(len(inlineTable.Get("name")), Equals, 2)
	c.Assert(inlineTable.Get("id")[0].Value, Equals, "1")
	c.Assert(inlineTable.Get("id")[0].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("id")[1].Value, Equals, "2")
	c.Assert(inlineTable.Get("id")[1].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("name")[0].Value, Equals, "foo")
	c.Assert(inlineTable.Get("name")[0].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("name")[1].Value, Equals, "bar")
	c.Assert(inlineTable.Get("name")[1].CellType, Equals, gauge.Static)
}

func (s *MySuite) TestStepWithInlineTableWithDynamicParam(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.TableHeader, Args: []string{"type1", "type2"}},
		&Token{Kind: gauge.TableRow, Args: []string{"1", "2"}},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "Step with inline table", LineNo: 3},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}},
		&Token{Kind: gauge.TableRow, Args: []string{"1", "<type1>"}},
		&Token{Kind: gauge.TableRow, Args: []string{"2", "<type2>"}},
		&Token{Kind: gauge.TableRow, Args: []string{"<2>", "<type3>"}},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, true)
	step := spec.Scenarios[0].Steps[0]

	c.Assert(step.Args[0].ArgType, Equals, gauge.TableArg)
	inlineTable := step.Args[0].Table
	c.Assert(inlineTable, NotNil)

	c.Assert(step.Value, Equals, "Step with inline table {}")
	c.Assert(len(inlineTable.Get("id")), Equals, 3)
	c.Assert(len(inlineTable.Get("name")), Equals, 3)
	c.Assert(inlineTable.Get("id")[0].Value, Equals, "1")
	c.Assert(inlineTable.Get("id")[0].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("id")[1].Value, Equals, "2")
	c.Assert(inlineTable.Get("id")[1].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("id")[2].Value, Equals, "<2>")
	c.Assert(inlineTable.Get("id")[2].CellType, Equals, gauge.Static)

	c.Assert(inlineTable.Get("name")[0].Value, Equals, "type1")
	c.Assert(inlineTable.Get("name")[0].CellType, Equals, gauge.Dynamic)
	c.Assert(inlineTable.Get("name")[1].Value, Equals, "type2")
	c.Assert(inlineTable.Get("name")[1].CellType, Equals, gauge.Dynamic)
	c.Assert(inlineTable.Get("name")[2].Value, Equals, "<type3>")
	c.Assert(inlineTable.Get("name")[2].CellType, Equals, gauge.Static)
}

func (s *MySuite) TestStepWithInlineTableWithUnResolvableDynamicParam(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.TableHeader, Args: []string{"type1", "type2"}},
		&Token{Kind: gauge.TableRow, Args: []string{"1", "2"}},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "Step with inline table", LineNo: 3},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}},
		&Token{Kind: gauge.TableRow, Args: []string{"1", "<invalid>"}},
		&Token{Kind: gauge.TableRow, Args: []string{"2", "<type2>"}},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())
	c.Assert(result.Ok, Equals, true)
	c.Assert(spec.Scenarios[0].Steps[0].Args[0].Table.Get("id")[0].Value, Equals, "1")
	c.Assert(spec.Scenarios[0].Steps[0].Args[0].Table.Get("name")[0].Value, Equals, "<invalid>")
	c.Assert(result.Warnings[0].Message, Equals, "Dynamic param <invalid> could not be resolved, Treating it as static param")
}

func (s *MySuite) TestContextWithInlineTable(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading"},
		&Token{Kind: gauge.StepKind, Value: "Context with inline table"},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}},
		&Token{Kind: gauge.TableRow, Args: []string{"1", "foo"}},
		&Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}},
		&Token{Kind: gauge.TableRow, Args: []string{"3", "not a <dynamic>"}},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading"},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())
	c.Assert(len(spec.Items), Equals, 2)
	c.Assert(spec.Items[0], DeepEquals, spec.Contexts[0])
	c.Assert(spec.Items[1], Equals, spec.Scenarios[0])

	c.Assert(result.Ok, Equals, true)
	context := spec.Contexts[0]

	c.Assert(context.Args[0].ArgType, Equals, gauge.TableArg)
	inlineTable := context.Args[0].Table

	c.Assert(inlineTable, NotNil)
	c.Assert(context.Value, Equals, "Context with inline table {}")
	c.Assert(len(inlineTable.Get("id")), Equals, 3)
	c.Assert(len(inlineTable.Get("name")), Equals, 3)
	c.Assert(inlineTable.Get("id")[0].Value, Equals, "1")
	c.Assert(inlineTable.Get("id")[0].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("id")[1].Value, Equals, "2")
	c.Assert(inlineTable.Get("id")[1].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("id")[2].Value, Equals, "3")
	c.Assert(inlineTable.Get("id")[2].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("name")[0].Value, Equals, "foo")
	c.Assert(inlineTable.Get("name")[0].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("name")[1].Value, Equals, "bar")
	c.Assert(inlineTable.Get("name")[1].CellType, Equals, gauge.Static)
	c.Assert(inlineTable.Get("name")[2].Value, Equals, "not a <dynamic>")
	c.Assert(inlineTable.Get("name")[2].CellType, Equals, gauge.Static)
}

func (s *MySuite) TestErrorWhenDataTableHasOnlyHeader(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading"},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}, LineNo: 3},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading"},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, false)
	c.Assert(result.ParseErrors[0].Message, Equals, "Data table should have at least 1 data row")
	c.Assert(result.ParseErrors[0].LineNo, Equals, 3)
}

func (s *MySuite) TestWarningWhenParsingMultipleDataTable(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading"},
		&Token{Kind: gauge.CommentKind, Value: "Comment before data table"},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}},
		&Token{Kind: gauge.TableRow, Args: []string{"1", "foo"}},
		&Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}},
		&Token{Kind: gauge.CommentKind, Value: "Comment before data table"},
		&Token{Kind: gauge.TableHeader, Args: []string{"phone"}, LineNo: 7},
		&Token{Kind: gauge.TableRow, Args: []string{"1"}},
		&Token{Kind: gauge.TableRow, Args: []string{"2"}},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, true)
	c.Assert(len(result.Warnings), Equals, 1)
	c.Assert(result.Warnings[0].String(), Equals, "line no: 7, Multiple data table present, ignoring table")

}

func (s *MySuite) TestWarningWhenParsingTableOccursWithoutStep(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "name"}, LineNo: 3},
		&Token{Kind: gauge.TableRow, Args: []string{"1", "foo"}, LineNo: 4},
		&Token{Kind: gauge.TableRow, Args: []string{"2", "bar"}, LineNo: 5},
		&Token{Kind: gauge.StepKind, Value: "Step", LineNo: 6},
		&Token{Kind: gauge.CommentKind, Value: "comment in between", LineNo: 7},
		&Token{Kind: gauge.TableHeader, Args: []string{"phone"}, LineNo: 8},
		&Token{Kind: gauge.TableRow, Args: []string{"1"}},
		&Token{Kind: gauge.TableRow, Args: []string{"2"}},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())
	c.Assert(result.Ok, Equals, true)
	c.Assert(len(result.Warnings), Equals, 2)
	c.Assert(result.Warnings[0].String(), Equals, "line no: 3, Table not associated with a step, ignoring table")
	c.Assert(result.Warnings[1].String(), Equals, "line no: 8, Table not associated with a step, ignoring table")

}

func (s *MySuite) TestAddSpecTags(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.TagKind, Args: []string{"tag1", "tag2"}, LineNo: 2},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 3},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.Tags.Values), Equals, 2)
	c.Assert(spec.Tags.Values[0], Equals, "tag1")
	c.Assert(spec.Tags.Values[1], Equals, "tag2")
}

func (s *MySuite) TestAddSpecTagsAndScenarioTags(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.TagKind, Args: []string{"tag1", "tag2"}, LineNo: 2},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 3},
		&Token{Kind: gauge.TagKind, Args: []string{"tag3", "tag4"}, LineNo: 2},
	}

	spec, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.Tags.Values), Equals, 2)
	c.Assert(spec.Tags.Values[0], Equals, "tag1")
	c.Assert(spec.Tags.Values[1], Equals, "tag2")

	tags := spec.Scenarios[0].Tags
	c.Assert(len(tags.Values), Equals, 2)
	c.Assert(tags.Values[0], Equals, "tag3")
	c.Assert(tags.Values[1], Equals, "tag4")
}

func (s *MySuite) TestErrorOnAddingDynamicParamterWithoutADataTable(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "Step with a {dynamic}", Args: []string{"foo"}, LineNo: 3, LineText: "*Step with a <foo>"},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, false)
	c.Assert(result.ParseErrors[0].Message, Equals, "Dynamic parameter <foo> could not be resolved")
	c.Assert(result.ParseErrors[0].LineNo, Equals, 3)

}

func (s *MySuite) TestErrorOnAddingDynamicParamterWithoutDataTableHeaderValue(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.TableHeader, Args: []string{"id, name"}, LineNo: 2},
		&Token{Kind: gauge.TableRow, Args: []string{"123, hello"}, LineNo: 3},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 4},
		&Token{Kind: gauge.StepKind, Value: "Step with a {dynamic}", Args: []string{"foo"}, LineNo: 5, LineText: "*Step with a <foo>"},
	}

	_, result := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())

	c.Assert(result.Ok, Equals, false)
	c.Assert(result.ParseErrors[0].Message, Equals, "Dynamic parameter <foo> could not be resolved")
	c.Assert(result.ParseErrors[0].LineNo, Equals, 5)

}

func (s *MySuite) TestCreateStepFromSimpleConcept(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "test concept step 1", LineNo: 3},
	}

	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "concept.cpt"))
	AddConcepts(path, conceptDictionary)
	spec, result := new(SpecParser).CreateSpecification(tokens, conceptDictionary)
	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.Scenarios[0].Steps), Equals, 1)
	specConceptStep := spec.Scenarios[0].Steps[0]
	c.Assert(specConceptStep.IsConcept, Equals, true)
	assertStepEqual(c, &gauge.Step{LineNo: 2, Value: "step 1", LineText: "step 1"}, specConceptStep.ConceptSteps[0])
}

func (s *MySuite) TestCreateStepFromConceptWithParameters(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "assign id {static} and name {static}", Args: []string{"foo", "foo1"}, LineNo: 3},
		&Token{Kind: gauge.StepKind, Value: "assign id {static} and name {static}", Args: []string{"bar", "bar1"}, LineNo: 4},
	}

	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "dynamic_param_concept.cpt"))
	AddConcepts(path, conceptDictionary)

	spec, result := new(SpecParser).CreateSpecification(tokens, conceptDictionary)
	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.Scenarios[0].Steps), Equals, 2)

	firstConceptStep := spec.Scenarios[0].Steps[0]
	c.Assert(firstConceptStep.IsConcept, Equals, true)
	c.Assert(firstConceptStep.ConceptSteps[0].Value, Equals, "add id {}")
	c.Assert(firstConceptStep.ConceptSteps[0].Args[0].Value, Equals, "userid")
	c.Assert(firstConceptStep.ConceptSteps[1].Value, Equals, "add name {}")
	c.Assert(firstConceptStep.ConceptSteps[1].Args[0].Value, Equals, "username")
	c.Assert(firstConceptStep.GetArg("username").Value, Equals, "foo1")
	c.Assert(firstConceptStep.GetArg("userid").Value, Equals, "foo")

	secondConceptStep := spec.Scenarios[0].Steps[1]
	c.Assert(secondConceptStep.IsConcept, Equals, true)
	c.Assert(secondConceptStep.ConceptSteps[0].Value, Equals, "add id {}")
	c.Assert(secondConceptStep.ConceptSteps[0].Args[0].Value, Equals, "userid")
	c.Assert(secondConceptStep.ConceptSteps[1].Value, Equals, "add name {}")
	c.Assert(secondConceptStep.ConceptSteps[1].Args[0].Value, Equals, "username")
	c.Assert(secondConceptStep.GetArg("username").Value, Equals, "bar1")
	c.Assert(secondConceptStep.GetArg("userid").Value, Equals, "bar")

}

func (s *MySuite) TestCreateStepFromConceptWithDynamicParameters(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "description"}, LineNo: 2},
		&Token{Kind: gauge.TableRow, Args: []string{"123", "Admin fellow"}, LineNo: 3},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 4},
		&Token{Kind: gauge.StepKind, Value: "assign id {dynamic} and name {dynamic}", Args: []string{"id", "description"}, LineNo: 5},
	}

	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "dynamic_param_concept.cpt"))
	AddConcepts(path, conceptDictionary)
	spec, result := new(SpecParser).CreateSpecification(tokens, conceptDictionary)
	c.Assert(result.Ok, Equals, true)

	c.Assert(len(spec.Items), Equals, 2)
	c.Assert(spec.Items[0], DeepEquals, &spec.DataTable)
	c.Assert(spec.Items[1], Equals, spec.Scenarios[0])

	scenarioItems := (spec.Items[1]).(*gauge.Scenario).Items
	c.Assert(scenarioItems[0], Equals, spec.Scenarios[0].Steps[0])

	c.Assert(len(spec.Scenarios[0].Steps), Equals, 1)

	firstConcept := spec.Scenarios[0].Steps[0]
	c.Assert(firstConcept.IsConcept, Equals, true)
	c.Assert(firstConcept.ConceptSteps[0].Value, Equals, "add id {}")
	c.Assert(firstConcept.ConceptSteps[0].Args[0].ArgType, Equals, gauge.Dynamic)
	c.Assert(firstConcept.ConceptSteps[0].Args[0].Value, Equals, "userid")
	c.Assert(firstConcept.ConceptSteps[1].Value, Equals, "add name {}")
	c.Assert(firstConcept.ConceptSteps[1].Args[0].Value, Equals, "username")
	c.Assert(firstConcept.ConceptSteps[1].Args[0].ArgType, Equals, gauge.Dynamic)

	arg1 := firstConcept.Lookup.GetArg("userid")
	c.Assert(arg1.Value, Equals, "id")
	c.Assert(arg1.ArgType, Equals, gauge.Dynamic)

	arg2 := firstConcept.Lookup.GetArg("username")
	c.Assert(arg2.Value, Equals, "description")
	c.Assert(arg2.ArgType, Equals, gauge.Dynamic)
}

func (s *MySuite) TestCreateStepFromConceptWithInlineTable(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 4},
		&Token{Kind: gauge.StepKind, Value: "assign id {static} and name", Args: []string{"sdf"}, LineNo: 3},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "description"}, LineNo: 4},
		&Token{Kind: gauge.TableRow, Args: []string{"123", "Admin"}, LineNo: 5},
		&Token{Kind: gauge.TableRow, Args: []string{"456", "normal fellow"}, LineNo: 6},
	}

	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "dynamic_param_concept.cpt"))
	AddConcepts(path, conceptDictionary)
	spec, result := new(SpecParser).CreateSpecification(tokens, conceptDictionary)
	c.Assert(result.Ok, Equals, true)

	steps := spec.Scenarios[0].Steps
	c.Assert(len(steps), Equals, 1)
	c.Assert(steps[0].IsConcept, Equals, true)
	c.Assert(steps[0].Value, Equals, "assign id {} and name {}")
	c.Assert(len(steps[0].Args), Equals, 2)
	c.Assert(steps[0].Args[1].ArgType, Equals, gauge.TableArg)
	c.Assert(len(steps[0].ConceptSteps), Equals, 2)
}

func (s *MySuite) TestCreateStepFromConceptWithInlineTableHavingDynamicParam(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "description"}, LineNo: 2},
		&Token{Kind: gauge.TableRow, Args: []string{"123", "Admin"}, LineNo: 3},
		&Token{Kind: gauge.TableRow, Args: []string{"456", "normal fellow"}, LineNo: 4},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 5},
		&Token{Kind: gauge.StepKind, Value: "assign id {static} and name", Args: []string{"sdf"}, LineNo: 6},
		&Token{Kind: gauge.TableHeader, Args: []string{"user-id", "description", "name"}, LineNo: 7},
		&Token{Kind: gauge.TableRow, Args: []string{"<id>", "<description>", "root"}, LineNo: 8},
	}

	conceptDictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "dynamic_param_concept.cpt"))
	AddConcepts(path, conceptDictionary)
	spec, result := new(SpecParser).CreateSpecification(tokens, conceptDictionary)
	c.Assert(result.Ok, Equals, true)

	steps := spec.Scenarios[0].Steps
	c.Assert(len(steps), Equals, 1)
	c.Assert(steps[0].IsConcept, Equals, true)
	c.Assert(steps[0].Value, Equals, "assign id {} and name {}")
	c.Assert(len(steps[0].Args), Equals, 2)
	c.Assert(steps[0].Args[1].ArgType, Equals, gauge.TableArg)
	table := steps[0].Args[1].Table
	c.Assert(table.Get("user-id")[0].Value, Equals, "id")
	c.Assert(table.Get("user-id")[0].CellType, Equals, gauge.Dynamic)
	c.Assert(table.Get("description")[0].Value, Equals, "description")
	c.Assert(table.Get("description")[0].CellType, Equals, gauge.Dynamic)
	c.Assert(table.Get("name")[0].Value, Equals, "root")
	c.Assert(table.Get("name")[0].CellType, Equals, gauge.Static)
	c.Assert(len(steps[0].ConceptSteps), Equals, 2)
}

func (s *MySuite) TestCreateConceptStep(c *C) {
	dictionary := gauge.NewConceptDictionary()
	path, _ := filepath.Abs(filepath.Join("testdata", "param_nested_concept.cpt"))
	AddConcepts(path, dictionary)

	argsInStep := []*gauge.StepArg{&gauge.StepArg{Name: "bar", Value: "first name", ArgType: gauge.Static}, &gauge.StepArg{Name: "far", Value: "last name", ArgType: gauge.Static}}
	originalStep := &gauge.Step{
		LineNo:         12,
		Value:          "create user {} {}",
		LineText:       "create user \"first name\" \"last name\"",
		Args:           argsInStep,
		IsConcept:      false,
		HasInlineTable: false}

	createConceptStep(new(gauge.Specification), dictionary.Search("create user {} {}").ConceptStep, originalStep)

	c.Assert(originalStep.IsConcept, Equals, true)
	c.Assert(len(originalStep.ConceptSteps), Equals, 1)
	c.Assert(originalStep.Args[0].Value, Equals, "first name")
	c.Assert(originalStep.Lookup.GetArg("bar").Value, Equals, "first name")
	c.Assert(originalStep.Args[1].Value, Equals, "last name")
	c.Assert(originalStep.Lookup.GetArg("far").Value, Equals, "last name")

	nestedConcept := originalStep.ConceptSteps[0]
	c.Assert(nestedConcept.IsConcept, Equals, true)
	c.Assert(len(nestedConcept.ConceptSteps), Equals, 1)

	c.Assert(nestedConcept.Args[0].ArgType, Equals, gauge.Dynamic)
	c.Assert(nestedConcept.Args[0].Name, Equals, "bar")

	c.Assert(nestedConcept.Args[1].ArgType, Equals, gauge.Dynamic)
	c.Assert(nestedConcept.Args[1].Name, Equals, "far")

	c.Assert(nestedConcept.ConceptSteps[0].Args[0].ArgType, Equals, gauge.Dynamic)
	c.Assert(nestedConcept.ConceptSteps[0].Args[0].Name, Equals, "baz")

	c.Assert(nestedConcept.Lookup.GetArg("baz").ArgType, Equals, gauge.Dynamic)
	c.Assert(nestedConcept.Lookup.GetArg("baz").Value, Equals, "bar")
}

func (s *MySuite) TestCreateInValidSpecialArgInStep(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.TableHeader, Args: []string{"unknown:foo", "description"}, LineNo: 2},
		&Token{Kind: gauge.TableRow, Args: []string{"123", "Admin"}, LineNo: 3},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "Example {special} step", LineNo: 3, Args: []string{"unknown:foo"}},
	}
	spec, parseResults := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())
	c.Assert(spec.Scenarios[0].Steps[0].Args[0].ArgType, Equals, gauge.Dynamic)
	c.Assert(len(parseResults.Warnings), Equals, 1)
	c.Assert(parseResults.Warnings[0].Message, Equals, "Could not resolve special param type <unknown:foo>. Treating it as dynamic param.")
}

func (s *MySuite) TestTearDownSteps(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.CommentKind, Value: "A comment with some text and **bold** characters", LineNo: 2},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 3},
		&Token{Kind: gauge.CommentKind, Value: "Another comment", LineNo: 4},
		&Token{Kind: gauge.StepKind, Value: "Example step", LineNo: 5},
		&Token{Kind: gauge.CommentKind, Value: "Third comment", LineNo: 6},
		&Token{Kind: gauge.TearDownKind, Value: "____", LineNo: 7},
		&Token{Kind: gauge.StepKind, Value: "Example step1", LineNo: 8},
		&Token{Kind: gauge.CommentKind, Value: "Fourth comment", LineNo: 9},
		&Token{Kind: gauge.StepKind, Value: "Example step2", LineNo: 10},
	}

	spec, _ := new(SpecParser).CreateSpecification(tokens, gauge.NewConceptDictionary())
	c.Assert(len(spec.TearDownSteps), Equals, 2)
	c.Assert(spec.TearDownSteps[0].Value, Equals, "Example step1")
	c.Assert(spec.TearDownSteps[0].LineNo, Equals, 8)
	c.Assert(spec.TearDownSteps[1].Value, Equals, "Example step2")
	c.Assert(spec.TearDownSteps[1].LineNo, Equals, 10)
}

func (s *MySuite) TestParsingOfTableWithHyphens(c *C) {
	p := new(SpecParser)

	text := SpecBuilder().specHeading("My Spec Heading").text("|id|").text("|--|").text("|1 |").text("|- |").String()
	tokens, _ := p.GenerateTokens(text)

	spec, _ := p.CreateSpecification(tokens, gauge.NewConceptDictionary())
	c.Assert((len(spec.DataTable.Table.Get("id"))), Equals, 2)
	c.Assert(spec.DataTable.Table.Get("id")[0].Value, Equals, "1")
	c.Assert(spec.DataTable.Table.Get("id")[1].Value, Equals, "-")
}

func (s *MySuite) TestCreateStepWithNewlineBetweenTextAndTable(c *C) {
	tokens := []*Token{
		&Token{Kind: gauge.SpecKind, Value: "Spec Heading", LineNo: 1},
		&Token{Kind: gauge.ScenarioKind, Value: "Scenario Heading", LineNo: 2},
		&Token{Kind: gauge.StepKind, Value: "some random step", LineNo: 3},
		&Token{Kind: gauge.NewLineKind, Value: "\n", LineNo: 4},
		&Token{Kind: gauge.TableHeader, Args: []string{"id", "description"}, LineNo: 5},
		&Token{Kind: gauge.TableRow, Args: []string{"123", "Admin"}, LineNo: 6},
		&Token{Kind: gauge.TableRow, Args: []string{"456", "normal fellow"}, LineNo: 7},
	}

	conceptDictionary := gauge.NewConceptDictionary()
	spec, _ := new(SpecParser).CreateSpecification(tokens, conceptDictionary)

	c.Assert(spec.Scenarios[0].Steps[0].HasInlineTable, Equals, true)
}
