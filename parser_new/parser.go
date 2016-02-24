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

import "strings"

type nodeType int

const (
	nodeConcept nodeType = iota
	nodeStep
)

// Node represents the node of AST
type Node struct {
	nodeType
	value    string
	children []*Node
}

func newNode(typ nodeType, value string) *Node {
	return &Node{typ, value, make([]*Node, 0)}
}

// Parser represents the parser object
type Parser struct {
	name string
	text string
	// Parsing only
	lex       *lexer
	token     [3]item // three-token lookahead for parser.
	peekCount int
}

// next returns the next token from the lexer channel for processing
func (p *Parser) next() item {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.token[0] = p.lex.nextItem()
	}
	return p.token[p.peekCount]
}

// peek is a lookahead for the next item on the lexer channel
func (p *Parser) peek() item {
	if p.peekCount > 0 {
		return p.token[p.peekCount-1]
	}
	p.peekCount = 1
	p.token[0] = p.lex.nextItem()
	return p.token[0]
}

// New returns the new parser object
func New(name, text string) *Parser {
	return &Parser{
		name: name,
		text: text,
		lex:  newLexer(name, text),
	}
}

// Concept takes in the contents of concept file and returns the root node
// of the concept AST
func Concept(filename, text string) *Node {
	p := New(filename, text)
	cptHeading := p.parseConceptHeading()
	conceptNode := newNode(nodeConcept, cptHeading)
	conceptNode.children = p.parseSteps()
	return conceptNode
}

func (p *Parser) parseConceptHeading() string {
	token := p.next()
	var heading string
	switch {
	case token.typ == itemH1Hash:
		heading = strings.TrimSpace(p.next().val)
	default:
		heading = strings.TrimSpace(token.val)
	}
	p.next()
	return heading
}

func (p *Parser) parseSteps() []*Node {
	var steps []*Node
	for p.peek().typ != itemEOF {
		token := p.next()
		if token.typ == itemAsterisk {
			steps = append(steps, newNode(nodeStep, strings.TrimSpace(p.next().val)))
		}
	}
	return steps
}
