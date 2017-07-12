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

	"strings"

	"sync"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
	"github.com/sourcegraph/jsonrpc2"
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

func (f *files) update(uri, text string) {
	f.add(uri, text)
}

func (f *files) char(uri string, line, char int) (asciiCode byte) {
	f.Lock()
	defer f.Unlock()
	return f.cache[uri][line][char]
}

var f = &files{cache: make(map[string][]string)}

func openFile(req *jsonrpc2.Request) {
	var params lsp.DidOpenTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return
	}
	f.add(params.TextDocument.URI, params.TextDocument.Text)
}

func closeFile(req *jsonrpc2.Request) {
	var params lsp.DidCloseTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return
	}
	f.remove(params.TextDocument.URI)
	delete(f.cache, params.TextDocument.URI)
}

func changeFile(req *jsonrpc2.Request) {
	var params lsp.DidChangeTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		return
	}
	f.add(params.TextDocument.URI, params.ContentChanges[0].Text)
}

func getChar(uri string, line, char int) (asciiCode byte) {
	return f.char(uri, line, char)
}
