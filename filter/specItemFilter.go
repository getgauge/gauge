/*----------------------------------------------------------------
 *  Copyright (c) ThoughtWorks, Inc.
 *  Licensed under the Apache License, Version 2.0
 *  See LICENSE in the project root for license information.
 *----------------------------------------------------------------*/

package filter

import (
	"errors"
	"go/constant"
	"go/token"
	"go/types"
	"regexp"
	"strconv"
	"strings"

	"github.com/getgauge/gauge/env"

	"fmt"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
)

type scenarioFilterBasedOnSpan struct {
	lineNumbers []int
}
type ScenarioFilterBasedOnTags struct {
	specTags      []string
	tagExpression string
}

type scenarioFilterBasedOnName struct {
	scenariosName []string
}

func NewScenarioFilterBasedOnSpan(lineNumbers []int) *scenarioFilterBasedOnSpan {
	return &scenarioFilterBasedOnSpan{lineNumbers}
}

func (filter *scenarioFilterBasedOnSpan) Filter(item gauge.Item) bool {
	for _, lineNumber := range filter.lineNumbers {
		if item.(*gauge.Scenario).InSpan(lineNumber) {
			return false
		}
	}
	return true
}

func NewScenarioFilterBasedOnTags(specTags []string, tagExp string) *ScenarioFilterBasedOnTags {
	return &ScenarioFilterBasedOnTags{specTags, tagExp}
}

func (filter *ScenarioFilterBasedOnTags) Filter(item gauge.Item) bool {
	tags := item.(*gauge.Scenario).Tags
	if tags == nil {
		return !filter.filterTags(filter.specTags)
	}
	return !filter.filterTags(append(tags.Values(), filter.specTags...))
}

func newScenarioFilterBasedOnName(scenariosName []string) *scenarioFilterBasedOnName {
	return &scenarioFilterBasedOnName{scenariosName}
}

func (filter *scenarioFilterBasedOnName) Filter(item gauge.Item) bool {
	return !item.(*gauge.Scenario).HasAnyHeading(filter.scenariosName)
}

func sanitize(tag string) string {
	if _, err := strconv.ParseBool(tag); err == nil {
		return fmt.Sprintf("{%s}", tag)
	}
	if env.AllowCaseSensitiveTags() {
		return tag
	}
	return strings.ToLower(tag)
}

func (filter *ScenarioFilterBasedOnTags) filterTags(stags []string) bool {
	tagsMap := make(map[string]bool)
	for _, tag := range stags {
		tag = sanitize(tag)
		tagsMap[strings.Replace(tag, " ", "", -1)] = true
	}
	filter.replaceSpecialChar()
	value, _ := filter.formatAndEvaluateExpression(tagsMap, filter.isTagPresent)
	return value
}

func (filter *ScenarioFilterBasedOnTags) replaceSpecialChar() {
	filter.tagExpression = strings.Replace(strings.Replace(strings.Replace(strings.Replace(filter.tagExpression, " ", "", -1), ",", "&", -1), "&&", "&", -1), "||", "|", -1)
}

func (filter *ScenarioFilterBasedOnTags) formatAndEvaluateExpression(tagsMap map[string]bool, isTagQualified func(tagsMap map[string]bool, tagName string) bool) (bool, error) {
	tagExpressionParts, tags := filter.parseTagExpression()
	for _, tag := range tags {
		for i, txp := range tagExpressionParts {
			if strings.TrimSpace(txp) == strings.TrimSpace(tag) {
				tagExpressionParts[i] = strconv.FormatBool(isTagQualified(tagsMap, strings.TrimSpace(tag)))
			}
		}
	}
	return filter.evaluateExp(filter.handleNegation(strings.Join(tagExpressionParts, "")))
}

func (filter *ScenarioFilterBasedOnTags) handleNegation(tagExpression string) string {
	tagExpression = strings.Replace(strings.Replace(tagExpression, "!true", "false", -1), "!false", "true", -1)
	for strings.Contains(tagExpression, "!(") {
		tagExpression = filter.evaluateBrackets(tagExpression)
	}
	return tagExpression
}

func (filter *ScenarioFilterBasedOnTags) evaluateBrackets(tagExpression string) string {
	if strings.Contains(tagExpression, "!(") {
		innerText := filter.resolveBracketExpression(tagExpression)
		return strings.Replace(tagExpression, "!("+innerText+")", filter.evaluateBrackets(innerText), -1)
	}
	value, _ := filter.evaluateExp(tagExpression)
	return strconv.FormatBool(!value)
}

