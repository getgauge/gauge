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

const htmltemplate = `
{{$maxRows := 25}}
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
    </div>
{{end}}
<!DOCTYPE html>
<html>
	<head>
		<title>Gauge - Insights</title>
    	<link href="https://getgauge.io/assets/images/favicons/favicon.ico" rel="shortcut icon" type="image/ico" />
		<style>
			body,
			html {
				font-family: monospace;
				text-align: center;
				margin: 0;
				overflow: hidden;
				-webkit-transition: opacity 400ms;
				-moz-transition: opacity 400ms;
				transition: opacity 400ms;
			}

			body,
			.onepage-wrapper,
			html {
				display: block;
				position: static;
				padding: 0;
				width: 100%;
				height: 100%;
			}

			.onepage-wrapper {
				width: 100%;
				height: 100%;
				display: block;
				position: relative;
				padding: 0;
				-webkit-transform-style: preserve-3d;
			}

			.onepage-wrapper .ops-section {
				width: 100%;
				height: 100%;
				position: relative;
			}

			.onepage-pagination {
				position: absolute;
				right: 10px;
				top: 50%;
				z-index: 5;
				list-style: none;
				margin: 0;
				padding: 0;
			}

			.onepage-pagination li {
				padding: 0;
				text-align: center;
			}

			.onepage-pagination li a {
				padding: 10px;
				width: 4px;
				height: 4px;
				display: block;
			}

			.onepage-pagination li a:before {
				content: '';
				position: absolute;
				width: 4px;
				height: 4px;
				background: rgba(0, 0, 0, 0.85);
				border-radius: 10px;
				-webkit-border-radius: 10px;
				-moz-border-radius: 10px;
			}

			.onepage-pagination li a.active:before {
				width: 10px;
				height: 10px;
				background: none;
				border: 1px solid black;
				margin-top: -4px;
				left: 8px;
			}

			.disabled-onepage-scroll,
			.disabled-onepage-scroll .wrapper {
				overflow: auto;
			}

			.disabled-onepage-scroll .onepage-wrapper .ops-section {
				position: relative !important;
				top: auto !important;
			}

			.disabled-onepage-scroll .onepage-wrapper {
				-webkit-transform: none !important;
				-moz-transform: none !important;
				transform: none !important;
				-ms-transform: none !important;
				min-height: 100%;
			}

			.disabled-onepage-scroll .onepage-pagination {
				display: none;
			}

			body.disabled-onepage-scroll,
			.disabled-onepage-scroll .onepage-wrapper,
			html {
				position: inherit;
			}

			.header {
				background: #F5C20F;
				color: #343131;
			}

			.header .heading {
				font-size: 1.8em;
				font-weight: bold;
				font-family: monospace;
			}

			.header h2 {
				margin-top: 0.2em;
			}

			.logo {
				height: 4em;
				margin-top: 1em;
			}

			h1 {
				font-size: 2.5em;
				margin-top: 0.5em;
				color: #343131;
			}

			h2 {
				font-size: 1.8em;
				font-family: "Open Sans";
				margin-top: 0;
				padding-top: 0.5em;
			}

			.table-container {
				display: inline-block;
				overflow: hidden;
				padding-bottom: 1em;
			}

			table,
			th,
			td {
				border: 1px solid #755C07;
				border-collapse: collapse;
				font-size: 1em;
				text-align: left;
				margin-left: 0.5em;
			}

			.truncated:after {
				content: '<Truncated. Download CSV for Full data>';
				font-size: 1em;
				font-family: monospace;
			}

			th {
				background-color: #343131;
				color: #b3b3b3;
				text-align: center;
				padding: 0.5em;
			}

			td {
				padding: 0.3em;
			}

			tr:nth-child(n+{{$maxRows}}){
				display:none;
			}

			footer {
				margin-left: 22em;
				margin-top: 2em;
				margin-bottom: 1.5em;
			}

			.caption {
				font-size: 2em;
				margin: 0;
				padding: 1em;
			}

			section {
				margin: 0;
				background-color: #F5C20F;
			}

			@media only screen and (min-width: 768px) {
				.mobile {
					display: none;
				}
			}

			@media only screen and (max-width: 768px) {
				.content {
					display: none;
				}
				.mobile {
					margin-top: 3em;
				}
			}
			#map{
				height: 70em;
				padding-top: 6em;
			}

		</style>
	</head>
	<body>
		<div class="mobile">
			<div class="header">
				<img class="logo" src="https://docs.getgauge.io/_static/img/Gauge-Logo.svg">
				<div class="heading">Insights</div>
			</div>
			<p>This page is not supported in this resolution.</p>
		</div>
		<div class="content">
			<section>
				<div class="header">
					<img class="logo" src="https://docs.getgauge.io/_static/img/Gauge-Logo.svg">
					<div class="heading">Insights</div>
				</div>
				<h3 class="caption">Report period: {{reportPeriod}}</h3>
				<div id="map"></div>
			</section>
			<section>
				<h2 id="console">Command Usage - Console</h2>
				{{template "table" .Console}}
				<div class="actions">
					<a href="./console.csv">Download</a>
				</div>
			</section>
			<section>
				<h2 id="ci">Command Usage - CI</h2>
				{{template "table" .CI}}
				<div class="actions">
					<a href="./ci.csv">Download</a>
				</div>
			</section>
			<section>
				<h2 id="api">Command Usage - API</h2>
				{{template "table" .API}}
				<div class="actions">
					<a href="./api.csv">Download</a>
				</div>
			</section>
			<footer>Report generated on {{now}}</footer>
		</div>
		<script src="https://cdn.rawgit.com/peachananr/purejs-onepage-scroll/master/onepagescroll.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/d3/3.5.3/d3.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/topojson/1.6.9/topojson.min.js"></script>
		<script src="https://cdnjs.cloudflare.com/ajax/libs/datamaps/0.5.8/datamaps.world.min.js"></script>
		<script type="text/javascript">
			(function () {
				if (window.matchMedia("(max-width: 768px)").matches) {
					return;
				}
				var countries = [
					{{range .CountryWiseUsers}}
						{name: {{.Country.Name}}, latitude: {{.Country.Lat}}, longitude: {{.Country.Long}}, count: {{.UserCount}}, radius: {{.Radius}}},
					{{end}}
				];
				var map = new Datamap({
					scope: 'world',
					element: document.getElementById('map'),
					projection: 'mercator',
					width: 750,
					height: 500,
					fills: {
						defaultFill: '#755C07'
					},
					geographyConfig: {
						borderWidth: 1,
						borderOpacity: 1,
						borderColor: '#755C07',
						popupOnHover: false,
						highlightOnHover: false,
					},
					bubblesConfig: {
						popupTemplate: function (geo, data) {
							console.log(geo);
							return "<div class='hoverinfo'>" + data.name + " : " + data.count + "</div>";
						},
						highlightFillColor: '#DBAD0D',
        				highlightBorderColor: '#B58F0B',
					},					
				});
				map.bubbles(countries);
				onePageScroll(".content", { loop: true });
				var tables= document.getElementsByTagName("table");
				for(i=0; i< tables.length; i++){
					var table = tables[i];
					console.log(table);
					if (table.getElementsByTagName("tr").length > {{$maxRows}}) {
						table.parentElement.classList.add("truncated");
					}
				};
				})();
		</script>
	</body>
</html>
`
