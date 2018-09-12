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
	"github.com/getgauge/gauge/gauge"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestParsingSimpleStep(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading with hash ").scenarioHeading("Scenario Heading").step("sample step").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	stepToken := tokens[2]
	c.Assert(stepToken.Kind, Equals, gauge.StepKind)
	c.Assert(stepToken.Value, Equals, "sample step")
}

func (s *MySuite) TestParsingEmptyStepTextShouldThrowError(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading with hash ").scenarioHeading("Scenario Heading").step("").String()

	_, errs := parser.GenerateTokens(specText, "foo.spec")

	c.Assert(len(errs) > 0, Equals, true)
	c.Assert(errs[0].Error(), Equals, "foo.spec:3 Step should not be blank => ''")
}

func (s *MySuite) TestParsingStepWithParams(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading with hash ").scenarioHeading("Scenario Heading").step("enter user \"john\"").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	stepToken := tokens[2]
	c.Assert(stepToken.Kind, Equals, gauge.StepKind)
	c.Assert(stepToken.Value, Equals, "enter user {static}")
	c.Assert(len(stepToken.Args), Equals, 1)
	c.Assert(stepToken.Args[0], Equals, "john")
}

func (s *MySuite) TestParsingStepWithParametersWithQuotes(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading with hash ").scenarioHeading("Scenario Heading").step("\"param \\\"in quote\\\"\" step ").step("another * step with \"john 12 *-_{} \\\\ './;[]\" and \"second\"").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 4)

	firstStepToken := tokens[2]
	c.Assert(firstStepToken.Kind, Equals, gauge.StepKind)
	c.Assert(firstStepToken.Value, Equals, "{static} step")
	c.Assert(len(firstStepToken.Args), Equals, 1)
	c.Assert(firstStepToken.Args[0], Equals, "param \"in quote\"")

	secondStepToken := tokens[3]
	c.Assert(secondStepToken.Kind, Equals, gauge.StepKind)
	c.Assert(secondStepToken.Value, Equals, "another * step with {static} and {static}")
	c.Assert(len(secondStepToken.Args), Equals, 2)
	c.Assert(secondStepToken.Args[0], Equals, "john 12 *-_{} \\ './;[]")
	c.Assert(secondStepToken.Args[1], Equals, "second")

}

func (s *MySuite) TestParsingStepWithUnmatchedOpeningQuote(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading with hash ").scenarioHeading("Scenario Heading").step("sample step \"param").String()

	_, errs := parser.GenerateTokens(specText, "foo.spec")

	c.Assert(len(errs) > 0, Equals, true)
	c.Assert(errs[0].Error(), Equals, "foo.spec:3 String not terminated => 'sample step \"param'")
}

func (s *MySuite) TestParsingStepWithEscaping(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading with hash ").scenarioHeading("Scenario Heading").step("step with \\").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	stepToken := tokens[2]
	c.Assert(stepToken.Value, Equals, "step with")
}

func (s *MySuite) TestParsingExceptionIfStepContainsReservedChars(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading with hash ").scenarioHeading("Scenario Heading").step("step with {braces}").String()

	_, errs := parser.GenerateTokens(specText, "foo.spec")

	c.Assert(len(errs) > 0, Equals, true)
	c.Assert(errs[0].Error(), Equals, "foo.spec:3 '{' is a reserved character and should be escaped => 'step with {braces}'")
}

func (s *MySuite) TestParsingStepContainsEscapedReservedChars(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading with hash ").scenarioHeading("Scenario Heading").step("step with \\{braces\\}").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	stepToken := tokens[2]
	c.Assert(stepToken.Value, Equals, "step with {braces}")
}

func (s *MySuite) TestParsingSimpleStepWithDynamicParameter(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading with hash ").scenarioHeading("Scenario Heading").step("Step with \"static param\" and <name1>").String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	stepToken := tokens[2]
	c.Assert(stepToken.Value, Equals, "Step with {static} and {dynamic}")
	c.Assert(stepToken.Args[0], Equals, "static param")
	c.Assert(stepToken.Args[1], Equals, "name1")
}

