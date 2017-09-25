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
