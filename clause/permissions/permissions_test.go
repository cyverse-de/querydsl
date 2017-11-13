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

//type permissionTestCase struct {
//	users         interface{}
//	permission    string
//	exact         bool
//	expectedQuery string
//	shouldErr     bool
//}

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
	fmt.Printf("clause: %+v\n", clause)
	term, termOk := clause.(map[string]interface{})["term"]
	terms, termsOk := clause.(map[string]interface{})["terms"]
	wildcard, wildcardOk := clause.(map[string]interface{})["wildcard"]
	if termOk {
		ownPart, ok := term.(map[string]interface{})["userPermissions.permission"]
		if !ok {
			fmt.Println("Term was not for userPermissions.permission field")
			t.Error("Term was not for userPermissions.permission field")
		} else if ownPart.(string) != c.permission {
			fmt.Printf("Term query for userPermissions.permission was not for the %q permission\n", c.permission)
			t.Errorf("Term query for userPermissions.permission was not for the %q permission", c.permission)
		}
		return true, false, false, ""
	} else if termsOk {
		userPart, ok := terms.(map[string]interface{})["userPermissions.user"]
		if !ok {
			fmt.Println("Terms query was not for the userPermissions.user field")
			t.Error("Terms query was not for the userPermissions.user field")
		}
		fmt.Printf("users: %+v\n", userPart)

		userList, ok := userPart.([]interface{})
		if !ok {
			fmt.Println("user terms were not an array")
			t.Error("user terms were  was not an array")
		}
		fmt.Printf("users list: %+v\n", userList)

		setEqual, missing := stringSliceSetEqual(userList, c.expectedTerms)
		if !setEqual {
			fmt.Printf("Expected user list %+v to contain %s but did not\n", userList, missing)
			t.Errorf("Expected user list %+v to contain %s but did not", userList, missing)
		}
		return false, true, false, ""
	} else if wildcardOk {
		userPart, ok := wildcard.(map[string]interface{})["userPermissions.user"]
		if !ok {
			fmt.Println("Wildcard was not for the userPermissions.user field")
			t.Error("Wildcard was not for the userPermissions.user field")
		}

		userWildcard, ok := userPart.(map[string]interface{})["wildcard"]
		if !ok {
			fmt.Println("Wildcard for userPermissions.user lacks 'wildcard' child")
			t.Error("Wildcard for userPermissions.user lacks 'wildcard' child")
		}

		expectedWildcardSet := makeSSet(c.expectedWildcards)
		if !expectedWildcardSet[userWildcard.(string)] {
			fmt.Printf("Wildcard query part %q did not match any of expected values %v\n", userWildcard, c.expectedWildcards)
			t.Errorf("Wildcard query part %q did not match any of expected values %v", userWildcard, c.expectedWildcards)
		}
		return false, false, true, userWildcard.(string)
	} else {
		fmt.Println("A clause is none of term, terms, or wildcard")
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
			//t.Parallel()
			args := make(map[string]interface{})

			args["users"] = c.users
			args["permission"] = c.permission
			args["exact"] = c.exact

			fmt.Println("----------------------------------------")
			fmt.Printf("args: %+v\n", args)

			query, err := PermissionsProcessor(args)
			fmt.Printf("query: %+v\n", query)
			fmt.Printf("processor error (should?: %t): %+v\n", c.shouldErr, err)
			if c.shouldErr && err == nil {
				t.Errorf("PermissionsProcessor should have failed, instead returned nil error and query %+v", query)
			} else if !c.shouldErr && err != nil {
				t.Errorf("PermissionsProcessor failed with error: %q", err)
			} else if !c.shouldErr {
				source, err := query.Source()
				if err != nil {
					fmt.Printf("Source get failed with error: %q\n", err)
					t.Errorf("Source get failed with error: %q", err)
				}
				fmt.Printf("source: %+v\n", source)

				nested, ok := source.(map[string]interface{})["nested"]
				if !ok {
					fmt.Println("Source did not contain 'nested'")
					t.Error("Source did not contain 'nested'")
				}
				fmt.Printf("nested: %+v\n", nested)

				path, ok := nested.(map[string]interface{})["path"]
				if !ok {
					fmt.Println("nested query did not include a path")
					t.Error("nested query did not include a path")
				}
				fmt.Printf("path: %+v\n", path)
				if path.(string) != "userPermissions" {
					fmt.Println("nested query path was not 'userPermissions'")
					t.Error("nested query path was not 'userPermissions'")
				}

				nestedQuery, ok := nested.(map[string]interface{})["query"]
				if !ok {
					fmt.Println("nested query did not include a query")
					t.Error("nested query did not include a query")
				}
				fmt.Printf("nestedQuery: %+v\n", nestedQuery)

				boolQuery, ok := nestedQuery.(map[string]interface{})["bool"]
				if !ok {
					fmt.Println("nested query was not a bool query")
					t.Error("nested query was not a bool query")
				}
				fmt.Printf("boolQuery: %+v\n", boolQuery)

				if len(c.users.([]string)) == 1 && len(boolQuery.(map[string]interface{})) > 1 {
					fmt.Println("bool query had more than one subkey (should only have 'must')")
					t.Error("bool query had more than one subkey (should only have 'must')")
				}

				mustQuery, ok := boolQuery.(map[string]interface{})["must"]
				if !ok {
					fmt.Println("bool query did not have 'must' subkey")
					t.Error("bool query did not have 'must' subkey")
				}
				fmt.Printf("mustQuery: %+v\n", mustQuery)

				mustShouldBeArray := len(c.users.([]string)) == 1

				mustQueryArray, ok := mustQuery.([]interface{})
				if mustShouldBeArray && !ok {
					fmt.Println("'must' was not an array")
					t.Error("'must' was not an array")
				}
				fmt.Printf("mustQueryArray: %+v\n", mustQueryArray)

				if mustShouldBeArray && len(mustQueryArray) != 2 {
					fmt.Println("'must' was not two elements long")
					t.Error("'must' was not two elements long")
				}

				if !mustShouldBeArray {
					mustQueryArray = []interface{}{
						mustQuery.(interface{}),
					}
				}

				fmt.Printf("mustQueryArray: %+v\n", mustQueryArray)

				hasTerm := false
				hasTerms := false
				hasWildcard := false
				foundWildcards := make([]string, 0, 0)
				for _, clause := range mustQueryArray {
					indivTerm, indivTerms, indivWildcard, foundWildcardString := individualClause(t, c, clause)
					foundWildcards = append(foundWildcards, foundWildcardString)
					fmt.Printf("this clause: %t %t %t\n", indivTerm, indivTerms, indivWildcard)
					if indivTerm {
						hasTerm = true
					}
					if indivTerms {
						hasTerms = true
					}
					if indivWildcard {
						hasWildcard = true
					}
					fmt.Printf("bool state: %t %t %t\n", hasTerm, hasTerms, hasWildcard)
				}
				if !mustShouldBeArray {
					shouldQuery, ok := boolQuery.(map[string]interface{})["should"]
					if !ok {
						fmt.Println("bool query did not have 'should' subkey")
						t.Error("bool query did not have 'should' subkey")
					}
					fmt.Printf("shouldQuery: %+v\n", shouldQuery)

					shouldQueryArray, ok := shouldQuery.([]interface{})
					if !ok {
						fmt.Println("'should' was not an array")
						t.Error("'should' was not an array")
					}
					fmt.Printf("shouldQueryArray: %+v\n", shouldQueryArray)
					for _, clause := range shouldQueryArray {
						indivTerm, indivTerms, indivWildcard, foundWildcardString := individualClause(t, c, clause)
						foundWildcards = append(foundWildcards, foundWildcardString)
						fmt.Printf("this clause: %t %t %t\n", indivTerm, indivTerms, indivWildcard)
						if indivTerm {
							hasTerm = true
						}
						if indivTerms {
							hasTerms = true
						}
						if indivWildcard {
							hasWildcard = true
						}
						fmt.Printf("bool state: %t %t %t\n", hasTerm, hasTerms, hasWildcard)
					}
				}

				if !(hasTerm && (hasWildcard || hasTerms)) {
					fmt.Println("query did not have both a term and either a wildcard or a terms query")
					t.Error("query did not have both a term and either a wildcard or a terms query")
				}

				fmt.Printf("found wildcards: %+v\n", foundWildcards)
				foundWildcardSet := makeSSet(foundWildcards)
				for _, item := range c.expectedWildcards {
					if !foundWildcardSet[item] {
						fmt.Printf("Expected to find %q in the list of wildcard queries but it was absent\n", item)
						t.Errorf("Expected to find %q in the list of wildcard queries but it was absent", item)
					}
				}
			}
		})
	}
}
