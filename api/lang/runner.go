package lang

import (
	"os"

	"github.com/getgauge/gauge/api"
	"github.com/getgauge/gauge/conn"
	gm "github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/runner"
	"github.com/getgauge/gauge/util"
)

type langRunner struct {
	runner   runner.Runner
	killChan chan bool
}

var lRunner langRunner

func startRunner() {
	lRunner.killChan = make(chan bool)
	var err error
	lRunner.runner, err = connectToRunner(lRunner.killChan)
	if err != nil {
		logger.APILog.Infof("Unable to connect to runner : %s", err.Error())
	}
}

func connectToRunner(killChan chan bool) (runner.Runner, error) {
	outfile, err := os.OpenFile(logger.GetLogFilePath(logger.GaugeLogFileName), os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		logger.APILog.Infof("%s", err.Error())
		return nil, err
	}
	runner, err := api.ConnectToRunner(killChan, false, outfile)
	if err != nil {
		logger.APILog.Infof("%s", err.Error())
		return nil, err
	}
	return runner, nil
}

func cacheFileOnRunner(uri, text string) error {
	cacheFileRequest := &gm.Message{MessageType: gm.Message_CacheFileRequest, CacheFileRequest: &gm.CacheFileRequest{Content: text, FilePath: util.ConvertURItoFilePath(uri), IsClosed: false}}
	err := sendMessageToRunner(cacheFileRequest)
	return err
}

func sendMessageToRunner(cacheFileRequest *gm.Message) error {
	err := conn.WriteGaugeMessage(cacheFileRequest, lRunner.runner.Connection())
	if err != nil {
		logger.APILog.Infof("Error while connecting to runner : %s", err.Error())
	}
	return err
}

func killRunner() {
	lRunner.runner.Kill()
}
