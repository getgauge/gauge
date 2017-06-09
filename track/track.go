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

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/version"
	"github.com/jpillora/go-ogle-analytics"
)

const (
	gaTrackingID  = "UA-54838477-1"
	appName       = "Gauge Core"
	consoleMedium = "console"
	apiMedium     = "api"
)

func send(category, action, label, medium string) {
	if !config.AnalyticsEnabled() {
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

	if config.AnalyticsLogEnabled() {
		client.HttpClient = newlogEnabledHTTPClient()
	}

	if err != nil {
		logger.Warning("Unable to create ga client, %s", err)
	}

	ev := ga.NewEvent(category, action)
	if label != "" {
		ev.Label(label)
	}

	err = client.Send(ev)
	if err != nil {
		logger.Warning("Unable to send analytics data, %s", err)
	}
}

func trackConsole(category, action, label string) {
	go send(category, action, label, consoleMedium)
}

func trackAPI(category, action, label string) {
	go send(category, action, label, apiMedium)
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

	label := ""
	if tagged {
		label = "tagged,"
	}

	if sorted {
		label += "sorted,"
	}

	if simpleConsole {
		label += "simple-console,"
	}

	if verbose {
		label += "verbose"
	}

	trackManifest()
	trackConsole("execution", action, label)
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
	trackConsole("execution", "daemon", "")
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
		logger.Warning("Unable to dump analytics request, %s", err)
	}

	logger.Debug(fmt.Sprintf("%q", dump))
	return http.DefaultTransport.RoundTrip(req)
}
