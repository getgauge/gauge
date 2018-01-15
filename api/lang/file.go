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

func (file *files) add(uri, text string) {
	file.Lock()
	defer file.Unlock()
	text = strings.Replace(text, "\r\n", "\n", -1)
	file.cache[uri] = strings.Split(text, "\n")
}

func (file *files) remove(uri string) {
	file.Lock()
	defer file.Unlock()
	delete(file.cache, uri)
}

func (file *files) line(uri string, lineNo int) string {
	file.Lock()
	defer file.Unlock()
	return file.cache[uri][lineNo]
}

func (file *files) content(uri string) []string {
	file.Lock()
	defer file.Unlock()
	return file.cache[uri]
}

func (file *files) exists(uri string) bool {
	file.Lock()
	defer file.Unlock()
	_, ok := file.cache[uri]
	return ok
}

var openFilesCache = &files{cache: make(map[string][]string)}

func openFile(params lsp.DidOpenTextDocumentParams) {
	openFilesCache.add(params.TextDocument.URI, params.TextDocument.Text)
}

func closeFile(params lsp.DidCloseTextDocumentParams) {
	openFilesCache.remove(params.TextDocument.URI)
}

func changeFile(params lsp.DidChangeTextDocumentParams) {
	openFilesCache.add(params.TextDocument.URI, params.ContentChanges[0].Text)
}

func getLine(uri string, line int) string {
	return openFilesCache.line(uri, line)
}

func getContent(uri string) string {
	return strings.Join(openFilesCache.content(uri), "\n")
}

func getLineCount(uri string) int {
	return len(openFilesCache.content(uri))
}

func isOpen(uri string) bool {
	return openFilesCache.exists(uri)
}
