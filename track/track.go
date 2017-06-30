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

package track

import (
	"net/http"
	"net/http/httputil"
	"strings"

	"fmt"

	"os"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/version"
	"github.com/jpillora/go-ogle-analytics"
)

const (
	gaTrackingID  = "UA-100778536-1"
	appName       = "Gauge Core"
	consoleMedium = "console"
	apiMedium     = "api"
	ciMedium      = "CI"
)

func send(category, action, label, medium string) {
	if !config.TelemetryEnabled() {
		return
	}
	client, err := ga.NewClient(gaTrackingID)
	client.HttpClient = &http.Client{}
	client.ClientID(config.UniqueID())
	client.AnonymizeIP(true)
	client.ApplicationName(appName)
	client.ApplicationVersion(version.FullVersion())
	client.CampaignMedium(medium)
	client.CampaignSource(appName)

	if config.TelemetryLogEnabled() {
		client.HttpClient = newlogEnabledHTTPClient()
	}

	if err != nil {
		logger.Warningf("Unable to create ga client, %s", err)
	}

	ev := ga.NewEvent(category, action)
	if label != "" {
		ev.Label(label)
	}

	err = client.Send(ev)
	if err != nil {
		logger.Warningf("Unable to send analytics data, %s", err)
	}
}

func trackConsole(category, action, label string) {
	var medium = consoleMedium
	if isCI() {
		medium = ciMedium
	}
	go send(category, action, label, medium)
}

func trackAPI(category, action, label string) {
	var medium = apiMedium
	if isCI() {
		medium = ciMedium
	}
	go send(category, action, label, medium)
}

func isCI() bool {
	// Travis, AppVeyor, CircleCI, Wercket, drone.io, gitlab-ci
	if os.Getenv("CI") == "true" {
		return true
	}

	// GoCD
	if os.Getenv("GO_SERVER_URL") != "" {
		return true
	}

	// Teamcity
	if os.Getenv("TEAMCITY_VERSION") != "" {
		return true
	}

	// TFS
	if os.Getenv("TFS_BUILD") == "true" {
		return true
	}
	return false
}

func trackManifest() {
	m, err := manifest.ProjectManifest()
	if err == nil {
		trackConsole("manifest", "language", fmt.Sprintf("%s,%s", m.Language, strings.Join(m.Plugins, ",")))
	}
}

func Execution(parallel, tagged, sorted, simpleConsole, verbose bool, parallelExecutionStrategy string) {
	action := "serial"
	if parallel {
		action = "parallel"
	}
	flags := []string{action}

	if tagged {
		flags = append(flags, "tagged")
	}

	if sorted {
		flags = append(flags, "sorted")
	}

	if simpleConsole {
		flags = append(flags, "simple-console")
	}

	if verbose {
		flags = append(flags, "verbose")
	}

	trackManifest()
	trackConsole("execution", action, strings.Join(flags, ","))
}

func Validation() {
	trackConsole("validation", "validate", "")
}

func Docs(docs string) {
	trackConsole("docs", "generate", docs)
}

func Format() {
	trackConsole("formatting", "format", "")
}

func Refactor() {
	trackConsole("refactoring", "rephrase", "")
}

func ListTemplates() {
	trackConsole("templates", "list", "")
}

func AddPlugins(plugin string) {
	trackConsole("plugins", "add", plugin)
}

func UninstallPlugin(plugin string) {
	trackConsole("plugins", "uninstall", plugin)
}

func ProjectInit() {
	trackConsole("project", "init", "")
}

func Install(plugin string, zip bool) {
	if zip {
		trackConsole("plugins", "install-zip", plugin)
	} else {
		trackConsole("plugins", "install", plugin)
	}
}

func Update(plugin string) {
	trackConsole("plugins", "update", plugin)
}

func UpdateAll() {
	trackConsole("plugins", "update-all", "")
}

func InstallAll() {
	trackConsole("plugins", "install-all", "")
}

func CheckUpdates() {
	trackConsole("updates", "check", "")
}

func Daemon() {
	trackConsole("daemon", "api", "")
}

func Lsp() {
	trackConsole("daemon", "lsp", "")
}

func APIRefactoring() {
	trackAPI("refactoring", "rephrase", "")
}

func APIExtractConcept() {
	trackAPI("refactoring", "extract-concept", "")
}

func APIFormat() {
	trackAPI("formatting", "format", "")
}

func newlogEnabledHTTPClient() *http.Client {
	return &http.Client{
		Transport: logEnabledRoundTripper{},
	}
}

type logEnabledRoundTripper struct {
}

func (r logEnabledRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		logger.Warningf("Unable to dump analytics request, %s", err)
	}

	logger.Debugf(fmt.Sprintf("%q", dump))
	return http.DefaultTransport.RoundTrip(req)
}
