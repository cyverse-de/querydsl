// Package clauseutils provides various useful functions for clauses to use in their processing
package clauseutils

import (
	"regexp"
	"strings"
)

// AddImplicitWildcard takes a query string with OR operators and adds wildcards around each piece separated by OR, unless the query already has wildcard-y syntax
func AddImplicitWildcard(input string) string {
	haswild := regexp.MustCompile(`[*?\\"]`)
	if haswild.MatchString(input) {
		return input
	}

	splitRegex := regexp.MustCompile(`( OR |\s+)`)
	inputSplit := splitRegex.Split(input, -1)
	var rejoin []string

	blank := regexp.MustCompile(`^\s*$`)
	for _, part := range inputSplit {
		if !blank.MatchString(part) {
			rejoin = append(rejoin, "*"+strings.TrimSpace(part)+"*")
		}
	}

	return strings.Join(rejoin, " ")
}

func AddImplicitUsernameWildcard(input string) string {
	hasdelim := regexp.MustCompile(`[#]`)
	if hasdelim.MatchString(input) {
		return input
	} else {
		return input + "#*"
	}
}