func (filter *ScenarioFilterBasedOnTags) resolveBracketExpression(tagExpression string) string {
	indexOfOpenBracket := strings.Index(tagExpression, "!(") + 1
	bracketStack := make([]string, 0)
	i := indexOfOpenBracket
	for ; i < len(tagExpression); i++ {
		if tagExpression[i] == '(' {
			bracketStack = append(bracketStack, "(")
		} else if tagExpression[i] == ')' {
			bracketStack = bracketStack[:len(bracketStack)-1]
		}
		if len(bracketStack) == 0 {
			break
		}
	}
	return tagExpression[indexOfOpenBracket+1 : i]
}

func (filter *ScenarioFilterBasedOnTags) evaluateExp(tagExpression string) (bool, error) {
	tre := regexp.MustCompile("true")
	fre := regexp.MustCompile("false")

	s := fre.ReplaceAllString(tre.ReplaceAllString(tagExpression, "1"), "0")

	val, err := types.Eval(token.NewFileSet(), nil, 0, s)
	if err != nil {
		return false, errors.New("Invalid Expression.\n" + err.Error())
	}
	res, _ := constant.Uint64Val(val.Value)

	var final bool
	if res == 1 {
		final = true
	} else {
		final = false
	}

	return final, nil
}

func (filter *ScenarioFilterBasedOnTags) isTagPresent(tagsMap map[string]bool, tagName string) bool {
	_, ok := tagsMap[tagName]
	return ok
}

func (filter *ScenarioFilterBasedOnTags) parseTagExpression() (tagExpressionParts []string, tags []string) {
	isValidOperator := func(r rune) bool { return r == '&' || r == '|' || r == '(' || r == ')' || r == '!' }
	var word string
	var wordValue = func() string {
		return sanitize(strings.TrimSpace(word))
	}
	for _, c := range filter.tagExpression {
		c1, _ := strconv.Unquote(strconv.QuoteRuneToASCII(c))
		if isValidOperator(c) {
			if word != "" {
				tagExpressionParts = append(tagExpressionParts, wordValue())
				tags = append(tags, wordValue())
			}
			tagExpressionParts = append(tagExpressionParts, c1)
			word = ""
		} else {
			word += c1
		}
	}
	if word != "" {
		tagExpressionParts = append(tagExpressionParts, wordValue())
		tags = append(tags, wordValue())
	}
	return
}

func filterSpecsByTags(specs []*gauge.Specification, tagExpression string) ([]*gauge.Specification, []*gauge.Specification) {
	filteredSpecs := make([]*gauge.Specification, 0)
	otherSpecs := make([]*gauge.Specification, 0)
	for _, spec := range specs {
		tagValues := make([]string, 0)
		if spec.Tags != nil {
			tagValues = spec.Tags.Values()
		}
		specWithFilteredItems, specWithOtherItems := spec.Filter(NewScenarioFilterBasedOnTags(tagValues, tagExpression))
		if len(specWithFilteredItems.Scenarios) != 0 {
			filteredSpecs = append(filteredSpecs, specWithFilteredItems)
		}
		if len(specWithOtherItems.Scenarios) != 0 {
			otherSpecs = append(otherSpecs, specWithOtherItems)
		}
	}
	return filteredSpecs, otherSpecs
}

func validateTagExpression(tagExpression string) {
	filter := &ScenarioFilterBasedOnTags{tagExpression: tagExpression}
	filter.replaceSpecialChar()
	_, err := filter.formatAndEvaluateExpression(make(map[string]bool), func(a map[string]bool, b string) bool { return true })
	if err != nil {
		logger.Fatalf(true, err.Error())
	}
}

func filterSpecsByScenarioName(specs []*gauge.Specification, scenariosName []string) []*gauge.Specification {
	filteredSpecs := make([]*gauge.Specification, 0)
	scenarios := filterValidScenarios(specs, scenariosName)
	for _, spec := range specs {
		s, _ := spec.Filter(newScenarioFilterBasedOnName(scenarios))
		if len(s.Scenarios) != 0 {
			filteredSpecs = append(filteredSpecs, s)
		}
	}
	return filteredSpecs
}

func filterValidScenarios(specs []*gauge.Specification, headings []string) []string {
	filteredScenarios := make([]string, 0)
	allScenarios := GetAllScenarios(specs)
	var exists = func(scenarios []string, heading string) bool {
		for _, scenario := range scenarios {
			if strings.Compare(scenario, heading) == 0 {
				return true
			}
		}
		return false
	}
	for _, heading := range headings {
		if exists(allScenarios, heading) {
			filteredScenarios = append(filteredScenarios, heading)
		} else {
			logger.Warningf(true, "Warning: scenario name - \"%s\" not found", heading)
		}
	}
	return filteredScenarios
}

func GetAllScenarios(specs []*gauge.Specification) []string {
	allScenarios := []string{}
	for _, spec := range specs {
		for _, scenario := range spec.Scenarios {
			allScenarios = append(allScenarios, scenario.Heading.Value)
		}
	}
	return allScenarios
}
