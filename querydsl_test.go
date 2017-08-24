package querydsl

import (
	"testing"
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
