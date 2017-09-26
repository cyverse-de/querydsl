package label

import (
	"errors"
	"fmt"

	"github.com/cyverse-de/querydsl"
	"github.com/cyverse-de/querydsl/clause"
	"github.com/cyverse-de/querydsl/clauseutils"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/olivere/elastic.v5"
)

const (
	typeKey = "label"
)

var (
	documentation = clause.ClauseDocumentation{
		Summary: "Searches based on an object's label (typically, its filename)",
		Args: map[string]clause.ClauseArgumentDocumentation{
			"label": clause.ClauseArgumentDocumentation{Type: "string", Summary: "The label to search for"},
			"exact": clause.ClauseArgumentDocumentation{Type: "boolean", Summary: "Whether to search more precisely, or whether the query should be processed to add wildcards"},
		},
	}
)

type LabelArgs struct {
	Label string
	// Negation bool // TODO
	Exact bool
}

func LabelProcessor(args map[string]interface{}) (elastic.Query, error) {
	var realArgs LabelArgs
	err := mapstructure.Decode(args, &realArgs)
	if err != nil {
		return nil, err
	}

	if realArgs.Label == "" {
		return nil, errors.New("No label was passed, cannot create clause.")
	}

	var processedQuery string
	if realArgs.Exact {
		processedQuery = realArgs.Label
	} else {
		processedQuery = clauseutils.AddImplicitWildcard(clauseutils.AddOrOperator(realArgs.Label))
	}
	query := elastic.NewQueryStringQuery(processedQuery).Field("label").QueryName(fmt.Sprintf("label: %q", realArgs.Label))
	return query, nil
}

func Register() {
	querydsl.AddClauseType(typeKey, LabelProcessor, documentation)
}
