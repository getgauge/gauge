package print

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/version"
)

func Version() {
	fmt.Printf("Gauge version: %s\n\n", version.CurrentGaugeVersion.String())
	fmt.Println("Plugins\n-------")
	allPluginsWithVersion, err := plugin.GetAllInstalledPluginsWithVersion()
	if err != nil {
		fmt.Println("No plugins found")
		fmt.Println("Plugins can be installed with `gauge --install {plugin-name}`")
		os.Exit(0)
	}
	for _, pluginInfo := range allPluginsWithVersion {
		fmt.Printf("%s (%s)\n", pluginInfo.Name, pluginInfo.Version.String())
	}
}

func Usage() {
	fmt.Printf("gauge -version %s\n", version.CurrentGaugeVersion.String())
	fmt.Printf("Copyright %d Thoughtworks\n\n", time.Now().Year())
	fmt.Println("Usage:")
	fmt.Println("\tgauge specs/")
	fmt.Println("\tgauge specs/spec_name.spec")
	fmt.Println("\nOptions:")
	flag.PrintDefaults()
	os.Exit(2)
}
