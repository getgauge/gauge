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
		t.Errorf("The parameters are not replaced correctly")
	}
}

func TestGetFilterTextWithDynamicParam(t *testing.T) {
	got := getStepFilterText("Text with {}", []string{"param1"}, []gauge.StepArg{{Name: "Args1", Value: "Args1", ArgType: gauge.Dynamic}})
	want := `Text with <Args1>`
	if got != want {
		t.Errorf("The parameters are not replaced correctly")
	}
}

func TestGetFilterTextShouldNotReplaceIfNoStepArgsGiven(t *testing.T) {
	got := getStepFilterText("Text with {}", []string{"param1"}, []gauge.StepArg{})
	want := `Text with <param1>`
	if got != want {
		t.Errorf("The parameters are not replaced correctly")
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
		t.Errorf("The parameters are not replaced correctly")
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
		cursorPos: lsp.Position{Line: 0, Character: 1},
		wantStart: lsp.Position{Line: 0, Character: 1},
		wantEnd:   lsp.Position{Line: 0, Character: 1},
	},
	{
		input:     "* ",
		cursorPos: lsp.Position{Line: 0, Character: 1},
		wantStart: lsp.Position{Line: 0, Character: 1},
		wantEnd:   lsp.Position{Line: 0, Character: 2},
	},
	{
		input:     "* Step",
		cursorPos: lsp.Position{Line: 10, Character: 1},
		wantStart: lsp.Position{Line: 10, Character: 1},
		wantEnd:   lsp.Position{Line: 10, Character: 6},
	},
	{
		input:     "* Step",
		cursorPos: lsp.Position{Line: 0, Character: 2},
		wantStart: lsp.Position{Line: 0, Character: 2},
		wantEnd:   lsp.Position{Line: 0, Character: 6},
	},
	{
		input:     "* Step",
		cursorPos: lsp.Position{Line: 0, Character: 4},
		wantStart: lsp.Position{Line: 0, Character: 2},
		wantEnd:   lsp.Position{Line: 0, Character: 6},
	},
	{
		input:     "    * Step",
		cursorPos: lsp.Position{Line: 0, Character: 7},
		wantStart: lsp.Position{Line: 0, Character: 6},
		wantEnd:   lsp.Position{Line: 0, Character: 10},
	},
	{
		input:     " * Step ",
		cursorPos: lsp.Position{Line: 0, Character: 10},
		wantStart: lsp.Position{Line: 0, Character: 3},
		wantEnd:   lsp.Position{Line: 0, Character: 10},
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
