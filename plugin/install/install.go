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
	jsonExt    = ".json"
	x86        = "386"
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
	Info    string
	Success bool
	Skipped bool
}

func (installResult *InstallResult) getMessage() string {
	return installResult.Error.Error()
}

func installError(err error) InstallResult {
	return InstallResult{Error: err, Success: false}
}

func installSuccess(info string) InstallResult {
	return InstallResult{Info: info, Success: true}
}
func installSkipped(warning, info string) InstallResult {
	return InstallResult{Warning: warning, Info: info, Skipped: true}
}

// GaugePlugin represents any plugin to Gauge. It can be an language runner or any other plugin.
type GaugePlugin struct {
	ID          string
	Version     string
	Description string
	PreInstall  struct {
		Windows []string
		Linux   []string
		Darwin  []string
	}
	PostInstall struct {
		Windows []string
		Linux   []string
		Darwin  []string
	}
	PreUnInstall struct {
		Windows []string
		Linux   []string
		Darwin  []string
	}
	PostUnInstall struct {
		Windows []string
		Linux   []string
		Darwin  []string
	}
	GaugeVersionSupport version.VersionSupport
}

func getGoArch() string {
	arch := runtime.GOARCH
	if arch == x86 {
		return "x86"
	}
	return "x86_64"
}

func isPlatformIndependent(zipfile string) bool {
	re, _ := regexp.Compile("-([a-z]*)\\.")
	return !re.MatchString(zipfile)
}

func isOsOSCompatible(zipfile string) bool {
	os := runtime.GOOS
	arch := getGoArch()
	return strings.Contains(zipfile, fmt.Sprintf("%s.%s", os, arch))
}

// InstallPluginFromZipFile installs plugin from given zip file
func InstallPluginFromZipFile(zipFile string, pluginName string) InstallResult {
	tempDir := common.GetTempDir()
	defer common.Remove(tempDir)
	unzippedPluginDir, err := common.UnzipArchive(zipFile, tempDir)
	if err != nil {
		return installError(err)
	}
	if !isPlatformIndependent(zipFile) && !isOsOSCompatible(zipFile) {
		err := fmt.Errorf("Provided plugin is not compatible with OS %s_%s.", runtime.GOOS, runtime.GOARCH)
		return installError(err)
	}
	gp, err := parsePluginJSON(unzippedPluginDir, pluginName)
	if err != nil || gp.ID != pluginName {
		err := fmt.Errorf("Provided zip file is not a valid plugin of %s.", pluginName)
		return installError(err)
	}
	if err = runPlatformCommands(gp.PreInstall, unzippedPluginDir); err != nil {
		return installError(err)
	}

	pluginInstallDir, err := getPluginInstallDir(gp.ID, getVersionedPluginDirName(zipFile))
	if err != nil {
		return installError(err)
	}

	// copy files to gauge plugin install location
	logger.Debugf("Installing plugin %s %s", gp.ID, filepath.Base(pluginInstallDir))
	if _, err = common.MirrorDir(unzippedPluginDir, pluginInstallDir); err != nil {
		return installError(err)
	}

	if err = runPlatformCommands(gp.PostInstall, pluginInstallDir); err != nil {
		return installError(err)
	}
	return installSuccess("")
}

func getPluginInstallDir(pluginID, pluginDirName string) (string, error) {
	pluginsDir, err := common.GetPrimaryPluginsInstallDir()
	if err != nil {
		return "", err
	}
	pluginDirPath := filepath.Join(pluginsDir, pluginID, pluginDirName)
	if common.DirExists(pluginDirPath) {
		return "", fmt.Errorf("Plugin %s %s already installed at %s", pluginID, pluginDirName, pluginDirPath)
	}
	return pluginDirPath, nil
}

func parsePluginJSON(pluginDir, pluginName string) (*GaugePlugin, error) {
	var file string
	if common.FileExists(filepath.Join(pluginDir, pluginName+jsonExt)) {
		file = filepath.Join(pluginDir, pluginName+jsonExt)
	} else {
		file = filepath.Join(pluginDir, pluginJSON)
	}

	var gp GaugePlugin
	contents, err := common.ReadFileContents(file)
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal([]byte(contents), &gp); err != nil {
		return nil, err
	}
	return &gp, nil
}

// Plugin download and install the latest plugin(if version not specified) of given plugin name
func Plugin(pluginName, version string) InstallResult {
	logger.Debugf("Gathering metadata for %s", pluginName)
	installDescription, result := getInstallDescription(pluginName, false)
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
			return installError(err)
		}
		if compatibilityError := version.CheckCompatibility(version.CurrentGaugeVersion, &versionInstallDescription.GaugeVersionSupport); compatibilityError != nil {
			return installError(fmt.Errorf("Plugin Version %s-%s is not supported for gauge %s : %s", installDescription.Name, versionInstallDescription.Version, version.CurrentGaugeVersion.String(), compatibilityError.Error()))
		}
	} else {
		versionInstallDescription, err = installDescription.getLatestCompatibleVersionTo(version.CurrentGaugeVersion)
		if err != nil {
			return installError(fmt.Errorf("Could not find compatible version for plugin %s. : %s", installDescription.Name, err))
		}
	}
	return installPluginVersion(installDescription, versionInstallDescription)
}

