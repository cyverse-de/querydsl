package querydsl

import (
	"testing"

	"gopkg.in/olivere/elastic.v5"
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
		genericClause := GenericClause{Query: &c.query, Clause: &c.clause}
		isQuery := genericClause.IsQuery()
		if isQuery != c.isQuery {
			t.Errorf("GenericClause{Query: &%+v, Clause: &%+v} returned %q from IsQuery, not %q", c.query, c.clause, isQuery, c.isQuery)
		}
		isClause := genericClause.IsClause()
		if isClause != c.isClause {
			t.Errorf("GenericClause{Query: &%+v, Clause: &%+v} returned %q from IsClause, not %q", c.query, c.clause, isClause, c.isClause)
		}
	}
}

func TestTranslateClause(t *testing.T) {
	AddClauseType("foo", func(args map[string]interface{}) (elastic.Query, error) {
		return elastic.NewTermQuery("user", "arbitrary"), nil
	})

	clause := Clause{Type: "foo"}

	translated, err := clause.Translate()
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
