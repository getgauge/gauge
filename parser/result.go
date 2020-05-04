/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package parser

import "fmt"

// ParseError holds information about a parse failure
type ParseError struct {
	FileName string
	LineNo   int
	SpanEnd  int
	Message  string
	LineText string
}

// Error prints error with filename, line number, error message and step text.
func (se ParseError) Error() string {
	if se.LineNo == 0 && se.FileName == "" {
		return se.Message
	}
	return fmt.Sprintf("%s:%d %s => '%s'", se.FileName, se.LineNo, se.Message, se.LineText)
}

func (token *Token) String() string {
	return fmt.Sprintf("kind:%d, lineNo:%d, value:%s, line:%s, args:%s", token.Kind, token.LineNo, token.Value, token.LineText(), token.Args)
}

// ParseResult is a collection of parse errors and warnings in a file.
type ParseResult struct {
	ParseErrors []ParseError
	Warnings    []*Warning
	Ok          bool
	FileName    string
}

// Errors Prints parse errors and critical errors.
func (result *ParseResult) Errors() (errors []string) {
	for _, err := range result.ParseErrors {
		errors = append(errors, fmt.Sprintf("[ParseError] %s", err.Error()))
	}
	return
}

// Warning is used to indicate discrepancies that do not necessarily need to break flow.
type Warning struct {
	FileName    string
	LineNo      int
	LineSpanEnd int
	Message     string
}

func (warning *Warning) String() string {
	return fmt.Sprintf("%s:%d %s", warning.FileName, warning.LineNo, warning.Message)
}
