// Copyright 2015 ThoughtWorks, Inc.

// This file is part of Gauge.

// Gauge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// Gauge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with Gauge.  If not, see <http://www.gnu.org/licenses/>.

package install

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/dmotylev/goproperties"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/version"
)

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
			logger.Info(fmt.Sprintf("%-10s\t\t%-10s\t%s", update.Name, update.CompatibleVersion, update.Message))
		}
	} else {
		logger.Info("No Updates available.")
	}
}

func checkUpdates() []UpdateInfo {
	return append(checkGaugeUpdate(), checkPluginUpdates()...)
}

func printUpdateInfo(print chan bool, wg *sync.WaitGroup) {
	message := make(chan string)
	go func() {
		updates := checkUpdates()
		if len(updates) > 0 {
			message <- "Updates are available. Run gauge --check-updates for more info."
		}
	}()
	waitToPrint(message, print, "", wg)
}

func waitToPrint(messageChan chan string, printChan chan bool, message string, wg *sync.WaitGroup) {
	select {
	case <-printChan:
		if message != "" {
			logger.Info(message)
		}
		wg.Done()
	case message = <-messageChan:
		waitToPrint(messageChan, printChan, message, wg)
	}
}

func checkGaugeUpdate() []UpdateInfo {
	var updateInfos []UpdateInfo
	v, err := getLatestGaugeVersion(config.GaugeUpdateUrl())
	if err != nil {
		return updateInfos
	}
	latestVersion, err := version.ParseVersion(v)
	if err != nil {
		return updateInfos
	}
	isLatestVersion := version.CurrentGaugeVersion.IsLesserThan(latestVersion)
	if isLatestVersion {
		updateInfos = append(updateInfos, UpdateInfo{"Gauge", latestVersion.String(), "Download the installer from http://getgauge.io/get-started/"})
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
	plugins, err := plugin.GetAllInstalledPluginsWithVersion()
	if err != nil {
		return pluginsToUpdate
	}
	for _, plugin := range plugins {
		desc, result := getInstallDescription(plugin.Name)
		if result.Error != nil {
			continue
		}
		pluginsToUpdate = append(pluginsToUpdate, createPluginUpdateDetail(plugin.Version.String(), *desc)...)
	}
	return pluginsToUpdate
}

func createPluginUpdateDetail(currentVersion string, latestVersionDetails installDescription) []UpdateInfo {
	var updateInfo []UpdateInfo
	v, _ := version.ParseVersion(currentVersion)
	v1, _ := version.ParseVersion(latestVersionDetails.Versions[0].Version)
	if v.IsLesserThan(v1) {
		versionDesc, err := latestVersionDetails.getLatestCompatibleVersionTo(version.CurrentGaugeVersion)
		if err != nil {
			return updateInfo
		}
		updateInfo = append(updateInfo, UpdateInfo{latestVersionDetails.Name, versionDesc.Version, fmt.Sprintf("Run 'gauge --update %s'", latestVersionDetails.Name)})
	}
	return updateInfo
}

var getLatestGaugeVersion = func(url string) (string, error) {
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	v, err := getGaugeVersionProperty(res.Body)
	if err != nil {
		return "", err
	}
	return v, nil
}

func getGaugeVersionProperty(r io.Reader) (string, error) {
	properties := make(properties.Properties)

	err := properties.Load(r)
	if err != nil {
		return "", fmt.Errorf("Failed to parse: %s", err.Error())
	}
	return strings.TrimSpace(properties["version"]), nil
}
