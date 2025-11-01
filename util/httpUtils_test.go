/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package util

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	"github.com/getgauge/common"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestDownloadFailureIfFileNotFound(c *C) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	_, err := Download(server.URL, ".", "", false)

	c.Assert(err, NotNil)
}

func (s *MySuite) TestDownloadFailureIfServerError(c *C) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	_, err := Download(server.URL, ".", "", false)

	c.Assert(err, NotNil)
}

func (s *MySuite) TestDownloadFailureIfSomeHTTPError(c *C) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	_, err := Download(server.URL, ".", "", false)

	c.Assert(err, NotNil)
}

func (s *MySuite) TestDownloadFailureIfTargetDirDoesntExist(c *C) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "All OK", http.StatusOK)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	_, err := Download(server.URL, "/foo/bar", "", false)
	errMsg := fmt.Sprintf("Error downloading file: %s\nTarget dir /foo/bar doesn't exists.", server.URL)

	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, errMsg)
}

func (s *MySuite) TestDownloadSuccess(c *C) {
	err := os.Mkdir("temp", 0755)
	c.Assert(err, IsNil)
	defer func() {
		_ = os.RemoveAll("temp")
	}()

	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "All OK", http.StatusOK)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	actualDownloadedFilePath, err := Download(server.URL, "temp", "", false)
	expectedDownloadFilePath := filepath.Join("temp", strings.TrimPrefix(server.URL, "http://"))
	absoluteDownloadFilePath, _ := filepath.Abs(expectedDownloadFilePath)
	expectedFileContents := "All OK\n"

	c.Assert(err, Equals, nil)
	c.Assert(actualDownloadedFilePath, Equals, expectedDownloadFilePath)
	c.Assert(common.FileExists(absoluteDownloadFilePath), Equals, true)

	actualFileContents, err := common.ReadFileContents(absoluteDownloadFilePath)
	c.Assert(err, Equals, nil)
	c.Assert(actualFileContents, Equals, expectedFileContents)
}
