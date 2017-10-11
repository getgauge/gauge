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

package util

import (
	"strings"
)

const (
	uriPrefix      = "file://"
	unixSep        = "/"
	windowColonRep = "%3A"
	colon          = ":"
	windowsSep     = "\\"
)

// ConvertURItoFilePath - converts file uri (eg: file://example.spec) to OS specific file paths.
func ConvertURItoFilePath(uri string) string {
	if IsWindows() {
		return convertURIToWindowsPath(uri)
	}
	return convertURIToUnixPath(uri)
}

func convertURIToWindowsPath(uri string) string {
	uri = strings.TrimPrefix(uri, uriPrefix+unixSep)
	uri = strings.Replace(uri, windowColonRep, colon, -1)
	return strings.Replace(uri, unixSep, windowsSep, -1)
}

func convertURIToUnixPath(uri string) string {
	return strings.TrimPrefix(uri, uriPrefix)
}

// ConvertPathToURI - converts OS specific file paths to file uri (eg: file://example.spec).
func ConvertPathToURI(path string) string {
	if IsWindows() {
		return convertWindowsPathToURI(path)
	}
	return convertUnixPathToURI(path)
}

func convertWindowsPathToURI(path string) string {
	path = strings.Replace(path, colon, windowColonRep, -1)
	return uriPrefix + strings.Replace(path, windowsSep, unixSep, -1)
}

func convertUnixPathToURI(path string) string {
	return uriPrefix + path
}
