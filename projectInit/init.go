/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package projectInit

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/template"
	"github.com/getgauge/gauge/util"
)

const (
	specsDirName      = "specs"
	skelFileName      = "example.spec"
	gitignoreFileName = ".gitignore"
	envDefaultDirName = "default"
	metadataFileName  = "metadata.json"
	https             = "https"
)

var defaultPlugins = []string{"html-report"}

type templateMetadata struct {
	Name           string
	Description    string
	Version        string
	PostInstallCmd string
	PostInstallMsg string
}

func initializeTemplate(templateName, templateUrl string) error {
	tempDir := common.GetTempDir()
	defer util.Remove(tempDir)
	logger.Infof(true, "Initializing template from %s", templateUrl)
	unzippedTemplate, err := util.DownloadAndUnzip(templateUrl, tempDir)
	if err != nil {
		return err
	}

	wd := config.ProjectRoot

	if common.FileExists(gitignoreFileName) {
		templateGitIgnore := filepath.Join(unzippedTemplate, templateName, gitignoreFileName)
		if err := common.AppendToFile(gitignoreFileName, templateGitIgnore); err != nil {
			return err
		}
	}

	logger.Infof(true, "Copying Gauge template %s to current directory ...", templateName)
	filesAdded, err := common.MirrorDir(filepath.Join(unzippedTemplate, templateName), wd)
	if err != nil {
		return fmt.Errorf("Failed to copy Gauge template: %s", err.Error())
	}

	metadataFile := filepath.Join(wd, metadataFileName)
	metadataContents, err := common.ReadFileContents(metadataFile)
	if err != nil {
		return fmt.Errorf("Failed to read file contents of %s: %s", metadataFile, err.Error())
	}

	metadata := &templateMetadata{}
	err = json.Unmarshal([]byte(metadataContents), metadata)
	if err != nil {
		return err
	}

	if metadata.PostInstallCmd != "" {
		logger.Debugf(true, "Running post install command %s", metadata.PostInstallCmd)
		command := strings.Fields(metadata.PostInstallCmd)
		cmd, err := common.ExecuteSystemCommand(command, wd, os.Stdout, os.Stderr)
		if err != nil {
			for _, file := range filesAdded {
				pathSegments := strings.Split(file, string(filepath.Separator))
				util.Remove(filepath.Join(wd, pathSegments[0]))
			}
			return fmt.Errorf("Failed to run post install commands: %s", err.Error())
		}
		if err = cmd.Wait(); err != nil {
			return err
		}
	}
	logger.Infof(true, "Successfully initialized the project. %s", metadata.PostInstallMsg)

	util.Remove(metadataFile)
	return nil
}

func getTemplateLanguage(templateName string) string {
	return strings.Split(templateName, "_")[0]
}

func getTemplateName(templateURL string) string {
	parts := strings.Split(templateURL, "/")
	basename := parts[len(parts)-1]
	return strings.TrimSuffix(basename, filepath.Ext(basename))
}

func isGaugeProject() bool {
	m, err := manifest.ProjectManifest()
	if err != nil {
		return false
	}
	return m.Language != ""
}

func installRunner(templateName string, silent bool) {
	language := getTemplateLanguage(templateName)
	if !install.IsCompatiblePluginInstalled(language, true) {
		logger.Infof(true, "Compatible language plugin %s is not installed. Installing plugin...", language)
		install.HandleInstallResult(install.Plugin(language, "", silent), language, true)
	}
}

// InitializeProject initializes a Gauge project with specified template
func InitializeProject(templateName string, silent bool) {
	validateDirectory()
	templateURL, err := template.Get(templateName)
	logger.Debugf(true, "Failed to get template download info for '%s'.", templateName)
	if err == nil {
		checkURL(templateURL)
		err = initializeTemplate(templateName, templateURL)
		installRunner(templateName, silent)
	} else {
		installRunner(templateName, silent)
		err = createProjectTemplate(templateName)
	}
	if err != nil {
		logger.Fatalf(true, "Failed to initialize project. %s", err.Error())
	}
}

// InitializeProject initializes a Gauge project with specified template URL
func InitializeProjectFromURL(templateURL string, silent bool) {
	validateDirectory()
	name := getTemplateName(templateURL)
	checkURL(templateURL)
	if err := initializeTemplate(name, templateURL); err != nil {
		logger.Fatalf(true, "Failed to initialize project. %s", err.Error())
	}
	installRunner(name, silent)

}

