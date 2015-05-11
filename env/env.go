package env

import (
	"errors"
	"fmt"
	"github.com/dmotylev/goproperties"
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"os"
	"path"
	"path/filepath"
)

const (
	envDefaultDirName = "default"
)
var CurrentEnv = "default"

// Loading default environment and loading user specified env
// this way user specified env variable can override default if required
func LoadEnv(env string, shouldSkip bool) {
	err := loadEnvironment(envDefaultDirName)
	if err != nil {
		if !shouldSkip {
			logger.Log.Critical("Failed to load the default environment. %s\n", err.Error())
			os.Exit(1)
		}
	}

	if env != envDefaultDirName {
		err := loadEnvironment(env)
		if err != nil {
			if !shouldSkip {
				logger.Log.Critical("Failed to load the environment: %s. %s\n", env, err.Error())
				os.Exit(1)
			}
		}
		CurrentEnv = env
	}

}

// Loads all the properties files available in the specified env directory
func loadEnvironment(env string) error {
	envDir := filepath.Join(config.ProjectRoot, common.EnvDirectoryName)

	dirToRead := path.Join(envDir, env)
	if !common.DirExists(dirToRead) {
		return errors.New(fmt.Sprintf("%s is an invalid environment", env))
	}

	isProperties := func(fileName string) bool {
		return filepath.Ext(fileName) == ".properties"
	}

	err := filepath.Walk(dirToRead, func(path string, info os.FileInfo, err error) error {
		if isProperties(path) {
			p, e := properties.Load(path)
			if e != nil {
				return errors.New(fmt.Sprintf("Failed to parse: %s. %s", path, e.Error()))
			}

			for k, v := range p {
				err := common.SetEnvVariable(k, v)
				if err != nil {
					return errors.New(fmt.Sprintf("%s: %s", path, err.Error()))
				}
			}
		}
		return nil
	})

	return err
}
