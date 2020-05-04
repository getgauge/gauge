/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package util

import (
	"net/url"
	"strings"

	"github.com/sourcegraph/go-langserver/pkg/lsp"
)

const (
	uriPrefix      = "file://"
	UnixSep        = "/"
	windowColonRep = "%3A"
	colon          = ":"
	WindowsSep     = "\\"
)

// ConvertURItoFilePath - converts file uri (eg: file://example.spec) to OS specific file paths.
func ConvertURItoFilePath(uri lsp.DocumentURI) string {
	if IsWindows() {
		return convertURIToWindowsPath(string(uri))
	}
	return convertURIToUnixPath(string(uri))
}

func convertURIToWindowsPath(uri string) string {
	uri = strings.TrimPrefix(uri, uriPrefix+UnixSep)
	uri = strings.Replace(uri, windowColonRep, colon, -1)
	path, _ := url.PathUnescape(strings.Replace(uri, UnixSep, WindowsSep, -1))
	return path
}

func convertURIToUnixPath(uri string) string {
	path, _ := url.PathUnescape(uri)
	return strings.TrimPrefix(path, uriPrefix)
}

// ConvertPathToURI - converts OS specific file paths to file uri (eg: file://example.spec).
func ConvertPathToURI(path string) lsp.DocumentURI {
	if IsWindows() {
		return lsp.DocumentURI(convertWindowsPathToURI(path))
	}
	return lsp.DocumentURI(convertUnixPathToURI(path))
}

func convertWindowsPathToURI(path string) string {
	path = strings.Replace(path, WindowsSep, UnixSep, -1)
	encodedPath := url.URL{Path: path}
	path = strings.Replace(strings.TrimPrefix(encodedPath.String(), "./"), colon, windowColonRep, -1)
	return uriPrefix + UnixSep + path
}

func convertUnixPathToURI(path string) string {
	encodedPath := url.URL{Path: path}
	return uriPrefix + encodedPath.String()
}
