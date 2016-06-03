package filter

import (
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/parser"
	"github.com/getgauge/gauge/util"
)

var ExecuteTags string
var DoNotRandomize bool
var Distribute int
var NumberOfExecutionStreams int

func GetSpecsToExecute(conceptsDictionary *gauge.ConceptDictionary, specDirs []string) ([]*gauge.Specification, int, bool) {
	specsToExecute, failed := specsFromArgs(conceptsDictionary, specDirs)
	totalSpecs := specsToExecute
	specsToExecute = applyFilters(specsToExecute, specsFilters())
	return sortSpecsList(specsToExecute), len(totalSpecs) - len(specsToExecute), failed
}

func specsFilters() []specsFilter {
	return []specsFilter{&tagsFilter{ExecuteTags}, &specsGroupFilter{Distribute, NumberOfExecutionStreams}, &specRandomizer{DoNotRandomize}}
}

func applyFilters(specsToExecute []*gauge.Specification, filters []specsFilter) []*gauge.Specification {
	for _, specsFilter := range filters {
		specsToExecute = specsFilter.filter(specsToExecute)
	}
	return specsToExecute
}

func addSpecsToMap(specs []*gauge.Specification, specsMap map[string]*gauge.Specification) {
	for _, spec := range specs {
		if _, ok := specsMap[spec.FileName]; ok {
			specsMap[spec.FileName].Scenarios = append(specsMap[spec.FileName].Scenarios, spec.Scenarios...)
			continue
		}
		specsMap[spec.FileName] = spec
	}
}

func specsFromArgs(conceptDictionary *gauge.ConceptDictionary, specDirs []string) ([]*gauge.Specification, bool) {
	specsMap := make(map[string]*gauge.Specification)
	var specs []*gauge.Specification
	var specParseResults []*parser.ParseResult
	passed := true
	for _, arg := range specDirs {
		specSource := arg
		if isIndexedSpec(specSource) {
			specs, specParseResults = getSpecWithScenarioIndex(specSource, conceptDictionary)
		} else {
			specs, specParseResults = parser.ParseSpecFiles(util.GetSpecFiles(specSource), conceptDictionary)
		}
		passed = !parser.HandleParseResult(specParseResults...) && passed
		addSpecsToMap(specs, specsMap)
	}
	var allSpecs []*gauge.Specification
	for _, spec := range specsMap {
		allSpecs = append(allSpecs, spec)
	}
	return allSpecs, !passed
}

func getSpecWithScenarioIndex(specSource string, conceptDictionary *gauge.ConceptDictionary) ([]*gauge.Specification, []*parser.ParseResult) {
	specName, indexToFilter := GetIndexedSpecName(specSource)
	parsedSpecs, parseResult := parser.ParseSpecFiles(util.GetSpecFiles(specName), conceptDictionary)
	return filterSpecsItems(parsedSpecs, newScenarioFilterBasedOnSpan(indexToFilter)), parseResult
}
