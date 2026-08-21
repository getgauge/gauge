/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package util

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/getgauge/gauge/logger"

	"github.com/getgauge/common"
)

// Environment variables a GitHub token is read from, in order of precedence.
// GAUGE_GITHUB_TOKEN exists so a CI job can give Gauge a token without
// widening what every other tool in the job sees through GITHUB_TOKEN.
const (
	gaugeGitHubTokenEnv = "GAUGE_GITHUB_TOKEN"
	gitHubTokenEnv      = "GITHUB_TOKEN"
)

// gitHubHosts are the hosts a token may be sent to. Deliberately narrow:
// release assets on github.com redirect to objects.githubusercontent.com,
// which rejects a request carrying an Authorization header. net/http drops
// the header on that cross-host hop by itself, and this list makes sure we
// never put it there in the first place either.
var gitHubHosts = map[string]bool{
	"github.com":          true,
	"www.github.com":      true,
	"api.github.com":      true,
	"codeload.github.com": true,
}

// gitHubToken returns the token to authenticate rawURL with, or "" when the
// URL is not a GitHub host or no token is configured.
//
// Unauthenticated GitHub traffic is rate limited hard enough that plugin
// installs fail on shared CI runners — usually as a 504 while GitHub sheds
// load, rather than an explicit 403, which is why the failure does not look
// like a rate limit.
func gitHubToken(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil || !gitHubHosts[strings.ToLower(parsed.Hostname())] {
		return ""
	}
	for _, name := range []string{gaugeGitHubTokenEnv, gitHubTokenEnv} {
		if token := strings.TrimSpace(os.Getenv(name)); token != "" {
			return token
		}
	}
	return ""
}

// progressReader is for indicating the download / upload progress on the console
type progressReader struct {
	io.Reader
	bytesTransfered   int64
	totalBytes        int64
	progress          float64
	progressDisplayed bool
}

// Read overrides the underlying io.Reader's Read method.
// io.Copy() will be calling this method.
func (w *progressReader) Read(p []byte) (int, error) {
	n, err := w.Reader.Read(p)
	if n > 0 {
		w.bytesTransfered += int64(n)
		percent := float64(w.bytesTransfered) * float64(100) / float64(w.totalBytes)
		if percent-w.progress > 4 {
			fmt.Print(".")
			w.progress = percent
			w.progressDisplayed = true
		}
	}
	return n, err
}

// Download fires a HTTP GET request to download a resource to target directory
func Download(url, targetDir, fileName string, silent bool) (string, error) {
	if !common.DirExists(targetDir) {
		return "", fmt.Errorf("Error downloading file: %s\nTarget dir %s doesn't exists.", url, targetDir)
	}

	if fileName == "" {
		fileName = filepath.Base(url)
	}
	targetFile := filepath.Join(targetDir, fileName)

	logger.Debugf(true, "Downloading %s", url)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	if token := gitHubToken(url); token != "" {
		// Never logged: the header is set on the request and nothing prints
		// req.Header. The token is not in the URL either, so the Debugf above
		// and every error below carry the URL only.
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("Error downloading file: %s.\n%s", url, resp.Status)
	}

	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	out, err := os.Create(targetFile)
	if err != nil {
		return "", err
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)
	if silent {
		_, err = io.Copy(out, resp.Body)
	} else {
		progressReader := &progressReader{Reader: resp.Body, totalBytes: resp.ContentLength}
		_, err = io.Copy(out, progressReader)
		if progressReader.progressDisplayed {
			fmt.Println()
		}
	}
	return targetFile, err
}
