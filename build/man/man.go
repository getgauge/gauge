/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package main

import (
	"log"

	"strings"

	"os"
	"path/filepath"

	"math"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/cmd"
	"github.com/russross/blackfriday/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const (
	maxDescLength = 46
	maxLineLength = 77
)

type text struct {
	name    string
	content string
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
	for _, t := range texts {
		name := strings.TrimSuffix(t.name, filepath.Ext(t.name)) + ".html"
		output := strings.ReplaceAll(html, "<!--CONTENT-->", string(blackfriday.Run([]byte(t.content))))
		p := filepath.Join(htmlPath, name)
		err := os.WriteFile(p, []byte(output), 0644)
		if err != nil {
			log.Fatalf("Unable to write file %s: %s", p, err.Error())
		}
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

func indentText(p string) (texts []text) {
	err := filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), ".md") {
			bytes, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			var lines []string
			for _, l := range strings.Split(string(bytes), string("\n")) {
				tLine := strings.TrimSpace(l)
				if strings.HasPrefix(tLine, "-") && len(tLine) > maxLineLength {
					lines = append(lines, indentFlag(l, tLine)...)
				} else {
					lines = append(lines, strings.ReplaceAll(l, ".md", ".html"))
				}
			}
			texts = append(texts, text{name: info.Name(), content: strings.Join(lines, "\n")})
		}
		return nil
	})
	if err != nil {
		log.Fatal(err.Error())
	}

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
		prefix := strings.ReplaceAll(line, desc, strings.Join(dWords[:till], " "))
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
    </style>
</head>

<body id='manpage'>
    <div class='mp' id='man'>
        <!--CONTENT-->
		<div><b>Complete documentation is available <a href="https://docs.gauge.org/">here</a>.</b></div>
    </div>

</body>

</html>
`
