package path

import (
	"context"
	"errors"
	"fmt"

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
			"prefix": {Type: "string", Summary: "The path prefix to search for"},
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

func PathSummary(_ context.Context, args map[string]interface{}) (string, error) {
	var realArgs PathArgs
	err := mapstructure.Decode(args, &realArgs)
	if err != nil {
		return "", err
	}

	if realArgs.Prefix == "" {
		return "", errors.New("No prefix was passed, cannot create clause.")
	}

	return fmt.Sprintf("path=\"%s\"", realArgs.Prefix), nil
}

func Register(qd *querydsl.QueryDSL) {
	qd.AddClauseTypeSummarized(typeKey, PathProcessor, documentation, PathSummary)
}
