package main

import "fmt"

type specBuilder struct {
	lines []string
}

func SpecBuilder() *specBuilder {
	return &specBuilder{lines: make([]string, 0)}
}

func (specBuilder *specBuilder) addPrefix(prefix string, line string) string {
	return fmt.Sprintf("%s%s\n", prefix, line)
}

func (specBuilder *specBuilder) String() string {
	var result string
	for _, line := range specBuilder.lines {
		result = fmt.Sprintf("%s%s", result, line)
	}
	return result
}

func (specBuilder *specBuilder) specHeading(heading string) *specBuilder {
	line := specBuilder.addPrefix("#", heading)
	specBuilder.lines = append(specBuilder.lines, line)
	return specBuilder
}

func (specBuilder *specBuilder) scenarioHeading(heading string) *specBuilder {
	line := specBuilder.addPrefix("##", heading)
	specBuilder.lines = append(specBuilder.lines, line)
	return specBuilder
}

func (specBuilder *specBuilder) step(stepText string) *specBuilder {
	line := specBuilder.addPrefix("* ", stepText)
	specBuilder.lines = append(specBuilder.lines, line)
	return specBuilder
}

func (specBuilder *specBuilder) tags(tags ...string) *specBuilder {
	tagText := ""
	for i, tag := range tags {
		tagText = fmt.Sprintf("%s%s", tagText, tag)
		if i != len(tags)-1 {
			tagText = fmt.Sprintf("%s,", tagText)
		}
	}
	line := specBuilder.addPrefix("tags: ", tagText)
	specBuilder.lines = append(specBuilder.lines, line)
	return specBuilder
}

func (specBuilder *specBuilder) text(comment string) *specBuilder {
	specBuilder.lines = append(specBuilder.lines, fmt.Sprintf("%s\n", comment))
	return specBuilder
}

func isInState(currentState int, statesToCheck ...int) bool {
	var mask int
	for _, value := range statesToCheck {
		mask |= value
	}
	return (mask & currentState) != 0
}

func isInAnyState(currentState int, statesToCheck ...int) bool {
	for _, value := range statesToCheck {
		if (currentState & value) != 0 {
			return true
		}
	}
	return false
}

func retainStates(currentState *int, statesToKeep ...int) {
	var mask int
	for _, value := range statesToKeep {
		mask |= value
	}
	*currentState = mask & *currentState
}

func addStates(currentState *int, states ...int) {
	var mask int
	for _, value := range states {
		mask |= value
	}
	*currentState = mask | *currentState
}

func isUnderline(text string, underlineChar rune) bool {
	if len(text) == 0 || rune(text[0]) != underlineChar {
		return false
	}
	for _, value := range text {
		if rune(value) != underlineChar {
			return false
		}
	}
	return true
}

func areUnderlined(values []string) bool {
	if len(values) == 0 {
		return false
	}
	for _, value := range values {
		if len(value) == 0 {
			continue
		}
		if !isUnderline(value, rune('-')) {
			return false
		}
	}
	return true
}

func arrayContains(array []string, toFind string) bool {
	for _, value := range array {
		if value == toFind {
			return true
		}
	}
	return false
}
