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
	"strings"

	"sync"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

type files struct {
	cache map[string][]string
	sync.Mutex
}

func (f *files) add(uri, text string) {
	f.Lock()
	defer f.Unlock()
	f.cache[uri] = strings.Split(text, "\n")
}

func (f *files) remove(uri string) {
	f.Lock()
	defer f.Unlock()
	delete(f.cache, uri)
}

func (f *files) line(uri string, lineNo int) string {
	f.Lock()
	defer f.Unlock()
	return f.cache[uri][lineNo]
}

func (files *files) content(uri string) []string {
	f.Lock()
	defer f.Unlock()
	return f.cache[uri]
}

func (files *files) exists(uri string) bool {
	f.Lock()
	defer f.Unlock()
	_, ok := f.cache[uri]
	return ok

}

var f = &files{cache: make(map[string][]string)}

func openFile(params lsp.DidOpenTextDocumentParams) {
	f.add(params.TextDocument.URI, params.TextDocument.Text)
}

func closeFile(params lsp.DidCloseTextDocumentParams) {
	f.remove(params.TextDocument.URI)
	delete(f.cache, params.TextDocument.URI)
}

func changeFile(params lsp.DidChangeTextDocumentParams) {
	f.add(params.TextDocument.URI, params.ContentChanges[0].Text)
}

func getLine(uri string, line int) string {
	return f.line(uri, line)
}

func getContent(uri string) string {
	return strings.Join(f.content(uri), "\n")
}

func isOpen(uri string) bool {
	return f.exists(uri)
}
