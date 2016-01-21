package filter

import (
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/parser"
)

var ExecuteTags string
var DoNotRandomize bool
var Distribute int
var NumberOfExecutionStreams int

func GetSpecsToExecute(conceptsDictionary *gauge.ConceptDictionary, args []string) ([]*gauge.Specification, int) {
	specsToExecute := specsFromArgs(conceptsDictionary, args)
	totalSpecs := specsToExecute
	specsToExecute = applyFilters(specsToExecute, specsFilters())
	return sortSpecsList(specsToExecute), len(totalSpecs) - len(specsToExecute)
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

func specsFromArgs(conceptDictionary *gauge.ConceptDictionary, args []string) []*gauge.Specification {
	allSpecs := make([]*gauge.Specification, 0)
	specs := make([]*gauge.Specification, 0)
	var specParseResults []*parser.ParseResult
	for _, arg := range args {
		specSource := arg
		if isIndexedSpec(specSource) {
			specs, specParseResults = getSpecWithScenarioIndex(specSource, conceptDictionary)
		} else {
			specs, specParseResults = parser.FindSpecs(specSource, conceptDictionary)
		}
		parser.HandleParseResult(specParseResults...)
		allSpecs = append(allSpecs, specs...)
	}
	return allSpecs
}

func getSpecWithScenarioIndex(specSource string, conceptDictionary *gauge.ConceptDictionary) ([]*gauge.Specification, []*parser.ParseResult) {
	specName, indexToFilter := GetIndexedSpecName(specSource)
	parsedSpecs, parseResult := parser.FindSpecs(specName, conceptDictionary)
	return filterSpecsItems(parsedSpecs, newScenarioIndexFilterToRetain(indexToFilter)), parseResult
}
