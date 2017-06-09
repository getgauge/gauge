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
	gaTrackingID = "UA-100778536-1"
	appName      = "Gauge Core"
	medium       = "in-app"
)

// Send sends one event to ga, with category, action and label parameters.
func send(category, action, label string) {
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

func trackManifest() {
	m, err := manifest.ProjectManifest()
	if err == nil {
		go send("manifest", "language", fmt.Sprintf("%s,%s", m.Language, strings.Join(m.Plugins, ",")))
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
	go send("execution", action, label)
}

func Validation() {
	go send("validation", "validate", "")
}

func Docs(docs string) {
	go send("docs", "generate", docs)
}

func Format() {
	go send("formatting", "format", "")
}

func Refactor() {
	go send("refactoring", "rephrase", "")
}

func ListTemplates() {
	go send("templates", "list", "")
}

func AddPlugins(plugin string) {
	go send("plugins", "add", plugin)
}

func UninstallPlugin(plugin string) {
	go send("plugins", "uninstall", plugin)
}

func ProjectInit() {
	go send("project", "init", "")
}

func Install(plugin string, zip bool) {
	if zip {
		go send("plugins", "install-zip", plugin)
	} else {
		go send("plugins", "install", plugin)
	}
}

func Update(plugin string) {
	go send("plugins", "update", plugin)
}

func UpdateAll() {
	go send("plugins", "update-all", "")
}

func InstallAll() {
	go send("plugins", "install-all", "")
}

func CheckUpdates() {
	go send("updates", "check", "")
}

func Daemon() {
	go send("execution", "daemon", "")
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
