package util

import (
	"github.com/getgauge/common"
	"github.com/getgauge/gauge/logger"
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
