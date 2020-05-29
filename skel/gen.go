/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package skel

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/template"
)

func CreateSkelFilesIfRequired() {
	p, err := common.GetConfigurationDir()
	if err != nil {
		logger.Debugf(true, "Unable to get path to config. Error: %s", err.Error())
		return
	}
	if _, err := os.Stat(p); os.IsNotExist(err) {
		err = os.MkdirAll(p, common.NewDirectoryPermissions)
		if err != nil {
			logger.Debugf(true, "Unable to create config path dir `%s`. Error: %s", p, err.Error())
			return
		}
	}
	err = config.Merge()
	if err != nil {
		logger.Debugf(true, "Unable to create gauge.properties. Error: %s", err.Error())
	}

	err = template.Generate()
	if err != nil {
		logger.Debugf(true, "Unable to create tempate.properties. Error: %s", err.Error())
	}
	writeFile(filepath.Join(p, "notice.md"), Notice, true)
	writeFile(filepath.Join(p, "skel", "example.spec"), ExampleSpec, false)
	writeFile(filepath.Join(p, "skel", "env", "default.properties"), DefaultProperties, false)
	writeFile(filepath.Join(p, "skel", ".gitignore"), Gitignore, false)
}

func writeFile(path, text string, overwrite bool) {
	dirPath := filepath.Dir(path)
	if !common.DirExists(dirPath) {
		err := os.MkdirAll(dirPath, common.NewDirectoryPermissions)
		if err != nil {
			logger.Debugf(true, "Unable to create dir `%s`. Error: %s", dirPath, err.Error())
			return
		}
	}
	if !common.FileExists(path) || overwrite {
		err := ioutil.WriteFile(path, []byte(text), common.NewFilePermissions)
		if err != nil {
			logger.Debugf(true, "Unable to create file `%s`. Error: %s", path, err.Error())
		}
	}
}
