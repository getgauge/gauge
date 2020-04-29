/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

// Package common functions shared across all files
package common

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	properties "github.com/dmotylev/goproperties"
)

const (
	ManifestFile            = "manifest.json"
	PluginJSONFile          = "plugin.json"
	NewDirectoryPermissions = 0755
	NewFilePermissions      = 0644
	DefaultEnvFileName      = "default.properties"
	EnvDirectoryName        = "env"
	DefaultEnvDir           = "default"
	ProductName             = "gauge"
	DotGauge                = ".gauge"
	config                  = "config"
	SpecsDirectoryName      = "specs"
	ConceptFileExtension    = ".cpt"
	Plugins                 = "plugins"
	appData                 = "APPDATA"
	GaugePropertiesFile     = "gauge.properties"
	NightlyDatelayout       = "2006-01-02"
)

const (
	GaugeProjectRootEnv      = "GAUGE_PROJECT_ROOT"
	GaugeHome                = "GAUGE_HOME" //specifies the plugin installation path if installs to non-standard location
	GaugePortEnvName         = "GAUGE_PORT" // user specifies this to use a specific port
	GaugeInternalPortEnvName = "GAUGE_INTERNAL_PORT"
	APIPortEnvVariableName   = "GAUGE_API_PORT"
	APIV2PortEnvVariableName = "GAUGE_API_V2_PORT"
	GaugeDebugOptsEnv        = "GAUGE_DEBUG_OPTS" //specify the debug options to be used while launching the runner
)

// Property represents a single property in the properties file
type Property struct {
	Name         string
	Comment      string
	DefaultValue string
}

// GetProjectRoot returns the Gauge project root
// A project root is where a manifest.json files exists
// this routine keeps going upwards searching for manifest.json
func GetProjectRoot() (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("Failed to find project root directory. %s\n", err.Error())
	}
	return findManifestInPath(pwd)
}

func findManifestInPath(pwd string) (string, error) {
	wd, err := filepath.Abs(pwd)
	if err != nil {
		return "", fmt.Errorf("Failed to find project directory: %s", err)
	}
	manifestExists := func(dir string) bool {
		return FileExists(filepath.Join(dir, ManifestFile))
	}
	dir := wd

	for {
		if manifestExists(dir) {
			return dir, nil
		}
		if dir == filepath.Clean(fmt.Sprintf("%c", os.PathSeparator)) || dir == "" {
			return "", fmt.Errorf("Failed to find project directory")
		}
		oldDir := dir
		dir = filepath.Clean(fmt.Sprintf("%s%c..", dir, os.PathSeparator))
		if dir == oldDir {
			return "", fmt.Errorf("Failed to find project directory")
		}
	}
}

// GetDirInProject returns the path of a particular directory in a Gauge project
func GetDirInProject(dirName string, specPath string) (string, error) {
	projectRoot, err := GetProjectRootFromSpecPath(specPath)
	if err != nil {
		return "", err
	}
	requiredDir := filepath.Join(projectRoot, dirName)
	if !DirExists(requiredDir) {
		return "", fmt.Errorf("Could not find %s directory. %s does not exist", dirName, requiredDir)
	}

	return requiredDir, nil
}

// GetProjectRootFromSpecPath returns the path of the project root from a given spec path
func GetProjectRootFromSpecPath(specPath string) (string, error) {
	projectRoot, err := GetProjectRoot()
	if err != nil {
		dir, _ := filepath.Split(specPath)
		fullPath, pathErr := filepath.Abs(dir)
		if pathErr != nil {
			return "", fmt.Errorf("Unable to get absolute path to specifications. %s", err)
		}
		return findManifestInPath(fullPath)
	}
	return projectRoot, err
}

// GetDefaultPropertiesFile returns the path of the default.properties file in the default env
func GetDefaultPropertiesFile() (string, error) {
	envDir, err := GetDirInProject(EnvDirectoryName, "")
	if err != nil {
		return "", err
	}
	defaultEnvFile := filepath.Join(envDir, DefaultEnvDir, DefaultEnvFileName)
	if !FileExists(defaultEnvFile) {
		return "", fmt.Errorf("Default environment file does not exist: %s \n", defaultEnvFile)
	}
	return defaultEnvFile, nil
}

