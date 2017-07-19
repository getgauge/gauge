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

package cmd

import (
	"fmt"
	"path"
	"strings"

	"os"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/logger"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

var (
	manCmd = &cobra.Command{
		Use:   "man [flags]",
		Short: "Generate man pages",
		Long:  `Generate man pages.`,
		Example: `  gauge man
  gauge man --html
  gauge man --html man/`,
		Hidden: true,
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if md {
				out := "man"
				if len(args) > 0 {
					out = args[0]
				}
				if err := genMarkdownManPages(out); err != nil {
					logger.Fatalf(err.Error())
				}
			} else {
				logger.Infof("Not available.")
			}
		},
		DisableAutoGenTag: true,
	}
	md bool
)

func genMarkdownManPages(out string) error {
	c := setupCmd()
	if err := os.MkdirAll(out, common.NewDirectoryPermissions); err != nil {
		return err
	}
	if err := doc.GenMarkdownTreeCustom(c, out, func(s string) string { return "" }, func(s string) string {
		return fmt.Sprintf("%s.html", strings.TrimSuffix(s, path.Ext(s)))
	}); err != nil {
		return err
	}
	return nil
}

func init() {
	GaugeCmd.AddCommand(manCmd)
	manCmd.Flags().BoolVarP(&md, "md", "", false, "Generate man pagess in markdown format")
}

func setupCmd() *cobra.Command {
	GaugeCmd.Short = "A light-weight cross-platform test automation tool"
	GaugeCmd.Long = "Gauge is a light-weight cross-platform test automation tool with the ability to author test cases in the business language."
	return GaugeCmd
}
