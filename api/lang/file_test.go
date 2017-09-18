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
)

func TestConvertURItoWindowsFilePath(t *testing.T) {
	uri := `file:///c%3A/Users/gauge/project/example.spec`
	want := `c:\Users\gauge\project\example.spec`
	got := convertURIToWindowsPath(uri)
	if want != got {
		t.Errorf("got : %s, want : %s", got, want)
	}
}

func TestConvertURItoUnixFilePath(t *testing.T) {
	uri := `file:///Users/gauge/project/example.spec`
	want := `/Users/gauge/project/example.spec`
	got := convertURIToUnixPath(uri)
	if want != got {
		t.Errorf("got : %s, want : %s", got, want)
	}
}