func installPluginVersion(installDesc *installDescription, versionInstallDescription *versionInstallDescription) InstallResult {
	if common.IsPluginInstalled(installDesc.Name, versionInstallDescription.Version) {
		return installSkipped("", fmt.Sprintf("Plugin %s %s is already installed.", installDesc.Name, versionInstallDescription.Version))
	}

	downloadLink, err := getDownloadLink(versionInstallDescription.DownloadUrls)
	if err != nil {
		return installError(fmt.Errorf("Could not get download link: %s", err.Error()))
	}

	tempDir := common.GetTempDir()
	defer common.Remove(tempDir)
	logger.Debugf("Downloading %s", filepath.Base(downloadLink))
	pluginZip, err := util.Download(downloadLink, tempDir, "", false)
	if err != nil {
		return installError(fmt.Errorf("Failed to download the plugin. %s", err.Error()))
	}
	return InstallPluginFromZipFile(pluginZip, installDesc.Name)
}

func runPlatformCommands(commands platformSpecificCommand, workingDir string) error {
	command := []string{}
	switch runtime.GOOS {
	case "windows":
		command = commands.Windows
		break
	case "darwin":
		command = commands.Darwin
		break
	default:
		command = commands.Linux
		break
	}

	if len(command) == 0 {
		return nil
	}

	logger.Debugf("Running plugin hook command => %s", command)
	cmd, err := common.ExecuteSystemCommand(command, workingDir, os.Stdout, os.Stderr)

	if err != nil {
		return err
	}

	return cmd.Wait()
}

// UninstallPlugin uninstall the given plugin of the given uninstallVersion
// If uninstallVersion is not specified, it uninstalls all the versions of given plugin
func UninstallPlugin(pluginName string, uninstallVersion string) {
	pluginsHome, err := common.GetPrimaryPluginsInstallDir()
	if err != nil {
		logger.Fatalf("Failed to uninstall plugin %s. %s", pluginName, err.Error())
	}
	if !common.DirExists(filepath.Join(pluginsHome, pluginName, uninstallVersion)) {
		logger.Errorf("Plugin %s not found.", strings.TrimSpace(pluginName+" "+uninstallVersion))
		os.Exit(0)
	}
	var failed bool
	pluginsDir := filepath.Join(pluginsHome, pluginName)
	filepath.Walk(pluginsDir, func(dir string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() && dir != pluginsDir {
			if matchesUninstallVersion(filepath.Base(dir), uninstallVersion) {
				if err := uninstallVersionOfPlugin(dir, pluginName, filepath.Base(dir)); err != nil {
					logger.Errorf("Failed to uninstall plugin %s %s. %s", pluginName, uninstallVersion, err.Error())
					failed = true
				}
			}
		}
		return nil
	})
	if failed {
		os.Exit(1)
	}
	if uninstallVersion == "" {
		if err := os.RemoveAll(pluginsDir); err != nil {
			logger.Fatalf("Failed to remove directory %s. %s", pluginsDir, err.Error())
		}
	}
}

func matchesUninstallVersion(pluginDirPath, uninstallVersion string) bool {
	if uninstallVersion == "" {
		return true
	}
	return pluginDirPath == uninstallVersion
}