// AppendProperties appends the given properties to the end of the properties file.
func AppendProperties(propertiesFile string, properties ...*Property) error {
	file, err := os.OpenFile(propertiesFile, os.O_RDWR|os.O_APPEND, NewFilePermissions)
	if err != nil {
		return err
	}
	for _, property := range properties {
		file.WriteString(fmt.Sprintf("\n%s\n", property.String()))
	}
	return file.Close()
}

// FindFilesInDir returns a list of files for which isValidFile func returns true
func FindFilesInDir(dirPath string, isValidFile func(path string) bool, shouldSkip func(path string, f os.FileInfo) bool) []string {
	files := []string{}
	filepath.Walk(dirPath, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if shouldSkip(path, f) {
			return filepath.SkipDir
		}
		if !f.IsDir() && isValidFile(path) {
			files = append(files, path)
		}
		return nil
	})
	return files
}

// GetConfigurationPrefix returns the configuration directory prefix
// $GAUGE_HOME or $home/.gauge/config
func GetConfigurationDir() (string, error) {
	gaugeHome := os.Getenv(GaugeHome)
	if gaugeHome != "" {
		return filepath.Join(gaugeHome, config), nil
	}

	var configDir string

	if isWindows() {
		appDataPath := os.Getenv(appData)
		if appDataPath == "" {
			return "", fmt.Errorf("Cannot locate gauge shared file. Could not find App Data directory.")
		}
		configDir = filepath.Join(appDataPath, ProductName, config)
	} else {
		home := os.Getenv("HOME")
		configDir = filepath.Join(home, DotGauge, config)
	}

	if configDir != "" {
		return configDir, nil
	}

	return "", fmt.Errorf("Can't find configuration files")
}

// GetInstallationPrefix returns the installation directory prefix
// /usr or /usr/local
func GetInstallationPrefix() (string, error) {
	var possibleInstallationPrefixes []string
	if isWindows() {
		programFilesPath := os.Getenv("PROGRAMFILES")
		if programFilesPath == "" {
			return "", fmt.Errorf("Cannot locate gauge shared file. Could not find Program Files directory.")
		}
		possibleInstallationPrefixes = []string{filepath.Join(programFilesPath, ProductName)}
	} else {
		possibleInstallationPrefixes = []string{"/usr/local", "/usr"}
	}

	for _, p := range possibleInstallationPrefixes {
		if FileExists(filepath.Join(p, "bin", ExecutableName())) {
			return p, nil
		}
	}

	return "", fmt.Errorf("Can't find installation files")
}

// ExecutableName returns the Gauge executable name based on user's OS
func ExecutableName() string {
	if isWindows() {
		return "gauge.exe"
	}
	return "gauge"
}

// GetSkeletonFilePath returns the path skeleton file
func GetSkeletonFilePath(filename string) (string, error) {
	searchPath, err := GetConfigurationDir()
	if err != nil {
		return "", err
	}
	skelFile := filepath.Join(searchPath, "skel", filename)
	if FileExists(skelFile) {
		return skelFile, nil
	}

	return "", fmt.Errorf("Failed to find the skeleton file: %s", filename)
}

// GetPluginsInstallDir returns the plugin installation directory
func GetPluginsInstallDir(pluginName string) (string, error) {
	pluginInstallPrefixes, err := GetPluginInstallPrefixes()
	if err != nil {
		return "", err
	}

	for _, prefix := range pluginInstallPrefixes {
		if SubDirectoryExists(prefix, pluginName) {
			return prefix, nil
		}
	}
	return "", fmt.Errorf("Plugin '%s' not installed on following locations : %s", pluginName, pluginInstallPrefixes)
}

// SubDirectoryExists checks if a dir for given plugin exists in the plugin directory
func SubDirectoryExists(pluginDir string, pluginName string) bool {
	files, err := ioutil.ReadDir(pluginDir)
	if err != nil {
		return false
	}

	for _, f := range files {
		if f.Name() == pluginName && f.IsDir() {
			return true
		}
	}
	return false
}

// GetPluginInstallPrefixes returns the installation prefix path for the plugins
func GetPluginInstallPrefixes() ([]string, error) {
	primaryPluginInstallDir, err := GetPrimaryPluginsInstallDir()
	if err != nil {
		return nil, err
	}
	return []string{primaryPluginInstallDir}, nil
}

