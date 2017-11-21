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
	"encoding/json"
	"reflect"
	"testing"

	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func TestStepReferences(t *testing.T) {
	provider = &dummyInfoProvider{}
	specText := `Specification Heading
=====================

Scenario Heading
----------------

* Say <hello> to <gauge>`

	uri := util.ConvertPathToURI("foo.spec")
	f = &files{cache: make(map[string][]string)}
	f.add(uri, specText)

	b, _ := json.Marshal("Say {} to {}")
	params := json.RawMessage(b)
	want := []lsp.Location{
		{URI: uri, Range: lsp.Range{
			Start: lsp.Position{Line: 6, Character: 0},
			End:   lsp.Position{Line: 6, Character: 24},
		}},
	}
	got, err := getStepReferences(&jsonrpc2.Request{Params: &params})
	if err != nil {
		t.Fatalf("Got error %s", err.Error())
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("get step references failed, want: `%s`, got: `%s`", want, got)
	}
}

func TestStepValueAtShouldGive(t *testing.T) {
	provider = &dummyInfoProvider{}
	params := lsp.TextDocumentPositionParams{
		TextDocument: lsp.TextDocumentIdentifier{URI: "step_impl.js"},
		Position: lsp.Position{
			Line:      3,
			Character: 3,
		},
	}

	b, _ := json.Marshal(params)
	p := json.RawMessage(b)

	GetResponseFromRunner = func(m *gauge_messages.Message) (*gauge_messages.Message, error) {
		response := &gauge_messages.Message{
			MessageType: gauge_messages.Message_StepPositionsResponse,
			StepPositionsResponse: &gauge_messages.StepPositionsResponse{
				StepPositions: []*gauge_messages.StepPositionsResponse_StepPosition{
					{
						Span:      &gauge_messages.Span{Start: 2, End: 4},
						StepValue: "Step value at line {} and character {}",
					},
				},
			},
		}
		return response, nil
	}
	stepValue, err := getStepValueAt(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Fatalf("Got error %s", err.Error())
	}

	want := "Step value at line {} and character {}"

	if !reflect.DeepEqual(stepValue, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, stepValue)
	}
}
