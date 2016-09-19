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

package stream

import (
	"github.com/getgauge/gauge/execution"
	"github.com/getgauge/gauge/filter"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/getgauge/gauge/reporter"
	"github.com/getgauge/gauge/util"
	"github.com/golang/protobuf/proto"
	. "gopkg.in/check.v1"
)

func (s *MySuite) TestParallelFlagOperations(c *C) {
	err := flagsMap[parallelFlag](&gauge_messages.ExecutionRequestFlag{Name: proto.String(parallelFlag), Value: proto.String("true")})

	c.Assert(err, Equals, nil)
	c.Assert(reporter.IsParallel, Equals, true)
	c.Assert(execution.InParallel, Equals, true)
}

func (s *MySuite) TestParallelFlagOperationsWithInvalidValue(c *C) {
	err := flagsMap[parallelFlag](&gauge_messages.ExecutionRequestFlag{Name: proto.String(parallelFlag), Value: proto.String("sdfsdf")})

	c.Assert(err.Error(), Equals, "Invalid value for --parallel flag. Error: strconv.ParseBool: parsing \"sdfsdf\": invalid syntax")
}

func (s *MySuite) TestVerboseFlagOperations(c *C) {
	err := flagsMap[verboseFlag](&gauge_messages.ExecutionRequestFlag{Name: proto.String(verboseFlag), Value: proto.String("false")})

	c.Assert(err, Equals, nil)
	c.Assert(reporter.Verbose, Equals, false)
}

func (s *MySuite) TestVerboseFlagOperationsWithInvalidValue(c *C) {
	err := flagsMap[verboseFlag](&gauge_messages.ExecutionRequestFlag{Name: proto.String(verboseFlag), Value: proto.String("dfdsg")})

	c.Assert(err.Error(), Equals, "Invalid value for --verbose flag. Error: strconv.ParseBool: parsing \"dfdsg\": invalid syntax")
}

func (s *MySuite) TestTagsFlagOperations(c *C) {
	err := flagsMap[tagsFlag](&gauge_messages.ExecutionRequestFlag{Name: proto.String(tagsFlag), Value: proto.String("tag1 & tag2")})

	c.Assert(err, Equals, nil)
	c.Assert(filter.ExecuteTags, Equals, "tag1 & tag2")
}

func (s *MySuite) TestTableRowsFlagOperations(c *C) {
	err := flagsMap[tableRowsFlag](&gauge_messages.ExecutionRequestFlag{Name: proto.String(tableRowsFlag), Value: proto.String("1-2")})

	c.Assert(err, Equals, nil)
	c.Assert(execution.TableRows, Equals, "1-2")
}

func (s *MySuite) TestNumberOfStreamsFlagOperations(c *C) {
	err := flagsMap[nFlag](&gauge_messages.ExecutionRequestFlag{Name: proto.String(nFlag), Value: proto.String("3")})

	c.Assert(err, Equals, nil)
	c.Assert(execution.NumberOfExecutionStreams, Equals, 3)
	c.Assert(filter.NumberOfExecutionStreams, Equals, 3)
}

func (s *MySuite) TestNumberOfStreamsFlagOperationsWithInvalidValue(c *C) {
	err := flagsMap[nFlag](&gauge_messages.ExecutionRequestFlag{Name: proto.String(nFlag), Value: proto.String("ssdf123")})

	c.Assert(err.Error(), Equals, "Invalid value for -n flag. Error: strconv.ParseInt: parsing \"ssdf123\": invalid syntax")
}

func (s *MySuite) TestStrategyFlagOperations(c *C) {
	err := flagsMap[strategyFlag](&gauge_messages.ExecutionRequestFlag{Name: proto.String(strategyFlag), Value: proto.String("eager")})

	c.Assert(err, Equals, nil)
	c.Assert(execution.Strategy, Equals, "eager")
}

func (s *MySuite) TestSortFlagOperations(c *C) {
	err := flagsMap[sortFlag](&gauge_messages.ExecutionRequestFlag{Name: proto.String(sortFlag), Value: proto.String("true")})

	c.Assert(err, Equals, nil)
	c.Assert(filter.DoNotRandomize, Equals, true)
}