// GetGaugeHomeDirectory returns GAUGE_HOME. This is where all the plugins are installed
func GetGaugeHomeDirectory() (string, error) {
	customPluginRoot := os.Getenv(GaugeHome)
	if customPluginRoot != "" {
		return customPluginRoot, nil
	}
	if isWindows() {
		appDataDir := os.Getenv(appData)
		if appDataDir == "" {
			return "", fmt.Errorf("Failed to find plugin installation path. Could not get APPDATA")
		}
		return filepath.Join(appDataDir, ProductName), nil
	}
	userHome := getUserHomeFromEnv()
	return filepath.Join(userHome, DotGauge), nil
}

// GetPrimaryPluginsInstallDir returns the primary plugin installation dir
func GetPrimaryPluginsInstallDir() (string, error) {
	gaugeHome, err := GetGaugeHomeDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(gaugeHome, Plugins), nil
}

// IsPluginInstalled checks if the given Gauge plugin version is installed
func IsPluginInstalled(name, version string) bool {
	pluginsDir, err := GetPluginsInstallDir(name)
	if err != nil {
		return false
	}
	return DirExists(filepath.Join(pluginsDir, name, version))
}

// GetGaugeConfiguration parsed the gauge.properties file from GAUGE_HOME and returns the contents
func GetGaugeConfiguration() (properties.Properties, error) {
	fmt.Println("[DEPRECATED]: Please use GetGaugeConfigurationFor(propertiesFileName)")
	return GetGaugeConfigurationFor(GaugePropertiesFile)
}

// GetGaugeConfiguration parsed the given properties file from GAUGE_HOME and returns the contents
func GetGaugeConfigurationFor(propertiesFileName string) (properties.Properties, error) {
	configDir, err := GetConfigurationDir()
	if err != nil {
		return nil, err
	}
	propertiesFile := filepath.Join(configDir, propertiesFileName)
	config, err := properties.Load(propertiesFile)
	if err != nil {
		return nil, err
	}
	return config, nil
}

// ReadFileContents returns the contents of the file
func ReadFileContents(file string) (string, error) {
	if !FileExists(file) {
		return "", fmt.Errorf("File %s doesn't exist.", file)
	}
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("Failed to read the file %s.", file)
	}
	return strings.TrimLeft(string(bytes), "\xef\xbb\xbf"), nil
}

// FileExists checks if the given file exists
func FileExists(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

// DirExists checks if the given directory exists
func DirExists(dirPath string) bool {
	if stat, err := os.Stat(dirPath); err == nil && stat.IsDir() {
		return true
	}
	return false
}

// MirrorDir creates an exact copy of source dir to destination dir
// Modified version of bradfitz's camlistore (https://github.com/bradfitz/camlistore/blob/master/make.go)
func MirrorDir(src, dst string) ([]string, error) {
	var filesAdded []string
	err := filepath.Walk(src, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		suffix, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("Failed to find Rel(%q, %q): %v", src, path, err)
		}

		err = MirrorFile(path, filepath.Join(dst, suffix))
		filesAdded = append(filesAdded, suffix)
		return err
	})
	return filesAdded, err
}

// MirrorFile creates an exact copy of source file to destination file
// Modified version of bradfitz's camlistore (https://github.com/bradfitz/camlistore/blob/master/make.go)
func MirrorFile(src, dst string) error {
	sfi, err := os.Stat(src)
	if err != nil {
		return err
	}
	if sfi.Mode()&os.ModeType != 0 {
		log.Fatalf("mirrorFile can't deal with non-regular file %s", src)
	}
	dfi, err := os.Stat(dst)
	if err == nil &&
		isExecMode(sfi.Mode()) == isExecMode(dfi.Mode()) &&
		(dfi.Mode()&os.ModeType == 0) &&
		dfi.Size() == sfi.Size() &&
		dfi.ModTime().Unix() == sfi.ModTime().Unix() {
		// Seems to not be modified.
		return nil
	}

	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}

	df, err := os.Create(dst)
	if err != nil {
		return err
	}
	sf, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sf.Close()

	n, err := io.Copy(df, sf)
	if err == nil && n != sfi.Size() {
		err = fmt.Errorf("copied wrong size for %s -> %s: copied %d; want %d", src, dst, n, sfi.Size())
	}
	cerr := df.Close()
	if err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Chmod(dst, sfi.Mode())
	}
	if err == nil {
		err = os.Chtimes(dst, sfi.ModTime(), sfi.ModTime())
	}
	return err
}

