package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/getgauge/common"
	"runtime"
	"strings"
)

const (
	gaugeRepositoryUrl = "http://raw.github.com/getgauge/gauge-repository/master"
)

type installDescription struct {
	Name        string
	Description string
	Versions    []versionInstallDescription
}

type versionInstallDescription struct {
	Version             string
	GaugeVersionSupport versionSupport
	Install             platformSpecifics
	DownloadUrls        downloadUrls
}

type downloadUrls struct {
	X86 platformSpecifics
	X64 platformSpecifics
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

func installPlugin(pluginName, version string) {
	installDescription, err := getInstallDescription(pluginName)
	if err != nil {
		fmt.Printf("Error installing plugin %s %s : %s \n", pluginName, version, err)
	}
	fmt.Println(installDescription)
	installPluginWithDescription(installDescription, version)
}

func installPluginWithDescription(installDescription *installDescription, version string) {
	var versionInstallDescription *versionInstallDescription
	var err error
	if version != "" {
		versionInstallDescription, err = installDescription.getVersion(version)
		if err != nil {
			fmt.Printf("Could not install plugin %s %s : %s", installDescription.Name, version, err)
		}
		if compatibilityError := checkCompatiblity(currentGaugeVersion, &versionInstallDescription.GaugeVersionSupport); compatibilityError != nil {
			fmt.Printf("Could not install plugin %s %s. Plugin Version %s is not supported for gauge %s : %s \n", installDescription.Name, versionInstallDescription.Version, versionInstallDescription.Version, currentGaugeVersion.String(), compatibilityError.Error())
		}
	} else {
		versionInstallDescription, err = installDescription.getLatestCompatibleVersionTo(currentGaugeVersion)
		if (err != nil) {
			fmt.Printf("Could not install plugin %s. : %s", installDescription.Name, err)
		}
	}
	installPluginVersion(versionInstallDescription)
}

func installPluginVersion(versionInstallDescription *versionInstallDescription) error {
	pluginZip, err := downloadPluginZip(versionInstallDescription.DownloadUrls)
	if (err != nil) {
		return errors.New(fmt.Sprintf("Failed to download plugin zip %s.", err))
	}
	unzipDir, err := common.UnzipArchive(pluginZip)
	if (err != nil) {
		return errors.New(fmt.Sprintf("Failed to Unzip plugin-zip file %s.", err))
	}
	return copyPluginFilesToGauge(unzipDir)
}

func copyPluginFilesToGauge(pluginDir string) error {
	return nil
}

func downloadPluginZip(downloadUrls downloadUrls) (string, error) {
	var platformLinks *platformSpecifics
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

	downloadedFile, err := common.DownloadToTempDir(downloadLink)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to download File %s. %s", downloadLink, err.Error()))
	}
	return downloadedFile, err
}

func getInstallDescription(plugin string) (*installDescription, error) {
	installJson, err := getPluginInstallJson(plugin)
	if err != nil {
		return nil, err
	}
	InstallJsonContents, readErr := common.ReadFileContents(installJson)
	if readErr != nil {
		return nil, readErr
	}
	installDescription := &installDescription{}
	if err = json.Unmarshal([]byte(InstallJsonContents), installDescription); err != nil {
		return nil, err
	}
	return installDescription, nil
}

func getPluginInstallJson(plugin string) (string, error) {
	versionInstallDescriptionJsonFile := plugin + "-install.json"
	versionInstallDescriptionJsonUrl := constructPluginInstallJsonUrl(plugin)

	downloadedFile, downloadErr := common.DownloadToTempDir(versionInstallDescriptionJsonUrl)
	if downloadErr != nil {
		return "", errors.New(fmt.Sprintf("Could not find %s file. Check install name and version. %s", versionInstallDescriptionJsonFile, downloadErr.Error()))
	}
	return downloadedFile, nil
}

func constructPluginInstallJsonUrl(plugin string) string {
	installJsonFile := plugin + "-install.json"
	return fmt.Sprintf("%s/%s", gaugeRepositoryUrl, installJsonFile)
}

func (installDesc *installDescription) getVersion(version string) (*versionInstallDescription, error) {
	for _, versionInstallDescription := range installDesc.Versions {
		if versionInstallDescription.Version == version {
			return &versionInstallDescription, nil
		}
	}
	return nil, errors.New("Could not find install description for Version "+version)
}

func (installDesc *installDescription) getLatestCompatibleVersionTo(version *version) (*versionInstallDescription, error) {
	return nil, nil
}

func checkCompatiblity(version *version, versionSupport *versionSupport) error {
	minSupportVersion, err := parseVersion(versionSupport.Minimum)
	if err != nil {
		return errors.New(fmt.Sprintf("Invalid minimum support version %s. : %s. ", versionSupport.Minimum, err))
	}
	if versionSupport.Maximum != "" {
		maxSupportVersion, err := parseVersion(versionSupport.Maximum)
		if err != nil {
			return errors.New(fmt.Sprintf("Invalid maximum support version %s. : %s. ", versionSupport.Maximum, err))
		}
		if version.isBetween(minSupportVersion, maxSupportVersion) {
			return nil
		} else {
			return errors.New(fmt.Sprintf("Version %s is not between %s and %s", version, minSupportVersion, maxSupportVersion))
		}
	}

	if minSupportVersion.isLesserThanEqualTo(version) {
		return nil
	}
	return errors.New(fmt.Sprintf("Incompatible version. Minimun version %s is higher than current version %s", minSupportVersion, version))
}
