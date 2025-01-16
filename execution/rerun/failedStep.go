package rerun

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/getgauge/common"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/execution/event"
	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/util"
)

const (
	failedStepsFile = "failed_steps.json"
)

type StepFailureMetadata struct {
	SpecFile    string `json:"specFile"`
	ScenarioPos int    `json:"scenarioPos"`
	StepPos     int    `json:"stepPos"`
	StepText    string `json:"stepText"`
	StackTrace  string `json:"stackTrace"`
}

type StepFailureStore struct {
	FailedSteps []StepFailureMetadata `json:"failedSteps"`
}

func ListenFailedSteps(wg *sync.WaitGroup, specDirs []string) {
	ch := make(chan event.ExecutionEvent)

	event.Register(ch, event.StepEnd)

	store := &StepFailureStore{FailedSteps: make([]StepFailureMetadata, 0)}

	wg.Add(1)

	go func() {
		for e := range ch {
			if e.Topic == event.StepEnd {
				stepRes := e.Result.(*result.StepResult)
				if stepRes.GetFailed() {
					stepInfo := e.Item.(gauge.Step)
					executionInfo := e.ExecutionInfo

					metadata := StepFailureMetadata{
						SpecFile: util.RelPathToProjectRoot(executionInfo.CurrentSpec.FileName),
						StepPos:  int(stepInfo.LineNo),
						StepText: stepInfo.Value,
					}
					store.FailedSteps = append(store.FailedSteps, metadata)
				}
			}

			if e.Topic == event.SuiteEnd {
				writeFailedSteps(store)
			}
		}
	}()

}

func writeFailedSteps(store *StepFailureStore) {
	content, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		logger.Warningf(true, "Failed to marshal failed steps data: %s", err.Error())
		return
	}

	failedStepsPath := filepath.Join(config.ProjectRoot, common.DotGauge, failedStepsFile)
	if err := os.WriteFile(failedStepsPath, content, common.NewFilePermissions); err != nil {
		logger.Warningf(true, "Failed to write failed steps file: %s", err.Error())
	}
}

func GetFailedSteps() ([]StepFailureMetadata, error) {
	failedStepsPath := filepath.Join(config.ProjectRoot, common.DotGauge, failedStepsFile)
	content, err := common.ReadFileContents(failedStepsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read failed steps file: %s", err.Error())
	}

	var store StepFailureStore
	if err := json.Unmarshal([]byte(content), &store); err != nil {
		return nil, fmt.Errorf("failed to parse failed steps data: %s", err.Error())
	}

	return store.FailedSteps, nil
}

func RerunStep(stepMeta StepFailureMetadata) []string {
	return []string{
		stepMeta.SpecFile,
		fmt.Sprintf("--scenario-index=%d", stepMeta.ScenarioPos),
		fmt.Sprintf("--step-index=%d", stepMeta.StepPos),
	}
}
