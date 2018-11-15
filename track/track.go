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
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/getgauge/gauge/env"

	"fmt"

	"os"

	"sync"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/version"
	"github.com/jpillora/go-ogle-analytics"
)

const (
	gaTrackingID     = "UA-54838477-1"
	gaTestTrackingID = "UA-100778536-1"
	appName          = "Gauge Core"
	consoleMedium    = "console"
	apiMedium        = "api"
	ciMedium         = "CI"
	timeout          = 1
	// GaugeTelemetryMessageHeading is the header printed for telemetry warning
	GaugeTelemetryMessageHeading = `
Telemetry
---------
`
	// GaugeTelemetryMessage is the message printed when user has not explicitly opted in/out
	// of telemetry. Printed only in CLI.
	GaugeTelemetryMessage = `This installation of Gauge collects usage data in order to help us improve your experience.
The data is anonymous and doesn't include command-line arguments.
To turn this message off opt in or out by running 'gauge telemetry on' or 'gauge telemetry off'.

Read more about Gauge telemetry at https://gauge.org/telemetry
`
	// GaugeTelemetryMachineRedableMessage is the message printed when user has not explicitly opted in/out
	// of telemetry. Printed only in CLI.
	GaugeTelemetryMachineRedableMessage = `This installation of Gauge collects usage data in order to help us improve your experience.
<a href="https://gauge.org/telemetry">Read more here</a> about Gauge telemetry.`

	// GaugeTelemetryLSPMessage is the message printed when user has not explicitly opted in/out
	// of telemetry. Displayed only in LSP Client.
	GaugeTelemetryLSPMessage = `This installation of Gauge collects usage data in order to help us improve your experience.
[Read more here](https://gauge.org/telemetry) about Gauge telemetry.
Would you like to participate?`
)

var gaHTTPTransport = http.DefaultTransport

var telemetryEnabled, telemetryLogEnabled bool

func Init() {
	telemetryEnabled = config.TelemetryEnabled()
	telemetryLogEnabled = config.TelemetryLogEnabled()
}

func send(category, action, label, medium string, wg *sync.WaitGroup) bool {
	if !telemetryEnabled {
		wg.Done()
		return false
	}
	label = strings.Trim(fmt.Sprintf("%s,%s", label, runtime.GOOS), ",")
	sendChan := make(chan bool, 1)
	go func(c chan<- bool) {
		defer recoverPanic()
		t := gaTrackingID
		if env.UseTestGA() {
			t = gaTestTrackingID
		}
		client, err := ga.NewClient(t)
		if err != nil {
			logger.Debugf(true, "Unable to create ga client, %s", err)
		}
		client.HttpClient = &http.Client{}
		client.ClientID(config.UniqueID())
		client.AnonymizeIP(true)
		client.ApplicationName(appName)
		client.ApplicationVersion(version.FullVersion())
		client.CampaignMedium(medium)
		client.CampaignSource(appName)
		client.HttpClient.Transport = gaHTTPTransport
		if telemetryLogEnabled {
			client.HttpClient.Transport = newlogEnabledHTTPTransport()
		}
		ev := ga.NewEvent(category, action)
		if label != "" {
			ev.Label(label)
		}
		err = client.Send(ev)
		if err != nil {
			logger.Debugf(true, "Unable to send analytics data, %s", err)
		}
		c <- true
	}(sendChan)

	for {
		select {
		case <-sendChan:
			wg.Done()
			return true
		case <-time.After(timeout * time.Second):
			logger.Debugf(true, "Unable to send analytics data, timed out")
			wg.Done()
			return false
		}
	}
}

func recoverPanic() {
	if r := recover(); r != nil {
		logger.Errorf(true, "%v\n%s", r, string(debug.Stack()))
	}
}

func track(medium, category, action, label string){
	if isCI() {
		medium = ciMedium
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()
	go send(category, action, label, medium, wg)
}

func trackConsole(category, action, label string) {
	track(consoleMedium, category, action, label)
}

func trackAPI(category, action, label string) {
	track(apiMedium, category, action, label)
}

func isCI() bool {
	// Travis, AppVeyor, CircleCI, Wercket, drone.io, gitlab-ci
	if ci, _ := strconv.ParseBool(os.Getenv("CI")); ci {
		return true
	}

	// GoCD
	if os.Getenv("GO_SERVER_URL") != "" {
		return true
	}

	// Jenkins
	if os.Getenv("JENKINS_URL") != "" {
		return true
	}

	// Teamcity
	if os.Getenv("TEAMCITY_VERSION") != "" {
		return true
	}

	// TFS
	if ci, _ := strconv.ParseBool(os.Getenv("TFS_BUILD")); ci {
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

func Execution(parallel, tagged, sorted, simpleConsole, verbose, hideSuggestion bool, parallelExecutionStrategy string) {
	action := "serial"
	if parallel {
		action = "parallel"
	}
	flags := []string{}

	if tagged {
		flags = append(flags, "tagged")
	}

	if sorted {
		flags = append(flags, "sorted")
	}

	if simpleConsole {
		flags = append(flags, "simple-console")
	} else {
		flags = append(flags, "rich-console")
	}

	if verbose {
		flags = append(flags, "verbose")
	}
	if hideSuggestion {
		flags = append(flags, "hide-suggestion")
	}

	trackManifest()
	trackConsole("execution", action, strings.Join(flags, ","))
}

func Validation(hideSuggestion bool) {
	if hideSuggestion {
		trackConsole("validation", "validate", "hide-suggestion")
	} else {
		trackConsole("validation", "validate", "")
	}
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

func ListScenarios() {
	trackConsole("list", "scenarios", "")
}

func ListTags() {
	trackConsole("list", "tags", "")
}

func ListSpecifications() {
	trackConsole("list", "specifications", "")
}

func ListTemplates() {
	trackConsole("init", "templates", "")
}

func UninstallPlugin(plugin string) {
	trackConsole("plugins", "uninstall", plugin)
}

func ProjectInit(lang string) {
	trackConsole("project", "init", lang)
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
	trackConsole("plugins", "update-all", "all")
}

func InstallAll() {
	trackConsole("plugins", "install-all", "all")
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

func newlogEnabledHTTPTransport() http.RoundTripper {
	return &logEnabledRoundTripper{}
}

type logEnabledRoundTripper struct {
}

func (r logEnabledRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		logger.Debugf(true, "Unable to dump analytics request, %s", err)
	}

	logger.Debugf(true, fmt.Sprintf("%q", dump))
	return http.DefaultTransport.RoundTrip(req)
}
