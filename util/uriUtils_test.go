/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package util

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

func TestConvertWindowsFilePathToURI(t *testing.T) {
	path := `c:\Users\gauge\project\example.spec`
	want := `file:///c%3A/Users/gauge/project/example.spec`
	got := convertWindowsPathToURI(path)
	if want != got {
		t.Errorf("got : %s, want : %s", got, want)
	}
}

func TestConvertUnixFilePathToURI(t *testing.T) {
	path := `/Users/gauge/project/example.spec`
	want := `file:///Users/gauge/project/example.spec`
	got := convertUnixPathToURI(path)
	if want != got {
		t.Errorf("got : %s, want : %s", got, want)
	}
}