func (s *MySuite) TestSortFlagOperationsWithInvalidValue(c *C) {
	err := flagsMap[sortFlag](&gauge_messages.ExecutionRequestFlag{Name: proto.String(sortFlag), Value: proto.String("sdfsdf")})

	c.Assert(err.Error(), Equals, "Invalid value for --sort flag. Error: strconv.ParseBool: parsing \"sdfsdf\": invalid syntax")
}

func (s *MySuite) TestSetFlags(c *C) {
	nFlag := &gauge_messages.ExecutionRequestFlag{Name: proto.String(nFlag), Value: proto.String("4")}
	pFlag := &gauge_messages.ExecutionRequestFlag{Name: proto.String(parallelFlag), Value: proto.String("true")}
	sFlag := &gauge_messages.ExecutionRequestFlag{Name: proto.String(strategyFlag), Value: proto.String("eager")}
	vFlag := &gauge_messages.ExecutionRequestFlag{Name: proto.String(verboseFlag), Value: proto.String("true")}

	errs := setFlags([]*gauge_messages.ExecutionRequestFlag{nFlag, pFlag, sFlag, vFlag})

	c.Assert(len(errs), Equals, 0)
}

func (s *MySuite) TestSetFlagsWithInvalidValues(c *C) {
	nFlag := &gauge_messages.ExecutionRequestFlag{Name: proto.String(nFlag), Value: proto.String("true")}
	pFlag := &gauge_messages.ExecutionRequestFlag{Name: proto.String(parallelFlag), Value: proto.String("123")}
	sFlag := &gauge_messages.ExecutionRequestFlag{Name: proto.String(strategyFlag), Value: proto.String("lazy")}
	vFlag := &gauge_messages.ExecutionRequestFlag{Name: proto.String(verboseFlag), Value: proto.String("true")}

	errs := setFlags([]*gauge_messages.ExecutionRequestFlag{nFlag, pFlag, sFlag, vFlag})

	c.Assert(len(errs), Equals, 2)
}

func (s *MySuite) TestSetFlagsWithNumberOfStreamValidationError(c *C) {
	nFlag := &gauge_messages.ExecutionRequestFlag{Name: proto.String(nFlag), Value: proto.String("-2")}
	pFlag := &gauge_messages.ExecutionRequestFlag{Name: proto.String(parallelFlag), Value: proto.String("true")}

	errs := setFlags([]*gauge_messages.ExecutionRequestFlag{nFlag, pFlag})

	c.Assert(len(errs), Equals, 1)
}

func (s *MySuite) TestSetFlagsWithStrategyValidationError(c *C) {
	nFlag := &gauge_messages.ExecutionRequestFlag{Name: proto.String(nFlag), Value: proto.String("2")}
	sFlag := &gauge_messages.ExecutionRequestFlag{Name: proto.String(strategyFlag), Value: proto.String("wrong")}
	pFlag := &gauge_messages.ExecutionRequestFlag{Name: proto.String(parallelFlag), Value: proto.String("true")}

	errs := setFlags([]*gauge_messages.ExecutionRequestFlag{nFlag, pFlag, sFlag})

	c.Assert(len(errs), Equals, 1)
}

func (s *MySuite) TestResetFlags(c *C) {
	execution.Strategy = "HAHAH"
	reporter.IsParallel = true
	execution.InParallel = false
	reporter.Verbose = true
	filter.ExecuteTags = "sdfdsf"
	execution.TableRows = "1323"
	execution.NumberOfExecutionStreams = 1
	reporter.NumberOfExecutionStreams = 2
	filter.NumberOfExecutionStreams = 3
	filter.DoNotRandomize = true
	resetFlags()

	cores := util.NumberOfCores()

	c.Assert(execution.Strategy, Equals, "lazy")
	c.Assert(execution.NumberOfExecutionStreams, Equals, cores)
	c.Assert(reporter.NumberOfExecutionStreams, Equals, cores)
	c.Assert(filter.NumberOfExecutionStreams, Equals, cores)
	c.Assert(filter.DoNotRandomize, Equals, false)
	c.Assert(execution.TableRows, Equals, "")
	c.Assert(filter.ExecuteTags, Equals, "")
	c.Assert(reporter.Verbose, Equals, false)
	c.Assert(reporter.IsParallel, Equals, false)
	c.Assert(execution.InParallel, Equals, false)
}
