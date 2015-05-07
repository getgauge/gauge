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

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/version"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const (
	pluginJson = "plugin.json"
	jsonExt    = ".json"
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

type installResult struct {
	error   error
	warning string
	success bool
}

func (self *installResult) getMessage() string {
	return self.error.Error()
}

func installError(err string) installResult {
	return installResult{error: errors.New(err), success: false}
}

func installSuccess(warning string) installResult {
	return installResult{warning: warning, success: true}
}

func installPlugin(pluginName, version string) installResult {
	installDescription, result := getInstallDescription(pluginName)
	if !result.success {
		return result
	}
	return installPluginWithDescription(installDescription, version)
}

func installPluginWithDescription(installDescription *installDescription, currentVersion string) installResult {
	var versionInstallDescription *versionInstallDescription
	var err error
	if currentVersion != "" {
		versionInstallDescription, err = installDescription.getVersion(currentVersion)
		if err != nil {
			return installError(err.Error())
		}
		if compatibilityError := checkCompatiblity(version.CurrentGaugeVersion, &versionInstallDescription.GaugeVersionSupport); compatibilityError != nil {
			return installError(fmt.Sprintf("Plugin Version %s-%s is not supported for gauge %s : %s", installDescription.Name, versionInstallDescription.Version, version.CurrentGaugeVersion.String(), compatibilityError.Error()))
		}
	} else {
		versionInstallDescription, err = installDescription.getLatestCompatibleVersionTo(version.CurrentGaugeVersion)
		if err != nil {
			return installError(fmt.Sprintf("Could not find compatible version for plugin %s. : %s", installDescription.Name, err))
		}
	}
	return installPluginVersion(installDescription, versionInstallDescription)
}

func installPluginVersion(installDesc *installDescription, versionInstallDescription *versionInstallDescription) installResult {
	if common.IsPluginInstalled(installDesc.Name, versionInstallDescription.Version) {
		return installSuccess(fmt.Sprintf("Plugin %s %s is already installed.", installDesc.Name, versionInstallDescription.Version))
	}

	logger.Log.Info("Installing Plugin => %s %s\n", installDesc.Name, versionInstallDescription.Version)
	pluginZip, err := downloadPluginZip(versionInstallDescription.DownloadUrls)
	if err != nil {
		return installError(fmt.Sprintf("Could not download plugin zip: %s.", err))
	}
	unzippedPluginDir, err := common.UnzipArchive(pluginZip)
	if err != nil {
		return installError(fmt.Sprintf("Failed to Unzip plugin-zip file %s.", err))
	}
	logger.Log.Info("Plugin unzipped to => %s\n", unzippedPluginDir)
	if err := runInstallCommands(versionInstallDescription.Install, unzippedPluginDir); err != nil {
		return installError(fmt.Sprintf("Failed to Run install command. %s.", err))
	}
	err = copyPluginFilesToGauge(installDesc, versionInstallDescription, unzippedPluginDir)
	if err != nil {
		installError(err.Error())
	}
	return installSuccess("")
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

	logger.Log.Info("Running plugin install command => %s\n", command)
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
	logger.Log.Info("Downloading Plugin... => %s", downloadLink)
	downloadedFile, err := common.DownloadToTempDir(downloadLink)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Could not download File %s: %s", downloadLink, err.Error()))
	}
	return downloadedFile, err
}

func getInstallDescription(plugin string) (*installDescription, installResult) {
	installJson, result := getPluginInstallJson(plugin)
	if !result.success {
		return nil, result
	}

	return getInstallDescriptionFromJson(installJson)
}

func getInstallDescriptionFromJson(installJson string) (*installDescription, installResult) {
	InstallJsonContents, readErr := common.ReadFileContents(installJson)
	if readErr != nil {
		return nil, installError(readErr.Error())
	}
	installDescription := &installDescription{}
	if err := json.Unmarshal([]byte(InstallJsonContents), installDescription); err != nil {
		return nil, installError(err.Error())
	}
	return installDescription, installSuccess("")
}

func getPluginInstallJson(plugin string) (string, installResult) {
	versionInstallDescriptionJsonFile := plugin + "-install.json"
	versionInstallDescriptionJsonUrl, result := constructPluginInstallJsonUrl(plugin)
	if !result.success {
		return "", installError(fmt.Sprintf("Could not construct plugin install json file URL. %s", result.error))
	}
	downloadedFile, downloadErr := common.DownloadToTempDir(versionInstallDescriptionJsonUrl)
	if downloadErr != nil {
		return "", installError(fmt.Sprintf("Could not download %s file. Invalid plugin name", versionInstallDescriptionJsonFile))
	}
	return downloadedFile, installSuccess("")
}

func constructPluginInstallJsonUrl(plugin string) (string, installResult) {
	installJsonFile := plugin + "-install.json"
	repoUrl := config.GaugeRepositoryUrl()
	if repoUrl == "" {
		return "", installError("Could not find gauge repository url from configuration.")
	}
	return fmt.Sprintf("%s/%s", repoUrl, installJsonFile), installSuccess("")
}

func (installDesc *installDescription) getVersion(version string) (*versionInstallDescription, error) {
	for _, versionInstallDescription := range installDesc.Versions {
		if versionInstallDescription.Version == version {
			return &versionInstallDescription, nil
		}
	}
	return nil, errors.New("Could not find install description for Version " + version)
}

