package plugin

import (
	"fmt"
	"github.com/getgauge/gauge/logger/execLogger"
	"github.com/getgauge/gauge/util"
)

type pluginLogger struct {
	pluginName string
}

func (writer *pluginLogger) Write(b []byte) (int, error) {
	message := string(b)
	prefixedMessage := util.AddPrefixToEachLine(message, fmt.Sprintf("[%s Plugin] : ", writer.pluginName))
	gaugeConsoleWriter := execLogger.Current()
	_, err := gaugeConsoleWriter.Write([]byte(prefixedMessage))
	return len(message), err
}
