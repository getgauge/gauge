package util

import (
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/logger"
	"path/filepath"
	"runtime"
	"strings"
)

func SaveFile(fileName string, content string, backup bool) {
	err := common.SaveFile(fileName, content, backup)
	if err != nil {
		logger.Log.Error("Failed to refactor '%s': %s\n", fileName, err)
	}
}

func AddPrefixToEachLine(text string, template string) string {
	lines := strings.Split(text, "\n")
	prefixedLines := make([]string, 0)
	for i, line := range lines {
		if (i == len(lines)-1) && line == "" {
			prefixedLines = append(prefixedLines, line)
		} else {
			prefixedLines = append(prefixedLines, template+line)
		}
	}
	return strings.Join(prefixedLines, "\n")
}

func NumberOfCores() int {
	return runtime.NumCPU()
}

func GetPathToFile(path string) string {
	if filepath.IsAbs(path) {
		return path
	}
	return filepath.Join(config.ProjectRoot, path)
}