func (installDesc *installDescription) getLatestCompatibleVersionTo(currentVersion *version.Version) (*versionInstallDescription, error) {
	installDesc.sortVersionInstallDescriptions()
	for _, versionInstallDesc := range installDesc.Versions {
		if err := checkCompatiblity(currentVersion, &versionInstallDesc.GaugeVersionSupport); err == nil {
			return &versionInstallDesc, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("Compatible version to %s not found", currentVersion))

}

func (installDescription *installDescription) sortVersionInstallDescriptions() {
	sort.Sort(ByDecreasingVersion(installDescription.Versions))
}

func checkCompatiblity(currentVersion *version.Version, versionSupport *versionSupport) error {
	minSupportVersion, err := version.ParseVersion(versionSupport.Minimum)
	if err != nil {
		return errors.New(fmt.Sprintf("Invalid minimum support version %s. : %s. ", versionSupport.Minimum, err))
	}
	if versionSupport.Maximum != "" {
		maxSupportVersion, err := version.ParseVersion(versionSupport.Maximum)
		if err != nil {
			return errors.New(fmt.Sprintf("Invalid maximum support version %s. : %s. ", versionSupport.Maximum, err))
		}
		if currentVersion.IsBetween(minSupportVersion, maxSupportVersion) {
			return nil
		} else {
			return errors.New(fmt.Sprintf("Version %s is not between %s and %s", currentVersion, minSupportVersion, maxSupportVersion))
		}
	}

	if minSupportVersion.IsLesserThanEqualTo(currentVersion) {
		return nil
	}
	return errors.New(fmt.Sprintf("Incompatible version. Minimum support version %s is higher than current version %s", minSupportVersion, currentVersion))
}

func installPluginFromZip(zipFile string, language string) error {
	unzippedPluginDir, err := common.UnzipArchive(zipFile)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to Unzip plugin-zip file %s.", err))
	}
	logger.Log.Info("Plugin unzipped to => %s\n", unzippedPluginDir)

	hasPluginJson := common.FileExists(filepath.Join(unzippedPluginDir, pluginJson))
	if hasPluginJson {
		return installPluginFromDir(unzippedPluginDir)
	} else {
		return installRunnerFromDir(unzippedPluginDir, language)
	}
}

func installRunnerFromDir(unzippedPluginDir string, language string) error {
	var r runner
	contents, err := common.ReadFileContents(filepath.Join(unzippedPluginDir, language+jsonExt))
	if err != nil {
		return err
	}
	err = json.Unmarshal([]byte(contents), &r)
	if err != nil {
		return err
	}
	return copyPluginFilesToGaugeInstallDir(unzippedPluginDir, r.Id, r.Version)
}

func copyPluginFilesToGaugeInstallDir(unzippedPluginDir string, pluginId string, version string) error {
	logger.Log.Info("Installing Plugin => %s %s\n", pluginId, version)

	pluginsDir, err := common.GetPrimaryPluginsInstallDir()
	if err != nil {
		return err
	}
	versionedPluginDir := path.Join(pluginsDir, pluginId, version)
	if common.DirExists(versionedPluginDir) {
		return errors.New(fmt.Sprintf("Plugin %s %s already installed at %s", pluginId, version, versionedPluginDir))
	}
	return common.MirrorDir(unzippedPluginDir, versionedPluginDir)
}

func installPluginFromDir(unzippedPluginDir string) error {
	pd, err := getPluginDescriptorFromJson(filepath.Join(unzippedPluginDir, pluginJson))
	if err != nil {
		return err
	}
	return copyPluginFilesToGaugeInstallDir(unzippedPluginDir, pd.Id, pd.Version)
}

func installAllPlugins() {
	manifest, err := getProjectManifest()
	if err != nil {
		handleCriticalError(errors.New(fmt.Sprintf("manifest.json not found : --install-all requires manifest.json in working directory.")))
	}
	installPluginsFromManifest(manifest)
}

func updatePlugin(plugin string) {
	downloadAndInstall(plugin, "", fmt.Sprintf("Successfully updated plugin => %s", plugin))
}

func downloadAndInstallPlugin(plugin, version string) {
	downloadAndInstall(plugin, version, fmt.Sprintf("Successfully installed plugin => %s", plugin))
}

func downloadAndInstall(plugin, version string, successMessage string) {
	result := installPlugin(plugin, version)
	if !result.success {
		logger.Log.Error("%s : %s\n", plugin, result.getMessage())
		os.Exit(1)
	}
	if result.warning != "" {
		logger.Log.Warning(result.warning)
		os.Exit(0)
	}
	logger.Log.Info(successMessage)
}

func installPluginZip(zipFile string, pluginName string) {
	if err := installPluginFromZip(zipFile, pluginName); err != nil {
		logger.Log.Warning("Failed to install plugin from zip file. Invalid zip file : %s\n", err)
	} else {
		logger.Log.Info("Successfully installed plugin from file")
	}
}

func installPluginsFromManifest(manifest *manifest) {
	plugins := []string{manifest.Language}
	plugins = append(plugins, manifest.Plugins...)

	for _, pluginName := range plugins {
		if !isPluginInstalledAlready(pluginName) {
			installResult := installPlugin(pluginName, "")
			if !installResult.success {
				getCurrentLogger().Error("Failed to install the %s plugin.", pluginName)
			}
		}
	}
}

func isPluginInstalledAlready(name string) bool {
	pluginsDir, err := common.GetPluginsInstallDir(name)
	if err != nil {
		return false
	}
	return common.DirExists(filepath.Join(pluginsDir, name))
}

type ByDecreasingVersion []versionInstallDescription

func (a ByDecreasingVersion) Len() int      { return len(a) }
func (a ByDecreasingVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDecreasingVersion) Less(i, j int) bool {
	version1, _ := version.ParseVersion(a[i].Version)
	version2, _ := version.ParseVersion(a[j].Version)
	return version1.IsGreaterThan(version2)
}
