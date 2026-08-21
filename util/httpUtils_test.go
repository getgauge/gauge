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

// setEnv sets an environment variable and returns a function restoring
// whatever was there before. gocheck's *C has neither Setenv nor Cleanup, so
// the caller defers the restore.
func setEnv(c *C, name, value string) func() {
	previous, existed := os.LookupEnv(name)
	c.Assert(os.Setenv(name, value), IsNil)
	return func() {
		if existed {
			_ = os.Setenv(name, previous)
			return
		}
		_ = os.Unsetenv(name)
	}
}

func (s *MySuite) TestGitHubTokenSentToGitHubHosts(c *C) {
	defer setEnv(c, gitHubTokenEnv, "ghp_secret")()

	for _, u := range []string{
		"https://github.com/getgauge/gauge-java/releases/download/v1.0.3/x.zip",
		"https://www.github.com/getgauge/gauge-java",
		"https://api.github.com/repos/getgauge/gauge-java/releases/latest",
		"https://codeload.github.com/getgauge/gauge-java/zip/refs/tags/v1.0.3",
		"https://GitHub.com/getgauge/gauge-java",
	} {
		c.Assert(gitHubToken(u), Equals, "ghp_secret", Commentf("url %s", u))
	}
}

func (s *MySuite) TestGitHubTokenNotSentElsewhere(c *C) {
	defer setEnv(c, gitHubTokenEnv, "ghp_secret")()

	for _, u := range []string{
		// Release assets redirect here, and the CDN rejects a request that
		// carries an Authorization header.
		"https://objects.githubusercontent.com/some/asset.zip",
		"https://example.com/gauge-java.zip",
		// A host merely ending in the GitHub name is not GitHub.
		"https://evilgithub.com/x.zip",
		"https://github.com.attacker.test/x.zip",
		"http://127.0.0.1:8080/x.zip",
		"://not a url",
	} {
		c.Assert(gitHubToken(u), Equals, "", Commentf("url %s", u))
	}
}

func (s *MySuite) TestGitHubTokenAbsentWhenUnset(c *C) {
	defer setEnv(c, gitHubTokenEnv, "")()
	defer setEnv(c, gaugeGitHubTokenEnv, "")()

	c.Assert(gitHubToken("https://github.com/getgauge/gauge-java"), Equals, "")
}

func (s *MySuite) TestGaugeGitHubTokenWinsOverGitHubToken(c *C) {
	defer setEnv(c, gitHubTokenEnv, "generic")()
	defer setEnv(c, gaugeGitHubTokenEnv, "gauge-specific")()

	c.Assert(gitHubToken("https://github.com/getgauge/gauge-java"), Equals, "gauge-specific")
}

func (s *MySuite) TestBlankGitHubTokenIsIgnored(c *C) {
	defer setEnv(c, gaugeGitHubTokenEnv, "   ")()
	defer setEnv(c, gitHubTokenEnv, "ghp_secret")()

	c.Assert(gitHubToken("https://github.com/getgauge/gauge-java"), Equals, "ghp_secret")
}

func (s *MySuite) TestDownloadSendsNoAuthorizationToNonGitHubHosts(c *C) {
	defer setEnv(c, gitHubTokenEnv, "ghp_secret")()
	err := os.Mkdir("temp-auth", 0755)
	c.Assert(err, IsNil)
	defer func() {
		_ = os.RemoveAll("temp-auth")
	}()

	var seen string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = r.Header.Get("Authorization")
		http.Error(w, "All OK", http.StatusOK)
	}))
	defer server.Close()

	_, err = Download(server.URL, "temp-auth", "asset", false)

	c.Assert(err, IsNil)
	c.Assert(seen, Equals, "")
}

func (s *MySuite) TestDownloadErrorNeverCarriesTheToken(c *C) {
	defer setEnv(c, gitHubTokenEnv, "ghp_secret")()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	_, err := Download(server.URL, ".", "", false)

	c.Assert(err, NotNil)
	c.Assert(strings.Contains(err.Error(), "ghp_secret"), Equals, false)
}
