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
		cursorPos:   lsp.Position{Line: 0, Character: 19},
		wantRange:   lsp.Range{Start: lsp.Position{Line: 0, Character: 13}, End: lsp.Position{Line: 0, Character: 20}},
		wantArgType: gauge.Static,
		wantSuffix:  "\"",
	},
	{
		input:       `* Step with <dynamic> param"`,
		cursorPos:   lsp.Position{Line: 0, Character: 13},
		wantRange:   lsp.Range{Start: lsp.Position{Line: 0, Character: 13}, End: lsp.Position{Line: 0, Character: 21}},
		wantArgType: gauge.Dynamic,
		wantSuffix:  ">",
	},
	{
		input:       `* Step with <dynamic> param and "static" param`,
		cursorPos:   lsp.Position{Line: 0, Character: 14},
		wantRange:   lsp.Range{Start: lsp.Position{Line: 0, Character: 13}, End: lsp.Position{Line: 0, Character: 21}},
		wantArgType: gauge.Dynamic,
		wantSuffix:  ">",
	},
	{
		input:       `* Step with <dynamic> param and "static" param`,
		cursorPos:   lsp.Position{Line: 0, Character: 36},
		wantRange:   lsp.Range{Start: lsp.Position{Line: 0, Character: 33}, End: lsp.Position{Line: 0, Character: 40}},
		wantArgType: gauge.Static,
		wantSuffix:  "\"",
	},
	{
		input:       `* Step with "static" param and <dynamic> param`,
		cursorPos:   lsp.Position{Line: 0, Character: 35},
		wantRange:   lsp.Range{Start: lsp.Position{Line: 0, Character: 32}, End: lsp.Position{Line: 0, Character: 40}},
		wantArgType: gauge.Dynamic,
		wantSuffix:  ">",
	},
	{
		input:       `* Step with "static" param and <dynamic> param`,
		cursorPos:   lsp.Position{Line: 0, Character: 13},
		wantRange:   lsp.Range{Start: lsp.Position{Line: 0, Character: 13}, End: lsp.Position{Line: 0, Character: 20}},
		wantArgType: gauge.Static,
		wantSuffix:  "\"",
	},

	{
		input:       `* Incomplete step with <para`,
		cursorPos:   lsp.Position{Line: 0, Character: 28},
		wantRange:   lsp.Range{Start: lsp.Position{Line: 0, Character: 24}, End: lsp.Position{Line: 0, Character: 28}},
		wantArgType: gauge.Dynamic,
		wantSuffix:  ">",
	},
	{
		input:       `* Incomplete step with <para`,
		cursorPos:   lsp.Position{Line: 0, Character: 26},
		wantRange:   lsp.Range{Start: lsp.Position{Line: 0, Character: 24}, End: lsp.Position{Line: 0, Character: 28}},
		wantArgType: gauge.Dynamic,
		wantSuffix:  ">",
	},
	{
		input:       `* Step with "one" and "two" static params`,
		cursorPos:   lsp.Position{Line: 0, Character: 15},
		wantRange:   lsp.Range{Start: lsp.Position{Line: 0, Character: 13}, End: lsp.Position{Line: 0, Character: 17}},
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

func TestShouldAddParam(t *testing.T) {
	if !shouldAddParam(gauge.Static) {
		t.Errorf("should add static params")
	}
	if !shouldAddParam(gauge.Dynamic) {
		t.Errorf("should add dynamic params")
	}
	if shouldAddParam(gauge.TableArg) {
		t.Errorf("should not add table params")
	}
	if !shouldAddParam(gauge.SpecialTable) {
		t.Errorf("should add special table params")
	}
	if !shouldAddParam(gauge.SpecialString) {
		t.Errorf("should add special string params")
	}
}
