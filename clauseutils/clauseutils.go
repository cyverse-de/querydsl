// Package clauseutils provides various useful functions for clauses to use in their processing
package clauseutils

import (
	"regexp"
	"strings"
)

// AddOrOperator takes whitespace and turns it into the OR operator, for a query string
func AddOrOperator(input string) string {
	inputSplit := strings.Split(input, " ")
	var rejoin []string
	blank := regexp.MustCompile(`^\s*$`)
	for _, part := range inputSplit {
		if !blank.MatchString(part) {
			rejoin = append(rejoin, strings.TrimSpace(part))
		}
	}

	return strings.Join(rejoin, " OR ")
}

// AddImplicitWildcard takes a query string with OR operators and adds wildcards around each piece separated by OR, unless the query already has wildcard-y syntax
func AddImplicitWildcard(input string) string {
	haswild := regexp.MustCompile(`[*?\\]`)
	if haswild.MatchString(input) {
		return input
	}

	inputSplit := strings.Split(input, " OR ")
	var rejoin []string

	blank := regexp.MustCompile(`^\s*$`)
	for _, part := range inputSplit {
		if !blank.MatchString(part) {
			rejoin = append(rejoin, "*"+strings.TrimSpace(part)+"*")
		}
	}

	return strings.Join(rejoin, " OR ")
}
