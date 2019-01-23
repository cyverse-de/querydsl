package querydsl

import (
	"context"
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
		{Query{}, Clause{Type: "arbitrary"}, false, true},          // arbitrary type in clause
		{Query{All: []*GenericClause{{}}}, Clause{}, true, false},  // arbitrary clause in All
		{Query{Any: []*GenericClause{{}}}, Clause{}, true, false},  // arbitrary clause in Any
		{Query{None: []*GenericClause{{}}}, Clause{}, true, false}, // arbitrary clause in None
	}
	for _, c := range cases {
		t.Run(fmt.Sprintf("GenericClause{Query: &%+v, Clause: &%+v}", c.query, c.clause), func(t *testing.T) {
			genericClause := GenericClause{Query: &c.query, Clause: &c.clause}
			isQuery := genericClause.IsQuery()
			if isQuery != c.isQuery {
				t.Errorf("returned %v from IsQuery, not %v", isQuery, c.isQuery)
			}
			isClause := genericClause.IsClause()
			if isClause != c.isClause {
				t.Errorf("returned %v from IsClause, not %v", isClause, c.isClause)
			}
		})
	}
}

func addTestingClauseType() (*QueryDSL, Clause) {
	qd := New()
	qd.AddClauseTypeSummarized("foo", func(_ context.Context, args map[string]interface{}) (elastic.Query, error) {
		return elastic.NewTermQuery("user", "arbitrary"), nil
	}, clause.ClauseDocumentation{}, func(_ context.Context, args map[string]interface{}) (string, error) {
		return fmt.Sprintf("foo:%+v", args), nil
	})

	return qd, Clause{Type: "foo"}
}

func TestSummarizeNoTypes(t *testing.T) {
	cases := []struct {
		clause   GenericClause
		expected string
	}{
		{GenericClause{}, "GenericClause &{Clause:<nil> Query:<nil>} is neither a properly-formatted Query nor a Clause"},
		{GenericClause{Clause: &Clause{Type: "arbitrary"}}, "{clause:arbitrary}"},
		{GenericClause{Query: &Query{All: []*GenericClause{&GenericClause{Clause: &Clause{Type: "arbitrary"}}}}}, "All:[{clause:arbitrary}]"},
		{GenericClause{Query: &Query{All: []*GenericClause{&GenericClause{Clause: &Clause{Type: "arbitrary"}}}, Any: []*GenericClause{&GenericClause{Clause: &Clause{Type: "arbitrary"}}}}}, "All:[{clause:arbitrary}] Any:[{clause:arbitrary}]"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s", c.expected), func(t *testing.T) {
			summary := c.clause.Summarize(context.Background(), New())
			if summary != c.expected {
				t.Errorf("Got '%s' from summarize, not '%s'", summary, c.expected)
			}
		})
	}
}

func TestSummarize(t *testing.T) {
	qd, clause := addTestingClauseType()
	cases := []struct {
		clause   GenericClause
		expected string
	}{
		{GenericClause{}, "GenericClause &{Clause:<nil> Query:<nil>} is neither a properly-formatted Query nor a Clause"},
		{GenericClause{Clause: &clause}, "foo:map[]"},
		{GenericClause{Query: &Query{All: []*GenericClause{&GenericClause{Clause: &clause}}}}, "All:[foo:map[]]"},
		{GenericClause{Query: &Query{All: []*GenericClause{&GenericClause{Clause: &Clause{Type: "arbitrary"}}}, Any: []*GenericClause{&GenericClause{Clause: &clause}}}}, "All:[{clause:arbitrary}] Any:[foo:map[]]"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s", c.expected), func(t *testing.T) {
			summary := c.clause.Summarize(context.Background(), qd)
			if summary != c.expected {
				t.Errorf("Got '%s' from summarize, not '%s'", summary, c.expected)
			}
		})
	}
}

func TestTranslateClauseNoTypes(t *testing.T) {
	clause := Clause{Type: "type-that-doesnt-exist"}

	_, err := clause.Translate(context.Background(), New())
	if err == nil {
		t.Errorf("Translate did not return error using nonexistent clause type, which it should")
	}
}

func TestTranslateClause(t *testing.T) {
	qd, clause := addTestingClauseType()

	translated, err := clause.Translate(context.Background(), qd)
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
	query := Query{All: []*GenericClause{{Clause: &clause}}}

	_, err := query.Translate(context.Background(), New())
	if err == nil {
		t.Errorf("Translate did not return error using nonexistent clause type, which it should")
	}
}

func TestTranslateGenericClauseEmpty(t *testing.T) {
	query := GenericClause{}

	_, err := query.Translate(context.Background(), New())
	if err == nil {
		t.Errorf("Translate did not return an error when passed an empty query")
	}
}

func testSection(t *testing.T, source interface{}, subfield string) {
	boolQuery, ok := source.(map[string]interface{})["bool"]
	if !ok {
		t.Error("did not contain 'bool'")
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

func TestTranslateQuery(t *testing.T) {
	qd, clause := addTestingClauseType()

	queryAll := Query{All: []*GenericClause{{Clause: &clause}}}
	queryAny := Query{Any: []*GenericClause{{Clause: &clause}}}
	queryNone := Query{None: []*GenericClause{{Clause: &clause}}}

	testGivenQuery := func(t *testing.T, query Query, subfield string) {
		translated, err := query.Translate(context.Background(), qd)
		if err != nil {
			t.Errorf("Translate failed with error: %q", err)
		}
		querySource, err := translated.Source()
		if err != nil {
			t.Errorf("Source get failed with error: %q", err)
		}

		testSection(t, querySource, subfield)
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

	queryNested := Query{Any: []*GenericClause{{Query: &Query{All: []*GenericClause{{Clause: &clause}}}}}}
	t.Run("nested_query", func(t *testing.T) {
		translated, err := queryNested.Translate(context.Background(), qd)
		if err != nil {
			t.Errorf("Translate failed with error: %q", err)
		}
		querySource, err := translated.Source()
		if err != nil {
			t.Errorf("Source get failed with error: %q", err)
		}
		boolQuery, ok := querySource.(map[string]interface{})["bool"]
		if !ok {
			t.Error("did not contain 'bool'")
		}

		section, ok := boolQuery.(map[string]interface{})["should"]
		if !ok {
			t.Error("bool did not contain 'should'")
		}
		testSection(t, section, "must")
	})
}
