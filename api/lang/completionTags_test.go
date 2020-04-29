// Copyright 2015 ThoughtWorks, Inc.
/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package lang

import (
	"testing"

	"reflect"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

func TestGetTagsCompletion(t *testing.T) {
	lineNumber := 1
	provider = &dummyInfoProvider{}

	line := "tags:foo,"
	pLine := "tags:foo,"
	param := lsp.TextDocumentPositionParams{
		Position: lsp.Position{
			Line:      lineNumber,
			Character: len("tags:foo,"),
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
					SortText:   "ahello",
					Label:      "hello",
					FilterText: "hello",
					Detail:     tag,
					Kind:       lsp.CIKVariable,
					TextEdit: &lsp.TextEdit{
						Range: lsp.Range{
							Start: lsp.Position{
								Line:      lineNumber,
								Character: len("tags:foo,"),
							},
							End: lsp.Position{
								Line:      lineNumber,
								Character: len("tags:foo,"),
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
	lineNumber := 1
	provider = &dummyInfoProvider{}

	line := "tags:foo, bar, blah"
	pLine := "tags:foo,"
	param := lsp.TextDocumentPositionParams{
		Position: lsp.Position{
			Line:      lineNumber,
			Character: len("tags:foo,"),
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
					SortText:   "ahello",
					Label:      "hello",
					FilterText: "hello,",
					Detail:     tag,
					Kind:       lsp.CIKVariable,
					TextEdit: &lsp.TextEdit{
						Range: lsp.Range{
							Start: lsp.Position{
								Line:      lineNumber,
								Character: len("tags:foo,"),
							},
							End: lsp.Position{
								Line:      lineNumber,
								Character: len("tags:foo, bar,"),
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
