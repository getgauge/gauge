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

package skel

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/satori/go.uuid"
)

func CreateSkelFilesIfRequired() {
	p, err := common.GetConfigurationDir()
	if err != nil {
		logger.GaugeLog.Errorf("Unable to get path to config. Error: %s", err.Error())
		return
	}
	if _, err := os.Stat(p); os.IsNotExist(err) {
		logger.GaugeLog.Infof("Config directory does not exist. Setting up config directory `%s`.", p)
		err = os.MkdirAll(p, common.NewDirectoryPermissions)
		if err != nil {
			logger.GaugeLog.Errorf("Unable to create config path dir `%s`. Error: %s", p, err.Error())
			return
		}
	}
	err = config.Merge()
	if err != nil {
		logger.GaugeLog.Errorf("Unable to create gauge.properties. Error: %s", err.Error())
	}
	writeFile(filepath.Join(p, "notice.md"), Notice)
	writeFile(filepath.Join(p, "skel", "example.spec"), ExampleSpec)
	writeFile(filepath.Join(p, "skel", "env", "default.properties"), DefaultProperties)
	writeFile(filepath.Join(p, "skel", ".gitignore"), Gitignore)

	idFile := filepath.Join(p, "id")
	if !common.FileExists(idFile) {
		writeFile(idFile, uuid.NewV4().String())
	}
}

func writeFile(path, text string) {
	dirPath := filepath.Dir(path)
	if !common.DirExists(dirPath) {
		err := os.MkdirAll(dirPath, common.NewDirectoryPermissions)
		if err != nil {
			logger.GaugeLog.Errorf("Unable to create dir `%s`. Error: %s", dirPath, err.Error())
			return
		}
	}
	if !common.FileExists(path) {
		err := ioutil.WriteFile(path, []byte(text), common.NewFilePermissions)
		if err != nil {
			logger.GaugeLog.Errorf("Unable to create file `%s`. Error: %s", path, err.Error())
		}
	}
}
