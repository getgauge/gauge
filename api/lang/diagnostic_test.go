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

func TestDiagnostic(t *testing.T) {
	specText := `Specification Heading
=====================

Scenario Heading
================

* Step text`

	uri := "foo.spec"

	f = &files{cache: make(map[string][]string)}
	f.add(uri, specText)

	want := []lsp.Diagnostic{
		{
			Range: lsp.Range{
				Start: lsp.Position{0, 0},
				End:   lsp.Position{0, 21},
			},
			Message:  "Spec should have atleast one scenario",
			Severity: 1,
		},
		{
			Range: lsp.Range{
				Start: lsp.Position{3, 0},
				End:   lsp.Position{3, 16},
			},
			Message:  "Multiple spec headings found in same file",
			Severity: 1,
		},
	}

	got := createDiagnostics(uri)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, got)
	}
}

func TestDiagnosticWithoughtParseError(t *testing.T) {
	specText := `Specification Heading
=====================

Scenario Heading
----------------

* Step text`

	uri := "foo.spec"

	f = &files{cache: make(map[string][]string)}
	f.add(uri, specText)

	want := []lsp.Diagnostic{}

	got := createDiagnostics(uri)

	if !reflect.DeepEqual(got, want) {
		t.Errorf("want: `%s`,\n got: `%s`", want, got)
	}
}
