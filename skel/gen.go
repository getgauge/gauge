package skel

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/logger"
)

func CreateSkelFilesIfRequired() {
	p, err := common.GetConfigurationDir()
	if err != nil {
		logger.GaugeLog.Info("Unable to get path to config. Error: %s", err.Error())
		return
	}
	if _, err := os.Stat(p); os.IsNotExist(err) {
		logger.GaugeLog.Info("Config directory does not exist. Setting up config directory `%s`.", p)
		err = os.MkdirAll(p, common.NewDirectoryPermissions)
		if err != nil {
			logger.GaugeLog.Error("Unable to create config path dir `%s`. Error: %s", p, err.Error())
			return
		}
	}
	WriteFile(filepath.Join(p, "gauge.properties"), gaugeProperties)
	WriteFile(filepath.Join(p, "notice.md"), notice)
	WriteFile(filepath.Join(p, "skel", "example.spec"), exampleSpec)
	WriteFile(filepath.Join(p, "skel", "env", "default.properties"), defaultProperties)
	WriteFile(filepath.Join(p, "skel", ".gitignore"), gitignore)
}

func WriteFile(path, text string) {
	dirPath := filepath.Dir(path)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, common.NewDirectoryPermissions)
		if err != nil {
			logger.GaugeLog.Error("Unable to create dir `%s`. Error: %s", dirPath, err.Error())
			return
		}
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := ioutil.WriteFile(path, []byte(text), common.NewFilePermissions)
		if err != nil {
			logger.GaugeLog.Error("Unable to create file `%s`. Error: %s", path, err.Error())
		}
	}
}
