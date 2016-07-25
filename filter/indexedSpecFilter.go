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
	"regexp"
	"strconv"
)

func isIndexedSpec(specSource string) bool {
	return getIndex(specSource) != 0
}

func getIndexedSpecName(indexedSpec string) (string, int) {
	index := getIndex(indexedSpec)
	specName := indexedSpec[:index]
	scenarioNum := indexedSpec[index+1:]
	scenarioNumber, _ := strconv.Atoi(scenarioNum)
	return specName, scenarioNumber
}

func getIndex(specSource string) int {
	re, _ := regexp.Compile(":[0-9]+$")
	index := re.FindStringSubmatchIndex(specSource)
	if index != nil {
		return index[0]
	}
	return 0
}
