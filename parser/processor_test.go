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
)

func (s *MySuite) TestProcessingTokensGivesErrorWhenSpecHeadingIsEmpty(c *C) {
	parser := &SpecParser{}
	parser.lineNo = 2
	token := &Token{Value: ""}

	err, shouldSkip := processSpec(parser, token)

	c.Assert(shouldSkip, Equals, true)
	c.Assert(err.LineNo, Equals, 2)
	c.Assert(err.Message, Equals, "Spec heading should have at least one character")
	c.Assert(err.LineText, Equals, "")
}

func (s *MySuite) TestProcessingTokensGivesErrorWhenSpecHeadingHasOnlySpaces(c *C) {
	parser := &SpecParser{}
	parser.lineNo = 2
	spaces := "               "
	token := &Token{Value: spaces}

	err, shouldSkip := processSpec(parser, token)

	c.Assert(shouldSkip, Equals, true)
	c.Assert(err.LineNo, Equals, 2)
	c.Assert(err.Message, Equals, "Spec heading should have at least one character")
	c.Assert(err.LineText, Equals, spaces)
}

func (s *MySuite) TestProcessingTokensGivesNoErrorForValidSpecHeading(c *C) {
	parser := &SpecParser{}
	parser.lineNo = 2
	token := &Token{Value: "SPECHEADING"}
	var nilErr *ParseError

	err, shouldSkip := processSpec(parser, token)

	c.Assert(shouldSkip, Equals, false)
	c.Assert(err, Equals, nilErr)
}

func (s *MySuite) TestProcessingTokensGivesErrorWhenScenarioHeadingIsEmpty(c *C) {
	parser := &SpecParser{}
	parser.lineNo = 2
	token := &Token{Value: ""}

	err, shouldSkip := processScenario(parser, token)

	c.Assert(shouldSkip, Equals, true)
	c.Assert(err.LineNo, Equals, 2)
	c.Assert(err.Message, Equals, "Scenario heading should have at least one character")
	c.Assert(err.LineText, Equals, "")
}

func (s *MySuite) TestProcessingTokensGivesErrorWhenScenarioHeadingHasOnlySpaces(c *C) {
	parser := &SpecParser{}
	parser.lineNo = 2
	spaces := "            "
	token := &Token{Value: spaces}

	err, shouldSkip := processScenario(parser, token)

	c.Assert(shouldSkip, Equals, true)
	c.Assert(err.LineNo, Equals, 2)
	c.Assert(err.Message, Equals, "Scenario heading should have at least one character")
	c.Assert(err.LineText, Equals, spaces)
}

func (s *MySuite) TestProcessingTokensGivesNoErrorForValidScenarioHeading(c *C) {
	parser := &SpecParser{}
	parser.lineNo = 2
	token := &Token{Value: "SCENARIOHEADING"}
	var nilErr *ParseError

	err, shouldSkip := processScenario(parser, token)

	c.Assert(shouldSkip, Equals, false)
	c.Assert(err, Equals, nilErr)
}
