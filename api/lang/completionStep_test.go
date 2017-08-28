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

package lang

import (
	"testing"

	"github.com/getgauge/gauge/gauge"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

func TestGetPrefix(t *testing.T) {
	want := " "
	got := getPrefix("line1\n*")

	if got != want {
		t.Errorf("GetPrefix failed for autocomplete, want: `%s`, got: `%s`", want, got)
	}
}

func TestGetPrefixWithSpace(t *testing.T) {
	want := ""
	got := getPrefix("* ")

	if got != want {
		t.Errorf("GetPrefix failed for autocomplete, want: `%s`, got: `%s`", want, got)
	}
}

func TestGetFilterTextWithStaticParam(t *testing.T) {
	got := getStepFilterText("Text with {}", []string{"param1"}, []gauge.StepArg{{Name: "Args1", Value: "Args1", ArgType: gauge.Static}})
	want := `Text with "Args1"`
	if got != want {
		t.Errorf("Parameters not replaced properly; got : %+v, want : %+v", got, want)
	}
}

func TestGetFilterTextWithDynamicParam(t *testing.T) {
	got := getStepFilterText("Text with {}", []string{"param1"}, []gauge.StepArg{{Name: "Args1", Value: "Args1", ArgType: gauge.Dynamic}})
	want := `Text with <Args1>`
	if got != want {
		t.Errorf("Parameters not replaced properly; got : %+v, want : %+v", got, want)
	}
}

func TestGetFilterTextShouldNotReplaceIfNoStepArgsGiven(t *testing.T) {
	got := getStepFilterText("Text with {}", []string{"param1"}, []gauge.StepArg{})
	want := `Text with <param1>`
	if got != want {
		t.Errorf("Parameters not replaced properly; got : %+v, want : %+v", got, want)
	}
}

func TestGetFilterTextWithLesserNumberOfStepArgsGiven(t *testing.T) {
	stepArgs := []gauge.StepArg{
		{Name: "Args1", Value: "Args1", ArgType: gauge.Dynamic},
		{Name: "Args2", Value: "Args2", ArgType: gauge.Static},
	}
	got := getStepFilterText("Text with {} {} and {}", []string{"param1", "param2", "param3"}, stepArgs)
	want := `Text with <Args1> "Args2" and <param3>`
	if got != want {
		t.Errorf("Parameters not replaced properly; got : %+v, want : %+v", got, want)
	}
}

var testEditPosition = []struct {
	input     string
	cursorPos lsp.Position
	wantStart lsp.Position
	wantEnd   lsp.Position
}{
	{
		input:     "*",
		cursorPos: lsp.Position{Line: 0, Character: len(`*`)},
		wantStart: lsp.Position{Line: 0, Character: len(`*`)},
		wantEnd:   lsp.Position{Line: 0, Character: len(`*`)},
	},
	{
		input:     "* ",
		cursorPos: lsp.Position{Line: 0, Character: len(`*`)},
		wantStart: lsp.Position{Line: 0, Character: len(`*`)},
		wantEnd:   lsp.Position{Line: 0, Character: len(`* `)},
	},
	{
		input:     "* Step",
		cursorPos: lsp.Position{Line: 10, Character: len(`*`)},
		wantStart: lsp.Position{Line: 10, Character: len(`*`)},
		wantEnd:   lsp.Position{Line: 10, Character: len(`* Step`)},
	},
	{
		input:     "* Step",
		cursorPos: lsp.Position{Line: 0, Character: len(`* `)},
		wantStart: lsp.Position{Line: 0, Character: len(`* `)},
		wantEnd:   lsp.Position{Line: 0, Character: len(`* Step`)},
	},
	{
		input:     "* Step",
		cursorPos: lsp.Position{Line: 0, Character: len(`* St`)},
		wantStart: lsp.Position{Line: 0, Character: len(`* `)},
		wantEnd:   lsp.Position{Line: 0, Character: len(`* Step`)},
	},
	{
		input:     "    * Step",
		cursorPos: lsp.Position{Line: 0, Character: len(`    * S`)},
		wantStart: lsp.Position{Line: 0, Character: len(`    * `)},
		wantEnd:   lsp.Position{Line: 0, Character: len(`    * Step`)},
	},
	{
		input:     " * Step ",
		cursorPos: lsp.Position{Line: 0, Character: len(` * Step `) + 2},
		wantStart: lsp.Position{Line: 0, Character: len(` * `)},
		wantEnd:   lsp.Position{Line: 0, Character: len(` * Step `) + 2},
	},
}

func TestGetEditPosition(t *testing.T) {
	for _, test := range testEditPosition {
		gotRange := getStepEditRange(test.input, test.cursorPos)
		if gotRange.Start.Line != test.wantStart.Line || gotRange.Start.Character != test.wantStart.Character {
			t.Errorf(`Incorrect Edit Start Position got: %+v , want : %+v, input : "%s"`, gotRange.Start, test.wantStart, test.input)
		}
		if gotRange.End.Line != test.wantEnd.Line || gotRange.End.Character != test.wantEnd.Character {
			t.Errorf(`Incorrect Edit End Position got: %+v , want : %+v, input : "%s"`, gotRange.End, test.wantEnd, test.input)
		}
	}
}
