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

package install

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
	"github.com/getgauge/gauge/version"
)

const (
	pluginJSON = "plugin.json"
	setupScope = "setup"
	jsonExt    = ".json"
)

type installDescription struct {
	Name        string
	Description string
	Versions    []versionInstallDescription
}

type versionInstallDescription struct {
	Version             string
	GaugeVersionSupport version.VersionSupport
	Install             platformSpecificCommand
	DownloadUrls        downloadUrls
}

type downloadUrls struct {
	X86 platformSpecificURL
	X64 platformSpecificURL
}

type platformSpecificCommand struct {
	Windows []string
	Linux   []string
	Darwin  []string
}

type platformSpecificURL struct {
	Windows string
	Linux   string
	Darwin  string
}

type installResult struct {
	Error   error
	Warning string
	Success bool
}

func (installResult *installResult) getMessage() string {
	return installResult.Error.Error()
}

func installError(err string) installResult {
	return installResult{Error: errors.New(err), Success: false}
}

func installSuccess(warning string) installResult {
	return installResult{Warning: warning, Success: true}
}

func InstallPlugin(pluginName, version string) installResult {
	installDescription, result := getInstallDescription(pluginName)
	if !result.Success {
		return result
	}
	defer util.RemoveTempDir()
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
		if compatibilityError := version.CheckCompatibility(version.CurrentGaugeVersion, &versionInstallDescription.GaugeVersionSupport); compatibilityError != nil {
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

	logger.Info("Installing Plugin => %s %s\n", installDesc.Name, versionInstallDescription.Version)
	downloadLink, err := getDownloadLink(versionInstallDescription.DownloadUrls)
	if err != nil {
		return installError(fmt.Sprintf("Could not get download link: %s", err.Error()))
	}

	unzippedPluginDir, err := util.DownloadAndUnzip(downloadLink)
	if err != nil {
		return installError(err.Error())
	}

	if err := runInstallCommands(versionInstallDescription.Install, unzippedPluginDir); err != nil {
		return installError(fmt.Sprintf("Failed to Run install command. %s.", err.Error()))
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

	logger.Info("Running plugin install command => %s\n", command)
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
		return fmt.Errorf("Plugin %s %s already installed at %s", installDesc.Name, versionInstallDesc.Version, versionedPluginDir)
	}
	_, err = common.MirrorDir(pluginContents, versionedPluginDir)
	return err
}

func UninstallPlugin(pluginName string, version string) {
	pluginsDir, err := common.GetPrimaryPluginsInstallDir()
	if err != nil {
		handleUninstallFailure(err, pluginName)
	}
	pluginInfo := pluginName
	if version != "" {
		pluginInfo = fmt.Sprintf("%s(%s)", pluginName, version)
	}
	pluginInstallationDir := path.Join(pluginsDir, pluginName, version)
	if common.DirExists(pluginInstallationDir) {
		if err = os.RemoveAll(pluginInstallationDir); err != nil {
			handleUninstallFailure(err, pluginInfo)
		} else {
			logger.Info("%s plugin uninstalled successfully", pluginInfo)
		}
	} else {
		logger.Info("%s plugin is not installed", pluginInfo)
	}
}

func handleUninstallFailure(err error, pluginName string) {
	logger.Error("%s plugin uninstallation failed", pluginName)
	logger.Fatal(err.Error())
}

func getDownloadLink(downloadUrls downloadUrls) (string, error) {
	var platformLinks *platformSpecificURL
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
		return "", fmt.Errorf("Platform not supported for %s. Download URL not specified.", runtime.GOOS)
	}
	return downloadLink, nil
}

func getInstallDescription(plugin string) (*installDescription, installResult) {
	installJSON, result := getPluginInstallJSON(plugin)
	if !result.Success {
		return nil, result
	}

	return getInstallDescriptionFromJSON(installJSON)
}

func getInstallDescriptionFromJSON(installJSON string) (*installDescription, installResult) {
	InstallJSONContents, readErr := common.ReadFileContents(installJSON)
	if readErr != nil {
		return nil, installError(readErr.Error())
	}
	installDescription := &installDescription{}
	if err := json.Unmarshal([]byte(InstallJSONContents), installDescription); err != nil {
		return nil, installError(err.Error())
	}
	return installDescription, installSuccess("")
}

