/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package lang

import (
	"fmt"
	"testing"
	"time"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/runner"
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

func TestGetAllImplementedStepValues(t *testing.T) {
	stepValues := []gauge.StepValue{
		{
			StepValue:              "hello world",
			Args:                   []string{},
			ParameterizedStepValue: "hello world",
		},
		{
			StepValue:              "hello {}",
			Args:                   []string{"world"},
			ParameterizedStepValue: "hello <world>",
		},
	}
	responses := map[gauge_messages.Message_MessageType]interface{}{}
	responses[gauge_messages.Message_StepNamesResponse] = &gauge_messages.StepNamesResponse{
		Steps: []string{
			"hello world",
			"hello <world>",
		},
	}
	lRunner.runner = &runner.GrpcRunner{LegacyClient: &mockClient{responses: responses}, Timeout: time.Second * 30}

	got, err := allImplementedStepValues()

	if err != nil {
		t.Errorf("expected getAllImplementedStepValues() to not have errors, got %v", err)
	}
	for _, sv := range stepValues {
		if !contains(got, sv) {
			t.Errorf("expected getAllImplementedStepValues() to contain %v.\ngetAllImplementedStepValues() == %v", sv, got)
		}
	}
}

func TestGetAllImplementedStepValuesShouldGivesEmptyIfRunnerRespondWithError(t *testing.T) {
	responses := map[gauge_messages.Message_MessageType]interface{}{}
	responses[gauge_messages.Message_StepNamesResponse] = &gauge_messages.StepNamesResponse{}
	lRunner.runner = &runner.GrpcRunner{Timeout: time.Second * 30, LegacyClient: &mockClient{responses: responses, err: fmt.Errorf("can't get steps")}}
	got, err := allImplementedStepValues()

	if err == nil {
		t.Error("expected getAllImplementedStepValues() to have errors, got nil")
	}
	if len(got) > 0 {
		t.Errorf("expected 0 values. got %v", len(got))
	}
}

func TestRemoveDuplicateStepValues(t *testing.T) {
	stepValues := []gauge.StepValue{
		{
			Args:                   []string{},
			ParameterizedStepValue: "hello world",
			StepValue:              "hello world",
		}, {
			Args:                   []string{"world"},
			ParameterizedStepValue: "hello <world>",
			StepValue:              "hello {}",
		},
		{
			Args:                   []string{"gauge"},
			ParameterizedStepValue: "hello <gauge>",
			StepValue:              "hello {}",
		},
	}

	result := removeDuplicates(stepValues)

	if len(result) != 2 {
		t.Errorf("exppected 2 steps got %v", len(result))
	}
}

func contains(list []gauge.StepValue, v gauge.StepValue) bool {
	for _, e := range list {
		if e.ParameterizedStepValue == v.ParameterizedStepValue && e.StepValue == v.StepValue && len(e.Args) == len(v.Args) {
			return true
		}
	}
	return false
}