func validateDirectory() {
	wd, err := os.Getwd()
	if err != nil {
		logger.Fatalf(true, "Failed to find working directory. %s", err.Error())
	}
	config.ProjectRoot = wd
	if isGaugeProject() {
		logger.Fatalf(true, "This is already a Gauge Project. Please try to initialize a Gauge project in a different location.")
	}
}

func checkURL(templateURL string) {
	u, err := url.ParseRequestURI(templateURL)
	if err != nil {
		logger.Fatalf(true, "Failed to parse template URL '%s'. The template location must be a valid (https) URI", templateURL)
	}
	if u.Scheme != https && !config.AllowInsecureDownload() {
		logger.Fatalf(true, "The url '%s' in not secure and 'allow_insecure_download' is set to false.\n"+
			"To allow insecure downloads set 'allow_insecure_download' configuration to true.\n"+
			"Run 'gauge config allow_insecure_download true' to the same.", templateURL)
	}

}

func showMessage(action, filename string) {
	logger.Infof(true, " %s  %s", action, filename)
}

func createProjectTemplate(language string) error {
	if err := runner.ExecuteInitHookForRunner(language); err != nil {
		return err
	}
	if err := createManifestFile(language); err != nil {
		return err
	}
	if err := createSpecDirectory(); err != nil {
		return err
	}
	if err := createSkeletonFile(); err != nil {
		return err
	}
	if err := createOrAppendGitignoreFile(); err != nil {
		return err
	}
	if err := createEnvDirectory(); err != nil {
		return err
	}
	logger.Infof(true, "Successfully initialized the project. Run specifications with \"gauge run specs/\".")
	return nil
}

func createEnvDirectory() error {
	showMessage("create", common.EnvDirectoryName)
	if !common.DirExists(common.EnvDirectoryName) {
		err := os.Mkdir(common.EnvDirectoryName, common.NewDirectoryPermissions)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to create %s. %s", common.EnvDirectoryName, err.Error()))
		}
	}
	defaultEnv := filepath.Join(common.EnvDirectoryName, envDefaultDirName)
	showMessage("create", defaultEnv)
	if !common.DirExists(defaultEnv) {
		err := os.Mkdir(defaultEnv, common.NewDirectoryPermissions)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to create %s. %s", defaultEnv, err.Error()))
		}
	}
	defaultJSON, err := common.GetSkeletonFilePath(filepath.Join(common.EnvDirectoryName, common.DefaultEnvFileName))
	if err != nil {
		return err
	}
	defaultJSONDest := filepath.Join(defaultEnv, common.DefaultEnvFileName)
	showMessage("create", defaultJSONDest)
	err = common.CopyFile(defaultJSON, defaultJSONDest)
	if err != nil {
		showMessage("error", fmt.Sprintf("Failed to create %s. %s", defaultJSONDest, err.Error()))
		return err
	}
	return nil
}

func createOrAppendGitignoreFile() error {
	destFile := filepath.Join(gitignoreFileName)
	srcFile, err := common.GetSkeletonFilePath(gitignoreFileName)
	if err != nil {
		showMessage("error", fmt.Sprintf("Failed to read .gitignore file. %s", err.Error()))
		return err
	}
	showMessage("create", destFile)
	if err := common.AppendToFile(srcFile, destFile); err != nil {
		showMessage("error", err.Error())
		return err
	}
	return nil
}

func createSkeletonFile() error {
	skelFile, err := common.GetSkeletonFilePath(skelFileName)
	if err != nil {
		return err
	}
	specFile := filepath.Join(specsDirName, skelFileName)
	showMessage("create", specFile)
	if common.FileExists(specFile) {
		showMessage("skip", specFile)
		return err
	} else {
		err = common.CopyFile(skelFile, specFile)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to create %s. %s", specFile, err.Error()))
			return err
		}
	}
	return nil
}

func createSpecDirectory() error {
	showMessage("create", specsDirName)
	if !common.DirExists(specsDirName) {
		err := os.Mkdir(specsDirName, common.NewDirectoryPermissions)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to create %s. %s", specsDirName, err.Error()))
			return err
		}
	} else {
		showMessage("skip", specsDirName)
	}
	return nil
}

func createManifestFile(language string) error {
	showMessage("create", common.ManifestFile)
	if common.FileExists(common.ManifestFile) {
		showMessage("skip", common.ManifestFile)
	}
	manifest := &manifest.Manifest{Language: language, Plugins: defaultPlugins}
	return manifest.Save()
}
