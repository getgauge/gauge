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
