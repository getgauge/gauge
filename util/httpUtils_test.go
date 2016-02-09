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

	_, err := Download(server.URL, ".")

	c.Assert(err, NotNil)
}

func (s *MySuite) TestDownloadFailureIfServerError(c *C) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	_, err := Download(server.URL, ".")

	c.Assert(err, NotNil)
}

func (s *MySuite) TestDownloadFailureIfSomeHTTPError(c *C) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	_, err := Download(server.URL, ".")

	c.Assert(err, NotNil)
}

func (s *MySuite) TestDownloadFailureIfTargetDirDoesntExist(c *C) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "All OK", http.StatusOK)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	_, err := Download(server.URL, "/foo/bar")
	errMsg := fmt.Sprintf("Error downloading file: %s\nTarget dir /foo/bar doesn't exists.", server.URL)

	c.Assert(err, NotNil)
	c.Assert(err.Error(), Equals, errMsg)
}

func (s *MySuite) TestDownloadSuccess(c *C) {
	os.Mkdir("temp", 0755)
	defer os.RemoveAll("temp")

	handler := func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "All OK", http.StatusOK)
	}

	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	actualDownloadedFilePath, err := Download(server.URL, "temp")
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
