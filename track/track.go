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
	"github.com/getgauge/gauge/version"
	ga "github.com/jpillora/go-ogle-analytics"
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

// Init sets flags used by the package methods.
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

func trackConsole(category, action, label string) {
	var medium = consoleMedium
	if isCI() {
		medium = ciMedium
	}
	wg := &sync.WaitGroup{}
	wg.Add(1)
	defer wg.Wait()
	go send(category, action, label, medium, wg)
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

func daemon(mode, lang string) {
	trackConsole("daemon", mode, lang)
}

// ScheduleDaemonTracking sends pings to GA at regular intervals. This is used to flag active usage.
func ScheduleDaemonTracking(mode, lang string) {
	daemon(mode, lang)
	ticker := time.NewTicker(28 * time.Minute)
	if env.UseTestGA() && env.TelemetryInterval() != "" {
		duration, _ := strconv.Atoi(env.TelemetryInterval())
		ticker = time.NewTicker(time.Duration(duration) * time.Minute)
	}
	for {
		<-ticker.C
		daemon(mode, lang)
	}
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
