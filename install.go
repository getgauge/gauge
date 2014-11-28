package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/getgauge/common"
	"os"
	"path"
	"runtime"
	"sort"
	"strings"
)

type installDescription struct {
	Name        string
	Description string
	Versions    []versionInstallDescription
}

type versionInstallDescription struct {
	Version             string
	GaugeVersionSupport versionSupport
	Install             platformSpecificCommand
	DownloadUrls        downloadUrls
}

type downloadUrls struct {
	X86 platformSpecificUrl
	X64 platformSpecificUrl
}

type platformSpecificCommand struct {
	Windows []string
	Linux   []string
	Darwin  []string
}

type platformSpecificUrl struct {
	Windows string
	Linux   string
	Darwin  string
}

type versionSupport struct {
	Minimum string
	Maximum string
}

func installPlugin(pluginName, version string) error {
	installDescription, err := getInstallDescription(pluginName)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not find install description for Plugin: '%s' %s. : %s \n", pluginName, version, err))
	}
	if err := installPluginWithDescription(installDescription, version); err != nil {
		return err
	}

	return nil
}

func installPluginWithDescription(installDescription *installDescription, version string) error {
	var versionInstallDescription *versionInstallDescription
	var err error
	if version != "" {
		versionInstallDescription, err = installDescription.getVersion(version)
		if err != nil {
			return err
		}
		if compatibilityError := checkCompatiblity(currentGaugeVersion, &versionInstallDescription.GaugeVersionSupport); compatibilityError != nil {
			return errors.New(fmt.Sprintf("Plugin Version %s is not supported for gauge %s : %s", installDescription.Name, versionInstallDescription.Version, versionInstallDescription.Version, currentGaugeVersion.String(), compatibilityError.Error()))
		}
	} else {
		versionInstallDescription, err = installDescription.getLatestCompatibleVersionTo(currentGaugeVersion)
		if err != nil {
			return errors.New(fmt.Sprintf("Could not find compatible version for plugin %s. : %s", installDescription.Name, err))
		}
	}
	return installPluginVersion(installDescription, versionInstallDescription)
}

func installPluginVersion(installDesc *installDescription, versionInstallDescription *versionInstallDescription) error {
	if common.IsPluginInstalled(installDesc.Name, versionInstallDescription.Version) {
		return errors.New(fmt.Sprintf("Plugin %s %s is already installed.", installDesc.Name, versionInstallDescription.Version))
	}

	fmt.Printf("Installing Plugin => %s %s\n", installDesc.Name, versionInstallDescription.Version)
	pluginZip, err := downloadPluginZip(versionInstallDescription.DownloadUrls)
	if err != nil {
		return errors.New(fmt.Sprintf("Could not download plugin zip: %s.", err))
	}
	unzippedPluginDir, err := common.UnzipArchive(pluginZip)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to Unzip plugin-zip file %s.", err))
	}
	fmt.Printf("Plugin unzipped to => %s\n", unzippedPluginDir)
	if err := runInstallCommands(versionInstallDescription.Install, unzippedPluginDir); err != nil {
		return errors.New(fmt.Sprintf("Failed to Run install command. %s.", err))
	}
	return copyPluginFilesToGauge(installDesc, versionInstallDescription, unzippedPluginDir)
}

func runInstallCommands(installCommands platformSpecificCommand, workingDir string) error {
	command := []string{}
	switch runtime.GOOS {
	case "windows":
		command = installCommands.Windows
		break
	case "darwin":
		command = installCommands.Darwin
		break
	default:
		command = installCommands.Linux
		break
	}

	if len(command) == 0 {
		return nil
	}

	fmt.Printf("Running plugin install command => %s\n", command)
	cmd, err := common.ExecuteCommand(command, workingDir, os.Stdout, os.Stderr)

	if err != nil {
		return err
	}

	return cmd.Wait()
}

func copyPluginFilesToGauge(installDesc *installDescription, versionInstallDesc *versionInstallDescription, pluginContents string) error {
	pluginsDir, err := common.GetPrimaryPluginsInstallDir()
	if err != nil {
		return err
	}
	versionedPluginDir := path.Join(pluginsDir, installDesc.Name, versionInstallDesc.Version)
	if common.DirExists(versionedPluginDir) {
		return errors.New(fmt.Sprintf("Plugin %s %s already installed at %s", installDesc.Name, versionInstallDesc.Version, versionedPluginDir))
	}
	return common.MirrorDir(pluginContents, versionedPluginDir)

}

