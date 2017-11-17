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
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/util"
)

const (
	dotGauge     = ".gauge"
	infoFileName = "lastRunInfo.json"
)

var lastRunInfo *lastRunMeta

func init() {
	lastRunInfo = newLastRunMetaData()
}

type lastRunMeta struct {
	Args           []string
	failedItemsMap map[string]map[string]bool
	FailedItems    []string
	LastRunItems   []string
}

func (m *lastRunMeta) failedArgs() []string {
	return append(m.Args, m.FailedItems...)
}

func (m *lastRunMeta) lastRunArgs() []string {
	return append(m.Args, m.LastRunItems...)
}

func (m *lastRunMeta) getFailedItems() []string {
	failedItems := []string{}
	for _, v := range m.failedItemsMap {
		for k := range v {
			failedItems = append(failedItems, k)
		}
	}
	return failedItems
}

func (m *lastRunMeta) aggregateFailedItems() {
	m.FailedItems = m.getFailedItems()
}

func newLastRunMetaData() *lastRunMeta {
	return &lastRunMeta{Args: make([]string, 0), failedItemsMap: make(map[string]map[string]bool), FailedItems: []string{}}
}

func (m *lastRunMeta) addFailedItem(itemName string, item string) {
	if _, ok := m.failedItemsMap[itemName]; !ok {
		m.failedItemsMap[itemName] = make(map[string]bool, 0)
	}
	m.failedItemsMap[itemName][item] = true
}

func addLastItem(items []string) {
	lastRunInfo.LastRunItems = items
}

// ListenFailedScenarios listens to execution events and writes the failed scenarios to JSON file
func ListenFailedScenarios(wg *sync.WaitGroup, specDirs []string) {
	ch := make(chan event.ExecutionEvent, 0)
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
				addLastItem(specDirs)
				lastRunInfo.aggregateFailedItems()
				writeFailedMeta(getJSON(lastRunInfo))
				wg.Done()
			}
		}
	}()
}

func prepareScenarioFailedMetadata(res *result.ScenarioResult, sce *gauge.Scenario, executionInfo gauge_messages.ExecutionInfo) {
	if res.GetFailed() {
		specPath := executionInfo.GetCurrentSpec().GetFileName()
		failedScenario := util.RelPathToProjectRoot(specPath)
		lastRunInfo.addFailedItem(specPath, fmt.Sprintf("%s:%v", failedScenario, sce.Span.Start))
	}
}

func addSpecFailedMetadata(res result.Result, args []string) {
	fileName := util.RelPathToProjectRoot(res.(*result.SpecResult).ProtoSpec.GetFileName())
	if _, ok := lastRunInfo.failedItemsMap[fileName]; ok {
		delete(lastRunInfo.failedItemsMap, fileName)
	}
	lastRunInfo.addFailedItem(fileName, fileName)
}

func addSuiteFailedMetadata(res result.Result, args []string) {
	lastRunInfo.failedItemsMap = make(map[string]map[string]bool)
	for _, arg := range args {
		path, err := filepath.Abs(arg)
		path = util.RelPathToProjectRoot(path)
		if err == nil {
			lastRunInfo.addFailedItem(path, path)
		}
	}
}

func addFailedMetadata(res result.Result, args []string, add func(res result.Result, args []string)) {
	if len(res.GetPostHook()) > 0 || len(res.GetPreHook()) > 0 {
		add(res, args)
	}
}

func writeFailedMeta(contents string) {
	failuresFile := filepath.Join(config.ProjectRoot, dotGauge, infoFileName)
	dotGaugeDir := filepath.Join(config.ProjectRoot, dotGauge)
	if err := os.MkdirAll(dotGaugeDir, common.NewDirectoryPermissions); err != nil {
		logger.Fatalf("Failed to create directory in %s. Reason: %s", dotGaugeDir, err.Error())
	}
	err := ioutil.WriteFile(failuresFile, []byte(contents), common.NewFilePermissions)
	if err != nil {
		logger.Fatalf("Failed to write to %s. Reason: %s", failuresFile, err.Error())
	}
}

func getJSON(failedMeta *lastRunMeta) string {
	j, err := json.MarshalIndent(failedMeta, "", "\t")
	if err != nil {
		logger.Warningf("Failed to save run info. Reason: %s", err.Error())
	}
	return string(j)
}

func GetLastState(repeat bool) ([]string, error) {
	meta := readLastState()
	util.SetWorkingDir(config.ProjectRoot)
	if repeat {
		return meta.lastRunArgs(), nil
	}
	if len(meta.FailedItems) == 0 {
		return nil, errors.New("No failed tests found.")
	}
	return meta.failedArgs(), nil
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
			lastRunInfo.Args = append(lastRunInfo.Args, a)
		}
	}
}

func readLastState() *lastRunMeta {
	contents, err := common.ReadFileContents(filepath.Join(config.ProjectRoot, dotGauge, infoFileName))
	if err != nil {
		logger.Fatalf("Failed to read last run information. Reason: %s", err.Error())
	}
	meta := newLastRunMetaData()
	if err = json.Unmarshal([]byte(contents), meta); err != nil {
		logger.Fatalf("Invalid last run information. Reason: %s", err.Error())
	}
	return meta
}
