package filter

import (
	"github.com/getgauge/gauge/parser"
)

var ExecuteTags string
var DoNotRandomize bool
var Distribute int
var NumberOfExecutionStreams int

func GetSpecsToExecute(conceptsDictionary *parser.ConceptDictionary, args []string) ([]*parser.Specification, int) {
	specsToExecute := specsFromArgs(conceptsDictionary, args)
	totalSpecs := specsToExecute
	specsToExecute = applyFilters(specsToExecute, specsFilters())
	return sortSpecsList(specsToExecute), len(totalSpecs) - len(specsToExecute)
}

func specsFilters() []specsFilter {
	return []specsFilter{&tagsFilter{ExecuteTags}, &specsGroupFilter{Distribute, NumberOfExecutionStreams}, &specRandomizer{DoNotRandomize}}
}

func applyFilters(specsToExecute []*parser.Specification, filters []specsFilter) []*parser.Specification {
	for _, specsFilter := range filters {
		specsToExecute = specsFilter.filter(specsToExecute)
	}
	return specsToExecute
}

func specsFromArgs(conceptDictionary *parser.ConceptDictionary, args []string) []*parser.Specification {
	allSpecs := make([]*parser.Specification, 0)
	specs := make([]*parser.Specification, 0)
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

func getSpecWithScenarioIndex(specSource string, conceptDictionary *parser.ConceptDictionary) ([]*parser.Specification, []*parser.ParseResult) {
	specName, indexToFilter := GetIndexedSpecName(specSource)
	parsedSpecs, parseResult := parser.FindSpecs(specName, conceptDictionary)
	return filterSpecsItems(parsedSpecs, newScenarioIndexFilterToRetain(indexToFilter)), parseResult
}
