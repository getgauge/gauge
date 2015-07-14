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
	"github.com/getgauge/gauge/util"
	"regexp"
	"strconv"
)

func isIndexedSpec(specSource string) bool {
	return getIndex(specSource) != nil
}

func GetIndexedSpecName(IndexedSpec string) (string, int) {
	var specName, scenarioNum string
	index := getIndex(IndexedSpec)
	for i := 0; i < index[0]; i++ {
		specName += string(IndexedSpec[i])
	}
	typeOfSpec := getTypeOfSpecFile(IndexedSpec)
	for i := index[0] + len(typeOfSpec) + 1; i < index[1]; i++ {
		scenarioNum += string(IndexedSpec[i])
	}
	scenarioNumber, _ := strconv.Atoi(scenarioNum)
	return specName + typeOfSpec, scenarioNumber
}

func getIndex(specSource string) []int {
	re, _ := regexp.Compile(getTypeOfSpecFile(specSource) + ":[0-9]+$")
	index := re.FindStringSubmatchIndex(specSource)
	if index != nil {
		return index
	}
	return nil
}

func getTypeOfSpecFile(specSource string) string {
	for ext, accepted := range util.AcceptedExtensions {
		if accepted {
			re, _ := regexp.Compile(ext)
			if re.FindStringSubmatchIndex(specSource) != nil {
				return ext
			}
		}
	}
	return ""
}
