package util

import (
	"os"
	"runtime"
	"strings"
)

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

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func IsProcessRunning(processID int) bool {
	_, err := os.FindProcess(processID)
	return err == nil
}
