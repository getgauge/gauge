package main

import (
	"strings"
	"strconv"
	"fmt"
)

func evaluateTags(tags string) string{
	listOfOperators := make([]string,0)
	listOfTags := strings.FieldsFunc(tags, func(r rune) bool {
			isValidOperator := r == '&' || r == '|'
			if isValidOperator {
				operator, _ := strconv.Unquote(strconv.QuoteRuneToASCII(r))
				listOfOperators = append(listOfOperators, operator)
				return isValidOperator
			}
			return false
		})
	fmt.Println(listOfTags)
	fmt.Println(listOfOperators)
	return strings.Replace(tags,"&", ",",-1)
}
