//go:build linux || darwin

/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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
				logger.Fatal(true, "Cannot find the gauge home directory.")
			}
			if err := genManPages(out); err != nil {
				logger.Fatal(true, err.Error())
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
	logger.Infof(true, "To view gauge man pages, add `%s` to `MANPATH` environment variable.", p)
	return nil
}

func setupCmd() *cobra.Command {
	GaugeCmd.Short = "A light-weight cross-platform test automation tool"
	GaugeCmd.Long = "Gauge is a light-weight cross-platform test automation tool with the ability to author test cases in the business language."
	return GaugeCmd
}