func isExecMode(mode os.FileMode) bool {
	return (mode & 0111) != 0
}

var uniqeID = int64(1)
var m = sync.Mutex{}

// GetUniqueID returns a unique id for the proto messages
func GetUniqueID() int64 {
	m.Lock()
	defer m.Unlock()
	uniqeID++
	return uniqeID
}

// CopyFile creates a copy of source file to destination file
func CopyFile(src, dest string) error {
	if !FileExists(src) {
		return fmt.Errorf("%s doesn't exist", src)
	}

	b, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(dest, b, NewFilePermissions)
	if err != nil {
		return err
	}

	return nil
}

//Appends contents of source file to destination file.
// If destination file is not present, Copy file action is performed
func AppendToFile(srcFile, destFile string) error {
	if FileExists(destFile) {
		f, err := os.OpenFile(destFile, os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return fmt.Errorf("Failed to open %s. %s \n", destFile, err.Error())
		}

		defer f.Close()

		srcFileContent, err := ReadFileContents(srcFile)
		if err != nil {
			return fmt.Errorf("Failed to read %s. %s \n", srcFile, err.Error())
		}
		srcFileContent = fmt.Sprintf("\n%s\n", srcFileContent)
		if _, err = f.WriteString(srcFileContent); err != nil {
			return fmt.Errorf("Failed to append from %s. %s \n", srcFile, err.Error())
		}
	} else {
		err := CopyFile(srcFile, destFile)
		if err != nil {
			return fmt.Errorf("Failed to copy %s. %s \n", srcFile, err.Error())
		}
	}
	return nil
}

// SetEnvVariable is a wrapper around os.SetEnv to set env variable
func SetEnvVariable(key, value string) error {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	err := os.Setenv(key, value)
	if err != nil {
		return fmt.Errorf("Failed to set: %s = %s. %s", key, value, err.Error())
	}
	return nil
}

// ExecuteCommand executes the given command in the working directory.
func ExecuteCommand(command []string, workingDir string, outputStreamWriter io.Writer, errorStreamWriter io.Writer) (*exec.Cmd, error) {
	cmd := prepareCommand(false, command, workingDir, outputStreamWriter, errorStreamWriter)
	err := cmd.Start()
	return cmd, err
}

// ExecuteSystemCommand executes the given system command in the working directory.
func ExecuteSystemCommand(command []string, workingDir string, outputStreamWriter io.Writer, errorStreamWriter io.Writer) (*exec.Cmd, error) {
	cmd := prepareCommand(true, command, workingDir, outputStreamWriter, errorStreamWriter)
	err := cmd.Start()
	return cmd, err
}

// ExecuteCommandWithEnv executes command after setting the given environment
func ExecuteCommandWithEnv(command []string, workingDir string, outputStreamWriter io.Writer, errorStreamWriter io.Writer, env []string) (*exec.Cmd, error) {
	cmd := prepareCommand(false, command, workingDir, outputStreamWriter, errorStreamWriter)
	cmd.Env = env
	err := cmd.Start()
	return cmd, err
}

func prepareCommand(isSystemCommand bool, command []string, workingDir string, outputStreamWriter io.Writer, errorStreamWriter io.Writer) *exec.Cmd {
	cmd := GetExecutableCommand(isSystemCommand, command...)
	cmd.Dir = workingDir
	cmd.Stdout = outputStreamWriter
	cmd.Stderr = errorStreamWriter
	cmd.Stdin = os.Stdin
	return cmd
}

// GetExecutableCommand returns the path of the executable file
func GetExecutableCommand(isSystemCommand bool, command ...string) *exec.Cmd {
	if len(command) == 0 {
		panic(fmt.Errorf("Invalid executable command"))
	}
	cmd := &exec.Cmd{Path: command[0]}
	if len(command) > 1 {
		if isSystemCommand {
			cmd = exec.Command(command[0], command[1:]...)
		}
		cmd.Args = append([]string{command[0]}, command[1:]...)
	} else {
		if isSystemCommand {
			cmd = exec.Command(command[0])
		}
		cmd.Args = append([]string{command[0]})
	}
	return cmd
}

