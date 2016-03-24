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

package execution

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/getgauge/gauge/execution/result"
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/gauge_messages"
	"github.com/golang/protobuf/proto"

	. "gopkg.in/check.v1"
)

type testLogger struct {
	output string
}

func (l *testLogger) Write(b []byte) (int, error) {
	l.output = string(b)
	return len(b), nil
}

func (s *MySuite) TestFunctionsOfTypeSpecList(c *C) {
	mySpecs := gauge.NewSpecCollection(createSpecsList(4))
	c.Assert(mySpecs.Next().FileName, Equals, "spec0")
	c.Assert(mySpecs.Next().FileName, Equals, "spec1")
	c.Assert(mySpecs.HasNext(), Equals, true)
	c.Assert(mySpecs.Size(), Equals, 4)
	c.Assert(mySpecs.Next().FileName, Equals, "spec2")
	c.Assert(mySpecs.Next().FileName, Equals, "spec3")
	c.Assert(mySpecs.HasNext(), Equals, false)
}

func createSpecsList(number int) []*gauge.Specification {
	var specs []*gauge.Specification
	for i := 0; i < number; i++ {
		specs = append(specs, &gauge.Specification{FileName: fmt.Sprint("spec", i)})
	}
	return specs
}

func (s *MySuite) TestValidateFlagsIfNotParallel(c *C) {
	if os.Getenv("EXIT_VALIDATE") == "1" {
		InParallel = false
		validateFlags()
		return
	}
	cmd := exec.Command(os.Args[0], "-check.f=MySuite.TestValidateFlagsIfNotParallel")
	cmd.Env = append(os.Environ(), "EXIT_VALIDATE=1")
	err := cmd.Run()
	c.Assert(err, Equals, nil)
}

func (s *MySuite) TestValidateFlagsWithStartegyEager(c *C) {
	if os.Getenv("EXIT_VALIDATE") == "1" {
		InParallel = true
		Strategy = "eager"
		NumberOfExecutionStreams = 1
		validateFlags()
		return
	}
	cmd := exec.Command(os.Args[0], "-check.f=MySuite.TestValidateFlagsWithStartegyEager")
	cmd.Env = append(os.Environ(), "EXIT_VALIDATE=1")
	err := cmd.Run()
	c.Assert(err, Equals, nil)
}

func (s *MySuite) TestValidateFlagsWithStartegyLazy(c *C) {
	if os.Getenv("EXIT_VALIDATE") == "1" {
		InParallel = true
		Strategy = "lazy"
		NumberOfExecutionStreams = 1
		validateFlags()
		return
	}
	cmd := exec.Command(os.Args[0], "-check.f=MySuite.TestValidateFlagsWithStartegyLazy")
	cmd.Env = append(os.Environ(), "EXIT_VALIDATE=1")
	err := cmd.Run()
	c.Assert(err, Equals, nil)
}

func (s *MySuite) TestValidateFlagsWithInvalidStrategy(c *C) {
	if os.Getenv("EXIT_VALIDATE") == "1" {
		InParallel = true
		Strategy = "sdf"
		NumberOfExecutionStreams = 1
		validateFlags()
		return
	}
	cmd := exec.Command(os.Args[0], "-check.f=MySuite.TestValidateFlagsWithInvalidStrategy")
	cmd.Env = append(os.Environ(), "EXIT_VALIDATE=1")
	logger := &testLogger{}
	cmd.Stdout = logger
	err := cmd.Run()
	e, ok := err.(*exec.ExitError)
	c.Assert(ok, Equals, true)
	c.Assert(e.Success(), Equals, false)
	c.Assert(strings.TrimSpace(logger.output), Equals, "Invalid input(sdf) to --strategy flag.")
}

func (s *MySuite) TestValidateFlagsWithInvalidStream(c *C) {
	if os.Getenv("EXIT_VALIDATE") == "1" {
		InParallel = true
		NumberOfExecutionStreams = -1
		validateFlags()
		return
	}
	cmd := exec.Command(os.Args[0], "-check.f=MySuite.TestValidateFlagsWithInvalidStream")
	cmd.Env = append(os.Environ(), "EXIT_VALIDATE=1")
	logger := &testLogger{}
	cmd.Stdout = logger
	err := cmd.Run()
	e, ok := err.(*exec.ExitError)
	c.Assert(ok, Equals, true)
	c.Assert(e.Success(), Equals, false)
	c.Assert(strings.TrimSpace(logger.output), Equals, "Invalid input(-1) to --n flag.")
}

// Result Builders
type scenarioBuilder struct {
	scn *gauge_messages.ProtoScenario
}

func newScenarioBuilder() *scenarioBuilder {
	return &scenarioBuilder{
		scn: &gauge_messages.ProtoScenario{},
	}
}

func (sb *scenarioBuilder) heading(h string) *scenarioBuilder {
	sb.scn.ScenarioHeading = proto.String(h)
	return sb
}

func (sb *scenarioBuilder) failed(f bool) *scenarioBuilder {
	sb.scn.Failed = proto.Bool(f)
	return sb
}

func (sb *scenarioBuilder) build() *gauge_messages.ProtoScenario {
	return sb.scn
}

// Suite Result Builder
type protoSpecBuilder struct {
	spec *gauge_messages.ProtoSpec
}

func newSpecBuilder() *protoSpecBuilder {
	return &protoSpecBuilder{spec: &gauge_messages.ProtoSpec{}}
}

func (sb *protoSpecBuilder) tableDriven(f bool) *protoSpecBuilder {
	sb.spec.IsTableDriven = proto.Bool(f)
	return sb
}

func (sb *protoSpecBuilder) scenario(scn *gauge_messages.ProtoScenario) *protoSpecBuilder {
	sb.spec.Items = append(sb.spec.Items, &gauge_messages.ProtoItem{
		ItemType: gauge_messages.ProtoItem_Scenario.Enum(),
		Scenario: scn,
	})
	return sb
}

func (sb *protoSpecBuilder) tableDrivenScenario(scns []*gauge_messages.ProtoScenario) *protoSpecBuilder {
	sb.spec.Items = append(sb.spec.Items, &gauge_messages.ProtoItem{
		ItemType:            gauge_messages.ProtoItem_TableDrivenScenario.Enum(),
		TableDrivenScenario: &gauge_messages.ProtoTableDrivenScenario{Scenarios: scns},
	})
	return sb
}

func (sb *protoSpecBuilder) build() *gauge_messages.ProtoSpec {
	return sb.spec
}

type specResultBuilder struct {
	sr *result.SpecResult
}

func newSpecResultBuilder() *specResultBuilder {
	return &specResultBuilder{sr: &result.SpecResult{}}
}

func (s *specResultBuilder) spec(spec *gauge_messages.ProtoSpec) *specResultBuilder {
	s.sr.ProtoSpec = spec
	return s
}

func (s *specResultBuilder) failed(f bool) *specResultBuilder {
	s.sr.IsFailed = f
	return s
}

func (s *specResultBuilder) build() *result.SpecResult {
	return s.sr
}

type suiteResultBuilder struct {
	sr *result.SuiteResult
}

func newSuiteResultBuilder() *suiteResultBuilder {
	return &suiteResultBuilder{sr: &result.SuiteResult{}}
}

func (s *suiteResultBuilder) spec(res []*result.SpecResult) *suiteResultBuilder {
	s.sr.SpecResults = res
	return s
}

func (s *suiteResultBuilder) failed(f bool) *suiteResultBuilder {
	s.sr.IsFailed = f
	return s
}

func (s *suiteResultBuilder) build() *result.SuiteResult {
	return s.sr
}