func getPluginInstallJSON(plugin string) (string, installResult) {
	versionInstallDescriptionJSONFile := plugin + "-install.json"
	versionInstallDescriptionJSONUrl, result := constructPluginInstallJsonUrl(plugin)
	if !result.Success {
		return "", installError(fmt.Sprintf("Could not construct plugin install json file URL. %s", result.Error))
	}
	downloadedFile, downloadErr := common.DownloadToTempDir(versionInstallDescriptionJSONUrl)
	if downloadErr != nil {
		return "", installError(fmt.Sprintf("Invalid plugin : Could not download %s file. %s", versionInstallDescriptionJSONFile, downloadErr))
	}
	return downloadedFile, installSuccess("")
}

func constructPluginInstallJsonUrl(plugin string) (string, installResult) {
	installJSONFile := plugin + "-install.json"
	repoURL := config.GaugeRepositoryUrl()
	if repoURL == "" {
		return "", installError("Could not find gauge repository url from configuration.")
	}
	return fmt.Sprintf("%s/%s", repoURL, installJSONFile), installSuccess("")
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
		if err := version.CheckCompatibility(currentVersion, &versionInstallDesc.GaugeVersionSupport); err == nil {
			return &versionInstallDesc, nil
		}
	}
	return nil, fmt.Errorf("Compatible version to %s not found", currentVersion)
}

func (installDescription *installDescription) sortVersionInstallDescriptions() {
	sort.Sort(ByDecreasingVersion(installDescription.Versions))
}

func installPluginFromZip(zipFile string, language string) error {
	unzippedPluginDir, err := common.UnzipArchive(zipFile)
	if err != nil {
		return fmt.Errorf("Failed to unzip plugin-zip file %s.", err.Error())
	}
	logger.Info("Plugin unzipped to => %s\n", unzippedPluginDir)

	hasPluginJSON := common.FileExists(filepath.Join(unzippedPluginDir, pluginJSON))
	if hasPluginJSON {
		return installPluginFromDir(unzippedPluginDir)
	}
	return installRunnerFromDir(unzippedPluginDir, language)
}