func (s *MySuite) TestParsingStepWithUnmatchedDynamicParameterCharacter(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading with hash ").scenarioHeading("Scenario Heading").step("Step with \"static param\" and <name1").String()

	_, errs := parser.GenerateTokens(specText, "foo.spec")

	c.Assert(len(errs) > 0, Equals, true)
	c.Assert(errs[0].Error(), Equals, "foo.spec:3 Dynamic parameter not terminated => 'Step with \"static param\" and <name1'")

}

func (s *MySuite) TestParsingContext(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading with hash ").step("Context with \"param\"").scenarioHeading("Scenario Heading").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	contextToken := tokens[1]
	c.Assert(contextToken.Kind, Equals, gauge.StepKind)
	c.Assert(contextToken.Value, Equals, "Context with {static}")
	c.Assert(contextToken.Args[0], Equals, "param")
}

func (s *MySuite) TestParsingThrowsErrorWhenStepIsPresentWithoutStep(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().step("step without spec heading").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(tokens[0].Kind, Equals, gauge.StepKind)
	c.Assert(tokens[0].Value, Equals, "step without spec heading")

}

func (s *MySuite) TestParsingStepWithSimpleSpecialParameter(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading with hash ").scenarioHeading("Scenario Heading").step("Step with special parameter <table:user.csv>").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[2].Kind, Equals, gauge.StepKind)
	c.Assert(tokens[2].Value, Equals, "Step with special parameter {special}")
	c.Assert(len(tokens[2].Args), Equals, 1)
	c.Assert(tokens[2].Args[0], Equals, "table:user.csv")
}

func (s *MySuite) TestParsingStepWithSpecialParametersWithWhiteSpaces(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().specHeading("Spec heading with hash ").step("Step with \"first\" and special parameter <table : user.csv>").step("Another with <name> and <file  :something.txt>").String()

	tokens, err := parser.GenerateTokens(specText, "")

	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 3)

	c.Assert(tokens[1].Kind, Equals, gauge.StepKind)
	c.Assert(tokens[1].Value, Equals, "Step with {static} and special parameter {special}")
	c.Assert(len(tokens[1].Args), Equals, 2)
	c.Assert(tokens[1].Args[0], Equals, "first")
	c.Assert(tokens[1].Args[1], Equals, "table : user.csv")

	c.Assert(tokens[2].Kind, Equals, gauge.StepKind)
	c.Assert(tokens[2].Value, Equals, "Another with {dynamic} and {special}")
	c.Assert(len(tokens[2].Args), Equals, 2)
	c.Assert(tokens[2].Args[0], Equals, "name")
	c.Assert(tokens[2].Args[1], Equals, "file  :something.txt")
}

func (s *MySuite) TestParsingStepWithStaticParamHavingEscapeChar(c *C) {
	tokenValue, args, err := processStepText(`step "a\nb" only`)
	c.Assert(err, IsNil)
	c.Assert(args[0], Equals, "a\nb")
	c.Assert(tokenValue, Equals, "step {static} only")
}

func (s *MySuite) TestParsingStepWithStaticParamHavingDifferentEscapeChar(c *C) {
	tokenValue, args, err := processStepText(`step "foo bar \"hello\" all \"he\n\n\tyy \"foo\"hjhj\"" only`)
	c.Assert(err, IsNil)
	c.Assert(args[0], Equals, "foo bar \"hello\" all \"he\n\n\tyy \"foo\"hjhj\"")
	c.Assert(tokenValue, Equals, "step {static} only")
}

func (s *MySuite) TestParsingStepWithStaticParamHavingNestedEscapeSequences(c *C) {
	tokenValue, args, err := processStepText(`step "foo \t tab \n \"a\"dd r \\n" only`)
	c.Assert(err, IsNil)
	c.Assert(args[0], Equals, "foo \t tab \n \"a\"dd r \\n")
	c.Assert(tokenValue, Equals, "step {static} only")
}

func (s *MySuite) TestParsingStepWithSlash(c *C) {
	tokenValue, args, err := processStepText(`step foo \ only`)
	c.Assert(err, IsNil)
	c.Assert(len(args), Equals, 0)
	c.Assert(tokenValue, Equals, "step foo \\ only")
}

func (s *MySuite) TestParsingStepWithTab(c *C) {
	tokenValue, args, err := processStepText("step foo \t only")
	c.Assert(err, IsNil)
	c.Assert(len(args), Equals, 0)
	c.Assert(tokenValue, Equals, "step foo \t only")
}
