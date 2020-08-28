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
	"github.com/getgauge/gauge/template"
	"github.com/getgauge/gauge/util"
)

const (
	gitignoreFileName = ".gitignore"
	metadataFileName  = "metadata.json"
	https             = "https"
)

type templateMetadata struct {
	Name           string
	Description    string
	Version        string
	PostInstallCmd string
	PostInstallMsg string
}

func initializeTemplate(templateUrl string) error {
	tempDir := common.GetTempDir()
	defer util.Remove(tempDir)
	logger.Infof(true, "Initializing template from %s", templateUrl)
	unzippedTemplate, err := util.DownloadAndUnzip(templateUrl, tempDir)
	if err != nil {
		return fmt.Errorf("%w. Please use a valid Gauge template URI or check network connection", err)
	}
	return copyTemplateContents(unzippedTemplate)
}

func copyTemplateContents(unzippedTemplate string) error {
	wd := config.ProjectRoot
	templateDir, err := getTemplateDir(unzippedTemplate)
	if err != nil {
		return fmt.Errorf("failed to copy template. The dir %s does not contain required files. %w", unzippedTemplate, err)
	}
	if common.FileExists(gitignoreFileName) {
		templateGitIgnore := filepath.Join(templateDir, gitignoreFileName)
		if err := common.AppendToFile(gitignoreFileName, templateGitIgnore); err != nil {
			return err
		}
	}

	logger.Infof(true, "Copying Gauge template %s to current directory ...", filepath.Base(templateDir))
	filesAdded, err := common.MirrorDir(templateDir, wd)
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

func getTemplateDir(unzippedTemplate string) (templateDir string, err error) {
	err = filepath.Walk(unzippedTemplate, func(path string, info os.FileInfo, err error) error {
		if err == nil && info.IsDir() && common.FileExists(filepath.Join(path, common.ManifestFile)) {
			templateDir = path
		}
		return err
	})
	return templateDir, err
}

func isGaugeProject() bool {
	m, err := manifest.ProjectManifest()
	if err != nil {
		return false
	}
	return m.Language != ""
}

func installRunner(silent bool) {
	m, err := manifest.ProjectManifest()
	if err != nil {
		logger.Errorf(true, "failed to install language runner. %s", err.Error())
		return
	}
	if !install.IsCompatiblePluginInstalled(m.Language, true) {
		logger.Infof(true, "Compatible language plugin %s is not installed. Installing plugin...", m.Language)
		install.HandleInstallResult(install.Plugin(m.Language, "", silent), m.Language, true)
	}
}

func Template(name string, silent bool) {
	validateDirectory()
	if common.FileExists(name) {
		fromZipFile(name, silent)
	} else if isURL(name) {
		fromURL(name, silent)
	} else {
		fromTemplateName(name, silent)
	}
}

// FromTemplate initializes a Gauge project with specified template
func fromTemplateName(templateName string, silent bool) {
	templateURL, err := template.Get(templateName)
	if err != nil {
		logger.Fatalf(true, fmt.Errorf("Failed to initialize project. %w", err).Error())
	}
	checkURL(templateURL)
	if err := initializeTemplate(templateURL); err != nil {
		logger.Fatalf(true, fmt.Errorf("Failed to initialize project. %w", err).Error())
	}
	installRunner(silent)
}

// FromURL initializes a Gauge project with specified template URL
func fromURL(templateURL string, silent bool) {
	checkURL(templateURL)
	if err := initializeTemplate(templateURL); err != nil {
		logger.Fatalf(true, "Failed to initialize project. %s", err.Error())
	}
	installRunner(silent)
}

// fromZipFile initializes a Gauge project with specified zip file
func fromZipFile(templateFile string, silent bool) {
	validateDirectory()
	tempDir := common.GetTempDir()
	defer util.Remove(tempDir)
	unzippedTemplateDir, err := common.UnzipArchive(templateFile, tempDir)
	if err != nil {
		logger.Fatalf(true, "Failed to initialize project. %s", err.Error())
	}
	err = copyTemplateContents(unzippedTemplateDir)
	if err != nil {
		logger.Fatalf(true, "Failed to initialize project. %s", err.Error())
	}
	installRunner(silent)
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

func isURL(name string) bool {
	u, err := url.ParseRequestURI(name)
	return err == nil && u.Scheme != ""
}

func checkURL(templateURL string) {
	u, err := url.ParseRequestURI(templateURL)
	if err != nil {
		logger.Fatalf(true, "Failed to parse template URL '%s'. The template location must be a secure and valid URI", templateURL)
	}
	if u.Scheme != https && !config.AllowInsecureDownload() {
		logger.Fatalf(true, "The url '%s' in not secure and 'allow_insecure_download' is set to false.\n"+
			"To allow insecure downloads set 'allow_insecure_download' configuration to true.\n"+
			"Run 'gauge config allow_insecure_download true' to the same.", templateURL)
	}

}
