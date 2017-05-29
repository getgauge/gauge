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
	"strings"

	"sync"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/util"
	flag "github.com/getgauge/mflag"
)

// RunFailed represents if this is a re-run of only failed scenarios or a new run
var RunFailed bool

const (
	dotGauge   = ".gauge"
	failedFile = "failures.json"
)

var failedMeta *failedMetadata

func init() {
	failedMeta = newFailedMetaData()
}

type failedMetadata struct {
	Flags          map[string]string
	failedItemsMap map[string]map[string]bool
	FailedItems    []string
}

func (m *failedMetadata) String() string {
	cmd := "gauge "
	for flag, value := range m.Flags {
		cmd += "-" + flag + "=" + value + " "
	}
	return cmd + strings.Join(m.FailedItems, " ")
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
	return &failedMetadata{Flags: make(map[string]string), failedItemsMap: make(map[string]map[string]bool), FailedItems: []string{}}
}

func (m *failedMetadata) addFailedItem(itemName string, item string) {
	if _, ok := m.failedItemsMap[itemName]; !ok {
		m.failedItemsMap[itemName] = make(map[string]bool, 0)
	}
	m.failedItemsMap[itemName][item] = true
}

// ListenFailedScenarios listens to execution events and writes the failed scenarios to JSON file
func ListenFailedScenarios(wg *sync.WaitGroup) {
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
				addFailedMetadata(e.Result, addSpecFailedMetadata)
			case event.SuiteEnd:
				addFailedMetadata(e.Result, addSuiteFailedMetadata)
				failedMeta.aggregateFailedItems()
				writeFailedMeta(getJSON(failedMeta))
				wg.Done()
			}
		}
	}()
}

func prepareScenarioFailedMetadata(res *result.ScenarioResult, sce *gauge.Scenario, executionInfo gauge_messages.ExecutionInfo) {
	if res.GetFailed() {
		specPath := executionInfo.GetCurrentSpec().GetFileName()
		failedScenario := util.RelPathToProjectRoot(specPath)
		failedMeta.addFailedItem(specPath, fmt.Sprintf("%s:%v", failedScenario, sce.Span.Start))
	}
}

func addSpecFailedMetadata(res result.Result) {
	fileName := util.RelPathToProjectRoot(res.(*result.SpecResult).ProtoSpec.GetFileName())
	if _, ok := failedMeta.failedItemsMap[fileName]; ok {
		delete(failedMeta.failedItemsMap, fileName)
	}
	failedMeta.addFailedItem(fileName, fileName)
}

func addSuiteFailedMetadata(res result.Result) {
	failedMeta.failedItemsMap = make(map[string]map[string]bool)
	for _, arg := range flag.Args() {
		path, err := filepath.Abs(arg)
		path = util.RelPathToProjectRoot(path)
		if err == nil {
			failedMeta.addFailedItem(path, path)
		}
	}
}

func addFailedMetadata(res result.Result, add func(res result.Result)) {
	if len(res.GetPostHook()) > 0 || len(res.GetPreHook()) > 0 {
		add(res)
	}
}

func writeFailedMeta(contents string) {
	failuresFile := filepath.Join(config.ProjectRoot, dotGauge, failedFile)
	dotGaugeDir := filepath.Join(config.ProjectRoot, dotGauge)
	if err := os.MkdirAll(dotGaugeDir, common.NewDirectoryPermissions); err != nil {
		logger.Fatalf("Failed to create directory in %s. Reason: %s", dotGaugeDir, err.Error())
	}
	err := ioutil.WriteFile(failuresFile, []byte(contents), common.NewFilePermissions)
	if err != nil {
		logger.Fatalf("Failed to write to %s. Reason: %s", failuresFile, err.Error())
	}
}

func getJSON(failedMeta *failedMetadata) string {
	json, err := json.MarshalIndent(failedMeta, "", "\t")
	if err != nil {
		logger.Warning("Failed to save run info. Reason: %s", err.Error())
	}
	return string(json)
}

func saveFlagState(f *flag.Flag) {
	failedMeta.Flags[f.Names[0]] = f.Value.String()
}

func setDefault(f *flag.Flag) {
	f.Value.Set(f.DefValue)
}

// setFlags sets the flags if its a re-run of failed scenarios. Else, it will save the current execution run for next re-run.
func setFlags(meta *failedMetadata) {
	flag.VisitAll(setDefault)
	contents, err := common.ReadFileContents(filepath.Join(config.ProjectRoot, dotGauge, failedFile))
	if err != nil {
		logger.Fatalf("Failed to read last run information. Reason: %s", err.Error())
	}
	if err = json.Unmarshal([]byte(contents), &meta); err != nil {
		logger.Fatalf("Invalid last run information. Reason: %s", err.Error())
	}
	failedMeta.Flags = meta.Flags
	for k, v := range meta.Flags {
		err = flag.Set(k, v)
		if err != nil {
			logger.Warning("Failed to set flag %v to %v. Reason: %v", k, v, err.Error())
		}
	}
	flag.CommandLine.Parse(meta.FailedItems)
}

// Initialize sets up rerun and determines whether any failed tests will be executed or not
func Initialize() error {
	if !RunFailed {
		flag.Visit(saveFlagState)
		return nil
	}
	meta := new(failedMetadata)
	setFlags(meta)
	util.SetWorkingDir(config.ProjectRoot)
	if RunFailed && len(meta.FailedItems) == 0 {
		return errors.New("No failed tests found.")
	}
	fmt.Printf("Executing => %s\n", meta.String())
	return nil
}
