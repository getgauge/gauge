/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package formatter

import (
	"bytes"

	"strings"

	"github.com/getgauge/gauge/config"
	"github.com/getgauge/gauge/gauge"
)

type formatter struct {
	buffer    bytes.Buffer
	itemQueue *gauge.ItemQueue
}

func (formatter *formatter) Specification(specification *gauge.Specification) {
}

func (formatter *formatter) Heading(heading *gauge.Heading) {
	switch heading.HeadingType {
	case gauge.SpecHeading:
		formatter.buffer.WriteString(FormatHeading(heading.Value, "#"))
	case gauge.ScenarioHeading:
		formatter.buffer.WriteString(FormatHeading(heading.Value, "##"))
	}
}

func (formatter *formatter) Tags(tags *gauge.Tags) {
	if !strings.HasSuffix(formatter.buffer.String(), "\n\n") && !config.CurrentGaugeSettings().Format.SkipEmptyLineInsertions {
		formatter.buffer.WriteString("\n")
	}
	formatter.buffer.WriteString(FormatTags(tags))
	if formatter.itemQueue.Peek() != nil && (formatter.itemQueue.Peek().Kind() != gauge.CommentKind || strings.TrimSpace(formatter.itemQueue.Peek().(*gauge.Comment).Value) != "") && !config.CurrentGaugeSettings().Format.SkipEmptyLineInsertions {
		formatter.buffer.WriteString("\n")
	}
}

func (formatter *formatter) Table(table *gauge.Table) {
	formatter.buffer.WriteString(strings.TrimPrefix(FormatTable(table), "\n"))
}

func (formatter *formatter) DataTable(dataTable *gauge.DataTable) {
	if !dataTable.IsExternal {
		formatter.Table(dataTable.Table)
	} else {
		formatter.buffer.WriteString(formatExternalDataTable(dataTable))
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
