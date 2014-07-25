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
	GaugeVersionSupport versionSupport
	Description         string
	Install             platformSpecifics
	DownloadUrls        struct {
		X86 platformSpecifics
		X64 platformSpecifics
	}
}

type platformSpecifics struct {
	Windows string
	Linux   string
	Darwin  string
}

type versionSupport struct {
	Minimum string
	Maximum string
}

func installRunner(language, version string) {
	installDescription, err := getInstallDescription(language, version)
	if err != nil {
		fmt.Printf("Error installing language %s %s : %s \n", language, version, err)
	}
	installRunnerWithDescription(installDescription)
}

func installRunnerWithDescription(installDescription *installDescription) {
	if err := checkVersionCompatibilityWithGauge(installDescription, currentGaugeVersion); err != nil {
		fmt.Printf("Error installing runner %s. %s \n", installDescription.Name, err)
		return
	}
	fmt.Printf("Succesfully installed language runner : %s %s \n", installDescription.Name, installDescription.Version)
}

func checkVersionCompatibilityWithGauge(installDescription *installDescription, gaugeVersion *version) error {
	if installDescription.GaugeVersionSupport.Minimum == "" {
		return errors.New(fmt.Sprintf("Supported Gauge Version numbers not found in %s install file. Cannot install runner", installDescription.Name))
	}

	minGaugeVersion, err := parseVersion(installDescription.GaugeVersionSupport.Minimum)
	if err != nil {
		return errors.New(fmt.Sprintf("Invalid minimum gauge support version %s in install file for %s : %s. ", installDescription.GaugeVersionSupport.Minimum, installDescription.Name, err))
	}
	if installDescription.GaugeVersionSupport.Maximum != "" {
		maxGaugeVersion, err := parseVersion(installDescription.GaugeVersionSupport.Maximum)
		if err != nil {
			return errors.New(fmt.Sprintf("Invalid maximum gauge support version %s in install file for %s : %s. ", installDescription.GaugeVersionSupport.Maximum, installDescription.Name, err))
		}
		if gaugeVersion.isBetween(minGaugeVersion, maxGaugeVersion) {
			return nil
		} else {
			return errors.New(fmt.Sprintf("Incompatible %s version %s. Gauge version %s is not supported for this runner.", installDescription.Name, installDescription.Version, gaugeVersion.String()))
		}
	}

	if minGaugeVersion.isLesserThanEqualTo(gaugeVersion) {
		return nil
	}
	return errors.New(fmt.Sprintf("Incompatible language %s version %s. Minimun supported gauge version %s is higher than current Gauge version %s", installDescription.Name, installDescription.Version, minGaugeVersion.String(), gaugeVersion.String()))
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
