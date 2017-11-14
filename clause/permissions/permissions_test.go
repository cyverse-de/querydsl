package permissions

import (
	"fmt"
	"testing"
)

type permissionTestCase struct {
	users             interface{}
	permission        string
	permissionRecurse bool
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

// individualClause takes a clause and returns if it was for permissions, a user terms clause, and/or a user wildcard clause, and if a wildcard it returns what the query string was, for comparing as a set.
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
		// a terms query might be either
		// * (most likely) a set of users matched exactly
		// * a permissions clause for the case of write + recursive
		userPart, ok := terms.(map[string]interface{})["userPermissions.user"]
		if ok {
			// Set of users
			userList, ok := userPart.([]interface{})
			if !ok {
				t.Error("user terms were  was not an array")
			}

			setEqual, missing := stringSliceSetEqual(userList, c.expectedTerms)
			if !setEqual {
				t.Errorf("Expected user list %+v to contain %s but did not", userList, missing)
			}
			return false, true, false, ""
		} else {
			// Permissions write + recursive, or error
			permsPart, ok := terms.(map[string]interface{})["userPermissions.permission"]
			if !ok {
				t.Error("A terms clause was neither for the user nor the permission field")
			}

			permsList, ok := permsPart.([]interface{})
			setEqual, _ := stringSliceSetEqual(permsList, []string{"write", "own"})
			if !setEqual {
				t.Errorf("Expected permission list %+v to contain write and own but did not", permsList)
			}
			return true, false, false, ""
		}
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
		{users: []string{"mian#foo"}, permission: "read", permissionRecurse: true, expectedTerms: []string{"mian#foo"}},
		{users: []string{"mian#foo", "ipctest#foo"}, permission: "read", permissionRecurse: true, expectedTerms: []string{"mian#foo", "ipctest#foo"}},
		{users: []string{"mian#foo", "ipctest"}, permission: "read", permissionRecurse: true, expectedTerms: []string{"mian#foo"}, expectedWildcards: []string{"ipctest#*"}},
		{users: []string{"mian", "ipctest"}, permission: "read", permissionRecurse: true, expectedWildcards: []string{"mian#*", "ipctest#*"}},
		{users: []string{"mian#foo"}, permission: "write", permissionRecurse: true, expectedTerms: []string{"mian#foo"}},
		{users: []string{"mian#foo", "ipctest#foo"}, permission: "write", permissionRecurse: true, expectedTerms: []string{"mian#foo", "ipctest#foo"}},
		{users: []string{"mian#foo", "ipctest"}, permission: "write", permissionRecurse: true, expectedTerms: []string{"mian#foo"}, expectedWildcards: []string{"ipctest#*"}},
		{users: []string{"mian", "ipctest"}, permission: "write", permissionRecurse: true, expectedWildcards: []string{"mian#*", "ipctest#*"}},
		{users: []string{"mian#foo", "mian"}, permission: "own", expectedTerms: []string{"mian#foo"}, expectedWildcards: []string{"mian#*"}},
		{users: []string{"ipctest", "mian"}, permission: "own", expectedWildcards: []string{"mian#*", "ipctest#*"}},
		{users: []string{"mian"}, shouldErr: true},                      // no permission
		{users: []string{"mian"}, permission: "wrong", shouldErr: true}, // bad permission
		{shouldErr: true},                                               // empty owner
		{users: []int{666}, shouldErr: true},                            // bad type
		{users: 444, shouldErr: true},                                   // bad type
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("permission:%s-exact:%t-recurse:%t-users:%+v", c.permission, c.exact, c.permissionRecurse, c.users), func(t *testing.T) {
			args := make(map[string]interface{})

			args["users"] = c.users
			args["permission"] = c.permission
			args["permission_recurse"] = c.permissionRecurse
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

				// there should be a 'should' clause if:
				// * if there is more than one inexact search
				// * there is a mix of exact and inexact searches
				shouldHaveShould := len(c.expectedWildcards) > 1 || (len(c.expectedWildcards) > 0 && len(c.expectedTerms) > 0)

				// A permission clause only *doesn't* exist if it's set to read + recursive
				hasPermissionClause := !(c.permission == "read" && c.permissionRecurse)

				// 'must' can only have two items if
				// a.) there is no should clause
				// b.) there is a permission clause
				mustShouldBeArray := !shouldHaveShould && hasPermissionClause

				// 'must' should exist if:
				// * there is a permission clause
				// * there is no should clause
				mustShouldExist := !shouldHaveShould || hasPermissionClause

				mustQuery, ok := boolQuery.(map[string]interface{})["must"]
				if !ok && mustShouldExist {
					t.Error("bool query did not have 'must' subkey")
				}

				mustQueryArray, ok := mustQuery.([]interface{})
				if mustShouldBeArray && !ok {
					t.Error("'must' was not an array")
				}

				if mustShouldBeArray && len(mustQueryArray) != 2 {
					t.Error("'must' was not two elements long")
				}

				if !mustShouldBeArray && mustShouldExist {
					mustQueryArray = []interface{}{
						mustQuery.(interface{}),
					}
				}

				hasPerm := false // also set to true if one is not expected
				hasUserTerms := false
				hasWildcard := false
				foundWildcards := make([]string, 0, 0)

				if !hasPermissionClause {
					// no permission clause needs to be present if none should be present
					hasPerm = true
				}

				for _, clause := range mustQueryArray {
					indivTerm, indivTerms, indivWildcard, foundWildcardString := individualClause(t, c, clause)
					foundWildcards = append(foundWildcards, foundWildcardString)
					if indivTerm {
						hasPerm = true
					}
					if indivTerms {
						hasUserTerms = true
					}
					if indivWildcard {
						hasWildcard = true
					}
				}
				if shouldHaveShould {
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
							hasPerm = true
						}
						if indivTerms {
							hasUserTerms = true
						}
						if indivWildcard {
							hasWildcard = true
						}
					}
				}

				if !(hasPerm && (hasWildcard || hasUserTerms)) {
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
