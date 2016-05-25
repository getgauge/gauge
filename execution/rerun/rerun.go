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
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/logger"
	flag "github.com/getgauge/mflag"
)

// RunFailed represents if this is a re-run of only failed scenarios or a new run
var RunFailed bool

const (
	dotGauge   = ".gauge"
	failedFile = "failed.json"
)

var failedMeta *failedMetadata

func init() {
	failedMeta = newFailedMetaData()
}

type failedMetadata struct {
	Flags          map[string]string
	failedItemsMap map[string][]string
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
		failedItems = append(failedItems, v...)
	}
	return failedItems
}

func (m *failedMetadata) aggregateFailedItems() {
	m.FailedItems = m.getFailedItems()
}

func newFailedMetaData() *failedMetadata {
	return &failedMetadata{Flags: make(map[string]string), failedItemsMap: make(map[string][]string), FailedItems: []string{}}
}

func (m *failedMetadata) addFailedItem(itemName string, item string) {
	if _, ok := m.failedItemsMap[itemName]; !ok {
		m.failedItemsMap[itemName] = make([]string, 0)
	}
	m.failedItemsMap[itemName] = append(m.failedItemsMap[itemName], item)
}

// ListenFailedScenarios listens to execution events and writes the failed scenarios to JSON file
func ListenFailedScenarios() {
	ch := make(chan event.ExecutionEvent, 0)
	event.Register(ch, event.ScenarioEnd)
	event.Register(ch, event.SpecEnd)
	event.Register(ch, event.SuiteEnd)

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
			}
		}
	}()
}

func prepareScenarioFailedMetadata(res *result.ScenarioResult, sce *gauge.Scenario, executionInfo gauge_messages.ExecutionInfo) {
	if res.GetFailed() {
		specPath := executionInfo.GetCurrentSpec().GetFileName()
		failedScenario := getRelativePath(specPath)
		failedMeta.addFailedItem(specPath, fmt.Sprintf("%s:%v", failedScenario, sce.Span.Start))
	}
}

func getRelativePath(path string) string {
	return strings.TrimPrefix(path, config.ProjectRoot+string(filepath.Separator))
}

func addSpecFailedMetadata(res result.Result) {
	fileName := getRelativePath(res.(*result.SpecResult).ProtoSpec.GetFileName())
	if _, ok := failedMeta.failedItemsMap[fileName]; ok {
		failedMeta.failedItemsMap[fileName] = []string{}
	}
	failedMeta.addFailedItem(fileName, fileName)
}

func addSuiteFailedMetadata(res result.Result) {
	failedMeta.failedItemsMap = make(map[string][]string)
	for _, arg := range flag.Args() {
		path, err := filepath.Abs(arg)
		path = getRelativePath(path)
		if err == nil {
			failedMeta.addFailedItem(path, path)
		}
	}
}

func addFailedMetadata(res result.Result, add func(res result.Result)) {
	if (res.GetPostHook() != nil && *res.GetPostHook() != nil) || (res.GetPreHook() != nil && *res.GetPreHook() != nil) {
		add(res)
	}
}

func writeFailedMeta(contents string) {
	failedPath := filepath.Join(config.ProjectRoot, dotGauge, failedFile)
	dotGaugeDir := filepath.Join(config.ProjectRoot, dotGauge)
	if err := os.MkdirAll(dotGaugeDir, common.NewDirectoryPermissions); err != nil {
		logger.Fatalf("Failed to create directory in %s. Reason: %s", dotGaugeDir, err.Error())
	}
	err := ioutil.WriteFile(failedPath, []byte(contents), common.NewFilePermissions)
	if err != nil {
		logger.Fatalf("Failed to write to %s. Reason: %s", failedPath, err.Error())
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

// SetFlags sets the flags if its a re-run of failed scenarios. Else, it will save the current execution run for next re-run.
func SetFlags() {
	if !RunFailed {
		flag.Visit(saveFlagState)
		return
	}
	flag.VisitAll(setDefault)
	contents, err := common.ReadFileContents(filepath.Join(config.ProjectRoot, dotGauge, failedFile))
	if err != nil {
		logger.Fatalf("Failed to read last run information. Reason: %s", err.Error())
	}
	var meta failedMetadata
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
	fmt.Printf("Executing => %s\n", meta.String())
}
