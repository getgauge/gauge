/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package rerun

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/util"
)

const (
	failedFile         = "failures.json"
	lastRunCmdFileName = "lastRunCmd.json"
)

var failedMeta *failedMetadata

func init() {
	failedMeta = newFailedMetaData()
}

type failedMetadata struct {
	Args           []string
	failedItemsMap map[string]map[string]bool
	FailedItems    []string
}

func (m *failedMetadata) args() []string {
	return append(m.Args, m.FailedItems...)
}

func (m *failedMetadata) getFailedItems() []string {
	failedItems := []string{}
	for _, v := range m.failedItemsMap {
		for k := range v {
			failedItems = append(failedItems, k)
		}
	}
	return failedItems
}

func (m *failedMetadata) aggregateFailedItems() {
	m.FailedItems = m.getFailedItems()
}

func newFailedMetaData() *failedMetadata {
	return &failedMetadata{Args: make([]string, 0), failedItemsMap: make(map[string]map[string]bool), FailedItems: []string{}}
}

func (m *failedMetadata) addFailedItem(itemName string, item string) {
	if _, ok := m.failedItemsMap[itemName]; !ok {
		m.failedItemsMap[itemName] = make(map[string]bool)
	}
	m.failedItemsMap[itemName][item] = true
}

// ListenFailedScenarios listens to execution events and writes the failed scenarios to JSON file
func ListenFailedScenarios(wg *sync.WaitGroup, specDirs []string) {
	ch := make(chan event.ExecutionEvent)
	event.Register(ch, event.ScenarioEnd)
	event.Register(ch, event.SpecEnd)
	event.Register(ch, event.SuiteEnd)
	wg.Add(1)

	go func() {
		for {
			e := <-ch
			switch e.Topic {
			case event.ScenarioEnd:
				prepareScenarioFailedMetadata(e.Result.(*result.ScenarioResult), e.Item.(*gauge.Scenario), e.ExecutionInfo)
			case event.SpecEnd:
				addFailedMetadata(e.Result, specDirs, addSpecFailedMetadata)
			case event.SuiteEnd:
				addFailedMetadata(e.Result, specDirs, addSuiteFailedMetadata)
				failedMeta.aggregateFailedItems()
				writeFailedMeta(getJSON(failedMeta))
				wg.Done()
			}
		}
	}()
}

func prepareScenarioFailedMetadata(res *result.ScenarioResult, sce *gauge.Scenario, executionInfo *gauge_messages.ExecutionInfo) {
	if res.GetFailed() {
		specPath := executionInfo.GetCurrentSpec().GetFileName()
		failedScenario := util.RelPathToProjectRoot(specPath)
		failedMeta.addFailedItem(specPath, fmt.Sprintf("%s:%v", failedScenario, sce.Span.Start))
	}
}

func addSpecFailedMetadata(res result.Result, args []string) {
	fileName := util.RelPathToProjectRoot(res.(*result.SpecResult).ProtoSpec.GetFileName())
	delete(failedMeta.failedItemsMap, fileName)
	failedMeta.addFailedItem(fileName, fileName)
}

func addSuiteFailedMetadata(res result.Result, args []string) {
	failedMeta.failedItemsMap = make(map[string]map[string]bool)
	for _, arg := range args {
		path, err := filepath.Abs(arg)
		path = util.RelPathToProjectRoot(path)
		if err == nil {
			failedMeta.addFailedItem(path, path)
		}
	}
}

func addFailedMetadata(res result.Result, args []string, add func(res result.Result, args []string)) {
	if len(res.GetPostHook()) > 0 || len(res.GetPreHook()) > 0 {
		add(res, args)
	}
}

func writeFailedMeta(contents string) {
	failuresFile := filepath.Join(config.ProjectRoot, common.DotGauge, failedFile)
	dotGaugeDir := filepath.Join(config.ProjectRoot, common.DotGauge)
	if err := os.MkdirAll(dotGaugeDir, common.NewDirectoryPermissions); err != nil {
		logger.Fatalf(true, "Failed to create directory in %s. Reason: %s", dotGaugeDir, err.Error())
	}
	err := ioutil.WriteFile(failuresFile, []byte(contents), common.NewFilePermissions)
	if err != nil {
		logger.Fatalf(true, "Failed to write to %s. Reason: %s", failuresFile, err.Error())
	}
}

func getJSON(failedMeta *failedMetadata) string {
	j, err := json.MarshalIndent(failedMeta, "", "\t")
	if err != nil {
		logger.Warningf(true, "Failed to save run info. Reason: %s", err.Error())
	}
	return string(j)
}

var GetLastFailedState = func() ([]string, error) {
	meta := readLastFailedState()
	util.SetWorkingDir(config.ProjectRoot)
	if len(meta.FailedItems) == 0 {
		return nil, errors.New("No failed tests found.")
	}
	return meta.args(), nil
}

func SaveState(args []string, specs []string) {
	isPresent := func(values []string, value string) bool {
		for _, v := range values {
			if v == value {
				return true
			}
		}
		return false
	}
	for _, a := range args {
		if !isPresent(specs, a) {
			failedMeta.Args = append(failedMeta.Args, a)
		}
	}
}

func readLastFailedState() *failedMetadata {
	contents, err := common.ReadFileContents(filepath.Join(config.ProjectRoot, common.DotGauge, failedFile))
	if err != nil {
		logger.Fatalf(true, "Failed to read last run information. Reason: %s", err.Error())
	}
	meta := newFailedMetaData()
	if err = json.Unmarshal([]byte(contents), meta); err != nil {
		logger.Fatalf(true, "Invalid last run information. Reason: %s", err.Error())
	}
	return meta
}

var ReadPrevArgs = func() []string {
	contents, err := common.ReadFileContents(filepath.Join(config.ProjectRoot, common.DotGauge, lastRunCmdFileName))
	if err != nil {
		logger.Fatalf(true, "Failed to read previous command information. Reason: %s", err.Error())
		return nil
	}
	var args []string
	if err = json.Unmarshal([]byte(contents), &args); err != nil {
		logger.Fatalf(true, "Invalid previous command information. Reason: %s", err.Error())
		return nil
	}
	return args
}

var WritePrevArgs = func(cmdArgs []string) {
	b, err := json.MarshalIndent(cmdArgs, "", "\t")
	if err != nil {
		logger.Fatalf(true, "Unable to parse last run command. Error : %v", err.Error())
	}
	prevCmdFile := filepath.Join(config.ProjectRoot, common.DotGauge, lastRunCmdFileName)
	dotGaugeDir := filepath.Join(config.ProjectRoot, common.DotGauge)
	if err = os.MkdirAll(dotGaugeDir, common.NewDirectoryPermissions); err != nil {
		logger.Fatalf(true, "Failed to create directory in %s. Reason: %s", dotGaugeDir, err.Error())
	}
	err = ioutil.WriteFile(prevCmdFile, b, common.NewFilePermissions)
	if err != nil {
		logger.Fatalf(true, "Failed to write to %s. Reason: %s", prevCmdFile, err.Error())
	}
}
