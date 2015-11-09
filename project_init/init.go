package project_init

import (
	"errors"
	"fmt"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/logger/execLogger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/plugin/install"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
	"os"
	"path"
	"path/filepath"
)

const (
	specsDirName      = "specs"
	skelFileName      = "hello_world.spec"
	envDefaultDirName = "default"
)

var defaultPlugins = []string{"html-report"}

func InitializeProject(language string) {
	wd, err := os.Getwd()
	if err != nil {
		execLogger.CriticalError(errors.New(fmt.Sprintf("Failed to find working directory. %s\n", err.Error())))
	}
	config.ProjectRoot = wd
	err = createProjectTemplate(language)
	if err != nil {
		execLogger.CriticalError(errors.New(fmt.Sprintf("Failed to initialize. %s\n", err.Error())))
	}
	logger.Log.Info("\nSuccessfully initialized the project. Run specifications with \"gauge specs/\"")
}

func showMessage(action, filename string) {
	logger.Log.Info(" %s  %s\n", action, filename)
}

func createProjectTemplate(language string) error {
	if !install.IsCompatibleLanguagePluginInstalled(language) {
		logger.Log.Info("Compatible %s plugin is not installed \n", language)
		logger.Log.Info("Installing plugin => %s ... \n\n", language)

		if result := install.InstallPlugin(language, ""); !result.Success {
			return errors.New(fmt.Sprintf("Failed to install plugin %s . %s \n", language, result.Error.Error()))
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
	defaultJson, err := common.GetSkeletonFilePath(path.Join(common.EnvDirectoryName, common.DefaultEnvFileName))
	if err != nil {
		return err
	}
	defaultJsonDest := path.Join(defaultEnv, common.DefaultEnvFileName)
	showMessage("create", defaultJsonDest)
	err = common.CopyFile(defaultJson, defaultJsonDest)
	if err != nil {
		showMessage("error", fmt.Sprintf("Failed to create %s. %s", defaultJsonDest, err.Error()))
	}

	return runner.ExecuteInitHookForRunner(language)
}

func SetWorkingDir(workingDir string) {
	targetDir, err := filepath.Abs(workingDir)
	if err != nil {
		execLogger.CriticalError(errors.New(fmt.Sprintf("Unable to set working directory : %s\n", err)))
	}
	if !common.DirExists(targetDir) {
		err = os.Mkdir(targetDir, 0777)
		if err != nil {
			execLogger.CriticalError(errors.New(fmt.Sprintf("Unable to set working directory : %s\n", err)))
		}
	}
	err = os.Chdir(targetDir)
	_, err = os.Getwd()
	if err != nil {
		execLogger.CriticalError(errors.New(fmt.Sprintf("Unable to set working directory : %s\n", err)))
	}
}
