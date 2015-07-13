package util

import (
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/logger"
)

func SaveFile(fileName string, content string, backup bool) {
	err := common.SaveFile(fileName, content, backup)
	if err != nil {
		logger.Log.Error("Failed to refactor '%s': %s\n", fileName, err)
	}
}
