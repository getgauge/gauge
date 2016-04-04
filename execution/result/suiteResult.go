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

package result

import (
	"path/filepath"
	"time"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/env"
	"github.com/getgauge/gauge/gauge_messages"
)

type SuiteResult struct {
	SpecResults       []*SpecResult
	PreSuite          *(gauge_messages.ProtoHookFailure)
	PostSuite         *(gauge_messages.ProtoHookFailure)
	IsFailed          bool
	SpecsFailedCount  int
	ExecutionTime     int64 //in milliseconds
	UnhandledErrors   []error
	Environment       string
	Tags              string
	ProjectName       string
	Timestamp         string
	SpecsSkippedCount int
}

func NewSuiteResult(tags string, startTime time.Time) *SuiteResult {
	result := new(SuiteResult)
	result.SpecResults = make([]*SpecResult, 0)
	result.Timestamp = startTime.Format(config.LayoutForTimeStamp)
	result.ProjectName = filepath.Base(config.ProjectRoot)
	result.Environment = env.CurrentEnv()
	result.Tags = tags
	return result
}

func (suiteResult *SuiteResult) SetFailure() {
	suiteResult.IsFailed = true
}

func (suiteResult *SuiteResult) AddSpecResult(specResult *SpecResult) {
	if specResult.IsFailed {
		suiteResult.IsFailed = true
		suiteResult.SpecsFailedCount++
	}
	suiteResult.ExecutionTime += specResult.ExecutionTime
	suiteResult.SpecResults = append(suiteResult.SpecResults, specResult)
}

func (suiteResult *SuiteResult) getPreHook() **(gauge_messages.ProtoHookFailure) {
	return &suiteResult.PreSuite
}

func (suiteResult *SuiteResult) getPostHook() **(gauge_messages.ProtoHookFailure) {
	return &suiteResult.PostSuite
}
