/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package install

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/plugin/pluginInfo"
	"github.com/getgauge/gauge/version"
)

const gauge_releases_url = "https://github.com/getgauge/gauge/releases"

type UpdateFacade struct {
	wg    *sync.WaitGroup
	print chan bool
}

func (u *UpdateFacade) BufferUpdateDetails() {
	var wg sync.WaitGroup
	u.print = make(chan bool)
	u.wg = &wg
	u.wg.Add(1)
	go printUpdateInfo(u.print, u.wg)
}

func (u *UpdateFacade) PrintUpdateBuffer() {
	u.print <- true
	u.wg.Wait()
}

func PrintUpdateInfoWithDetails() {
	updates := checkUpdates()
	if len(updates) > 0 {
		for _, update := range updates {
			logger.Infof(true, "%-10s\t\t%-10s\t%s", update.Name, update.CompatibleVersion, update.Message)
		}
	} else {
		logger.Info(true, "No Updates available.")
	}
}

func checkUpdates() []UpdateInfo {
	return append(checkGaugeUpdate(), checkPluginUpdates()...)
}

func recoverPanic() {
	if r := recover(); r != nil {
		logger.Fatalf(true, "%v\n%s", r, string(debug.Stack()))
	}
}

func printUpdateInfo(print chan bool, wg *sync.WaitGroup) {
	message := make(chan string)
	go func() {
		defer recoverPanic()
		updates := checkUpdates()
		if len(updates) > 0 {
			message <- "Updates are available. Run `gauge update -c` for more info."
		}
	}()
	waitToPrint(message, print, "", wg)
}

func waitToPrint(messageChan chan string, printChan chan bool, message string, wg *sync.WaitGroup) {
	select {
	case <-printChan:
		if message != "" {
			logger.Info(true, message)
		}
		wg.Done()
	case message = <-messageChan:
		waitToPrint(messageChan, printChan, message, wg)
	}
}

func checkGaugeUpdate() []UpdateInfo {
	var updateInfos []UpdateInfo
	v, err := getLatestGaugeVersion(gauge_releases_url + "/latest")
	if err != nil {
		return updateInfos
	}
	latestVersion, err := version.ParseVersion(v)
	if err != nil {
		return updateInfos
	}
	isLatestVersion := version.CurrentGaugeVersion.IsLesserThan(latestVersion)
	if isLatestVersion {
		updateInfos = append(updateInfos, UpdateInfo{"Gauge", latestVersion.String(), "Download the installer from https://gauge.org/get-started/"})
	}
	return updateInfos
}

type UpdateInfo struct {
	Name              string
	CompatibleVersion string
	Message           string
}

func checkPluginUpdates() []UpdateInfo {
	var pluginsToUpdate []UpdateInfo
	plugins, err := pluginInfo.GetAllInstalledPluginsWithVersion()
	if err != nil {
		return pluginsToUpdate
	}
	logger.Debugf(true, "Checking updates...")
	for _, plugin := range plugins {
		desc, result := getInstallDescription(plugin.Name, true)
		if result.Error != nil {
			continue
		}
		pluginsToUpdate = append(pluginsToUpdate, createPluginUpdateDetail(plugin.Version.String(), *desc)...)
	}
	return pluginsToUpdate
}

func createPluginUpdateDetail(currentVersion string, latestVersionDetails installDescription) []UpdateInfo {
	var updateInfo []UpdateInfo
	v, err := version.ParseVersion(currentVersion)
	if err != nil {
		return updateInfo
	}
	v1, err := version.ParseVersion(latestVersionDetails.Versions[0].Version)
	if err != nil {
		return updateInfo
	}
	if v.IsLesserThan(v1) {
		versionDesc, err := latestVersionDetails.getLatestCompatibleVersionTo(version.CurrentGaugeVersion)
		if err != nil {
			return updateInfo
		}
		updateInfo = append(updateInfo, UpdateInfo{latestVersionDetails.Name, versionDesc.Version, fmt.Sprintf("Run 'gauge update %s'", latestVersionDetails.Name)})
	}
	return updateInfo
}

var getLatestGaugeVersion = func(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(res.Body)

	v, err := getGaugeVersionFromURL(res.Request.URL.String())
	if err != nil {
		return "", err
	}
	return v, nil
}

func getGaugeVersionFromURL(url string) (string, error) {
	versionString := strings.ReplaceAll(url, gauge_releases_url, "")
	reg, err := regexp.Compile(`tag/v(\d.*)`)
	if err != nil {
		return "", fmt.Errorf("unable to compile regex 'tag/v(\\d.*)': %s", err.Error())
	}
	matches := reg.FindStringSubmatch(versionString)
	if len(matches) < 2 {
		return "", fmt.Errorf("failed to parse: %s", url)
	}
	return matches[1], nil
}