func uninstallVersionOfPlugin(pluginDir, pluginName, uninstallVersion string) error {
	gp, err := parsePluginJSON(pluginDir, pluginName)
	if err != nil {
		return err
	}
	if err := runPlatformCommands(gp.PreUnInstall, pluginDir); err != nil {
		return err
	}

	if err := os.RemoveAll(pluginDir); err != nil {
		return err
	}
	if err := runPlatformCommands(gp.PostUnInstall, path.Dir(pluginDir)); err != nil {
		return err
	}
	logger.Infof("Successfully uninstalled plugin %s %s.", pluginName, uninstallVersion)
	return nil
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

func getInstallDescription(plugin string, silent bool) (*installDescription, InstallResult) {
	versionInstallDescriptionJSONFile := plugin + "-install.json"
	versionInstallDescriptionJSONUrl, result := constructPluginInstallJSONURL(plugin)
	if !result.Success {
		return nil, installError(fmt.Errorf("Could not construct plugin install json file URL. %s", result.Error))
	}
	tempDir := common.GetTempDir()
	defer common.Remove(tempDir)

	downloadedFile, downloadErr := util.Download(versionInstallDescriptionJSONUrl, tempDir, versionInstallDescriptionJSONFile, silent)
	if downloadErr != nil {
		logger.Debugf("Failed to download %s file: %s", versionInstallDescriptionJSONFile, downloadErr)
		return nil, installError(fmt.Errorf("Invalid plugin. Could not download %s file.", versionInstallDescriptionJSONFile))
	}

	return getInstallDescriptionFromJSON(downloadedFile)
}

func getInstallDescriptionFromJSON(installJSON string) (*installDescription, InstallResult) {
	InstallJSONContents, readErr := common.ReadFileContents(installJSON)
	if readErr != nil {
		return nil, installError(readErr)
	}
	installDescription := &installDescription{}
	if err := json.Unmarshal([]byte(InstallJSONContents), installDescription); err != nil {
		return nil, installError(err)
	}
	return installDescription, installSuccess("")
}

func constructPluginInstallJSONURL(p string) (string, InstallResult) {
	repoURL := config.GaugeRepositoryUrl()
	if repoURL == "" {
		return "", installError(fmt.Errorf("Could not find gauge repository url from configuration."))
	}
	JSONURL := fmt.Sprintf("%s/%s", repoURL, p)
	if qp := plugin.QueryParams(); qp != "" {
		JSONURL += qp
	}
	return JSONURL, installSuccess("")
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

func getVersionedPluginDirName(pluginZip string) string {
	zipFileName := filepath.Base(pluginZip)
	if !strings.Contains(zipFileName, "nightly") {
		re, _ := regexp.Compile("[0-9]+\\.[0-9]+\\.[0-9]+")
		return re.FindString(zipFileName)
	}
	re, _ := regexp.Compile("[0-9]+\\.[0-9]+\\.[0-9]+\\.nightly-[0-9]+-[0-9]+-[0-9]+")
	return re.FindString(zipFileName)
}

func getRunnerJSONContents(file string) (*runner.RunnerInfo, error) {
	var r runner.RunnerInfo
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

// AllPlugins install the latest version of all plugins specified in Gauge project manifest file
func AllPlugins() {
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
		logger.Debugf("Updating plugin '%s'", pluginInfo.Name)
		passed := HandleUpdateResult(Plugin(pluginInfo.Name, ""), pluginInfo.Name, false)
		if !passed {
			failedPlugin = append(failedPlugin, pluginInfo.Name)
		}
	}
	if len(failedPlugin) > 0 {
		logger.Fatalf("Failed to update '%s' plugins.", strings.Join(failedPlugin, ", "))
	}
	logger.Infof("Successfully updated all the plugins.")
}

// HandleInstallResult handles the result of plugin Installation
// TODO: Merge both HandleInstallResult and HandleUpdateResult, eliminate boolean exitIfFailure
func HandleInstallResult(result InstallResult, pluginName string, exitIfFailure bool) bool {
	if result.Info != "" {
		logger.Debugf(result.Info)
	}
	if result.Warning != "" {
		logger.Warningf(result.Warning)
	}
	if result.Skipped {
		return true
	}
	if !result.Success {
		logger.Errorf("Failed to install plugin '%s'.\nReason: %s", pluginName, result.getMessage())
		if exitIfFailure {
			os.Exit(1)
		}
		return false
	}

	logger.Infof("Successfully installed plugin '%s'.", pluginName)
	return true
}

// HandleUpdateResult handles the result of plugin Installation
func HandleUpdateResult(result InstallResult, pluginName string, exitIfFailure bool) bool {
	if result.Info != "" {
		logger.Debugf(result.Info)
	}
	if result.Warning != "" {
		logger.Warningf(result.Warning)
	}
	if result.Skipped {
		return true
	}
	if !result.Success {
		logger.Errorf("Failed to update plugin '%s'.\nReason: %s", pluginName, result.getMessage())
		if exitIfFailure {
			os.Exit(1)
		}
		return false
	}
	logger.Infof("Successfully updated plugin '%s'.", pluginName)
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
			logger.Infof("Compatible version of plugin %s not found. Installing plugin %s...", pluginName, pluginName)
			HandleInstallResult(Plugin(pluginName, ""), pluginName, false)
		} else {
			logger.Debugf("Plugin %s is already installed.", pluginName)
		}
	}
}

// IsCompatiblePluginInstalled checks if a plugin compatible to gauge is installed
// TODO: This always checks if latest installed version of a given plugin is compatible. This should also check for older versions.
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

// AddPluginToProject adds the given plugin to current Gauge project.
func AddPluginToProject(pluginName string) error {
	m, err := manifest.ProjectManifest()
	if err != nil {
		return nil
	}
	if plugin.IsLanguagePlugin(pluginName) {
		return nil
	}
	pd, err := plugin.GetPluginDescriptor(pluginName, "")
	if err != nil {
		return err
	}
	if plugin.IsPluginAdded(m, pd) {
		logger.Debugf("Plugin %s is already added.", pd.Name)
		return nil
	}
	m.Plugins = append(m.Plugins, pd.ID)
	if err = m.Save(); err != nil {
		return err
	}
	logger.Infof("Plugin %s was successfully added to the project\n", pluginName)
	return nil
}
