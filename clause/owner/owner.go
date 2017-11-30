package owner

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
	typeKey = "owner"
)

var (
	documentation = clause.ClauseDocumentation{
		Summary: "Searches based on an object's owner(s)",
		Args: map[string]clause.ClauseArgumentDocumentation{
			"owner": {Type: "string", Summary: "The owner to search for. If it includes a # character, it will be searched exactly, otherwise the zone will be wildcarded."},
		},
	}
)

type OwnerArgs struct {
	Owner string
}

func OwnerProcessor(_ context.Context, args map[string]interface{}) (elastic.Query, error) {
	var realArgs OwnerArgs
	err := mapstructure.Decode(args, &realArgs)
	if err != nil {
		return nil, err
	}

	if realArgs.Owner == "" {
		return nil, errors.New("No owner was passed, cannot create clause.")
	}

	processedOwner := clauseutils.AddImplicitUsernameWildcard(realArgs.Owner)
	innerquery := elastic.NewBoolQuery().Must(elastic.NewTermQuery("userPermissions.permission", "own"), elastic.NewWildcardQuery("userPermissions.user", processedOwner))
	query := elastic.NewNestedQuery("userPermissions", innerquery)
	return query, nil
}

func Register(qd *querydsl.QueryDSL) {
	qd.AddClauseType(typeKey, OwnerProcessor, documentation)
}
