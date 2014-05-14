package main

import (
	"regexp"
	"github.com/twist2/common"
)

type resolverFn func(string) *stepArg
type specialTypeResolver struct {
	predefinedResolvers map[string]resolverFn
}

func newSpecialTypeResolver() *specialTypeResolver {
	resolver := new(specialTypeResolver)
	resolver.predefinedResolvers = initializePredefinedResolvers()
	return resolver
}

func initializePredefinedResolvers() map[string]resolverFn {
	return map[string]resolverFn {
		"file" : func(filePath string) *stepArg {
			fileContent := common.ReadFileContents(filePath)
			return &stepArg{value:fileContent, argType:static}
		},
	}
}

func (resolver *specialTypeResolver) resolve(arg string) *stepArg {
	regEx := regexp.MustCompile("(.*):(.*)")
	match := regEx.FindAllStringSubmatch(arg, -1)
	specialType := match[0][1]
	value := match[0][2]
	return resolver.getStepArg(specialType, value, arg)
}

func (resolver *specialTypeResolver) getStepArg(specialType string, value string, arg string) *stepArg {
	resolveFunc, found := resolver.predefinedResolvers[specialType]
	if found {
		return resolveFunc(value)
	}
	return &stepArg{value:arg, argType:static}
}

