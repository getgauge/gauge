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
		lsp.Location{URI: uri, Range: lsp.Range{
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
