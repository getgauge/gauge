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

func (formatter *formatter) Specification(specification *gauge.Specification) {
}

func (formatter *formatter) Heading(heading *gauge.Heading) {
	if heading.HeadingType == gauge.SpecHeading {
		formatter.buffer.WriteString(FormatHeading(heading.Value, "="))
	} else if heading.HeadingType == gauge.ScenarioHeading {
		formatter.buffer.WriteString(FormatHeading(heading.Value, "-"))
	}
}

func (formatter *formatter) Tags(tags *gauge.Tags) {
	formatter.buffer.WriteString(FormatTags(tags))
}

func (formatter *formatter) Table(table *gauge.Table) {
	formatter.buffer.WriteString(strings.TrimPrefix(FormatTable(table), "\n"))
}

func (formatter *formatter) DataTable(dataTable *gauge.DataTable) {
	if !dataTable.IsExternal {
		formatter.Table(&(dataTable.Table))
	} else {
		formatter.buffer.WriteString(FormatExternalDataTable(dataTable))
	}
}

func (formatter *formatter) TearDown(t *gauge.TearDown) {
	formatter.buffer.WriteString(t.Value + "\n")
}

func (formatter *formatter) Scenario(scenario *gauge.Scenario) {
}

func (formatter *formatter) Step(step *gauge.Step) {
	formatter.buffer.WriteString(FormatStep(step))
}

func (formatter *formatter) Comment(comment *gauge.Comment) {
	formatter.buffer.WriteString(FormatComment(comment))
}
