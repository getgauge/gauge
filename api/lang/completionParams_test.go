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

var testParamEditPosition = []struct {
	input       string
	cursorPos   lsp.Position
	wantRange   lsp.Range
	wantArgType gauge.ArgType
	wantSuffix  string
}{
	{
		input:       `* Step with "static" param`,
		cursorPos:   lsp.Position{Line: 0, Character: len(`* Step with "stat`)},
		wantRange:   lsp.Range{Start: lsp.Position{Line: 0, Character: len(`* Step with "`)}, End: lsp.Position{Line: 0, Character: len(`* Step with "static"`)}},
		wantArgType: gauge.Static,
		wantSuffix:  "\"",
	},
	{
		input:       `* Step with <dynamic> param"`,
		cursorPos:   lsp.Position{Line: 0, Character: len(`* Step with <`)},
		wantRange:   lsp.Range{Start: lsp.Position{Line: 0, Character: len(`* Step with <`)}, End: lsp.Position{Line: 0, Character: len(`* Step with <dynamic>`)}},
		wantArgType: gauge.Dynamic,
		wantSuffix:  ">",
	},
	{
		input:       `* Step with <dynamic> param and "static" param`,
		cursorPos:   lsp.Position{Line: 0, Character: len(`* Step with <dyna`)},
		wantRange:   lsp.Range{Start: lsp.Position{Line: 0, Character: len(`* Step with <`)}, End: lsp.Position{Line: 0, Character: len(`* Step with <dynamic>`)}},
		wantArgType: gauge.Dynamic,
		wantSuffix:  ">",
	},
	{
		input:       `* Step with <dynamic> param and "static" param`,
		cursorPos:   lsp.Position{Line: 0, Character: len(`* Step with <dynamic> param and "st`)},
		wantRange:   lsp.Range{Start: lsp.Position{Line: 0, Character: len(`* Step with <dynamic> param and "`)}, End: lsp.Position{Line: 0, Character: len(`* Step with <dynamic> param and "static"`)}},
		wantArgType: gauge.Static,
		wantSuffix:  "\"",
	},
	{
		input:       `* Step with "static" param and <dynamic> param`,
		cursorPos:   lsp.Position{Line: 0, Character: len(`* Step with "static" param and <dy`)},
		wantRange:   lsp.Range{Start: lsp.Position{Line: 0, Character: len(`* Step with "static" param and <`)}, End: lsp.Position{Line: 0, Character: len(`* Step with "static" param and <dynamic>`)}},
		wantArgType: gauge.Dynamic,
		wantSuffix:  ">",
	},
	{
		input:       `* Step with "static" param and <dynamic> param`,
		cursorPos:   lsp.Position{Line: 0, Character: len(`* Step with "`)},
		wantRange:   lsp.Range{Start: lsp.Position{Line: 0, Character: len(`* Step with "`)}, End: lsp.Position{Line: 0, Character: len(`* Step with "static"`)}},
		wantArgType: gauge.Static,
		wantSuffix:  "\"",
	},
	{
		input:       `* Incomplete step with <para`,
		cursorPos:   lsp.Position{Line: 0, Character: len(`* Incomplete step with <para`)},
		wantRange:   lsp.Range{Start: lsp.Position{Line: 0, Character: len(`* Incomplete step with <`)}, End: lsp.Position{Line: 0, Character: len(`* Incomplete step with <para`)}},
		wantArgType: gauge.Dynamic,
		wantSuffix:  ">",
	},
	{
		input:       `* Incomplete step with <para`,
		cursorPos:   lsp.Position{Line: 0, Character: len(`* Incomplete step with <`)},
		wantRange:   lsp.Range{Start: lsp.Position{Line: 0, Character: len(`* Incomplete step with <`)}, End: lsp.Position{Line: 0, Character: len(`* Incomplete step with <para`)}},
		wantArgType: gauge.Dynamic,
		wantSuffix:  ">",
	},
	{
		input:       `* Step with "one" and "two" static params`,
		cursorPos:   lsp.Position{Line: 0, Character: len(`* Step with "o`)},
		wantRange:   lsp.Range{Start: lsp.Position{Line: 0, Character: len(`* Step with "`)}, End: lsp.Position{Line: 0, Character: len(`* Step with "one"`)}},
		wantArgType: gauge.Dynamic,
		wantSuffix:  ">",
	},
}

func TestGetParamEditPosition(t *testing.T) {
	for _, test := range testParamEditPosition {
		pline := test.input
		if len(test.input) > test.cursorPos.Character {
			pline = test.input[:test.cursorPos.Character]
		}
		_, _, gotRange := getParamArgTypeAndEditRange(test.input, pline, test.cursorPos)
		if gotRange.Start.Line != test.wantRange.Start.Line || gotRange.Start.Character != test.wantRange.Start.Character {
			t.Errorf(`Incorrect Edit Start Position got: %+v , want : %+v, input : "%s", cursorPos : "%d"`, gotRange.Start, test.wantRange.Start, test.input, test.cursorPos.Character)
		}
		if gotRange.End.Line != test.wantRange.End.Line || gotRange.End.Character != test.wantRange.End.Character {
			t.Errorf(`Incorrect Edit End Position got: %+v , want : %+v, input : "%s", cursorPos : "%d"`, gotRange.End, test.wantRange.End, test.input, test.cursorPos.Character)
		}
	}
}

var shouldAddParamTests = []struct {
	aType gauge.ArgType
	want  bool
}{
	{gauge.Static, true},
	{gauge.Dynamic, true},
	{gauge.TableArg, false},
	{gauge.SpecialTable, true},
	{gauge.SpecialString, true},
}

func TestShouldAddParam(t *testing.T) {
	for _, test := range shouldAddParamTests {
		got := shouldAddParam(test.aType)
		if got != test.want {
			t.Errorf(`want %s for shouldAddParam for ArgType "%s"", but got %s`, test.want, test.aType, got)
		}
	}
}
