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

package analytics

import (
	"net/http"
	"net/http/httputil"

	"fmt"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/version"
	"github.com/jpillora/go-ogle-analytics"
)

const (
	gaTrackingID = "UA-54838477-1"
	appName      = "Gauge Core"
	medium       = "in-app"
)

// Send sends one event to ga, with category, action and label parameters.
func Send(category, action, label string) {
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
