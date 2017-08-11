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

package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"html/template"
	"log"
	"os"
	"strconv"
	"time"

	analytics "google.golang.org/api/analytics/v3"

	"golang.org/x/oauth2/jwt"

	"path/filepath"
	"sort"
	"strings"
)

const (
	datelayout            = "2006-01-02"
	dimensions            = "ga:eventCategory, ga:eventAction, ga:eventLabel,ga:medium"
	metric                = "ga:totalEvents"
	demographicDimensions = "ga:countryIsoCode"
	demographicMetric     = "ga:users"
	outDir                = "_insights"
)

var (
	enddate            = time.Now().Format(datelayout)
	startdate          = time.Now().Add(time.Hour * 24 * -30).Format(datelayout)
	tokenurl           = os.Getenv("GA_TOKEN_URL")             // (json:"token_uri") Google oauth2 Token URL
	gaServiceAcctEmail = os.Getenv("GA_SERVICE_ACCOUNT_EMAIL") // (json:"client_email") email address of registered application
	gaServiceAcctPEM   = os.Getenv("GA_PRIVATE_KEY")           // private key string (PEM format) from Google Cloud Console
	gaTableID          = os.Getenv("GA_VIEW_ID")               // namespaced profile (table) ID of your analytics account/property/profile
)

type eventSummary struct {
	Category string
	Action   string
	Labels   string
	Hits     int
	Percent  float64
}

func (e *eventSummary) toSlice() []string {
	return []string{e.Category, e.Action, e.Labels, strings.Replace(e.Labels, ",", " ", -1), fmt.Sprintf("%.2f", e.Percent)}
}

type countryWiseUser struct {
	Country   country
	UserCount int
	Radius    int
}

type events struct {
	Console          []*eventSummary
	CI               []*eventSummary
	API              []*eventSummary
	apiCount         int
	ciCount          int
	consoleCount     int
	CountryWiseUsers []*countryWiseUser
}

type eventSorter []*eventSummary

func (e eventSorter) Len() int           { return len(e) }
func (e eventSorter) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e eventSorter) Less(i, j int) bool { return e[i].Hits > e[j].Hits }

func newEvents() *events {
	return &events{
		Console:          make([]*eventSummary, 0),
		CI:               make([]*eventSummary, 0),
		API:              make([]*eventSummary, 0),
		consoleCount:     0,
		apiCount:         0,
		ciCount:          0,
		CountryWiseUsers: make([]*countryWiseUser, 0),
	}
}

func (e *events) populate(data, countryData *analytics.GaData) {
	for _, row := range data.Rows {
		h, _ := strconv.Atoi(row[4])
		switch row[3] {
		case "console":
			e.Console = append(e.Console, &eventSummary{Category: row[0], Action: row[1], Labels: row[2], Hits: h})
			e.consoleCount += h
			break
		case "CI":
			e.CI = append(e.CI, &eventSummary{Category: row[0], Action: row[1], Labels: row[2], Hits: h})
			e.ciCount += h
			break
		case "api":
			e.API = append(e.API, &eventSummary{Category: row[0], Action: row[1], Labels: row[2], Hits: h})
			e.apiCount += h
			break
		}
	}

	for _, ev := range e.Console {
		ev.Percent = float64(ev.Hits) / float64(e.consoleCount) * 100
	}
	for _, ev := range e.CI {
		ev.Percent = float64(ev.Hits) / float64(e.ciCount) * 100
	}
	for _, ev := range e.API {
		ev.Percent = float64(ev.Hits) / float64(e.apiCount) * 100
	}

	sort.Sort(eventSorter(e.API))
	sort.Sort(eventSorter(e.Console))
	sort.Sort(eventSorter(e.CI))
	var totalUsers = 0
	for _, c := range countryData.Rows {
		country, ok := countries()[c[0]]
		if ok {
			count, err := strconv.Atoi(c[1])
			if err == nil {
				e.CountryWiseUsers = append(e.CountryWiseUsers, &countryWiseUser{Country: country, UserCount: count})
				totalUsers += count
			}
		}
	}

	for _, c := range e.CountryWiseUsers {
		c.Radius = int(c.UserCount * 100 / totalUsers)
	}
}

func (e *events) writeHTML(dest string) {
	funcMap := template.FuncMap{
		"now": func() string {
			return time.Now().Format(datelayout)
		},
		"reportPeriod": func() string {
			return fmt.Sprintf("%s - %s", startdate, enddate)
		},
	}
	tmpl, err := template.New("html").Funcs(funcMap).Parse(htmltemplate)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.OpenFile(filepath.Join(dest, "index.html"), os.O_CREATE|os.O_WRONLY, 0777)
	defer f.Close()
	if err != nil {
		log.Fatal(err)
	}

	err = tmpl.Execute(f, e)
	if err != nil {
		log.Fatal(err)
	}
}

func (e *events) writeCSV(dest string) {
	var write = func(medium string, data []*eventSummary) {
		f, err := os.OpenFile(filepath.Join(dest, fmt.Sprintf("%s.csv", medium)), os.O_CREATE|os.O_WRONLY, 0777)
		defer f.Close()
		if err != nil {
			log.Fatal(err)
		}

		w := csv.NewWriter(f)
		for _, record := range data {
			if err := w.Write(record.toSlice()); err != nil {
				log.Fatalln("error writing event to csv:", err)
			}
		}
		w.Flush()
		if err := w.Error(); err != nil {
			log.Fatalln("error writing csv:", err)
		}
	}
	write("api", e.API)
	write("console", e.Console)
	write("ci", e.CI)
}

func main() {
	jwtc := jwt.Config{
		Email:      gaServiceAcctEmail,
		PrivateKey: []byte(gaServiceAcctPEM),
		Scopes:     []string{analytics.AnalyticsReadonlyScope},
		TokenURL:   tokenurl,
	}
	clt := jwtc.Client(context.Background())
	as, err := analytics.New(clt)
	if err != nil {
		log.Fatal("Error creating Analytics Service", err)
	}
	ads := analytics.NewDataGaService(as)
	gasetup := ads.Get(gaTableID, startdate, enddate, metric).Dimensions(dimensions)
	eventData, err := gasetup.Do()
	if err != nil {
		log.Fatal(err)
	}
	gasetup = ads.Get(gaTableID, startdate, enddate, demographicMetric).Dimensions(demographicDimensions)
	usersPerCountry, err := gasetup.Do()
	if err != nil {
		log.Fatal(err)
	}
	events := newEvents()
	events.populate(eventData, usersPerCountry)

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	dest := filepath.Join(pwd, outDir)
	err = os.MkdirAll(dest, 0755)
	if err != nil {
		log.Fatal(err)
	}
	events.writeHTML(dest)
	events.writeCSV(dest)
}
