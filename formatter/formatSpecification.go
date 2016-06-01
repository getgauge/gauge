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

package formatter

import (
	"bytes"

	"strings"

	"github.com/getgauge/gauge/gauge"
)

type formatter struct {
	buffer bytes.Buffer
}

func (formatter *formatter) SpecHeading(specHeading *gauge.Heading) {
	formatter.buffer.WriteString(FormatHeading(specHeading.Value, "="))
}

func (formatter *formatter) SpecTags(tags *gauge.Tags) {
	formatter.buffer.WriteString(FormatTags(tags))
}

func (formatter *formatter) DataTable(table *gauge.Table) {
	formatter.buffer.WriteString(strings.TrimPrefix(FormatTable(table), "\n"))
}

func (formatter *formatter) ExternalDataTable(extDataTable *gauge.DataTable) {
	formatter.buffer.WriteString(FormatExternalDataTable(extDataTable))
}

func (formatter *formatter) ContextStep(step *gauge.Step) {
	formatter.Step(step)
}

func (formatter *formatter) TearDown(t *gauge.TearDown) {
	formatter.buffer.WriteString(t.Value + "\n")
}

func (formatter *formatter) Scenario(scenario *gauge.Scenario) {
}

func (formatter *formatter) ScenarioHeading(scenarioHeading *gauge.Heading) {
	formatter.buffer.WriteString(FormatHeading(scenarioHeading.Value, "-"))
}

func (formatter *formatter) ScenarioTags(scenarioTags *gauge.Tags) {
	formatter.SpecTags(scenarioTags)
}

func (formatter *formatter) Step(step *gauge.Step) {
	formatter.buffer.WriteString(FormatStep(step))
}

func (formatter *formatter) Comment(comment *gauge.Comment) {
	formatter.buffer.WriteString(FormatComment(comment))
}
