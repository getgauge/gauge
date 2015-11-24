package projectInit

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
)

const (
	specsDirName      = "specs"
	skelFileName      = "hello_world.spec"
	envDefaultDirName = "default"
	metadataFileName  = "metadata.json"
)

var defaultPlugins = []string{"html-report"}

type templateMetadata struct {
	Name           string
	Description    string
	Version        string
	PostInstallCmd string
}

func initializeTemplate(templateName string) error {
	defer util.RemoveTempDir()

	unzippedTemplate, err := util.DownloadAndUnzip(getTemplateURL(templateName))
	if err != nil {
		return err
	}

	wd := config.ProjectRoot

	logger.Log.Info("Copying Gauge template %s to current directory ...", templateName)
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
		cmd, err := common.ExecuteCommand([]string{metadata.PostInstallCmd}, wd, os.Stdout, os.Stderr)
		cmd.Wait()
		if err != nil {
			for _, file := range filesAdded {
				pathSegments := strings.Split(file, string(filepath.Separator))
				util.Remove(filepath.Join(wd, pathSegments[0]))
			}
			return fmt.Errorf("Failed to run post install commands: %s", err.Error())
		}
	}

	util.Remove(metadataFile)
	return nil
}

func getTemplateURL(templateName string) string {
	//	filepath.Join(config.GaugeRepositoryUrl(), "templates", templateName, ".zip")
	return "https://github.com/getgauge/gauge-repository/raw/template/templates/java.zip"
}

// InitializeProject initializes a Gauge project with specified template
func InitializeProject(templateName string) {
	wd, err := os.Getwd()
	if err != nil {
		logger.Log.Critical("Failed to find working directory. %s\n", err.Error())
	}
	config.ProjectRoot = wd
	//	if templateName == "java" {
	//		err = initializeTemplate(templateName)
	//	} else {
	//		err = createProjectTemplate(templateName)
	//	}
	err = createProjectTemplate(templateName)
	if err != nil {
		logger.Log.Critical("Failed to initialize. %s\n", err.Error())
		return
	}
	logger.Log.Info("\nSuccessfully initialized the project. Run specifications with \"gauge specs/\"")
}

func showMessage(action, filename string) {
	logger.Log.Info(" %s  %s", action, filename)
}

func createProjectTemplate(language string) error {
	if !install.IsCompatibleLanguagePluginInstalled(language) {
		logger.Log.Info("Compatible %s plugin is not installed \n", language)
		logger.Log.Info("Installing plugin => %s ... \n\n", language)

		if result := install.InstallPlugin(language, ""); !result.Success {
			return fmt.Errorf("Failed to install plugin %s . %s \n", language, result.Error.Error())
		}
	}
	// Create the project manifest
	showMessage("create", common.ManifestFile)
	if common.FileExists(common.ManifestFile) {
		showMessage("skip", common.ManifestFile)
	}
	manifest := &manifest.Manifest{Language: language, Plugins: defaultPlugins}
	if err := manifest.Save(); err != nil {
		return err
	}

	// creating the spec directory
	showMessage("create", specsDirName)
	if !common.DirExists(specsDirName) {
		err := os.Mkdir(specsDirName, common.NewDirectoryPermissions)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to create %s. %s", specsDirName, err.Error()))
		}
	} else {
		showMessage("skip", specsDirName)
	}

	// Copying the skeleton file
	skelFile, err := common.GetSkeletonFilePath(skelFileName)
	if err != nil {
		return err
	}
	specFile := path.Join(specsDirName, skelFileName)
	showMessage("create", specFile)
	if common.FileExists(specFile) {
		showMessage("skip", specFile)
	} else {
		err = common.CopyFile(skelFile, specFile)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to create %s. %s", specFile, err.Error()))
		}
	}

	// Creating the env directory
	showMessage("create", common.EnvDirectoryName)
	if !common.DirExists(common.EnvDirectoryName) {
		err = os.Mkdir(common.EnvDirectoryName, common.NewDirectoryPermissions)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to create %s. %s", common.EnvDirectoryName, err.Error()))
		}
	}
	defaultEnv := path.Join(common.EnvDirectoryName, envDefaultDirName)
	showMessage("create", defaultEnv)
	if !common.DirExists(defaultEnv) {
		err = os.Mkdir(defaultEnv, common.NewDirectoryPermissions)
		if err != nil {
			showMessage("error", fmt.Sprintf("Failed to create %s. %s", defaultEnv, err.Error()))
		}
	}
	defaultJSON, err := common.GetSkeletonFilePath(path.Join(common.EnvDirectoryName, common.DefaultEnvFileName))
	if err != nil {
		return err
	}
	defaultJSONDest := path.Join(defaultEnv, common.DefaultEnvFileName)
	showMessage("create", defaultJSONDest)
	err = common.CopyFile(defaultJSON, defaultJSONDest)
	if err != nil {
		showMessage("error", fmt.Sprintf("Failed to create %s. %s", defaultJSONDest, err.Error()))
	}

	return runner.ExecuteInitHookForRunner(language)
}

// SetWorkingDir sets the current working directory to specified location
func SetWorkingDir(workingDir string) {
	targetDir, err := filepath.Abs(workingDir)
	if err != nil {
		logger.Log.Critical("Unable to set working directory : %s\n", err.Error())
	}
	if !common.DirExists(targetDir) {
		err = os.Mkdir(targetDir, 0777)
		if err != nil {
			logger.Log.Critical("Unable to set working directory : %s\n", err.Error())
		}
	}
	err = os.Chdir(targetDir)
	_, err = os.Getwd()
	if err != nil {
		logger.Log.Critical("Unable to set working directory : %s\n", err.Error())
	}
}
