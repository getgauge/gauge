package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/getgauge/common"
	"regexp"
	"strings"
)

type resolverFn func(string) (*stepArg, error)
type specialTypeResolver struct {
	predefinedResolvers map[string]resolverFn
}

func newSpecialTypeResolver() *specialTypeResolver {
	resolver := new(specialTypeResolver)
	resolver.predefinedResolvers = initializePredefinedResolvers()
	return resolver
}

func initializePredefinedResolvers() map[string]resolverFn {
	return map[string]resolverFn{
		"file": func(filePath string) (*stepArg, error) {
			fileContent, err := common.ReadFileContents(filePath)
			if err != nil {
				return nil, err
			}
			return &stepArg{value: fileContent, argType: static}, nil
		},
		"table": func(filePath string) (*stepArg, error) {
			csv, err := common.ReadFileContents(filePath)
			if err != nil {
				return nil, err
			}
			csvTable, err := convertCsvToTable(csv)
			if err != nil {
				return nil, err
			}
			return &stepArg{table: *csvTable, argType: tableArg}, nil
		},
	}
}

func convertCsvToTable(csvContents string) (*table, error) {
	r := csv.NewReader(strings.NewReader(csvContents))
	lines, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	table := new(table)
	for i, line := range lines {
		if i == 0 {
			table.addHeaders(line)
		} else {
			table.addRowValues(line)
		}
	}
	return table, nil
}

func (resolver *specialTypeResolver) resolve(arg string) (*stepArg, error) {
	regEx := regexp.MustCompile("(.*):(.*)")
	match := regEx.FindAllStringSubmatch(arg, -1)
	specialType := strings.TrimSpace(match[0][1])
	value := strings.TrimSpace(match[0][2])
	return resolver.getStepArg(specialType, value, arg)
}

func (resolver *specialTypeResolver) getStepArg(specialType string, value string, arg string) (*stepArg, error) {
	resolveFunc, found := resolver.predefinedResolvers[specialType]
	if found {
		return resolveFunc(value)
	}
	return nil, errors.New(fmt.Sprintf("Resolver not found for special param <%s>", arg))
}
