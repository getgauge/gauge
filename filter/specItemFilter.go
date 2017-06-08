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

package filter

import (
	"errors"
	"go/constant"
	"go/token"
	"go/types"
	"regexp"
	"strconv"
	"strings"

	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
)

type scenarioFilterBasedOnSpan struct {
	lineNumber int
}
type ScenarioFilterBasedOnTags struct {
	specTags      []string
	tagExpression string
}

func NewScenarioFilterBasedOnSpan(lineNumber int) *scenarioFilterBasedOnSpan {
	return &scenarioFilterBasedOnSpan{lineNumber}
}

func (filter *scenarioFilterBasedOnSpan) Filter(item gauge.Item) bool {
	return (item.Kind() == gauge.ScenarioKind) && !(item.(*gauge.Scenario).InSpan(filter.lineNumber))
}

func newScenarioFilterBasedOnTags(specTags []string, tagExp string) *ScenarioFilterBasedOnTags {
	return &ScenarioFilterBasedOnTags{specTags, tagExp}
}

func (filter *ScenarioFilterBasedOnTags) Filter(item gauge.Item) bool {
	if item.Kind() == gauge.ScenarioKind {
		tags := item.(*gauge.Scenario).Tags
		if tags == nil {
			return !filter.filterTags(filter.specTags)
		}
		return !filter.filterTags(append(tags.Values, filter.specTags...))
	}
	return false
}

func (filter *ScenarioFilterBasedOnTags) filterTags(stags []string) bool {
	tagsMap := make(map[string]bool, 0)
	for _, tag := range stags {
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
			bracketStack = append(bracketStack[:len(bracketStack)-1])
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
	for _, c := range filter.tagExpression {
		c1, _ := strconv.Unquote(strconv.QuoteRuneToASCII(c))
		if isValidOperator(c) {
			if word != "" {
				tagExpressionParts = append(tagExpressionParts, strings.TrimSpace(word))
				tags = append(tags, strings.TrimSpace(word))
			}
			tagExpressionParts = append(tagExpressionParts, c1)
			word = ""
		} else {
			word += c1
		}
	}
	if word != "" {
		tagExpressionParts = append(tagExpressionParts, strings.TrimSpace(word))
		tags = append(tags, strings.TrimSpace(word))
	}
	return
}

func FilterSpecsItems(specs []*gauge.Specification, filter gauge.SpecItemFilter) []*gauge.Specification {
	filteredSpecs := make([]*gauge.Specification, 0)
	for _, spec := range specs {
		spec.Filter(filter)
		if len(spec.Scenarios) != 0 {
			filteredSpecs = append(filteredSpecs, spec)
		}
	}
	return filteredSpecs
}

func filterSpecsByTags(specs []*gauge.Specification, tagExpression string) []*gauge.Specification {
	filteredSpecs := make([]*gauge.Specification, 0)
	for _, spec := range specs {
		tagValues := make([]string, 0)
		if spec.Tags != nil {
			tagValues = spec.Tags.Values
		}
		spec.Filter(newScenarioFilterBasedOnTags(tagValues, tagExpression))
		if len(spec.Scenarios) != 0 {
			filteredSpecs = append(filteredSpecs, spec)
		}
	}
	return filteredSpecs
}

func validateTagExpression(tagExpression string) {
	filter := &ScenarioFilterBasedOnTags{tagExpression: tagExpression}
	filter.replaceSpecialChar()
	_, err := filter.formatAndEvaluateExpression(make(map[string]bool, 0), func(a map[string]bool, b string) bool { return true })
	if err != nil {
		logger.Fatalf(err.Error())
	}
}
