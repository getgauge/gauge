package execution

import (
	"encoding/json"

	"github.com/getgauge/gauge/logger"
)

type executionStatus struct {
	Type          string `json:"type"`
	SpecsExecuted int    `json:"specsExecuted"`
	SpecsPassed   int    `json:"specsPassed"`
	SpecsFailed   int    `json:"specsFailed"`
	SpecsSkipped  int    `json:"specsSkipped"`
	SceExecuted   int    `json:"sceExecuted"`
	ScePassed     int    `json:"scePassed"`
	SceFailed     int    `json:"sceFailed"`
	SceSkipped    int    `json:"sceSkipped"`
}

func (status *executionStatus) getJSON() (string, error) {
	j, err := json.Marshal(status)
	if err != nil {
		return "", err
	}
	return string(j), nil
}

func statusJSON(executedSpecs, passedSpecs, failedSpecs, skippedSpecs, executedScenarios, passedScenarios, failedScenarios, skippedScenarios int) string {
	executionStatus := &executionStatus{}
	executionStatus.Type = "out"
	executionStatus.SpecsExecuted = executedSpecs
	executionStatus.SpecsPassed = passedSpecs
	executionStatus.SpecsFailed = failedSpecs
	executionStatus.SpecsSkipped = skippedSpecs
	executionStatus.SceExecuted = executedScenarios
	executionStatus.ScePassed = passedScenarios
	executionStatus.SceFailed = failedScenarios
	executionStatus.SceSkipped = skippedScenarios
	s, err := executionStatus.getJSON()
	if err != nil {
		logger.Fatalf(true, "Unable to parse execution status information : %v", err.Error())
	}
	return s
}
