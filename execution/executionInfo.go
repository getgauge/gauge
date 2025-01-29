/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package execution

import (
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/manifest"
	"github.com/getgauge/gauge/plugin"
	"github.com/getgauge/gauge/runner"
)

type executionInfo struct {
	manifest        *manifest.Manifest
	specs           *gauge.SpecCollection
	runner          runner.Runner
	pluginHandler   plugin.Handler
	errMaps         *gauge.BuildErrors
	inParallel      bool
	numberOfStreams int
	tagsToFilter    string
	stream          int
}

func newExecutionInfo(s *gauge.SpecCollection, r runner.Runner, ph plugin.Handler, e *gauge.BuildErrors, p bool, stream int) *executionInfo {
	m, err := manifest.ProjectManifest()
	if err != nil {
		logger.Fatal(true, err.Error())
	}
	return &executionInfo{
		manifest:        m,
		specs:           s,
		runner:          r,
		pluginHandler:   ph,
		errMaps:         e,
		inParallel:      p,
		numberOfStreams: NumberOfExecutionStreams,
		tagsToFilter:    TagsToFilterForParallelRun,
		stream:          stream,
	}
}

func (executionInfo *executionInfo) getExecutor() suiteExecutor {
	if executionInfo.inParallel {
		return newParallelExecution(executionInfo)
	}
	return newSimpleExecution(executionInfo, true, false)
}
