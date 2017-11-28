package path

import (
	"context"
	"errors"

	"github.com/cyverse-de/querydsl"
	"github.com/cyverse-de/querydsl/clause"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/olivere/elastic.v5"
)

const (
	typeKey = "path"
)

var (
	documentation = clause.ClauseDocumentation{
		Summary: "Searches based on an object's full path",
		Args: map[string]clause.ClauseArgumentDocumentation{
			"prefix": clause.ClauseArgumentDocumentation{Type: "string", Summary: "The path prefix to search for"},
		},
	}
)

type PathArgs struct {
	Prefix string
}

func PathProcessor(_ context.Context, args map[string]interface{}) (elastic.Query, error) {
	var realArgs PathArgs
	err := mapstructure.Decode(args, &realArgs)
	if err != nil {
		return nil, err
	}

	if realArgs.Prefix == "" {
		return nil, errors.New("No prefix was passed, cannot create clause.")
	}

	query := elastic.NewPrefixQuery("path", realArgs.Prefix)
	return query, nil
}

func Register(qd *querydsl.QueryDSL) {
	qd.AddClauseType(typeKey, PathProcessor, documentation)
}
