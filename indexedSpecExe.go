package main

import (
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
	for ext, accepted := range acceptedExtensions {
		if accepted {
			re, _ := regexp.Compile(ext)
			if re.FindStringSubmatchIndex(specSource) != nil {
				return ext
			}
		}
	}
	return ""
}
