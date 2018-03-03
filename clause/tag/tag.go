package tag

import (
	"context"
	"errors"

	"github.com/cyverse-de/querydsl"
	"github.com/cyverse-de/querydsl/clause"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/olivere/elastic.v5"
)

const (
	typeKey = "tag"
)

var (
	documentation = clause.ClauseDocumentation{
		Summary: "Searches based on a set of provided tag IDs",
		Args: map[string]clause.ClauseArgumentDocumentation{
			"tags": {Type: "[]string", Summary: "The tag UUIDs to search for"},
		},
	}
)

type TagArgs struct {
	Tags []string
}

func TagProcessor(_ context.Context, args map[string]interface{}) (elastic.Query, error) {
	var realArgs TagArgs
	err := mapstructure.Decode(args, &realArgs)
	if err != nil {
		return nil, err
	}

	if len(realArgs.Tags) == 0 {
		return nil, errors.New("No tags were passed, cannot create clause.")
	}

	query := elastic.NewBoolQuery()

	for _, tag := range realArgs.Tags {
		termsLookup := elastic.NewTermsLookup().Type("tag").Id(tag).Path("targets.id")
		query.Should(elastic.NewTermsQuery("id").TermsLookup(termsLookup))
	}

	return query, nil
}

func Register(qd *querydsl.QueryDSL) {
	qd.AddClauseType(typeKey, TagProcessor, documentation)
}
