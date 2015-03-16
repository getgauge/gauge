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

import "bytes"

type formatter struct {
	buffer bytes.Buffer
}

func (formatter *formatter) specHeading(specHeading *heading) {
	formatter.buffer.WriteString(formatHeading(specHeading.value, "="))
}

func (formatter *formatter) specTags(tags *tags) {
	formatter.buffer.WriteString(formatTags(tags))
}

func (formatter *formatter) dataTable(table *table) {
	formatter.buffer.WriteString(formatTable(table))
}

func (formatter *formatter) externalDataTable(extDataTable *dataTable) {
	formatter.buffer.WriteString(formatExternalDataTable(extDataTable))
}

func (formatter *formatter) contextStep(step *step) {
	formatter.step(step)
}

func (formatter *formatter) scenario(scenario *scenario) {
}

func (formatter *formatter) scenarioHeading(scenarioHeading *heading) {
	formatter.buffer.WriteString(formatHeading(scenarioHeading.value, "-"))
}

func (formatter *formatter) scenarioTags(scenarioTags *tags) {
	formatter.specTags(scenarioTags)
}

func (formatter *formatter) step(step *step) {
	formatter.buffer.WriteString(formatStep(step))
}

func (formatter *formatter) comment(comment *comment) {
	formatter.buffer.WriteString(formatComment(comment))
}
