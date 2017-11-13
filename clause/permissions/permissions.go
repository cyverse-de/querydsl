package permissions

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
	typeKey = "permissions"
)

var (
	documentation = clause.ClauseDocumentation{
		Summary: "Searches based on an object's permissions for specified users",
		Args: map[string]clause.ClauseArgumentDocumentation{
			"users":      clause.ClauseArgumentDocumentation{Type: "[]string", Summary: "The users to search for. If there is only one user listed, if that username is not qualified (does not contain a # character), a wildcard will be added. If there are multiple users, they must all be qualified."},
			"permission": clause.ClauseArgumentDocumentation{Type: "string", Summary: "The permission to check for; should be one of 'own', 'write', or 'read', with own implying write implying read. To search for objects where the user has no permissions, use 'read' in a negation."},
			"exact":      clause.ClauseArgumentDocumentation{Type: "bool", Summary: "If set to true, do not add implicit wildcards even to usernames without the # character. This will effectively ignore those arguments."},
		},
	}
)

type PermissionsArgs struct {
	Users      []string
	Permission string
	Exact      bool
}

func PermissionsProcessor(args map[string]interface{}) (elastic.Query, error) {
	var realArgs PermissionsArgs
	err := mapstructure.Decode(args, &realArgs)
	if err != nil {
		return nil, err
	}

	if len(realArgs.Users) == 0 {
		return nil, errors.New("No users were passed, cannot create clause.")
	}

	if realArgs.Permission == "" {
		return nil, errors.New("No permission was passed, cannot create clause.")
	}

	if realArgs.Permission != "own" && realArgs.Permission != "write" && realArgs.Permission != "read" {
		return nil, fmt.Errorf("Got a permission of %q, but expected read, write, or own.", realArgs.Permission)
	}

	var terms []interface{}
	var shoulds []elastic.Query

	for _, user := range realArgs.Users {
		processedUser := clauseutils.AddImplicitUsernameWildcard(user)
		if processedUser == user || realArgs.Exact {
			terms = append(terms, user)
		} else {
			shoulds = append(shoulds, elastic.NewWildcardQuery("userPermissions.user", processedUser))
		}
	}

	innerquery := elastic.NewBoolQuery().Must(elastic.NewTermQuery("userPermissions.permission", realArgs.Permission))

	if len(terms) > 0 {
		termsq := elastic.NewTermsQuery("userPermissions.user", terms...)
		if len(shoulds) == 0 {
			innerquery.Must(termsq)
		} else {
			innerquery.Should(termsq)
		}
	}
	if len(shoulds) == 1 && len(terms) == 0 {
		innerquery.Must(shoulds...)
	} else if len(shoulds) > 0 {
		innerquery.Should(shoulds...)
	}

	query := elastic.NewNestedQuery("userPermissions", innerquery)
	return query, nil
}

func Register(qd *querydsl.QueryDSL) {
	qd.AddClauseType(typeKey, PermissionsProcessor, documentation)
}
