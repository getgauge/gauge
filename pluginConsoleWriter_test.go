package main

import (
	. "launchpad.net/gocheck"
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
