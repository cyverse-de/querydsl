package querydsl

import (
	"fmt"
	"testing"

	"gopkg.in/olivere/elastic.v5"

	"github.com/cyverse-de/querydsl/clause"
)

func TestIsQuery_IsClause(t *testing.T) {
	cases := []struct {
		query    Query
		clause   Clause
		isQuery  bool
		isClause bool
	}{
		{Query{}, Clause{}, false, false},
		{Query{All: nil, Any: nil, None: nil}, Clause{}, false, false},
		{Query{}, Clause{Type: ""}, false, false},
		{Query{}, Clause{Type: "arbitrary"}, false, true},                        // arbitrary type in clause
		{Query{All: []*GenericClause{&GenericClause{}}}, Clause{}, true, false},  // arbitrary clause in All
		{Query{Any: []*GenericClause{&GenericClause{}}}, Clause{}, true, false},  // arbitrary clause in Any
		{Query{None: []*GenericClause{&GenericClause{}}}, Clause{}, true, false}, // arbitrary clause in None
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("GenericClause{Query: &%+v, Clause: &%+v}", c.query, c.clause), func(t *testing.T) {
			genericClause := GenericClause{Query: &c.query, Clause: &c.clause}
			isQuery := genericClause.IsQuery()
			if isQuery != c.isQuery {
				t.Errorf("returned %q from IsQuery, not %q", isQuery, c.isQuery)
			}
			isClause := genericClause.IsClause()
			if isClause != c.isClause {
				t.Errorf("returned %q from IsClause, not %q", isClause, c.isClause)
			}
		})
	}
}

func addTestingClauseType() (*QueryDSL, Clause) {
	qd := New()
	qd.AddClauseType("foo", func(args map[string]interface{}) (elastic.Query, error) {
		return elastic.NewTermQuery("user", "arbitrary"), nil
	}, clause.ClauseDocumentation{})

	return qd, Clause{Type: "foo"}
}

func TestTranslateClauseNoTypes(t *testing.T) {
	clause := Clause{Type: "type-that-doesnt-exist"}

	_, err := clause.Translate(New())
	if err == nil {
		t.Errorf("Translate did not return error using nonexistent clause type, which it should")
	}
}

func TestTranslateClause(t *testing.T) {
	qd, clause := addTestingClauseType()

	translated, err := clause.Translate(qd)
	if err != nil {
		t.Errorf("Translate failed with error: %q", err)
	}
	querySource, err := translated.Source()
	if err != nil {
		t.Errorf("Source get failed with error: %q", err)
	}

	termQuery, ok := querySource.(map[string]interface{})["term"]
	if !ok {
		t.Error("Source did not contain 'term'")
	}

	userValue, ok := termQuery.(map[string]interface{})["user"]
	if !ok {
		t.Error("term query did not contain 'user'")
	}
	if userValue.(string) != "arbitrary" {
		t.Errorf("term user query was %q rather than %q", userValue, "arbitrary")
	}
}

func TestTranslateQueryNoTypes(t *testing.T) {
	clause := Clause{Type: "type-that-doesnt-exist"}
	query := Query{All: []*GenericClause{&GenericClause{Clause: &clause}}}

	_, err := query.Translate(New())
	if err == nil {
		t.Errorf("Translate did not return error using nonexistent clause type, which it should")
	}
}

func TestTranslateQuery(t *testing.T) {
	qd, clause := addTestingClauseType()

	queryAll := Query{All: []*GenericClause{&GenericClause{Clause: &clause}}}
	queryAny := Query{Any: []*GenericClause{&GenericClause{Clause: &clause}}}
	queryNone := Query{None: []*GenericClause{&GenericClause{Clause: &clause}}}

	testGivenQuery := func(t *testing.T, query Query, subfield string) {
		translated, err := query.Translate(qd)
		if err != nil {
			t.Errorf("Translate failed with error: %q", err)
		}
		querySource, err := translated.Source()
		if err != nil {
			t.Errorf("Source get failed with error: %q", err)
		}

		boolQuery, ok := querySource.(map[string]interface{})["bool"]
		if !ok {
			t.Error("Source did not contain 'bool'")
		}

		section, ok := boolQuery.(map[string]interface{})[subfield]
		if !ok {
			t.Errorf("bool did not contain '%s'", subfield)
		}

		termQuery, ok := section.(map[string]interface{})["term"]
		if !ok {
			t.Error("subfield did not contain 'term'")
		}

		userValue, ok := termQuery.(map[string]interface{})["user"]
		if !ok {
			t.Error("term query did not contain 'user'")
		}
		if userValue.(string) != "arbitrary" {
			t.Errorf("term user query was %q rather than %q", userValue, "arbitrary")
		}
	}

	t.Run("must", func(t *testing.T) {
		testGivenQuery(t, queryAll, "must")
	})
	t.Run("should", func(t *testing.T) {
		testGivenQuery(t, queryAny, "should")
	})
	t.Run("must_not", func(t *testing.T) {
		testGivenQuery(t, queryNone, "must_not")
	})
}
