package permissions

import (
	"fmt"
	"testing"
)

type permissionTestCase struct {
	users             interface{}
	permission        string
	exact             bool
	expectedTerms     []string
	expectedWildcards []string
	shouldErr         bool

	expectedQuery string
}

func makeSSet(a []string) map[string]bool {
	r := make(map[string]bool)
	for _, item := range a {
		r[item] = true
	}
	return r
}

func makeISet(a []interface{}) map[string]bool {
	r := make(map[string]bool)
	for _, item := range a {
		r[item.(string)] = true
	}
	return r
}

// a string slice set-equality function suitable really only for this case
func stringSliceSetEqual(a []interface{}, b []string) (bool, string) {
	set_a := makeISet(a)
	for _, item := range b {
		if !set_a[item] {
			return false, item
		}
	}

	set_b := makeSSet(b)
	for _, item := range a {
		if !set_b[item.(string)] {
			return false, item.(string)
		}
	}
	return true, ""
}

// individualClause takes a clause and returns if it was a term clause, a terms clause, and/or a wildcard clause, and if a wildcard it returns what the query string was, for comparing as a set.
func individualClause(t *testing.T, c permissionTestCase, clause interface{}) (bool, bool, bool, string) {
	term, termOk := clause.(map[string]interface{})["term"]
	terms, termsOk := clause.(map[string]interface{})["terms"]
	wildcard, wildcardOk := clause.(map[string]interface{})["wildcard"]
	if termOk {
		ownPart, ok := term.(map[string]interface{})["userPermissions.permission"]
		if !ok {
			t.Error("Term was not for userPermissions.permission field")
		} else if ownPart.(string) != c.permission {
			t.Errorf("Term query for userPermissions.permission was not for the %q permission", c.permission)
		}
		return true, false, false, ""
	} else if termsOk {
		userPart, ok := terms.(map[string]interface{})["userPermissions.user"]
		if !ok {
			t.Error("Terms query was not for the userPermissions.user field")
		}

		userList, ok := userPart.([]interface{})
		if !ok {
			t.Error("user terms were  was not an array")
		}

		setEqual, missing := stringSliceSetEqual(userList, c.expectedTerms)
		if !setEqual {
			t.Errorf("Expected user list %+v to contain %s but did not", userList, missing)
		}
		return false, true, false, ""
	} else if wildcardOk {
		userPart, ok := wildcard.(map[string]interface{})["userPermissions.user"]
		if !ok {
			t.Error("Wildcard was not for the userPermissions.user field")
		}

		userWildcard, ok := userPart.(map[string]interface{})["wildcard"]
		if !ok {
			t.Error("Wildcard for userPermissions.user lacks 'wildcard' child")
		}

		expectedWildcardSet := makeSSet(c.expectedWildcards)
		if !expectedWildcardSet[userWildcard.(string)] {
			t.Errorf("Wildcard query part %q did not match any of expected values %v", userWildcard, c.expectedWildcards)
		}
		return false, false, true, userWildcard.(string)
	} else {
		t.Error("A clause is none of term, terms, or wildcard")
		return false, false, false, ""
	}
}

func TestPermissionsProcessor(t *testing.T) {
	cases := []permissionTestCase{
		{users: []string{"mian"}, permission: "own", expectedWildcards: []string{"mian#*"}},
		{users: []string{"mian"}, permission: "own", exact: true, expectedTerms: []string{"mian"}},
		{users: []string{"mian#foo"}, permission: "own", expectedTerms: []string{"mian#foo"}},
		{users: []string{"mian#foo", "mian"}, permission: "own", expectedTerms: []string{"mian#foo"}, expectedWildcards: []string{"mian#*"}},
		{users: []string{"ipctest", "mian"}, permission: "own", expectedWildcards: []string{"mian#*", "ipctest#*"}},
		{users: []int{666}, shouldErr: true},
		{shouldErr: true},             // empty owner
		{users: 444, shouldErr: true}, // bad type
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("permission:%s-exact:%t-users:%+v", c.permission, c.exact, c.users), func(t *testing.T) {
			args := make(map[string]interface{})

			args["users"] = c.users
			args["permission"] = c.permission
			args["exact"] = c.exact

			query, err := PermissionsProcessor(args)
			if c.shouldErr && err == nil {
				t.Errorf("PermissionsProcessor should have failed, instead returned nil error and query %+v", query)
			} else if !c.shouldErr && err != nil {
				t.Errorf("PermissionsProcessor failed with error: %q", err)
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

				if len(c.users.([]string)) == 1 && len(boolQuery.(map[string]interface{})) > 1 {
					t.Error("bool query had more than one subkey (should only have 'must')")
				}

				mustQuery, ok := boolQuery.(map[string]interface{})["must"]
				if !ok {
					t.Error("bool query did not have 'must' subkey")
				}

				mustShouldBeArray := len(c.users.([]string)) == 1

				mustQueryArray, ok := mustQuery.([]interface{})
				if mustShouldBeArray && !ok {
					t.Error("'must' was not an array")
				}

				if mustShouldBeArray && len(mustQueryArray) != 2 {
					t.Error("'must' was not two elements long")
				}

				if !mustShouldBeArray {
					mustQueryArray = []interface{}{
						mustQuery.(interface{}),
					}
				}

				hasTerm := false
				hasTerms := false
				hasWildcard := false
				foundWildcards := make([]string, 0, 0)
				for _, clause := range mustQueryArray {
					indivTerm, indivTerms, indivWildcard, foundWildcardString := individualClause(t, c, clause)
					foundWildcards = append(foundWildcards, foundWildcardString)
					if indivTerm {
						hasTerm = true
					}
					if indivTerms {
						hasTerms = true
					}
					if indivWildcard {
						hasWildcard = true
					}
				}
				if !mustShouldBeArray {
					shouldQuery, ok := boolQuery.(map[string]interface{})["should"]
					if !ok {
						t.Error("bool query did not have 'should' subkey")
					}

					shouldQueryArray, ok := shouldQuery.([]interface{})
					if !ok {
						t.Error("'should' was not an array")
					}
					for _, clause := range shouldQueryArray {
						indivTerm, indivTerms, indivWildcard, foundWildcardString := individualClause(t, c, clause)
						foundWildcards = append(foundWildcards, foundWildcardString)
						if indivTerm {
							hasTerm = true
						}
						if indivTerms {
							hasTerms = true
						}
						if indivWildcard {
							hasWildcard = true
						}
					}
				}

				if !(hasTerm && (hasWildcard || hasTerms)) {
					t.Error("query did not have both a term and either a wildcard or a terms query")
				}

				foundWildcardSet := makeSSet(foundWildcards)
				for _, item := range c.expectedWildcards {
					if !foundWildcardSet[item] {
						t.Errorf("Expected to find %q in the list of wildcard queries but it was absent", item)
					}
				}
			}
		})
	}
}
