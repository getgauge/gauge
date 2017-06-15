package cmd

import (
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/track"
	"github.com/spf13/cobra"
)

var (
	installCmd = &cobra.Command{
		Use:   "install",
		Short: "Downloads and installs a plugin.",
		Long:  "Downloads and installs a plugin.",
		Example: `  gauge install java
  gauge install java -f gauge-java-0.6.3-darwin.x86_64.zip`,
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if all {
				track.InstallAll()
				install.InstallAllPlugins()
				return
			}
			if len(args) < 1 {
				logger.Fatalf("Error: Missing argument <plugin name>.\n%s", cmd.UsageString())
			}
			if zip != "" {
				track.Install(args[0], true)
				install.HandleInstallResult(install.InstallPluginFromZipFile(zip, args[0]), args[0], true)
			} else {
				track.Install(args[0], false)
				install.HandleInstallResult(install.InstallPlugin(args[0], pVersion), args[0], true)
			}
		},
	}
	all      bool
	zip      string
	pVersion string
)

func init() {
	GaugeCmd.AddCommand(installCmd)
	installCmd.Flags().BoolVarP(&all, "all", "a", false, "Installs all the plugins specified in project manifest, if not installed.")
	installCmd.Flags().StringVarP(&zip, "file", "f", "", "Installs the plugin from zip file.")
	installCmd.Flags().StringVarP(&pVersion, "plugin-version", "v", "", "Version of plugin to be installed.")
}
