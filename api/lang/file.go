/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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
	if !file.exists(uri) || len(file.content(uri)) <= lineNo {
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

func isOpen(uri lsp.DocumentURI) bool {
	return openFilesCache.exists(uri)
}
