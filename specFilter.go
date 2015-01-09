package main

import (
	"strings"
	"strconv"
	"golang.org/x/tools/go/types"
	"golang.org/x/tools/go/exact"
	"log"
	"regexp"
)

type scenarioIndexFilterToRetain struct {
	indexToNotFilter     int
	currentScenarioIndex int
}
type ScenarioFilterBasedOnTags struct {
	specTags        []string
	tagExpression    string
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
		if filter.filter(filter.specTags) {
			return true
		}
		return filter.filterTags(item)
	}
	return false
}

func (filter *ScenarioFilterBasedOnTags) filterTags(item item) bool{
	tagsMap := make(map[string]bool,0)
	for _,tag := range item.(*scenario).tags.values{
		tagsMap[tag] = true
	}
	_, tags := getOperatorsAndOperands(filter.tagExpression)
	for _,tag := range tags{
		strings.Replace(filter.tagExpression,tag,strconv.FormatBool(isTagPresent(tagsMap,tag)),-1)
	}
	return evaluateExp(filter.tagExpression)
	
}

func evaluateExp(tagExpression string) bool {
	tre := regexp.MustCompile("true")
	fre := regexp.MustCompile("false")

	s := fre.ReplaceAllString(tre.ReplaceAllString(tagExpression, "1"), "0")

	_, val, err := types.Eval(s, nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	res, _ := exact.Uint64Val(val)

	var final bool
	if res == 1 {
		final = true
	} else {
		final = false
	}

	return final
}

func isTagPresent(tagsMap map[string]bool, tagName string) bool {
	_,ok := tagsMap[tagName]
	return ok
}

func getOperatorsAndOperands(tagExpression string) ([]string, []string){
	listOfOperators := make([]string,0)
	listOfTags := strings.FieldsFunc(tagExpression, func(r rune) bool {
			isValidOperator := r == '&' || r == '|'
			if isValidOperator {
				operator, _ := strconv.Unquote(strconv.QuoteRuneToASCII(r))
				listOfOperators = append(listOfOperators, operator)
				return isValidOperator
			}
			return false
		})
	return listOfOperators, listOfTags
}
