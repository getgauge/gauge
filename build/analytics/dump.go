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
	datelayout   = "2006-01-02"                                                // date format that Core Reporting API requires
	dimensions   = "ga:eventCategory, ga:eventAction, ga:eventLabel,ga:medium" // GA dimensions that we want
	metric       = "ga:totalEvents"                                            // GA metric that we want
	outDir       = "insights"
	htmltemplate = `
{{define "table"}}
    <div class="table-container">
        <table>
            <thead>
                <tr>
                    <th>Category</th>
                    <th>Action</th>
                    <th>Label</th>
                    <th>Hits</th>
                    <th>%</th>
                </tr>
            </thead>
			<tbody>
				{{range .}}
					<tr>
						<td>{{.Category}}</td>
						<td>{{.Action}}</td>
						<td>{{.Labels}}</td>
						<td>{{.Hits}}</td>
						<td>{{printf "%.2f" .Percent}}</td>
					</tr>
				{{end}}
            </tbody>
        </table>
		<div class="actions">
			<a href="./console.csv">Download</a>
		</div>
    </div>
{{end}}
<!DOCTYPE html>
<html>
	<head>
		<title>Gauge - Insights</title>
    	<link href="https://getgauge.io/assets/images/favicons/favicon.ico" rel="shortcut icon" type="image/ico" />
		<style>
			.side{
				position: fixed;
				top:0;
				left:0;
				background-color: #343131;
				min-height: 100%;
				z-index: 200;
				width: 22em;
				font-family: "Open Sans";
			}
			.side .header{
				background: #F5C20F;
				color: #343131;
			}
			.logo{
				background: url('data:image/svg+xml; ');
			}
			.side .header .heading{
				font-size: 1.8em;
				font-weight: bold;
				font-family: monospace;
			}
			.side .header h2 {
				margin-top: 0.2em;
			}
			.side ul{
				list-style: none;
				text-align: left;
				padding:0;
			}
			.side ul li{
				padding-bottom: 0.2em;
				padding-top: 0.2em;
				color: #b3b3b3;
			}
			.side ul li:hover{
				background-color: #4e4a4a;
			}
			.side ul li.active{
				background-color: rgb(252, 252, 252);
				color: rgb(64, 64, 64);
			}
			.side ul li a{
				text-decoration: none;
				font-size: 1.2em;
				display: block;
				margin-left: 2em;
				color: inherit;
			}		
			.logo{
				height: 3em;
				margin-top: 1em;
			}
			body{
				margin:0;
				background: rgb(252, 252, 252);
				color: rgb(64, 64, 64);
				font-family: monospace;
				text-align: center;
			}
			h1{
				font-size: 2.5em;
				margin-top: 0.5em;
				color: #343131;
			}
			h2{
				font-size: 1.8em;
				font-family: "Open Sans";
			}
			.content{
				margin-left: 22em;
			}
			.table-container{
				display: inline-block;
			}
			.table-container .actions{
				margin-top: 1em;
			}
			table, th, td{
				border: 1px solid #755C07;
				border-collapse: collapse;
				font-size: 1em;
				text-align: left;
				margin-left: 0.5em;
			}
			th {
				background-color: #343131;
				color: #b3b3b3;
				text-align: center;
			}
			footer{
				margin-left: 22em;
				margin-top: 2em;
				margin-bottom: 1.5em;
			}
			section {
				min-height: 96vh;
			}
		</style>
	</head>
	<body>
		<div class="side">
			<div class="header">
				<img class="logo" src='https://docs.getgauge.io/_static/img/Gauge-Logo.svg'>
				<div class="heading">Insights</div>
			</div>
			<ul class="menu">
				<li class="active"><a href="#console">Console</a></li>
				<li><a href="#ci">CI</a></li>
				<li><a href="#api">API</a></li>
			</ul>
		</div>
		<div class="content">
			<h3>Report period: {{reportPeriod}}</h3>
			<section>
				<h2 id="console">Command Usage - Console</h2>
				{{template "table" .Console}}
			</section>
			<section>
				<h2 id="ci">Command Usage - CI</h2>
				{{template "table" .CI}}
			</section>
			<section>
				<h2 id="api">Command Usage - API</h2>
				{{template "table" .API}}
			</section>
		</div>
		<footer>Report generated on {{now}}</footer>
		<script type="text/javascript">
			(function() {
				'use strict';
				var section = document.querySelectorAll("h2");
				var sections = {};
				var i = 0;

				Array.prototype.forEach.call(section, function(e) {
					sections[e.id] = e.offsetTop;
					document.querySelector('a[href*=' + e.id + ']').addEventListener("click", function(){
						document.querySelector('.active').setAttribute('class', ' ');
						this.parentNode.setAttribute('class', 'active');
					})
				});

				window.onscroll = function() {
					var scrollPosition = document.documentElement.scrollTop || document.body.scrollTop;
					for (i in sections) {
						if (sections[i] <= scrollPosition) {
							document.querySelector('.active').setAttribute('class', ' ');
							document.querySelector('a[href*=' + i + ']').parentNode.setAttribute('class', 'active');
						}
					}
				};
			})();
		</script>
	</body>
</html>
`
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

type events struct {
	Console      []*eventSummary
	CI           []*eventSummary
	API          []*eventSummary
	apiCount     int
	ciCount      int
	consoleCount int
}

type eventSorter []*eventSummary

func (e eventSorter) Len() int           { return len(e) }
func (e eventSorter) Swap(i, j int)      { e[i], e[j] = e[j], e[i] }
func (e eventSorter) Less(i, j int) bool { return e[i].Hits > e[j].Hits }

func newEvents() *events {
	return &events{
		Console:      make([]*eventSummary, 0),
		CI:           make([]*eventSummary, 0),
		API:          make([]*eventSummary, 0),
		consoleCount: 0,
		apiCount:     0,
		ciCount:      0,
	}
}

func (e *events) populate(data *analytics.GaData) {
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
	gadata, err := gasetup.Do()
	if err != nil {
		log.Fatal(err)
	}

	events := newEvents()
	events.populate(gadata)

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
