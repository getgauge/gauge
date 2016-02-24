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

package parse

import (
	"fmt"
	"unicode/utf8"
)

type item struct {
	typ     itemType
	pos     int
	lineNum int
	val     string
}

func (i item) String() string {
	return fmt.Sprintf("%q %d", i.val, i.lineNum)
}

// itemType identifies the type of lex items.
type itemType int

const (
	itemError itemType = iota // error occurred; value is text of error
	itemNewline
	itemEOF
	itemText
	itemH1Hash
	itemDoubleUnderline
	itemAsterisk
)

const (
	eof = -1
)

var (
	lineCtr = 1
)

// stateFn represents the state of the scanner as a function that returns the next state.
type stateFn func(*lexer) stateFn

// lexer holds the state of the scanner.
type lexer struct {
	name       string    // the name of the input; used only for error reports
	input      string    // the string being scanned
	state      stateFn   // the next lexing function to enter
	currentPos int       // current position in the input
	start      int       // start position of this item
	width      int       // width of current rune
	items      chan item // channel of scanned items
}

func newLexer(name, input string) *lexer {
	l := &lexer{
		name:       name,
		input:      input,
		currentPos: 0,
		start:      0,
		items:      make(chan item),
	}
	lineCtr = 1
	go l.run()
	return l
}

// // line returns the current line in the input.
// func (l *lexer) line() string {
// 	var r rune
// 	for r != eof && r != '\n' && r != '\r' {
// 		r = l.rune()
// 	}
// 	currLine := l.input[l.start:l.currentPos]
// 	l.start = l.currentPos
// 	return currLine
// }

// rune returns the next rune in the input.
func (l *lexer) rune() rune {
	if l.currentPos >= len(l.input) {
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.currentPos:])
	l.width = w
	l.currentPos += w
	return r
}

// peek returns but does not consume the next rune in the input.
func (l *lexer) peek() rune {
	r := l.rune()
	if r != eof {
		l.backup()
	}
	return r
}

// backup steps back one rune. Can only be called once per call of next.
func (l *lexer) backup() {
	l.currentPos -= l.width
}

// // peekNextLine returns the next line, without consuming it.
// func (l *lexer) peekNextLine() string {
// 	currLexerPos := l.currentPos
// 	currLexerStart := l.start

// 	currLine := l.line()
// 	currLine = l.line()

// 	l.currentPos = currLexerPos
// 	l.start = currLexerStart
// 	return currLine
// }

// emit passes an item back to the client.
func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, lineCtr, l.input[l.start:l.currentPos]}
	l.start = l.currentPos
}

// nextItem returns the next item from the input.
func (l *lexer) nextItem() item {
	item := <-l.items
	return item
}

// TODO: don't like this func name. need to change it
func (l *lexer) hasUnemittedText() bool {
	return l.currentPos-l.start > 1
}

// run runs the state machine for the lexer.
func (l *lexer) run() {
	for l.state = lexStart; l.state != nil; {
		l.state = l.state(l)
	}
	close(l.items)
}

func lexStart(l *lexer) stateFn {
	r := l.rune()
	for isSpace(r) || isNewLine(r) {
		r = l.rune()
		l.start = l.currentPos
	}

	switch {
	case isH1Hash(r):
		return lexH1Hash
	default:
		l.backup()
		return lexText
	}
}

func lexText(l *lexer) stateFn {
	r := l.rune()
	for !isEOF(r) {
		switch {
		case isDoubleUnderline(r):
			return lexDoubleUnderline
		case isAsterisk(r):
			return lexAsterisk
		case isNewLine(r):
			if l.hasUnemittedText() {
				l.backup()
				l.emit(itemText)
				l.rune()
			}
			return lexNewLine
		}

		r = l.rune()
	}

	if l.hasUnemittedText() {
		l.emit(itemText)
	}
	return lexEOF
}

func lexAsterisk(l *lexer) stateFn {
	l.emit(itemAsterisk)
	return lexText
}

func lexH1Hash(l *lexer) stateFn {
	l.emit(itemH1Hash)
	return lexText
}

func lexEOF(l *lexer) stateFn {
	l.emit(itemEOF)
	return nil
}

func lexNewLine(l *lexer) stateFn {
	l.emit(itemNewline)
	lineCtr++
	return lexText
}

func lexDoubleUnderline(l *lexer) stateFn {
	for isDoubleUnderline(l.peek()) {
		l.rune()
	}
	l.emit(itemDoubleUnderline)
	return lexText
}

// ---------------------------------------------
// helper funcs

func isEOF(r rune) bool {
	return r == eof
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isNewLine(r rune) bool {
	return r == '\n' || r == '\r'
}

func isGaugeParamChar(r rune) bool {
	return r == '"' || r == '<' || r == '>'
}

func isAsterisk(r rune) bool {
	return r == '*'
}

func isDoubleUnderline(r rune) bool {
	return r == '='
}

func isH1Hash(r rune) bool {
	return r == '#'
}
