package main

import (
	"log"

	"fmt"
	"path"
	"strings"

	"io/ioutil"
	"os"
	"path/filepath"

	"math"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/cmd"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const (
	templateFile = "template.html"
)

func main() {
	c := setupCmd()
	p := filepath.Join("_man", "docs")
	if err := os.MkdirAll(p, common.NewDirectoryPermissions); err != nil {
		log.Fatal(err.Error())
	}
	if err := ioutil.WriteFile(filepath.Join(p, templateFile), []byte(html), common.NewFilePermissions); err != nil {
		log.Fatal(err.Error())
	}
	if err := doc.GenMarkdownTreeCustom(c, p, func(s string) string { return "" }, func(s string) string {
		return fmt.Sprintf("%s.html", strings.TrimSuffix(s, path.Ext(s)))
	}); err != nil {
		log.Fatal(err.Error())
	}
	indent(p)
	log.Printf("Markdown man pages are available in %s dir\n", p)
}

func setupCmd() *cobra.Command {
	cmd.InitHelp(cmd.GaugeCmd)
	cmd.GaugeCmd.Short = "A light-weight cross-platform test automation tool"
	cmd.GaugeCmd.Long = "Gauge is a light-weight cross-platform test automation tool with the ability to author test cases in the business language."
	return cmd.GaugeCmd
}

func indent(p string) {
	filepath.Walk(p, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(info.Name(), ".md") {
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			var lines []string
			for _, l := range strings.Split(string(bytes), string("\n")) {
				tLine := strings.TrimSpace(l)
				if strings.HasPrefix(tLine, "-") && len(tLine) > 77 {
					lines = append(lines, indentFlag(l, tLine)...)
				} else {
					lines = append(lines, l)
				}
			}
			err = ioutil.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
func indentFlag(line, tLine string) (lines []string) {
	words := strings.Split(tLine, "  ")
	desc := strings.TrimSpace(words[len(words)-1])
	dWords := strings.Split(desc, " ")
	times := math.Ceil(float64(len(desc)) / 46)
	for i := 0; float64(i) < times; i++ {
		till := 0
		length := 0
		for i, v := range dWords {
			length += len(v)
			if length > 46 {
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

var html = `
<!DOCTYPE html>
<html>

<head>
    <title>Gauge - Manual</title>
    <link href="https://getgauge.io/assets/images/favicons/favicon.ico" rel="shortcut icon" type="image/ico" />
    <style type='text/css' media='all'>
        body#manpage {
            margin: 0
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
            top: 0;
            left: 113ex;
            height: 100%;
            width: 100%;
            padding: 48px 0 0 0;
            border-left: 1px solid #dbdbdb;
            background: #eee;
        }

        .man-navigation a,
        .man-navigation a:hover,
        .man-navigation a:link,
        .man-navigation a:visited {
            display: block;
            margin: 0;
            padding: 5px 2px 5px 0px;
            color: #999;
            text-decoration: none;
        }

        .man-navigation a:hover {
            color: #111;
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
            color: #717171 !important;
        }
    </style>
</head>

<body id='manpage'>
    <div class='mp' id='man'>
        <!--CONTENT-->
        <nav id="menu" class='man-navigation' style='display:none'>
            <!--NAV-->
        </nav>
    </div>
    <script>
        (function(window, document) {

            var layout = document.getElementById('layout'),
                menu = document.getElementById('menu');

            function toggleClass(element, className) {
                var classes = element.className.split(/\s+/),
                    length = classes.length,
                    i = 0;
                for (; i < length; i++) {
                    if (classes[i] === className) {
                        classes.splice(i, 1);
                        break;
                    }
                }
                if (length === classes.length) {
                    classes.push(className);
                }
                element.className = classes.join(' ');
            }
        }(this, this.document));
    </script>
</body>

</html>
`