func downloadPluginZip(downloadUrls downloadUrls) (string, error) {
	var platformLinks *platformSpecificUrl
	if strings.Contains(runtime.GOARCH, "64") {
		platformLinks = &downloadUrls.X64
	} else {
		platformLinks = &downloadUrls.X86
	}

	var downloadLink string
	switch runtime.GOOS {
	case "windows":
		downloadLink = platformLinks.Windows
		break
	case "darwin":
		downloadLink = platformLinks.Darwin
		break
	default:
		downloadLink = platformLinks.Linux
		break
	}
	if downloadLink == "" {
		return "", errors.New(fmt.Sprintf("Platform not supported for %s. Download URL not specified.", runtime.GOOS))
	}
	downloadedFile, err := common.DownloadToTempDir(downloadLink)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not download File %s: %s", downloadLink, err.Error()))
	}
	return downloadedFile, err
}

func getInstallDescription(plugin string) (*installDescription, error) {
	installJson, err := getPluginInstallJson(plugin)
	if err != nil {
		return nil, err
	}
	InstallJsonContents, readErr := common.ReadFileContents(installJson)
	if readErr != nil {
		return nil, readErr
	}
	installDescription := &installDescription{}
	if err = json.Unmarshal([]byte(InstallJsonContents), installDescription); err != nil {
		return nil, err
	}
	return installDescription, nil
}

func getPluginInstallJson(plugin string) (string, error) {
	versionInstallDescriptionJsonFile := plugin + "-install.json"
	versionInstallDescriptionJsonUrl, err := constructPluginInstallJsonUrl(plugin)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not construct plugin install json file URL. %s", err))
	}
	downloadedFile, downloadErr := common.DownloadToTempDir(versionInstallDescriptionJsonUrl)
	if downloadErr != nil {
		return "", errors.New(fmt.Sprintf("Could not find %s file. Check install name and version. %s", versionInstallDescriptionJsonFile, downloadErr.Error()))
	}
	return downloadedFile, nil
}

func constructPluginInstallJsonUrl(plugin string) (string, error) {
	installJsonFile := plugin + "-install.json"
	repoUrl, err := getGaugeRepositoryUrl()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", repoUrl, installJsonFile), nil
}

func getGaugeRepositoryUrl() (string, error) {
	config, err := common.GetGaugeConfiguration()
	if err != nil {
		return "", err
	}
	return config[common.GaugeRepositoryUrl], nil
}

func (installDesc *installDescription) getVersion(version string) (*versionInstallDescription, error) {
	for _, versionInstallDescription := range installDesc.Versions {
		if versionInstallDescription.Version == version {
			return &versionInstallDescription, nil
		}
	}
	return nil, errors.New("Could not find install description for Version " + version)
}

func (installDesc *installDescription) getLatestCompatibleVersionTo(version *version) (*versionInstallDescription, error) {
	installDesc.sortVersionInstallDescriptions()
	for _, versionInstallDesc := range installDesc.Versions {
		if err := checkCompatiblity(version, &versionInstallDesc.GaugeVersionSupport); err == nil {
			return &versionInstallDesc, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Compatible version to %s not found", version))

}

func (installDescription *installDescription) sortVersionInstallDescriptions() {
	sort.Sort(ByDecreasingVersion(installDescription.Versions))
}

func checkCompatiblity(version *version, versionSupport *versionSupport) error {
	minSupportVersion, err := parseVersion(versionSupport.Minimum)
	if err != nil {
		return errors.New(fmt.Sprintf("Invalid minimum support version %s. : %s. ", versionSupport.Minimum, err))
	}
	if versionSupport.Maximum != "" {
		maxSupportVersion, err := parseVersion(versionSupport.Maximum)
		if err != nil {
			return errors.New(fmt.Sprintf("Invalid maximum support version %s. : %s. ", versionSupport.Maximum, err))
		}
		if version.isBetween(minSupportVersion, maxSupportVersion) {
			return nil
		} else {
			return errors.New(fmt.Sprintf("Version %s is not between %s and %s", version, minSupportVersion, maxSupportVersion))
		}
	}

	if minSupportVersion.isLesserThanEqualTo(version) {
		return nil
	}
	return errors.New(fmt.Sprintf("Incompatible version. Minimum support version %s is higher than current version %s", minSupportVersion, version))
}

type ByDecreasingVersion []versionInstallDescription

func (a ByDecreasingVersion) Len() int      { return len(a) }
func (a ByDecreasingVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDecreasingVersion) Less(i, j int) bool {
	version1, _ := parseVersion(a[i].Version)
	version2, _ := parseVersion(a[j].Version)
	return version1.isGreaterThan(version2)
}
