// Package clauseutils provides various useful functions for clauses to use in their processing
package clauseutils

import (
	"regexp"
	"strings"

	"gopkg.in/olivere/elastic.v5"
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

// AddImplicitUsernameWildcard adds '#*' to input usernames which do not already contain a # character (which is the delimiter for qualified iRODS usernames)
func AddImplicitUsernameWildcard(input string) string {
	hasdelim := regexp.MustCompile(`[#]`)
	if hasdelim.MatchString(input) {
		return input
	}
	return input + "#*"
}

// RangeType specifies what sort of range to create for CreateRangeQuery
type RangeType int

const (
	// Both means add both lte and gte
	Both RangeType = iota
	// UpperOnly adds only lte
	UpperOnly
	// LowerOnly adds only gte
	LowerOnly
)

// CreateRangeQuery creates a simple range query for a field, and integer lower/upper limits, plus a RangeType to specify behavior
// Range values are int64s since this stuff deals with large numbers
func CreateRangeQuery(field string, rangetype RangeType, lower int64, upper int64) elastic.Query {
	rq := elastic.NewRangeQuery(field)
	if rangetype == Both || rangetype == UpperOnly {
		rq.Lte(upper)
	}
	if rangetype == Both || rangetype == LowerOnly {
		rq.Gte(lower)
	}
	return rq
}
