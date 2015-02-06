// Copyright 2014 ThoughtWorks, Inc.

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

package main

import (
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestPrefixingMessage(c *C) {
	writer := &pluginConsoleWriter{pluginName: "my-plugin"}
	prefixedLines := writer.addPrefixToEachLine("Hello\nWorld")
	c.Assert(prefixedLines, Equals, "[my-plugin Plugin] : Hello\n"+
		"[my-plugin Plugin] : World")
}

func (s *MySuite) TestPrefixingMessageEndingWithNewLine(c *C) {
	writer := &pluginConsoleWriter{pluginName: "my-plugin"}
	prefixedLines := writer.addPrefixToEachLine("Hello\nWorld\n")
	c.Assert(prefixedLines, Equals, "[my-plugin Plugin] : Hello\n"+
		"[my-plugin Plugin] : World\n")

}

func (s *MySuite) TestPrefixingMultiLineMessagWithNewLine(c *C) {
	writer := &pluginConsoleWriter{pluginName: "my-plugin"}
	prefixedLines := writer.addPrefixToEachLine("\nHello\nWorld\n\nFoo bar\n")
	c.Assert(prefixedLines, Equals, "[my-plugin Plugin] : \n"+
		"[my-plugin Plugin] : Hello\n"+
		"[my-plugin Plugin] : World\n"+
		"[my-plugin Plugin] : \n"+
		"[my-plugin Plugin] : Foo bar\n")

}
