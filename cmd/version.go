package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/version"
	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print Gauge and plugin versions.",
		Long:  "Print Gauge and plugin versions.",
		Example: `  gauge version
  gauge version -m`,
		Run: func(cmd *cobra.Command, args []string) {
			setGlobalFlags()
			if machineReadable {
				PrintJSONVersion()
				return
			}
			PrintVersion()
		},
	}
	machineReadable bool
)

func init() {
	GaugeCmd.AddCommand(versionCmd)
	versionCmd.Flags().BoolVarP(&machineReadable, "machine-readable", "m", false, "Outputs JSON output of currently installed Gauge and plugin versions.")
}

func PrintJSONVersion() {
	type pluginJSON struct {
		Name    string `json:"name"`
		Version string `json:"version"`
	}
	type versionJSON struct {
		Version string        `json:"version"`
		Plugins []*pluginJSON `json:"plugins"`
	}
	gaugeVersion := versionJSON{version.FullVersion(), make([]*pluginJSON, 0)}
	allPluginsWithVersion, err := plugin.GetAllInstalledPluginsWithVersion()
	for _, pluginInfo := range allPluginsWithVersion {
		gaugeVersion.Plugins = append(gaugeVersion.Plugins, &pluginJSON{pluginInfo.Name, filepath.Base(pluginInfo.Path)})
	}
	b, err := json.MarshalIndent(gaugeVersion, "", "    ")
	if err != nil {
		fmt.Println("error:", err)
	}
	fmt.Println(fmt.Sprintf("%s\n", string(b)))
}

func PrintVersion() {
	fmt.Printf("Gauge version: %s\n\n", version.FullVersion())
	fmt.Println("Plugins\n-------")
	allPluginsWithVersion, err := plugin.GetAllInstalledPluginsWithVersion()
	if err != nil {
		fmt.Println("No plugins found")
		fmt.Println("Plugins can be installed with `gauge --install {plugin-name}`")
		os.Exit(0)
	}
	for _, pluginInfo := range allPluginsWithVersion {
		fmt.Printf("%s (%s)\n", pluginInfo.Name, filepath.Base(pluginInfo.Path))
	}
}
