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

package main

import (
	"bytes"
	"github.com/getgauge/gauge/parser"
)

type formatter struct {
	buffer bytes.Buffer
}

func (formatter *formatter) SpecHeading(specHeading *parser.Heading) {
	formatter.buffer.WriteString(formatHeading(specHeading.Value(), "="))
}

func (formatter *formatter) SpecTags(tags *parser.Tags) {
	formatter.buffer.WriteString(formatTags(tags))
}

func (formatter *formatter) DataTable(table *parser.Table) {
	formatter.buffer.WriteString(formatTable(table))
}

func (formatter *formatter) ExternalDataTable(extDataTable *parser.DataTable) {
	formatter.buffer.WriteString(formatExternalDataTable(extDataTable))
}

func (formatter *formatter) ContextStep(step *parser.Step) {
	formatter.Step(step)
}

func (formatter *formatter) Scenario(scenario *parser.Scenario) {
}

func (formatter *formatter) ScenarioHeading(scenarioHeading *parser.Heading) {
	formatter.buffer.WriteString(formatHeading(scenarioHeading.Value(), "-"))
}

func (formatter *formatter) ScenarioTags(scenarioTags *parser.Tags) {
	formatter.SpecTags(scenarioTags)
}

func (formatter *formatter) Step(step *parser.Step) {
	formatter.buffer.WriteString(formatStep(step))
}

func (formatter *formatter) Comment(comment *parser.Comment) {
	formatter.buffer.WriteString(formatComment(comment))
}
