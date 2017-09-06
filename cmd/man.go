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

// +build linux darwin

package cmd

import (
	"strings"

	"os"

	"path/filepath"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/version"
	"github.com/spf13/cobra"
	"github.com/spf13/cobra/doc"
)

const manDir = "man"

var (
	manCmd = &cobra.Command{
		Use:     "man [flags]",
		Short:   "Generate man pages",
		Long:    `Generate man pages.`,
		Example: `  gauge man`,
		Run: func(cmd *cobra.Command, args []string) {
			out, err := getDefaultPath()
			if err != nil {
				logger.Fatalf("Cannot find the gauge home directory.")
			}
			if err := genManPages(out); err != nil {
				logger.Fatalf(err.Error())
			}
		},
		DisableAutoGenTag: true,
	}
)

func init() {
	GaugeCmd.AddCommand(manCmd)
}

func getDefaultPath() (string, error) {
	p, err := common.GetGaugeHomeDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(p, manDir, "man1"), nil
}

func genManPages(out string) error {
	if err := os.MkdirAll(out, common.NewDirectoryPermissions); err != nil {
		return err
	}
	if err := doc.GenManTreeFromOpts(setupCmd(), doc.GenManTreeOptions{
		Header: &doc.GenManHeader{
			Title:   "GAUGE",
			Section: "1",
			Manual:  "GAUGE MANUAL",
			Source:  "GAUGE " + version.CurrentGaugeVersion.String(),
		},
		Path:             out,
		CommandSeparator: "-",
	}); err != nil {
		return err
	}
	p := strings.TrimSuffix(out, filepath.Base(out))
	logger.Infof("To view gauge man pages, add `%s` to `MANPATH` environment variable.", p)
	return nil
}

func setupCmd() *cobra.Command {
	GaugeCmd.Short = "A light-weight cross-platform test automation tool"
	GaugeCmd.Long = "Gauge is a light-weight cross-platform test automation tool with the ability to author test cases in the business language."
	return GaugeCmd
}
