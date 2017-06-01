package filter

import (
	"github.com/getgauge/gauge/gauge"
	"github.com/getgauge/gauge/logger"
	"github.com/getgauge/gauge/util"
)

var ExecuteTags string
var Distribute int
var NumberOfExecutionStreams int

func FilterSpecs(specs []*gauge.Specification) []*gauge.Specification {
	specs = applyFilters(specs, specsFilters())
	if ExecuteTags != "" && len(specs) > 0 {
		logger.Debug("The following specifications satisfy filter criteria:")
		for _, s := range specs {
			logger.Debug(util.RelPathToProjectRoot(s.FileName))
		}
	}
	return specs
}

func specsFilters() []specsFilter {
	return []specsFilter{&tagsFilter{ExecuteTags}, &specsGroupFilter{Distribute, NumberOfExecutionStreams}}
}

func applyFilters(specsToExecute []*gauge.Specification, filters []specsFilter) []*gauge.Specification {
	for _, specsFilter := range filters {
		specsToExecute = specsFilter.filter(specsToExecute)
	}
	return specsToExecute
}
