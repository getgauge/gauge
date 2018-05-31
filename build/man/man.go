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
	"log"

	"strings"

	"io/ioutil"
	"os"
	"path/filepath"

	"math"

	"html/template"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/cmd"
	"github.com/russross/blackfriday"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const (
	maxDescLength = 46
	maxLineLength = 77
)

type link struct {
	Class string
	Link  string
	Name  string
}

type text struct {
	name    string
	content string
}

var indexTemplate, _ = template.New("test").Parse(`<ul>
{{ range . }}
	<li><a class="{{ .Class }}" href="{{ .Link }}">{{ .Name }}</a></li>
{{ end }}
</ul>`)

type writer struct {
	text string
}

func (w *writer) Write(b []byte) (int, error) {
	w.text += string(b)
	return 0, nil
}

func main() {
	mdPath := filepath.Join("_man", "md")
	htmlPath := filepath.Join("_man", "html")
	createDir(mdPath)
	createDir(htmlPath)
	if err := genMarkdownManPages(mdPath); err != nil {
		log.Fatal(err.Error())
	}
	texts := indentText(mdPath)
	links := getLinks(texts)
	for _, t := range texts {
		name := strings.TrimSuffix(t.name, filepath.Ext(t.name)) + ".html"
		var newLinks []link
		for _, l := range links {
			if l.Link == name {
				newLinks = append(newLinks, link{Name: l.Name, Link: l.Link, Class: "active"})
			} else {
				newLinks = append(newLinks, l)
			}
		}
		page := strings.Replace(html, "<!--NAV-->", prepareIndex(newLinks), -1)
		output := strings.Replace(page, "<!--CONTENT-->", string(blackfriday.MarkdownCommon([]byte(t.content))), -1)
		ioutil.WriteFile(filepath.Join(htmlPath, name), []byte(output), 0644)
	}
	log.Printf("HTML man pages are available in %s dir\n", htmlPath)
}
func createDir(p string) {
	if err := os.MkdirAll(p, common.NewDirectoryPermissions); err != nil {
		log.Fatal(err.Error())
	}
}

func genMarkdownManPages(out string) error {
	if err := doc.GenMarkdownTree(setupCmd(), out); err != nil {
		return err
	}
	log.Printf("Added markdown man pages to `%s`\n", out)
	return nil
}

func setupCmd() *cobra.Command {
	cmd.GaugeCmd.Short = "A light-weight cross-platform test automation tool"
	cmd.GaugeCmd.Long = "Gauge is a light-weight cross-platform test automation tool with the ability to author test cases in the business language."
	return cmd.GaugeCmd
}

func getLinks(texts []text) (links []link) {
	for _, t := range texts {
		name := strings.TrimSuffix(t.name, filepath.Ext(t.name))
		links = append(links, link{Class: "", Name: strings.Replace(name, "_", " ", -1), Link: name + ".html"})
	}
	return
}

func prepareIndex(links []link) string {
	w := &writer{}
	err := indexTemplate.Execute(w, links)
	if err != nil {
		log.Fatalf(err.Error())
	}
	return w.text
}

func indentText(p string) (texts []text) {
	filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), ".md") {
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			var lines []string
			for _, l := range strings.Split(string(bytes), string("\n")) {
				tLine := strings.TrimSpace(l)
				if strings.HasPrefix(tLine, "-") && len(tLine) > maxLineLength {
					lines = append(lines, indentFlag(l, tLine)...)
				} else {
					lines = append(lines, strings.Replace(l, ".md", ".html", -1))
				}
			}
			texts = append(texts, text{name: info.Name(), content: strings.Join(lines, "\n")})
		}
		return nil
	})
	return
}

func indentFlag(line, tLine string) (lines []string) {
	words := strings.Split(tLine, "  ")
	desc := strings.TrimSpace(words[len(words)-1])
	dWords := strings.Split(desc, " ")
	times := math.Ceil(float64(len(desc)) / maxDescLength)
	for i := 0; float64(i) < times; i++ {
		till := 0
		length := 0
		for i, v := range dWords {
			length += len(v)
			if length > maxDescLength {
				till = i - 1
				break
			}
			if i == len(dWords)-1 {
				till = len(dWords)
			}
		}
		if len(dWords) == 0 {
			continue
		}
		prefix := strings.Replace(line, desc, strings.Join(dWords[:till], " "), -1)
		if i != 0 {
			prefix = strings.Repeat(" ", strings.Index(line, desc)) + strings.Join(dWords[:till], " ")
		}
		lines = append(lines, prefix)
		dWords = dWords[till:]
	}
	return
}

