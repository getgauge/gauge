/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package result

import (
	"path/filepath"
	"time"

	"github.com/getgauge/gauge-proto/go/gauge_messages"
	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/env"
)

// SuitResult represents the result of suit execution
type SuiteResult struct {
	SpecResults             []*SpecResult
	PreSuite                *(gauge_messages.ProtoHookFailure)
	PostSuite               *(gauge_messages.ProtoHookFailure)
	IsFailed                bool
	SpecsFailedCount        int
	ExecutionTime           int64 //in milliseconds
	UnhandledErrors         []error
	Environment             string
	Tags                    string
	ProjectName             string
	Timestamp               string
	SpecsSkippedCount       int
	PreHookMessages         []string
	PostHookMessages        []string
	PreHookScreenshotFiles  []string
	PostHookScreenshotFiles []string
	PreHookScreenshots      [][]byte
	PostHookScreenshots     [][]byte
}

// NewSuiteResult is a constructor for SuitResult
func NewSuiteResult(tags string, startTime time.Time) *SuiteResult {
	result := new(SuiteResult)
	result.SpecResults = make([]*SpecResult, 0)
	result.Timestamp = startTime.Format(config.LayoutForTimeStamp)
	result.ProjectName = filepath.Base(config.ProjectRoot)
	result.Environment = env.CurrentEnvironments()
	result.Tags = tags
	return result
}

// SetFailure sets the result to failed
func (sr *SuiteResult) SetFailure() {
	sr.IsFailed = true
}

// SetSpecsSkippedCount sets the count of specs skipped.
func (sr *SuiteResult) SetSpecsSkippedCount() {
	sr.SpecsSkippedCount = 0
	for _, specRes := range sr.SpecResults {
		if specRes.Skipped {
			sr.SpecsSkippedCount++
		}
	}
}

// AddUnhandledError adds the unhandled error to suit result.
func (sr *SuiteResult) AddUnhandledError(err error) {
	sr.UnhandledErrors = append(sr.UnhandledErrors, err)
}

func (sr *SuiteResult) UpdateExecTime(startTime time.Time) {
	sr.ExecutionTime = int64(time.Since(startTime) / 1e6)
}

// AddSpecResult adds a specs result to suit result.
func (sr *SuiteResult) AddSpecResult(specResult *SpecResult) {
	if specResult.IsFailed {
		sr.IsFailed = true
		sr.SpecsFailedCount++
	}
	sr.ExecutionTime += specResult.ExecutionTime
	sr.SpecResults = append(sr.SpecResults, specResult)
}

// AddSpecResults adds multiple spec results to suit result.
func (sr *SuiteResult) AddSpecResults(specResults []*SpecResult) {
	for _, result := range specResults {
		sr.AddSpecResult(result)
	}
}

func (sr *SuiteResult) GetPreHook() []*gauge_messages.ProtoHookFailure {
	if sr.PreSuite == nil {
		return []*gauge_messages.ProtoHookFailure{}
	}
	return []*gauge_messages.ProtoHookFailure{sr.PreSuite}
}

func (sr *SuiteResult) GetPostHook() []*gauge_messages.ProtoHookFailure {
	if sr.PostSuite == nil {
		return []*gauge_messages.ProtoHookFailure{}
	}
	return []*gauge_messages.ProtoHookFailure{sr.PostSuite}
}

func (sr *SuiteResult) AddPreHook(f ...*gauge_messages.ProtoHookFailure) {
	sr.PreSuite = f[0]
}

func (sr *SuiteResult) AddPostHook(f ...*gauge_messages.ProtoHookFailure) {
	sr.PostSuite = f[0]
}

// ExecTime returns the time taken to execute the suit
func (sr *SuiteResult) ExecTime() int64 {
	return sr.ExecutionTime
}

// GetFailed returns the state of the result
func (sr *SuiteResult) GetFailed() bool {
	return sr.IsFailed
}

func (sr *SuiteResult) Item() interface{} {
	return nil
}
