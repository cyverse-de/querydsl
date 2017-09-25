package label

import (
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
		},
	}
)

type LabelArgs struct {
	Label string
	// Negation bool // TODO
	// Exact    bool // TODO
}

func LabelProcessor(args map[string]interface{}) (elastic.Query, error) {
	var realArgs LabelArgs
	err := mapstructure.Decode(args, &realArgs)
	if err != nil {
		return nil, err
	}

	processedQuery := clauseutils.AddImplicitWildcard(clauseutils.AddOrOperator(realArgs.Label))
	query := elastic.NewQueryStringQuery(processedQuery).Field("label")
	return query, nil
}

func Register() {
	querydsl.AddClauseType(typeKey, LabelProcessor, documentation)
}
