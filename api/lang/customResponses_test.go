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
	"fmt"
	"os"
	"github.com/getgauge/common"
	"io/ioutil"
	"path/filepath"
	"encoding/json"
	"testing"

	"reflect"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func TestGetScenariosShouldGiveTheScenarioAtCurrentCursorPosition(t *testing.T) {
	specText := `Specification Heading
=====================

Scenario Heading
----------------

* Step text

Scenario Heading2
-----------------

* Step text`

	uri := "foo.spec"
	f = &files{cache: make(map[string][]string)}
	f.add(uri, specText)

	position := lsp.Position{Line: 5, Character: 1}
	b, _ := json.Marshal(lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: uri}, Position: position})
	p := json.RawMessage(b)

	got, err := getScenarios(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected error to be nil. Got: \n%v", err.Error())
	}

	info := got.(ScenarioInfo)

	want := ScenarioInfo{
		Heading:             "Scenario Heading",
		LineNo:              4,
		ExecutionIdentifier: "foo.spec:4",
	}
	if !reflect.DeepEqual(info, want) {
		t.Errorf("expected %v to be equal %v", info, want)
	}
}

func TestGetScenariosShouldGiveTheScenariosIfCursorPositionIsNotInSpan(t *testing.T) {
	specText := `Specification Heading
=====================

Scenario Heading
----------------

* Step text

Scenario Heading2
-----------------

* Step text
`

	uri := "foo.spec"
	f = &files{cache: make(map[string][]string)}
	f.add(uri, specText)

	position := lsp.Position{Line: 2, Character: 1}
	b, _ := json.Marshal(lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: uri}, Position: position})
	p := json.RawMessage(b)

	got, err := getScenarios(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected error to be nil. Got: \n%v", err.Error())
	}

	info := got.([]ScenarioInfo)

	want := []ScenarioInfo{
		{
			Heading:             "Scenario Heading",
			LineNo:              4,
			ExecutionIdentifier: "foo.spec:4",
		},
		{
			Heading:             "Scenario Heading2",
			LineNo:              9,
			ExecutionIdentifier: "foo.spec:9",
		},
	}
	if !reflect.DeepEqual(info, want) {
		t.Errorf("expected %v to be equal %v", info, want)
	}
}

func TestGetScenariosShouldGiveTheScenariosIfDocumentIsNotOpened(t *testing.T) {
	specText := `Specification Heading
=====================

Scenario Heading
----------------

* Step text

Scenario Heading2
-----------------

* Step text
`

	uri := filepath.Join("_testdata", "foo.spec")
	ioutil.WriteFile(uri, []byte(specText), common.NewFilePermissions)

	position := lsp.Position{Line: 2, Character: 1}
	b, _ := json.Marshal(lsp.TextDocumentPositionParams{TextDocument: lsp.TextDocumentIdentifier{URI: uri}, Position: position})
	p := json.RawMessage(b)

	got, err := getScenarios(&jsonrpc2.Request{Params: &p})

	if err != nil {
		t.Errorf("expected error to be nil. Got: \n%v", err.Error())
	}

	info := got.([]ScenarioInfo)

	want := []ScenarioInfo{
		{
			Heading:             "Scenario Heading",
			LineNo:              4,
			ExecutionIdentifier: filepath.Join("_testdata","foo.spec:4"),
		},
		{
			Heading:             "Scenario Heading2",
			LineNo:              9,
			ExecutionIdentifier: filepath.Join("_testdata", "foo.spec:9"),
		},
	}
	if !reflect.DeepEqual(info, want) {
		t.Errorf("expected %v to be equal %v", info, want)
	}

	os.Remove(uri)
}


func TestGetSpecsShouldReturnAllSpecsInDirectory(t *testing.T) {
	specTextTemplate := `Specification %d
=====================

Scenario Heading
----------------

* Step text
`
	rootPath, err := filepath.Abs("_testdata")
	if err != nil {
		t.Error("Unable to get absolute path for _testdata")
	}
	os.Setenv(common.GaugeProjectRootEnv, rootPath)
	os.Mkdir(filepath.Join("_testdata", common.SpecsDirectoryName), common.NewDirectoryPermissions)
	want := make([]specInfo, 0)
	for i := 0; i < 2; i++ {
		uri := filepath.Join("_testdata", common.SpecsDirectoryName, fmt.Sprintf("foo%d.spec", i))
		ioutil.WriteFile(uri, []byte(fmt.Sprintf(specTextTemplate, i)), common.NewFilePermissions)			
		want = append(want, specInfo{
			Heading: fmt.Sprintf("Specification %d", i), 
			ExecutionIdentifier: filepath.Join(rootPath, common.SpecsDirectoryName, fmt.Sprintf("foo%d.spec", i)),
		})
	}

	got, err := getSpecs()

	if err != nil {
		t.Errorf("expected error to be nil. Got: \n%v", err.Error())
	}

	info := got.([]specInfo)

	if !reflect.DeepEqual(info, want) {
		t.Errorf("expected %v to be equal %v", info, want)
	}

	os.RemoveAll(filepath.Join("_testdata", common.SpecsDirectoryName))
	os.Setenv(common.GaugeProjectRootEnv, "")
}
