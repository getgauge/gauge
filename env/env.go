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

package env

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/dmotylev/goproperties"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
)

const (
	envDefaultDirName = "default"
)

var CurrentEnv = "default"
var ProjectEnv = "default"

// LoadEnv loads default and user specified env.
// This way user specified env variable can override default if required
func LoadEnv(isDefaultEnvRequired bool) {
	err := loadEnvironment(envDefaultDirName)
	if err != nil {
		if !isDefaultEnvRequired {
			logger.Fatalf("Failed to load the default environment. %s\n", err.Error())
		}
	}

	if ProjectEnv != envDefaultDirName {
		err := loadEnvironment(ProjectEnv)
		if err != nil {
			if !isDefaultEnvRequired {
				logger.Fatalf("Failed to load the environment: %s. %s\n", ProjectEnv, err.Error())
			}
		}
		CurrentEnv = ProjectEnv
	}

}

// Loads all the properties files available in the specified env directory
func loadEnvironment(env string) error {
	envDir := filepath.Join(config.ProjectRoot, common.EnvDirectoryName)

	dirToRead := path.Join(envDir, env)
	if !common.DirExists(dirToRead) {
		return fmt.Errorf("%s environment does not exist", env)
	}

	isProperties := func(fileName string) bool {
		return filepath.Ext(fileName) == ".properties"
	}

	err := filepath.Walk(dirToRead, func(path string, info os.FileInfo, err error) error {
		if isProperties(path) {
			p, e := properties.Load(path)
			if e != nil {
				return fmt.Errorf("Failed to parse: %s. %s", path, e.Error())
			}

			for k, v := range p {
				err := common.SetEnvVariable(k, v)
				if err != nil {
					return fmt.Errorf("%s: %s", path, err.Error())
				}
			}
		}
		return nil
	})

	return err
}
