/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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

func (s *MySuite) TestExtractStepArgsFromToken_WithImplicitMultiline(c *C) {
	token := &Token{
		Value: "simple step",
		Args:  []string{"multiline\ncontent"},
	}

	args, err := ExtractStepArgsFromToken(token)
	c.Assert(err, IsNil)
	c.Assert(args, HasLen, 1)
	c.Assert(args[0].ArgType, Equals, gauge.SpecialString)
	c.Assert(args[0].Value, Equals, "multiline\ncontent")
}

func (s *MySuite) TestExtractStepArgsFromToken_InvalidImplicitMultiline(c *C) {
	token := &Token{
		Value: "simple step", 
		Args:  []string{"arg1", "arg2"}, // Multiple args but no parameter types
	}

	_, err := ExtractStepArgsFromToken(token)
	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, "Multiline step should have exactly one argument")
}

func (s *MySuite) TestProcessStep_SkipsParameterProcessingForMultiline(c *C) {
	parser := &SpecParser{}
	token := &Token{
		Value: "step with multiline",
		Args:  []string{"multiline\ncontent"}, // Has args from lexer
	}

	errors, shouldContinue := processStep(parser, token)
	
	c.Assert(errors, HasLen, 0)
	c.Assert(shouldContinue, Equals, false)
	c.Assert(token.Value, Equals, "step with multiline") // Should remain unchanged
	c.Assert(token.Args, HasLen, 1) // Args should be preserved
}

func (s *MySuite) TestProcessStep_ProcessesParametersForRegularStep(c *C) {
	parser := &SpecParser{}
	token := &Token{
		Value: `enter user "john"`, // No args from lexer
		Args:  []string{},
	}

	errors, shouldContinue := processStep(parser, token)
	
	c.Assert(errors, HasLen, 0)
	c.Assert(shouldContinue, Equals, false)
	c.Assert(token.Value, Equals, "enter user {static}") // Should be processed
	c.Assert(token.Args, HasLen, 1)
	c.Assert(token.Args[0], Equals, "john")
}

// Integration test for the actual multiline string feature
// 1. Basic multiline string functionality
func (s *MySuite) TestParsingStepWithMultilineString(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().
		specHeading("Spec heading").
		scenarioHeading("Scenario Heading").
		step("the multiline step").
		text(`"""`).
		text("hello world").
		text("from multiline").
		text(`"""`).
		String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	
	stepToken := tokens[2]
	c.Assert(stepToken.Kind, Equals, gauge.StepKind)
	c.Assert(stepToken.Value, Equals, "the multiline step")
	c.Assert(stepToken.Args, HasLen, 1)
	c.Assert(stepToken.Args[0], Equals, "hello world\nfrom multiline")
}

// 2. Empty multiline string
func (s *MySuite) TestParsingStepWithEmptyMultilineString(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().
		specHeading("Spec heading").
		scenarioHeading("Scenario Heading").
		step("empty multiline step").
		text(`"""`).
		text(`"""`).
		String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	
	stepToken := tokens[2]
	c.Assert(stepToken.Kind, Equals, gauge.StepKind)
	c.Assert(stepToken.Value, Equals, "empty multiline step")
	c.Assert(stepToken.Args, HasLen, 1)
	c.Assert(stepToken.Args[0], Equals, "")
}
// 3. json file
func (s *MySuite) TestParsingStepWithJSONMultilineString(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().
		specHeading("Spec heading").
		scenarioHeading("Scenario Heading").
		step("json multiline step").
		text(`"""`).
		text(`{"name": "John", "age": 30}`).
		text(`"""`).
		String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	
	stepToken := tokens[2]
	c.Assert(stepToken.Kind, Equals, gauge.StepKind)
	c.Assert(stepToken.Value, Equals, "json multiline step")
	c.Assert(stepToken.Args, HasLen, 1)
	c.Assert(stepToken.Args[0], Equals, `{"name": "John", "age": 30}`)
}



// 4. Multiline string with special characters (edge case)
func (s *MySuite) TestParsingStepWithSpecialCharsMultilineString(c *C) {
	parser := new(SpecParser)
	specText := newSpecBuilder().
		step("special chars step").
		text(`"""`).
		text("line with {braces}").
		text("line with \"quotes\"").
		text(`"""`).String()

	tokens, err := parser.GenerateTokens(specText, "")
	c.Assert(err, IsNil)
	c.Assert(len(tokens), Equals, 1)

	c.Assert(tokens[0].Kind, Equals, gauge.StepKind)
	c.Assert(tokens[0].Value, Equals, "special chars step")
	c.Assert(len(tokens[0].Args), Equals, 1)
	c.Assert(tokens[0].Args[0], Equals, "line with {braces}\nline with \"quotes\"")
}