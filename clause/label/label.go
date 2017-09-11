package label

import (
	"github.com/cyverse-de/querydsl"
	"github.com/cyverse-de/querydsl/clause"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/olivere/elastic.v5"
)

const (
	typeKey = "label"
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

	query := elastic.NewQueryStringQuery(realArgs.Label).Field("label")
	return query, nil
}

func Register() {
	querydsl.AddClauseType(typeKey, LabelProcessor, clause.ClauseDocumentation{})
}