const html = `
<!DOCTYPE html>
<html>

<head>
    <title>Gauge - Manual</title>
    <link href="https://gauge.org/assets/images/favicons/favicon.ico" rel="shortcut icon" type="image/ico" />
    <style type='text/css' media='all'>
        body#manpage {
            margin: 0;
            border-top: 3px solid #f5c10e;
        }

        .mp {
            max-width: 100ex;
            padding: 0 9ex 1ex 4ex;
            margin-top: 1.5%;
        }

        .mp p,
        .mp pre,
        .mp ul,
        .mp ol,
        .mp dl {
            margin: 0 0 20px 0;
        }

        .mp h2 {
            margin: 10px 0 0 0
        }

        .mp h3 {
            margin: 0 0 0 0;
        }

        .mp dt {
            margin: 0;
            clear: left
        }

        .mp dt.flush {
            float: left;
            width: 8ex
        }

        .mp dd {
            margin: 0 0 0 9ex
        }

        .mp h1,
        .mp h2,
        .mp h3,
        .mp h4 {
            clear: left
        }

        .mp pre {
            margin-bottom: 20px;
        }

        .mp pre+h2,
        .mp pre+h3 {
            margin-top: 22px
        }

        .mp h2+pre,
        .mp h3+pre {
            margin-top: 5px
        }

        .mp img {
            display: block;
            margin: auto
        }

        .mp h1.man-title {
            display: none
        }

        .mp,
        .mp code,
        .mp pre,
        .mp tt,
        .mp kbd,
        .mp samp,
        .mp h3,
        .mp h4 {
            font-family: monospace;
            font-size: 14px;
            line-height: 1.42857142857143
        }

        .mp h2 {
            font-size: 16px;
            line-height: 1.25
        }

        .mp h1 {
            font-size: 20px;
            line-height: 2
        }

        .mp {
            text-align: justify;
            background: #fff
        }

        .mp,
        .mp code,
        .mp pre,
        .mp pre code,
        .mp tt,
        .mp kbd,
        .mp samp {
            color: #131211
        }

        .mp h1,
        .mp h2,
        .mp h3,
        .mp h4 {
            color: #030201
        }

        .mp u {
            text-decoration: underline
        }

        .mp code,
        .mp strong,
        .mp b {
            font-weight: bold;
            color: #131211
        }

        .mp em,
        .mp var {
            font-style: italic;
            color: #232221;
            text-decoration: none
        }

        .mp a,
        .mp a:link,
        .mp a:hover,
        .mp a code,
        .mp a pre,
        .mp a tt,
        .mp a kbd,
        .mp a samp {
            color: #0000ff
        }

        .mp b.man-ref {
            font-weight: normal;
            color: #434241
        }

        .mp pre code {
            font-weight: normal;
            color: #434241
        }

        .mp h2+pre,
        h3+pre {
            padding-left: 0
        }

        ol.man-decor,
        ol.man-decor li {
            margin: 3px 0 10px 0;
            padding: 0;
            float: left;
            width: 33%;
            list-style-type: none;
            text-transform: uppercase;
            color: #999;
            letter-spacing: 1px;
        }

        ol.man-decor {
            width: 100%;
        }

        ol.man-decor li.tl {
            text-align: left;
        }

        ol.man-decor li.tc {
            text-align: center;
            letter-spacing: 4px;
        }

        ol.man-decor li.tr {
            text-align: right;
            float: right;
        }

        .man-navigation ul {
            font-size: 16px;
        }
    </style>
    <style type='text/css' media='all'>
        .man-navigation {
            display: block !important;
            position: fixed;
            top: 3px;
            left: 113ex;
            height: 100%;
            width: 100%;
            padding: 48px 0 0 0;
            border-left: 1px solid #dbdbdb;
            background: #333333;
        }

        .man-navigation a,
        .man-navigation a:hover,
        .man-navigation a:link,
        .man-navigation a:visited {
            display: block;
            margin: 0;
            padding: 5px 2px 5px 0px;
            color: #ffffff;
            text-decoration: none;
        }

        .man-navigation a:hover {
            color: #f5c10e;
            text-decoration: underline;
        }

        li {
            list-style: none;
        }

        .mp li {
            margin-left: -3ex;
        }

        a.active {
            font-weight: bolder;
            color: #f5c10e !important;
        }
    </style>
</head>

<body id='manpage'>
    <div class='mp' id='man'>
        <!--CONTENT-->
		<div><b>Complete documentation is available <a href="https://docs.gauge.org/">here</a>.</b></div>
        <nav id="menu" class='man-navigation' style='display:none'>
            <!--NAV-->
        </nav>
    </div>

</body>

</html>
`
