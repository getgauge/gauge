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
	"strings"

	"github.com/getgauge/gauge/gauge"
)

type parser struct {
	name string
	text string
	// Parsing only
	lex       *lexer
	token     [3]item // three-token lookahead for parser.
	peekCount int
}

func (p *parser) next() item {
	if p.peekCount > 0 {
		p.peekCount--
	} else {
		p.token[0] = p.lex.nextItem()
	}
	return p.token[p.peekCount]
}

func (p *parser) peek() item {
	if p.peekCount > 0 {
		return p.token[p.peekCount-1]
	}
	p.peekCount = 1
	p.token[0] = p.lex.nextItem()
	return p.token[0]
}

func New(name, text string) *parser {
	return &parser{
		name: name,
		text: text,
		lex:  NewLexer(name, text),
	}
}

func ParseConcept(filename, text string) gauge.Concept {
	p := New(filename, text)
	cptHeading := p.parseConceptHeading()
	steps := p.parseSteps()

	return gauge.Concept{
		FileName: filename,
		LineNo:   0,
		Heading:  cptHeading,
		Steps:    steps,
	}
}

func (p *parser) parseConceptHeading() string {
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

func (p *parser) parseSteps() []gauge.Step {
	steps := make([]gauge.Step, 0)
	for p.peek().typ != itemEOF {
		token := p.next()
		if token.typ == itemAsterisk {
			steps = append(steps, gauge.Step{
				LineNo:     0,
				ActualText: strings.TrimSpace(p.next().val),
			})
		}
	}
	return steps
}
