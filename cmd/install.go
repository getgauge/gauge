/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package cmd

import (
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/plugin/install"
	"github.com/spf13/cobra"
)

var (
	installCmd = &cobra.Command{
		Use:   "install [flags] [plugin]",
		Short: "Download and install plugin(s)",
		Long:  `Download and install specified plugin or all plugins in the project's 'manifest.json' file.`,
		Example: `  gauge install
  gauge install java
  gauge install java -f gauge-java-0.6.3-darwin.x86_64.zip`,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) < 1 {
				install.AllPlugins(machineReadable, false)
				return
			}
			if zip != "" {
				install.HandleInstallResult(install.InstallPluginFromZipFile(zip, args[0]), args[0], true)
			} else {
				install.HandleInstallResult(install.Plugin(args[0], pVersion, machineReadable), args[0], true)
			}
			if err := config.SetProjectRoot(args); err == nil {
				if err := install.AddPluginToProject(args[0]); err != nil {
					logger.Fatalf(true, "Failed to add plugin %s to project : %s\n", args[0], err.Error())
				}
			}
		},
		DisableAutoGenTag: true,
	}
	zip      string
	pVersion string
)

func init() {
	GaugeCmd.AddCommand(installCmd)
	installCmd.Flags().StringVarP(&zip, "file", "f", "", "Installs the plugin from zip file")
	installCmd.Flags().StringVarP(&pVersion, "version", "v", "", "Version of plugin to be installed")
}
