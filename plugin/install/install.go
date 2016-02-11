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
	"regexp"
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

// InstallResult represents the result of plugin installation
type InstallResult struct {
	Error   error
	Warning string
	Success bool
}

func (installResult *InstallResult) getMessage() string {
	return installResult.Error.Error()
}

func installError(err string) InstallResult {
	return InstallResult{Error: errors.New(err), Success: false}
}

func installSuccess(warning string) InstallResult {
	return InstallResult{Warning: warning, Success: true}
}

// InstallPlugin download and install the latest plugin(if version not specified) of given plugin name
func InstallPlugin(pluginName, version string) InstallResult {
	installDescription, result := getInstallDescription(pluginName)
	defer util.RemoveTempDir()
	if !result.Success {
		return result
	}
	return installPluginWithDescription(installDescription, version)
}

func installPluginWithDescription(installDescription *installDescription, currentVersion string) InstallResult {
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

func installPluginVersion(installDesc *installDescription, versionInstallDescription *versionInstallDescription) InstallResult {
	if common.IsPluginInstalled(installDesc.Name, versionInstallDescription.Version) {
		return installSuccess(fmt.Sprintf("Plugin %s %s is already installed.", installDesc.Name, versionInstallDescription.Version))
	}

	logger.Info("Installing Plugin %s %s", installDesc.Name, versionInstallDescription.Version)
	downloadLink, err := getDownloadLink(versionInstallDescription.DownloadUrls)
	if err != nil {
		return installError(fmt.Sprintf("Could not get download link: %s", err.Error()))
	}

	tempDir := common.GetTempDir()
	defer common.Remove(tempDir)
	unzippedPluginDir, err := util.DownloadAndUnzip(downloadLink, tempDir)
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
		return fmt.Errorf("Plugin %s %s is already installed at %s", installDesc.Name, versionInstallDesc.Version, versionedPluginDir)
	}
	_, err = common.MirrorDir(pluginContents, versionedPluginDir)
	return err
}

// UninstallPlugin uninstall the given plugin of the given version
// If version is not specified, it uninstalls all the versions of given plugin
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
	logger.Errorf("%s plugin uninstallation failed", pluginName)
	logger.Fatalf(err.Error())
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

func getInstallDescription(plugin string) (*installDescription, InstallResult) {
	versionInstallDescriptionJSONFile := plugin + "-install.json"
	versionInstallDescriptionJSONUrl, result := constructPluginInstallJSONURL(plugin)
	if !result.Success {
		return nil, installError(fmt.Sprintf("Could not construct plugin install json file URL. %s", result.Error))
	}
	tempDir := common.GetTempDir()
	defer common.Remove(tempDir)

	logger.Info("Gathering metadata...")
	downloadedFile, downloadErr := util.Download(versionInstallDescriptionJSONUrl, tempDir)
	if downloadErr != nil {
		return nil, installError(fmt.Sprintf("Invalid plugin : Could not download %s file. %s", versionInstallDescriptionJSONFile, downloadErr))
	}

	return getInstallDescriptionFromJSON(downloadedFile)
}

