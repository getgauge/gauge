// Copyright 2018 ThoughtWorks, Inc.

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
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
)

func TestExtractToConcept(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("failed to set projectRoot. %v", err.Error())
	}
	config.ProjectRoot = filepath.Join(cwd, "_testdata")

	specFile = filepath.Join(config.ProjectRoot, "foo.spec")

	params := extractConceptInfo{
		ConceptFile: "New File",
		ConceptName: "new concept",
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      6,
				Character: 0,
			},
			End: lsp.Position{
				Line:      7,
				Character: 0,
			},
		},
		Uri: lsp.DocumentURI(specFile),
	}
	b, _ := json.Marshal(params)
	p := json.RawMessage(b)
	request := &jsonrpc2.Request{Params: &p}

	specText, _ := common.ReadFileContents(specFile)

	openFilesCache.add(lsp.DocumentURI(specFile), specText)

	expected := lsp.WorkspaceEdit{
		Changes: map[string][]lsp.TextEdit{
			specFile: []lsp.TextEdit{
				lsp.TextEdit{
					Range: lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 0,
						},
						End: lsp.Position{
							Line:      15,
							Character: 0,
						},
					},
					NewText: `Specification
=============

Scenario
--------

* new concept
* some step with

   |header|
   |------|
   |value |

* one more step`,
				},
			},
			"New File": []lsp.TextEdit{
				lsp.TextEdit{
					Range: lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 0,
						},
						End: lsp.Position{
							Line:      4,
							Character: 0,
						},
					},
					NewText: `# new concept
* some step
* some step
`,
				},
			},
		},
	}

	edits, err := extractConcept(request)

	if err != nil {
		t.Errorf("expected error to be nil. but got %v", err.Error())
	}

	if !reflect.DeepEqual(edits, expected) {
		t.Errorf("\n\nExpected: %vGot: %v", expected, edits)
	}
}

func TestExtractToConceptWithTable(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("failed to set projectRoot. %v", err.Error())
	}
	config.ProjectRoot = filepath.Join(cwd, "_testdata")

	specFile = filepath.Join(config.ProjectRoot, "foo.spec")

	params := extractConceptInfo{
		ConceptFile: "New File",
		ConceptName: "new concept",
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      7,
				Character: 0,
			},
			End: lsp.Position{
				Line:      13,
				Character: 0,
			},
		},
		Uri: lsp.DocumentURI(specFile),
	}
	b, _ := json.Marshal(params)
	p := json.RawMessage(b)
	request := &jsonrpc2.Request{Params: &p}

	specText, _ := common.ReadFileContents(specFile)

	openFilesCache.add(lsp.DocumentURI(specFile), specText)

	expected := lsp.WorkspaceEdit{
		Changes: map[string][]lsp.TextEdit{
			specFile: []lsp.TextEdit{
				lsp.TextEdit{
					Range: lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 0,
						},
						End: lsp.Position{
							Line:      15,
							Character: 0,
						},
					},
					NewText: `Specification
=============

Scenario
--------

* some step
* new concept
* one more step`,
				},
			},
			"New File": []lsp.TextEdit{
				lsp.TextEdit{
					Range: lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 0,
						},
						End: lsp.Position{
							Line:      8,
							Character: 0,
						},
					},
					NewText: "# new concept\n* some step\n* some step with \n\n   |header|\n   |------|\n   |value |\n",
				},
			},
		},
	}

	edits, err := extractConcept(request)

	if err != nil {
		t.Errorf("expected error to be nil. but got %v", err.Error())
	}

	if !reflect.DeepEqual(edits, expected) {
		t.Errorf("\n\nExpected: %vGot: %v", expected, edits)
	}
}

func TestExtractToConceptWithWithInvalidElements(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("failed to set projectRoot. %v", err.Error())
	}
	config.ProjectRoot = filepath.Join(cwd, "_testdata")

	specFile = filepath.Join(config.ProjectRoot, "foo.spec")

	params := extractConceptInfo{
		ConceptFile: "New File",
		ConceptName: "new concept",
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      1,
				Character: 0,
			},
			End: lsp.Position{
				Line:      3,
				Character: 0,
			},
		},
		Uri: lsp.DocumentURI(specFile),
	}
	b, _ := json.Marshal(params)
	p := json.RawMessage(b)
	request := &jsonrpc2.Request{Params: &p}

	specText, _ := common.ReadFileContents(specFile)

	openFilesCache.add(lsp.DocumentURI(specFile), specText)

	expectedError := "Can not extract to cencpet. Selected text contains invalid elements."

	_, err = extractConcept(request)

	if err == nil {
		t.Errorf("expected error but got nil")
	} else if err.Error() != expectedError {
		t.Errorf("\n\nExpected: %vGot: %v", expectedError, err.Error())
	}
}

func TestExtractToConceptWithInAExistingConceptFile(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("failed to set projectRoot. %v", err.Error())
	}
	config.ProjectRoot = filepath.Join(cwd, "_testdata")

	specFile = filepath.Join(config.ProjectRoot, "foo.spec")
	cptFile := filepath.Join(config.ProjectRoot, "some.cpt")

	params := extractConceptInfo{
		ConceptFile: lsp.DocumentURI(cptFile),
		ConceptName: "new concept",
		Range: lsp.Range{
			Start: lsp.Position{
				Line:      6,
				Character: 0,
			},
			End: lsp.Position{
				Line:      7,
				Character: 0,
			},
		},
		Uri: lsp.DocumentURI(specFile),
	}

	b, _ := json.Marshal(params)
	p := json.RawMessage(b)
	request := &jsonrpc2.Request{Params: &p}

	specText, _ := common.ReadFileContents(specFile)

	openFilesCache.add(lsp.DocumentURI(specFile), specText)

	expected := lsp.WorkspaceEdit{
		Changes: map[string][]lsp.TextEdit{
			specFile: []lsp.TextEdit{
				lsp.TextEdit{
					Range: lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 0,
						},
						End: lsp.Position{
							Line:      15,
							Character: 0,
						},
					},
					NewText: `Specification
=============

Scenario
--------

* new concept
* some step with

   |header|
   |------|
   |value |

* one more step`,
				},
			},
			cptFile: []lsp.TextEdit{
				lsp.TextEdit{
					Range: lsp.Range{
						Start: lsp.Position{
							Line:      0,
							Character: 0,
						},
						End: lsp.Position{
							Line:      7,
							Character: 0,
						},
					},
					NewText: "# concept heading\n* with a step\n\n# new concept\n* some step\n* some step\n",
				},
			},
		},
	}

	edits, err := extractConcept(request)

	if err != nil {
		t.Errorf("expected error to be nil. but got %v", err.Error())
	}

	if !reflect.DeepEqual(edits, expected) {
		t.Errorf("\n\nExpected: %vGot: %v", expected, edits)
	}
}
