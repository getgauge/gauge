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
	"strings"

	"sync"

	"github.com/getgauge/gauge/util"
	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

type files struct {
	cache map[lsp.DocumentURI][]string
	sync.Mutex
}

func (file *files) add(uri lsp.DocumentURI, text string) {
	file.Lock()
	defer file.Unlock()
	file.cache[uri] = util.GetLinesFromText(text)
}

func (file *files) remove(uri lsp.DocumentURI) {
	file.Lock()
	defer file.Unlock()
	delete(file.cache, uri)
}

func (file *files) line(uri lsp.DocumentURI, lineNo int) string {
	if !file.exists(uri) || (getLineCount(uri) <= lineNo) {
		return ""
	}
	file.Lock()
	defer file.Unlock()
	return file.cache[uri][lineNo]
}

func (file *files) content(uri lsp.DocumentURI) []string {
	file.Lock()
	defer file.Unlock()
	return file.cache[uri]
}

func (file *files) contentRange(uri lsp.DocumentURI, start, end int) []string {
	file.Lock()
	defer file.Unlock()
	return file.cache[uri][start-1 : end]
}

func (file *files) exists(uri lsp.DocumentURI) bool {
	file.Lock()
	defer file.Unlock()
	_, ok := file.cache[uri]
	return ok
}

var openFilesCache = &files{cache: make(map[lsp.DocumentURI][]string)}

func openFile(params lsp.DidOpenTextDocumentParams) {
	openFilesCache.add(params.TextDocument.URI, params.TextDocument.Text)
}

func closeFile(params lsp.DidCloseTextDocumentParams) {
	openFilesCache.remove(params.TextDocument.URI)
}

func changeFile(params lsp.DidChangeTextDocumentParams) {
	openFilesCache.add(params.TextDocument.URI, params.ContentChanges[0].Text)
}

func getLine(uri lsp.DocumentURI, line int) string {
	return openFilesCache.line(uri, line)
}

func getContent(uri lsp.DocumentURI) string {
	return strings.Join(openFilesCache.content(uri), "\n")
}

func getLineCount(uri lsp.DocumentURI) int {
	return len(openFilesCache.content(uri))
}

func isOpen(uri lsp.DocumentURI) bool {
	return openFilesCache.exists(uri)
}

func getContentRange(uri lsp.DocumentURI, start, end int) []string {
	return openFilesCache.contentRange(uri, start, end)
}
