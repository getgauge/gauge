package main

import (
	"errors"
	"golang.org/x/tools/go/exact"
	"golang.org/x/tools/go/types"
	"regexp"
	"strconv"
	"strings"
)

type scenarioIndexFilterToRetain struct {
	indexToNotFilter     int
	currentScenarioIndex int
}
type ScenarioFilterBasedOnTags struct {
	specTags      []string
	tagExpression string
}

func newScenarioIndexFilterToRetain(index int) *scenarioIndexFilterToRetain {
	return &scenarioIndexFilterToRetain{index, 0}
}

func (filter *scenarioIndexFilterToRetain) filter(item item) bool {
	if item.kind() == scenarioKind {
		if filter.currentScenarioIndex != filter.indexToNotFilter {
			filter.currentScenarioIndex++
			return true
		} else {
			filter.currentScenarioIndex++
			return false
		}
	}
	return false
}

func newScenarioFilterBasedOnTags(specTags []string, tagExp string) *ScenarioFilterBasedOnTags {
	return &ScenarioFilterBasedOnTags{specTags, tagExp}
}

func (filter *ScenarioFilterBasedOnTags) filter(item item) bool {
	if item.kind() == scenarioKind {
		if filter.filterTags(filter.specTags) {
			return false
		}
		tags := item.(*scenario).tags
		if tags == nil {
			return true
		}
		return !filter.filterTags(tags.values)
	}
	return true
}

func (filter *ScenarioFilterBasedOnTags) filterTags(stags []string) bool {
	tagsMap := make(map[string]bool, 0)
	for _, tag := range stags {
		tagsMap[strings.Replace(tag, " ", "", -1)] = true
	}
	value, _ := formatAndEvaluateExpression(strings.Replace(filter.tagExpression, " ", "", -1), tagsMap, isTagPresent)
	return value
}

func formatAndEvaluateExpression(tagExpression string, tagsMap map[string]bool, isTagQualified func(tagsMap map[string]bool, tagName string) bool) (bool, error) {
	_, tags := getOperatorsAndOperands(tagExpression)
	expToBeEvaluated := tagExpression
	for _, tag := range tags {
		expToBeEvaluated = strings.Replace(expToBeEvaluated, strings.TrimSpace(tag), strconv.FormatBool(isTagQualified(tagsMap, strings.TrimSpace(tag))), -1)
	}
	return evaluateExp(expToBeEvaluated)
}

func evaluateExp(tagExpression string) (bool, error) {
	tre := regexp.MustCompile("true")
	fre := regexp.MustCompile("false")

	s := fre.ReplaceAllString(tre.ReplaceAllString(tagExpression, "1"), "0")

	_, val, err := types.Eval(s, nil, nil)
	if err != nil {
		return false, errors.New("Invalid Expression.\n" + err.Error())
	}
	res, _ := exact.Uint64Val(val)

	var final bool
	if res == 1 {
		final = true
	} else {
		final = false
	}

	return final, nil
}

func isTagPresent(tagsMap map[string]bool, tagName string) bool {
	_, ok := tagsMap[tagName]
	return ok
}

func getOperatorsAndOperands(tagExpression string) ([]string, []string) {
	listOfOperators := make([]string, 0)
	listOfTags := strings.FieldsFunc(tagExpression, func(r rune) bool {
		isValidOperator := r == '&' || r == '|' || r == '(' || r == ')'
		if isValidOperator {
			operator, _ := strconv.Unquote(strconv.QuoteRuneToASCII(r))
			listOfOperators = append(listOfOperators, operator)
			return isValidOperator
		}
		return false
	})
	return listOfOperators, listOfTags
}
