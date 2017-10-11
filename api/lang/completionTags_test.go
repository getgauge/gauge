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

	"reflect"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

func TestGetTagsCompletion(t *testing.T) {
	provider = &dummyInfoProvider{}

	line := "tags:foo,"
	param := lsp.TextDocumentPositionParams{
		Position: lsp.Position{
			Line:      1,
			Character: 9,
		},
		TextDocument: lsp.TextDocumentIdentifier{URI: "foo.spec"},
	}
	got, err := tagsCompletion(line, line, param)
	if err != nil {
		t.Errorf("Autocomplete tags failed with error: %s", err.Error())
	}
	want := completionList{
		IsIncomplete: false,
		Items: []completionItem{
			{
				InsertTextFormat: text,
				CompletionItem: lsp.CompletionItem{
					Label:      "hello",
					FilterText: "hello",
					Detail:     tag,
					Kind:       lsp.CIKVariable,
					TextEdit: lsp.TextEdit{
						Range: lsp.Range{
							Start: lsp.Position{
								Line:      1,
								Character: 9,
							},
							End: lsp.Position{
								Line:      1,
								Character: 19,
							},
						},
						NewText: " hello"},
				},
			},
		},
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("want: %v\n but got: %v", want, got)
	}
}

func TestGetTagsCompletionWhenEditingInMiddle(t *testing.T) {
	provider = &dummyInfoProvider{}

	line := "tags:foo, bar, blah"
	pLine := "tags:foo,"
	param := lsp.TextDocumentPositionParams{
		Position: lsp.Position{
			Line:      1,
			Character: 9,
		},
		TextDocument: lsp.TextDocumentIdentifier{URI: "foo.spec"},
	}
	got, err := tagsCompletion(line, pLine, param)
	if err != nil {
		t.Errorf("Autocomplete tags failed with error: %s", err.Error())
	}
	want := completionList{
		IsIncomplete: false,
		Items: []completionItem{
			{
				InsertTextFormat: text,
				CompletionItem: lsp.CompletionItem{
					Label:      "hello",
					FilterText: "hello,",
					Detail:     tag,
					Kind:       lsp.CIKVariable,
					TextEdit: lsp.TextEdit{
						Range: lsp.Range{
							Start: lsp.Position{
								Line:      1,
								Character: 9,
							},
							End: lsp.Position{
								Line:      1,
								Character: 14,
							},
						},
						NewText: " hello,"},
				},
			},
		},
	}
	if !reflect.DeepEqual(want, got) {
		t.Errorf("want: %v\n but got: %v", want, got)
	}
}