// GetTempDir returns the system temp directory
func GetTempDir() string {
	tempGaugeDir := filepath.Join(os.TempDir(), "gauge_temp")
	tempGaugeDir += strconv.FormatInt(time.Now().UnixNano(), 10)
	if !exists(tempGaugeDir) {
		os.MkdirAll(tempGaugeDir, NewDirectoryPermissions)
	}
	return tempGaugeDir
}

// Remove removes all the files and directories recursively for the given path
func Remove(path string) error {
	return os.RemoveAll(path)
}

func exists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

// UnzipArchive extract the zip file to destination directory
func UnzipArchive(zipFile string, dest string) (string, error) {
	if !FileExists(zipFile) {
		return "", fmt.Errorf("ZipFile %s does not exist", zipFile)
	}

	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return "", err
	}
	defer r.Close()

	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		error := func() error {
			defer rc.Close()

			path := filepath.Join(dest, f.Name)
			os.MkdirAll(filepath.Dir(path), NewDirectoryPermissions)
			if f.FileInfo().IsDir() {
				os.MkdirAll(path, f.Mode())
			} else {
				f, err := os.OpenFile(
					path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
				if err != nil {
					return err
				}
				defer f.Close()

				_, err = io.Copy(f, rc)
				if err != nil {
					return err
				}
			}
			return nil
		}()
		if error != nil {
			return "", error

		}
	}

	return dest, nil
}

// SaveFile saves contents at the given filepath
func SaveFile(filePath, contents string, takeBackup bool) error {
	backupFile := ""
	if takeBackup {
		tmpDir := os.TempDir()
		fileName := fmt.Sprintf("%s_%v", filepath.Base(filePath), GetUniqueID())
		backupFile = filepath.Join(tmpDir, fileName)
		err := CopyFile(filePath, backupFile)
		if err != nil {
			return fmt.Errorf("Failed to make backup for '%s': %s", filePath, err.Error())
		}
	}
	err := ioutil.WriteFile(filePath, []byte(contents), NewFilePermissions)
	if err != nil {
		return fmt.Errorf("Failed to write to '%s': %s", filePath, err.Error())
	}

	return nil
}

func getUserHomeFromEnv() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func isWindows() bool {
	return runtime.GOOS == "windows"
}

// TrimTrailingSpace trims the trailing spaces in the given string
func TrimTrailingSpace(str string) string {
	var r = regexp.MustCompile(`[ \t]+$`)
	return r.ReplaceAllString(str, "")
}

func (property *Property) String() string {
	return fmt.Sprintf("#%s\n%s = %s", property.Comment, property.Name, property.DefaultValue)
}

// UrlExists checks if the given url exists
func UrlExists(url string) (bool, error) {
	resp, err := http.Head(url)
	if err != nil {
		return false, fmt.Errorf("Failed to resolve host")
	}
	if resp.StatusCode == 200 {
		return true, nil
	}
	return false, fmt.Errorf("Could not get %s, %d-%s", url, resp.StatusCode, resp.Status)
}

// GetPluginProperties returns the properties of the given plugin.
func GetPluginProperties(jsonPropertiesFile string) (map[string]interface{}, error) {
	pluginPropertiesJSON, err := ioutil.ReadFile(jsonPropertiesFile)
	if err != nil {
		return nil, fmt.Errorf("Could not read %s: %s\n", filepath.Base(jsonPropertiesFile), err)
	}
	var pluginJSON interface{}
	if err = json.Unmarshal([]byte(pluginPropertiesJSON), &pluginJSON); err != nil {
		return nil, fmt.Errorf("Could not read %s: %s\n", filepath.Base(jsonPropertiesFile), err)
	}
	return pluginJSON.(map[string]interface{}), nil
}

// GetGaugePluginVersion returns the latest version installed of the given plugin
func GetGaugePluginVersion(pluginName string) (string, error) {
	pluginProperties, err := GetPluginProperties(fmt.Sprintf("%s.json", pluginName))
	if err != nil {
		return "", fmt.Errorf("Failed to get gauge %s properties file. %s", pluginName, err)
	}
	return pluginProperties["version"].(string), nil
}
