package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/getgauge/common"
	"io/ioutil"
	"path"
)

const (
	gaugeRepositoryRunnersUrl = "http://raw.github.com/getgauge/gauge-repository/master/runners"
	currentDir                = "current"
)

type installDescription struct {
	Name                string
	Version             string
	GaugeVersionSupport struct {
		Minimum string
		Maximum string
	}
	Description  string
	Install      platformSpecifics
	DownloadUrls struct {
		X86 platformSpecifics
		X64 platformSpecifics
	}
}

type platformSpecifics struct {
	Windows string
	Linux   string
	Darwin  string
}

func installRunner(language, version string) {
	installDescription, err := getInstallDescription(language, version)
	if err != nil {
		fmt.Printf("Error installing language %s %s : %s \n", language, version, err)
	}
	installRunnerWithDescription(installDescription)
}

func installRunnerWithDescription(installDescription *installDescription) {
	if err := checkVersionCompatibilityWithGauge(installDescription); err != nil {
		fmt.Printf("Incompatible runner version. $s \n", err)
		return
	}
}

func checkVersionCompatibilityWithGauge(installDescription *installDescription) error {
	if installDescription.GaugeVersionSupport.Minimum == "" {
		return errors.New(fmt.Sprintf("Supported Gauge Version numbers not found in %s install file. Cannot install runner", installDescription.Name))
	}
	parseVersion(installDescription.GaugeVersionSupport.Minimum)
	return nil
}

func getInstallDescription(language, version string) (*installDescription, error) {
	languageJson, err := getLanguageJsonFromGaugeRepository(language, version)
	if err != nil {
		return nil, err
	}
	languageJsonContents, readErr := common.ReadFileContents(languageJson)
	if readErr != nil {
		return nil, readErr
	}
	installDescription := &installDescription{}
	if err = json.Unmarshal([]byte(languageJsonContents), installDescription); err != nil {
		return nil, err
	}
	return installDescription, nil
}

func getLanguageJsonFromGaugeRepository(language string, version string) (string, error) {
	languageInstallJsonFile := language + "-install.json"
	languageInstallJsonUrl := constructLanguageInstallJsonUrl(language, version)

	tempDir, err := ioutil.TempDir("", fmt.Sprintf("%d", common.GetUniqueId()))
	if err != nil {
		return "", err
	}
	if err = common.Download(languageInstallJsonUrl, tempDir); err != nil {
		return "", errors.New(fmt.Sprintf("Could not find %s file. Check install name and version. %s", languageInstallJsonFile, err.Error()))
	}
	return path.Join(tempDir, languageInstallJsonFile), nil
}

func constructLanguageInstallJsonUrl(language string, version string) string {
	languageInstallJsonFile := language + "-install.json"
	if version != "" {
		return fmt.Sprintf("%s/%s/%s/%s", gaugeRepositoryRunnersUrl, language, version, languageInstallJsonFile)
	} else {
		return fmt.Sprintf("%s/%s/%s/%s", gaugeRepositoryRunnersUrl, language, currentDir, languageInstallJsonFile)
	}
}
