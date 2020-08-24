/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

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
	"github.com/getgauge/gauge/plugin/pluginInfo"
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
	Version string
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
	re, err := regexp.Compile(`-([a-z]*)\.`)
	if err != nil {
		logger.Errorf(false, "unable to compile regex '-([a-z]*)\\.': %s", err.Error())
	}
	return !re.MatchString(zipfile)
}

func isOsOSCompatible(zipfile string) bool {
	os := runtime.GOOS
	arch := getGoArch()
	return strings.Contains(zipfile, fmt.Sprintf("%s.%s", os, arch))
}

// InstallPluginFromZipFile installs plugin from given zip file
func InstallPluginFromZipFile(zipFile string, pluginName string) InstallResult {
	if !isPlatformIndependent(zipFile) && !isOsOSCompatible(zipFile) {
		err := fmt.Errorf("provided plugin is not compatible with OS %s %s", runtime.GOOS, runtime.GOARCH)
		return installError(err)
	}
	tempDir := common.GetTempDir()
	defer func() {
		err := common.Remove(tempDir)
		if err != nil {
			logger.Errorf(false, "unable to remove temp directory: %s", err.Error())
		}
	}()

	unzippedPluginDir, err := common.UnzipArchive(zipFile, tempDir)
	if err != nil {
		return installError(err)
	}
	gp, err := parsePluginJSON(unzippedPluginDir, pluginName)
	if err != nil {
		return installError(err)
	}
	if gp.ID != pluginName {
		err := fmt.Errorf("provided zip file is not a valid plugin of %s", pluginName)
		return installError(err)
	}
	if err = runPlatformCommands(gp.PreInstall, unzippedPluginDir); err != nil {
		return installError(err)
	}

	if err = runPlatformCommands(gp.PostInstall, unzippedPluginDir); err != nil {
		return installError(err)
	}

	pluginInstallDir, err := getPluginInstallDir(gp.ID, getVersionedPluginDirName(zipFile))
	if err != nil {
		return installError(err)
	}

	// copy files to gauge plugin install location
	logger.Debugf(true, "Installing plugin %s %s", gp.ID, filepath.Base(pluginInstallDir))
	if _, err = common.MirrorDir(unzippedPluginDir, pluginInstallDir); err != nil {
		return installError(err)
	}
	installResult := installSuccess("")
	installResult.Version = gp.Version
	return installResult
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
func Plugin(pluginName, version string, silent bool) InstallResult {
	logger.Debugf(true, "Gathering metadata for %s", pluginName)
	installDescription, result := getInstallDescription(pluginName, false)
	defer util.RemoveTempDir()
	if !result.Success {
		return result
	}
	return installPluginWithDescription(installDescription, version, silent)
}

func installPluginWithDescription(installDescription *installDescription, currentVersion string, silent bool) InstallResult {
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
	return installPluginVersion(installDescription, versionInstallDescription, silent)
}

func installPluginVersion(installDesc *installDescription, versionInstallDescription *versionInstallDescription, silent bool) InstallResult {
	if common.IsPluginInstalled(installDesc.Name, versionInstallDescription.Version) {
		return installSkipped("", fmt.Sprintf("Plugin %s %s is already installed.", installDesc.Name, versionInstallDescription.Version))
	}

	downloadLink, err := getDownloadLink(versionInstallDescription.DownloadUrls)
	if err != nil {
		return installError(fmt.Errorf("Could not get download link: %s", err.Error()))
	}

	tempDir := common.GetTempDir()
	defer func() {
		err := common.Remove(tempDir)
		if err != nil {
			logger.Errorf(false, "unable to remove temp directory: %s", err.Error())
		}
	}()
	pluginZip, err := util.Download(downloadLink, tempDir, "", silent)
	if err != nil {
		return installError(fmt.Errorf("Failed to download the plugin. %s", err.Error()))
	}
	res := InstallPluginFromZipFile(pluginZip, installDesc.Name)
	res.Version = versionInstallDescription.Version
	return res
}

func runPlatformCommands(commands platformSpecificCommand, workingDir string) error {
	var command []string
	switch runtime.GOOS {
	case "windows":
		command = commands.Windows
	case "darwin":
		command = commands.Darwin
	default:
		command = commands.Linux
	}

	if len(command) == 0 {
		return nil
	}

	logger.Debugf(true, "Running plugin hook command => %s", command)
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
		logger.Fatalf(true, "Failed to uninstall plugin %s. %s", pluginName, err.Error())
	}
	if !common.DirExists(filepath.Join(pluginsHome, pluginName, uninstallVersion)) {
		logger.Errorf(true, "Plugin %s not found.", strings.TrimSpace(pluginName+" "+uninstallVersion))
		os.Exit(0)
	}
	var failed bool
	pluginsDir := filepath.Join(pluginsHome, pluginName)
	err = filepath.Walk(pluginsDir, func(dir string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() && dir != pluginsDir {
			if matchesUninstallVersion(filepath.Base(dir), uninstallVersion) {
				if err := uninstallVersionOfPlugin(dir, pluginName, filepath.Base(dir)); err != nil {
					failed = true
					return fmt.Errorf("failed to uninstall plugin %s %s. %s", pluginName, uninstallVersion, err.Error())
				}
			}
		}
		return nil
	})
	if err != nil {
		logger.Error(true, err.Error())
	}
	if failed {
		os.Exit(1)
	}
	if uninstallVersion == "" {
		if err := os.RemoveAll(pluginsDir); err != nil {
			logger.Fatalf(true, "Failed to remove directory %s. %s", pluginsDir, err.Error())
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
		logger.Infof(true, "Unable to read plugin's metadata, removing %s", pluginDir)
		return os.RemoveAll(pluginDir)
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
	logger.Infof(true, "Successfully uninstalled plugin %s %s.", pluginName, uninstallVersion)
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
	case "darwin":
		downloadLink = platformLinks.Darwin
	default:
		downloadLink = platformLinks.Linux
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
	defer func() {
		err := common.Remove(tempDir)
		if err != nil {
			logger.Errorf(false, "unable to remove temp directory: %s", err.Error())
		}
	}()

	downloadedFile, downloadErr := util.Download(versionInstallDescriptionJSONUrl, tempDir, versionInstallDescriptionJSONFile, silent)
	if downloadErr != nil {
		logger.Debugf(true, "Failed to download %s file: %s", versionInstallDescriptionJSONFile, downloadErr)
		return nil, installError(fmt.Errorf("Invalid plugin name or there's a network issue while fetching plugin details."))
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
	jsonURL := fmt.Sprintf("%s/%s", repoURL, p)
	if qp := plugin.QueryParams(); qp != "" {
		jsonURL += qp
	}
	return jsonURL, installSuccess("")
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
		rStr := `[0-9]+\.[0-9]+\.[0-9]+`
		re, err := regexp.Compile(rStr)
		if err != nil {
			logger.Errorf(false, "Unable to compile regex %s: %s", rStr, err.Error())
		}
		return re.FindString(zipFileName)
	}
	rStr := `[0-9]+\.[0-9]+\.[0-9]+\.nightly-[0-9]+-[0-9]+-[0-9]+`
	re, err := regexp.Compile(rStr)
	if err != nil {
		logger.Errorf(false, "Unable to compile regex %s: %s", rStr, err.Error())
	}
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
func AllPlugins(silent, languageOnly bool) {
	manifest, err := manifest.ProjectManifest()
	if err != nil {
		logger.Fatalf(true, err.Error())
	}
	installPluginsFromManifest(manifest, silent, languageOnly)
}

// UpdatePlugins updates all the currently installed plugins to its latest version
func UpdatePlugins(silent bool) {
	var failedPlugin []string
	pluginInfos, err := pluginInfo.GetPluginsInfo()
	if err != nil {
		logger.Infof(true, err.Error())
		os.Exit(0)
	}
	for _, pluginInfo := range pluginInfos {
		logger.Debugf(true, "Updating plugin '%s'", pluginInfo.Name)
		passed := HandleUpdateResult(Plugin(pluginInfo.Name, "", silent), pluginInfo.Name, false)
		if !passed {
			failedPlugin = append(failedPlugin, pluginInfo.Name)
		}
	}
	if len(failedPlugin) > 0 {
		logger.Fatalf(true, "Failed to update '%s' plugins.", strings.Join(failedPlugin, ", "))
	}
	logger.Infof(true, "Successfully updated all the plugins.")
}

// HandleInstallResult handles the result of plugin Installation
// TODO: Merge both HandleInstallResult and HandleUpdateResult, eliminate boolean exitIfFailure
func HandleInstallResult(result InstallResult, pluginName string, exitIfFailure bool) bool {
	if result.Info != "" {
		logger.Debugf(true, result.Info)
	}
	if result.Warning != "" {
		logger.Warningf(true, result.Warning)
	}
	if result.Skipped {
		return true
	}
	if !result.Success {
		if result.Version != "" {
			logger.Errorf(true, "Failed to install plugin '%s' version %s.\nReason: %s", pluginName, result.Version, result.getMessage())
		} else {
			logger.Errorf(true, "Failed to install plugin '%s'.\nReason: %s", pluginName, result.getMessage())
		}
		if exitIfFailure {
			os.Exit(1)
		}
		return false
	}
	if result.Version != "" {
		logger.Infof(true, "Successfully installed plugin '%s' version %s", pluginName, result.Version)
	} else {
		logger.Infof(true, "Successfully installed plugin '%s'", pluginName)
	}
	return true
}

// HandleUpdateResult handles the result of plugin Installation
func HandleUpdateResult(result InstallResult, pluginName string, exitIfFailure bool) bool {
	if result.Info != "" {
		logger.Debugf(true, result.Info)
	}
	if result.Warning != "" {
		logger.Warningf(true, result.Warning)
	}
	if result.Skipped {
		logger.Infof(true, "Plugin '%s' is up to date.", pluginName)
		return true
	}
	if !result.Success {
		logger.Errorf(true, "Failed to update plugin '%s'.\nReason: %s", pluginName, result.getMessage())
		if exitIfFailure {
			os.Exit(1)
		}
		return false
	}
	logger.Infof(true, "Successfully updated plugin '%s'.", pluginName)
	return true
}

func installPluginsFromManifest(manifest *manifest.Manifest, silent, languageOnly bool) {
	pluginsMap := make(map[string]bool)
	pluginsMap[manifest.Language] = true
	if !languageOnly {
		for _, plugin := range manifest.Plugins {
			pluginsMap[plugin] = false
		}
	}

	for pluginName, isRunner := range pluginsMap {
		if !IsCompatiblePluginInstalled(pluginName, isRunner) {
			logger.Infof(true, "Compatible version of plugin %s not found. Installing plugin %s...", pluginName, pluginName)
			HandleInstallResult(Plugin(pluginName, "", silent), pluginName, false)
		} else {
			logger.Debugf(true, "Plugin %s is already installed.", pluginName)
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
		logger.Debugf(true, "Plugin %s is already added.", pd.Name)
		return nil
	}
	m.Plugins = append(m.Plugins, pd.ID)
	if err = m.Save(); err != nil {
		return err
	}
	logger.Infof(true, "Plugin %s was successfully added to the project\n", pluginName)
	return nil
}