func getInstallDescriptionFromJSON(installJSON string) (*installDescription, InstallResult) {
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

func constructPluginInstallJSONURL(plugin string) (string, InstallResult) {
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

func (installDesc *installDescription) sortVersionInstallDescriptions() {
	sort.Sort(byDecreasingVersion(installDesc.Versions))
}

// InstallPluginFromZip installs the specified plugin using the given zip file
func InstallPluginFromZip(zipFile string, pluginName string) {
	tempDir := common.GetTempDir()
	defer common.Remove(tempDir)
	unzippedPluginDir, err := common.UnzipArchive(zipFile, tempDir)
	if err != nil {
		common.Remove(tempDir)
		logger.Fatalf("Failed to install plugin %s. Reason: %s", pluginName, err.Error())
	}
	logger.Info("Plugin unzipped to => %s\n", unzippedPluginDir)

	hasPluginJSON := common.FileExists(filepath.Join(unzippedPluginDir, pluginJSON))
	pluginDirName := getVersionedPluginDirName(zipFile)
	if hasPluginJSON {
		err = installPluginFromDir(unzippedPluginDir, pluginDirName)
	} else {
		err = installRunnerFromDir(unzippedPluginDir, pluginName, pluginDirName)
	}
	if err != nil {
		common.Remove(tempDir)
		logger.Fatalf("Failed to install plugin %s. Reason: %s", pluginName, err.Error())
	}
	logger.Info("Successfully installed plugin %s.", pluginName)
}

func getVersionedPluginDirName(pluginZip string) string {
	zipFileName := path.Base(pluginZip)
	if !strings.Contains(zipFileName, "nightly") {
		return ""
	}
	re, _ := regexp.Compile("[0-9]+.[0-9]+.[0-9]+.nightly-[0-9]+-[0-9]+-[0-9]+")
	return re.FindString(zipFileName)
}

func installPluginFromDir(unzippedPluginDir string, pluginDirName string) error {
	pd, err := plugin.GetPluginDescriptorFromJSON(filepath.Join(unzippedPluginDir, pluginJSON))
	if err != nil {
		return err
	}
	if pluginDirName == "" {
		pluginDirName = pd.Version
	}
	return copyPluginFilesToGaugeInstallDir(unzippedPluginDir, pd.ID, pluginDirName)
}

func installRunnerFromDir(unzippedPluginDir string, language string, pluginDirName string) error {
	r, err := getRunnerJSONContents(filepath.Join(unzippedPluginDir, language+jsonExt))
	if err != nil {
		return err
	}
	if pluginDirName == "" {
		pluginDirName = r.Version
	}
	return copyPluginFilesToGaugeInstallDir(unzippedPluginDir, r.Id, pluginDirName)
}

func getRunnerJSONContents(file string) (*runner.Runner, error) {
	var r runner.Runner
	contents, err := common.ReadFileContents(file)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(contents), &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func copyPluginFilesToGaugeInstallDir(unzippedPluginDir string, pluginID string, versionedPluginDirName string) error {
	logger.Info("Installing Plugin %s %s", pluginID, versionedPluginDirName)

	pluginsDir, err := common.GetPrimaryPluginsInstallDir()
	if err != nil {
		return err
	}
	versionedPluginDirPath := path.Join(pluginsDir, pluginID, versionedPluginDirName)
	if common.DirExists(versionedPluginDirPath) {
		return fmt.Errorf("Plugin %s %s already installed at %s", pluginID, versionedPluginDirName, versionedPluginDirPath)
	}
	_, err = common.MirrorDir(unzippedPluginDir, versionedPluginDirPath)
	return err
}

// InstallAllPlugins install the latest version of all plugins specified in Gauge project manifest file
func InstallAllPlugins() {
	manifest, err := manifest.ProjectManifest()
	if err != nil {
		logger.Fatalf(err.Error())
	}
	installPluginsFromManifest(manifest)
}

// UpdatePlugins updates all the currently installed plugins to its latest version
func UpdatePlugins() {
	var failedPlugin []string
	for _, pluginInfo := range plugin.GetPluginsInfo() {
		logger.Info("Updating plugin '%s'", pluginInfo.Name)
		passed := HandleUpdateResult(InstallPlugin(pluginInfo.Name, ""), pluginInfo.Name, false)
		if !passed {
			failedPlugin = append(failedPlugin, pluginInfo.Name)
		}
		fmt.Println()
	}
	if len(failedPlugin) > 0 {
		logger.Fatalf("Failed to update '%s' plugins.", strings.Join(failedPlugin, ", "))
	}
	logger.Info("Successfully updated all the plugins.")
}

// HandleInstallResult handles the result of plugin Installation
func HandleInstallResult(result InstallResult, pluginName string, exitIfFailure bool) bool {
	if !result.Success {
		logger.Errorf("Failed to install plugin '%s'.\nReason: %s", pluginName, result.getMessage())
		if exitIfFailure {
			os.Exit(1)
		}
		return false
	}
	if result.Warning != "" {
		logger.Warning(result.Warning)
	}
	logger.Info("Successfully installed plugin '%s'.", pluginName)
	return true
}

// HandleUpdateResult handles the result of plugin Installation
func HandleUpdateResult(result InstallResult, pluginName string, exitIfFailure bool) bool {
	if !result.Success {
		logger.Errorf("Failed to update plugin '%s'.\nReason: %s", pluginName, result.getMessage())
		if exitIfFailure {
			os.Exit(1)
		}
		return false
	}
	if result.Warning != "" {
		logger.Warning(result.Warning)
	}
	logger.Info("Successfully updated plugin '%s'.", pluginName)
	return true
}

func installPluginsFromManifest(manifest *manifest.Manifest) {
	pluginsMap := make(map[string]bool, 0)
	pluginsMap[manifest.Language] = true
	for _, plugin := range manifest.Plugins {
		pluginsMap[plugin] = false
	}

	for pluginName, isRunner := range pluginsMap {
		if !IsCompatiblePluginInstalled(pluginName, isRunner) {
			logger.Info("Compatible version of plugin %s not found. Installing plugin %s...", pluginName, pluginName)
			HandleInstallResult(InstallPlugin(pluginName, ""), pluginName, true)
		} else {
			logger.Info("Plugin %s is already installed.", pluginName)
		}
	}
}

// IsCompatiblePluginInstalled checks if a plugin compatible to gauge is installed
func IsCompatiblePluginInstalled(pluginName string, isRunner bool) bool {
	if isRunner {
		return isCompatibleLanguagePluginInstalled(pluginName)
	}
	pd, err := plugin.GetPluginDescriptor(pluginName, "")
	if err != nil {
		return false
	}
	return version.CheckCompatibility(version.CurrentGaugeVersion, &pd.GaugeVersionSupport) == nil
}

func isCompatibleLanguagePluginInstalled(name string) bool {
	jsonFilePath, err := plugin.GetLanguageJSONFilePath(name)
	if err != nil {
		return false
	}

	r, err := getRunnerJSONContents(jsonFilePath)
	if err != nil {
		return false
	}
	return version.CheckCompatibility(version.CurrentGaugeVersion, &r.GaugeVersionSupport) == nil
}

type byDecreasingVersion []versionInstallDescription

func (a byDecreasingVersion) Len() int      { return len(a) }
func (a byDecreasingVersion) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byDecreasingVersion) Less(i, j int) bool {
	version1, _ := version.ParseVersion(a[i].Version)
	version2, _ := version.ParseVersion(a[j].Version)
	return version1.IsGreaterThan(version2)
}

// AddPluginToProject adds the given plugin to current Gauge project. It installs the plugin if not installed.
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
		logger.Fatalf(err.Error())
	}
	if err := addPluginToTheProject(pluginName, additionalArgs, manifest); err != nil {
		logger.Fatalf(fmt.Sprintf("Failed to add plugin %s to project : %s\n", pluginName, err.Error()))
	} else {
		logger.Info("Plugin %s was successfully added to the project\n", pluginName)
	}
}

func addPluginToTheProject(pluginName string, pluginArgs map[string]string, manifest *manifest.Manifest) error {
	if !plugin.IsPluginInstalled(pluginName, pluginArgs["version"]) {
		logger.Info("Plugin %s %s is not installed. Downloading the plugin.... \n", pluginName, pluginArgs["version"])
		result := InstallPlugin(pluginName, pluginArgs["version"])
		if !result.Success {
			logger.Errorf(result.getMessage())
		}
	}
	pd, err := plugin.GetPluginDescriptor(pluginName, pluginArgs["version"])
	if err != nil {
		return err
	}
	if plugin.IsPluginAdded(manifest, pd) {
		return fmt.Errorf("Plugin %s is already added.", pd.Name)
	}
	manifest.Plugins = append(manifest.Plugins, pd.ID)
	return manifest.Save()
}
