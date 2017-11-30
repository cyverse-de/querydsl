package label

import (
	"context"
	"errors"

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
			"label": {Type: "string", Summary: "The label to search for"},
			"exact": {Type: "bool", Summary: "Whether to search more precisely, or whether the query should be processed to add wildcards"},
		},
	}
)

type LabelArgs struct {
	Label string
	Exact bool
}

func LabelProcessor(_ context.Context, args map[string]interface{}) (elastic.Query, error) {
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
		processedQuery = clauseutils.AddImplicitWildcard(realArgs.Label)
	}
	query := elastic.NewQueryStringQuery(processedQuery).Field("label")
	return query, nil
}

func Register(qd *querydsl.QueryDSL) {
	qd.AddClauseType(typeKey, LabelProcessor, documentation)
}