func installRunnerFromDir(unzippedPluginDir string, language string) error {
	var r runner.Runner
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

func copyPluginFilesToGaugeInstallDir(unzippedPluginDir string, pluginID string, version string) error {
	logger.Info("Installing Plugin => %s %s\n", pluginID, version)

	pluginsDir, err := common.GetPrimaryPluginsInstallDir()
	if err != nil {
		return err
	}
	versionedPluginDir := path.Join(pluginsDir, pluginID, version)
	if common.DirExists(versionedPluginDir) {
		return fmt.Errorf("Plugin %s %s already installed at %s", pluginID, version, versionedPluginDir)
	}
	_, err = common.MirrorDir(unzippedPluginDir, versionedPluginDir)
	return err
}

func installPluginFromDir(unzippedPluginDir string) error {
	pd, err := plugin.GetPluginDescriptorFromJSON(filepath.Join(unzippedPluginDir, pluginJSON))
	if err != nil {
		return err
	}
	return copyPluginFilesToGaugeInstallDir(unzippedPluginDir, pd.Id, pd.Version)
}

func InstallAllPlugins() {
	manifest, err := manifest.ProjectManifest()
	if err != nil {
		logger.Fatal(err.Error())
	}
	installPluginsFromManifest(manifest)
}

func UpdatePlugins() {
	var failedPlugin []string
	for _, pluginInfo := range plugin.GetPluginsInfo() {
		logger.Info("Updating plugin '%s'", pluginInfo.Name)
		err := downloadAndInstall(pluginInfo.Name, "", fmt.Sprintf("Successfully updated plugin => %s", pluginInfo.Name))
		if err != nil {
			logger.Error(err.Error())
			failedPlugin = append(failedPlugin, pluginInfo.Name)
		}
		fmt.Println()
	}
	if len(failedPlugin) > 0 {
		logger.Error("Failed to update '%s' plugins.", strings.Join(failedPlugin, ", "))
		os.Exit(1)
	}
	logger.Info("Successfully updated all the plugins.")
	os.Exit(0)
}

func DownloadAndInstallPlugin(plugin, version, messageFormat string) {
	err := downloadAndInstall(plugin, version, fmt.Sprintf(messageFormat, plugin))
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}

func downloadAndInstall(plugin, version string, successMessage string) error {
	result := InstallPlugin(plugin, version)
	if !result.Success {
		return fmt.Errorf("%s : %s\n", plugin, result.getMessage())
	}
	if result.Warning != "" {
		logger.Warning(result.Warning)
		return nil
	}
	logger.Info(successMessage)
	return nil
}

func InstallPluginZip(zipFile string, pluginName string) {
	if err := installPluginFromZip(zipFile, pluginName); err != nil {
		logger.Fatal("Failed to install plugin. %s\n", err.Error())
	} else {
		logger.Info("Successfully installed plugin from file.")
	}
}

func installPluginsFromManifest(manifest *manifest.Manifest) {
	pluginsMap := make(map[string]bool, 0)
	pluginsMap[manifest.Language] = true
	for _, plugin := range manifest.Plugins {
		pluginsMap[plugin] = false
	}

	for pluginName, isRunner := range pluginsMap {
		if !isCompatiblePluginInstalled(pluginName, "", isRunner) {
			logger.Info("Compatible version of plugin %s not found. Installing plugin %s...", pluginName, pluginName)
			installResult := InstallPlugin(pluginName, "")
			if installResult.Success {
				logger.Info("Successfully installed the plugin %s.", pluginName)
			} else {
				logger.Error("Failed to install the %s plugin.", pluginName)
			}
		} else {
			logger.Info("Plugin %s is already installed.", pluginName)
		}
	}
}

func isCompatiblePluginInstalled(pluginName string, pluginVersion string, isRunner bool) bool {
	if isRunner {
		return IsCompatibleLanguagePluginInstalled(pluginName)
	}
	pd, err := plugin.GetPluginDescriptor(pluginName, pluginVersion)
	if err != nil {
		return false
	}
	err = version.CheckCompatibility(version.CurrentGaugeVersion, &pd.GaugeVersionSupport)
	if err != nil {
		return false
	}
	return true
}

func IsCompatibleLanguagePluginInstalled(name string) bool {
	jsonFilePath, err := plugin.GetLanguageJSONFilePath(name)
	if err != nil {
		return false
	}
	var r runner.Runner
	contents, err := common.ReadFileContents(jsonFilePath)
	if err != nil {
		return false
	}
	err = json.Unmarshal([]byte(contents), &r)
	if err != nil {
		return false
	}
	return (version.CheckCompatibility(version.CurrentGaugeVersion, &r.GaugeVersionSupport) == nil)
}

type ByDecreasingVersion []versionInstallDescription

func (a ByDecreasingVersion) Len() int      { return len(a) }
func (a ByDecreasingVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByDecreasingVersion) Less(i, j int) bool {
	version1, _ := version.ParseVersion(a[i].Version)
	version2, _ := version.ParseVersion(a[j].Version)
	return version1.IsGreaterThan(version2)
}

func AddPluginToProject(pluginName string, pluginArgs string) {
	additionalArgs := make(map[string]string)
	if pluginArgs != "" {
		// plugin args will be comma separated values
		// eg: version=1.0, foo_version = 2.41
		args := strings.Split(pluginArgs, ",")
		for _, arg := range args {
			keyValuePair := strings.Split(arg, "=")
			if len(keyValuePair) == 2 {
				additionalArgs[strings.TrimSpace(keyValuePair[0])] = strings.TrimSpace(keyValuePair[1])
			}
		}
	}
	manifest, err := manifest.ProjectManifest()
	if err != nil {
		logger.Fatal(err.Error())
	}
	if err := addPluginToTheProject(pluginName, additionalArgs, manifest); err != nil {
		logger.Fatal(fmt.Sprintf("Failed to add plugin %s to project : %s\n", pluginName, err.Error()))
	} else {
		logger.Info("Plugin %s was successfully added to the project\n", pluginName)
	}
}

func addPluginToTheProject(pluginName string, pluginArgs map[string]string, manifest *manifest.Manifest) error {
	if !plugin.IsPluginInstalled(pluginName, pluginArgs["version"]) {
		logger.Info("Plugin %s %s is not installed. Downloading the plugin.... \n", pluginName, pluginArgs["version"])
		result := InstallPlugin(pluginName, pluginArgs["version"])
		if !result.Success {
			logger.Error(result.getMessage())
		}
	}
	pd, err := plugin.GetPluginDescriptor(pluginName, pluginArgs["version"])
	if err != nil {
		return err
	}
	if plugin.IsPluginAdded(manifest, pd) {
		logger.Info("Plugin " + pd.Name + " is already added.")
		return nil
	}

	action := setupScope
	if err := plugin.SetEnvForPlugin(action, pd, manifest, pluginArgs); err != nil {
		return err
	}
	if _, err := plugin.StartPlugin(pd, action, true); err != nil {
		return err
	}
	manifest.Plugins = append(manifest.Plugins, pd.Id)
	return manifest.Save()
}
