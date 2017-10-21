package owner

import (
	"testing"
)

func TestOwnerProcessor(t *testing.T) {
	cases := []struct {
		owner         interface{}
		expectedQuery string
		shouldErr     bool
	}{
		{owner: "mian", expectedQuery: "mian#*"},
		{owner: "mian#foo", expectedQuery: "mian#foo"},
		{shouldErr: true},             // empty owner
		{owner: 444, shouldErr: true}, // bad type
	}

	for _, c := range cases {
		args := make(map[string]interface{})

		args["owner"] = c.owner

		query, err := OwnerProcessor(args)
		if c.shouldErr && err == nil {
			t.Errorf("OwnerProcessor should have failed, instead returned nil error and query %+v", query)
		} else if !c.shouldErr && err != nil {
			t.Errorf("OwnerProcessor failed with error: %q", err)
		} else if !c.shouldErr {
			source, err := query.Source()
			if err != nil {
				t.Errorf("Source get failed with error: %q", err)
			}

			nested, ok := source.(map[string]interface{})["nested"]
			if !ok {
				t.Error("Source did not contain 'nested'")
			}

			path, ok := nested.(map[string]interface{})["path"]
			if !ok {
				t.Error("nested query did not include a path")
			}
			if path.(string) != "userPermissions" {
				t.Error("nested query path was not 'userPermissions'")
			}

			nestedQuery, ok := nested.(map[string]interface{})["query"]
			if !ok {
				t.Error("nested query did not include a query")
			}

			boolQuery, ok := nestedQuery.(map[string]interface{})["bool"]
			if !ok {
				t.Error("nested query was not a bool query")
			}

			if len(boolQuery.(map[string]interface{})) > 1 {
				t.Error("bool query had more than one subkey (should only have 'must')")
			}

			mustQuery, ok := boolQuery.(map[string]interface{})["must"]
			if !ok {
				t.Error("bool query did not have 'must' subkey")
			}

			mustQueryArray, ok := mustQuery.([]interface{})
			if !ok {
				t.Error("'must' was not an array")
			}

			if len(mustQueryArray) != 2 {
				t.Error("'must' was not two elements long")
			}

			hasTerm := false
			hasWildcard := false
			for _, clause := range mustQueryArray {
				term, ok := clause.(map[string]interface{})["term"]
				if ok {
					hasTerm = true
					ownPart, ok := term.(map[string]interface{})["userPermissions.permission"]
					if !ok {
						t.Error("Term was not for userPermissions.permission field")
					} else if ownPart.(string) != "own" {
						t.Error("Term query for userPermissions.permission was not for the 'own' permission")
					}
				} else {
					wildcard, ok := clause.(map[string]interface{})["wildcard"]
					if !ok {
						t.Error("A clause of the 'must' is neither a term nor a wildcard")
					} else {
						hasWildcard = true
						userPart, ok := wildcard.(map[string]interface{})["userPermissions.user"]
						if !ok {
							t.Error("Wildcard was not for the userPermissions.user field")
						}

						userWildcard, ok := userPart.(map[string]interface{})["wildcard"]
						if !ok {
							t.Error("Wildcard for userPermissions.user lacks 'wildcard' child")
						}

						if userWildcard.(string) != c.expectedQuery {
							t.Errorf("Wildcard query part %q did not match expected value %q", userWildcard, c.expectedQuery)
						}
					}
				}
			}

			if !(hasTerm && hasWildcard) {
				t.Error("'must' did not have both a term and a wildcard query")
			}
		}
	}
}
