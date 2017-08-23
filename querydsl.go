// Package querydsl provides programmatic mapping from the CyVerse search DSL to Elasticsearch queries
//package querydsl
package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"gopkg.in/olivere/elastic.v5"
)

// Query represents a boolean query (which can also act as a clause within a query)
type Query struct {
	All  []*GenericClause `json:"all,omitempty"`
	Any  []*GenericClause `json:"any,omitempty"`
	None []*GenericClause `json:"none,omitempty"`
}

// Clause represents a particular clause (including, potentially, a Query)
type Clause struct {
	Type string                 `json:"type,omitempty"`
	Args map[string]interface{} `json:"args,omitempty"`
}

type GenericClause struct {
	*Clause
	*Query
}

func (c *GenericClause) IsQuery() bool {
	return c.Query != nil && (len(c.All) > 0 || len(c.Any) > 0 || len(c.None) > 0)
}

func (c *GenericClause) IsClause() bool {
	return c.Clause != nil && len(c.Type) > 0
}

type Translatable interface {
	Translate() (elastic.Query, error)
}

// Translate turns a GenericClause into an elastic.Query
func (c *GenericClause) Translate() (elastic.Query, error) {
	if c.IsQuery() {
		// Looks like it's another nested query.
		query := Query{All: c.All, Any: c.Any, None: c.None}
		return query.Translate()
	} else if c.IsClause() {
		clause := Clause{Type: c.Type, Args: c.Args}
		return clause.Translate()
	} else {
		return nil, fmt.Errorf("GenericClause %+v is neither a properly-formatted Query nor a Clause", c)
	}
}

// Translate turns a regular Clause into an elastic.Query
func (c *Clause) Translate() (elastic.Query, error) {
	return elastic.NewTermQuery("user", "olivere"), nil
}

// launchClauseTranslators launches a set of goroutines to translate a set of Clauses
// internally, it uses a WaitGroup to track when all of the goroutines it
// launched are finished, and once they are signals to the WaitGroup passed as
// an argument; this way if several calls to this function all pass the same
// WaitGroup, that WaitGroup only shows up finished when every clause across
// all of the several calls is processed
//
// This long comment brought to you by the author not wanting to forget how this works
func launchClauseTranslators(clauses []*GenericClause, waitgroup *sync.WaitGroup, resultsChan chan elastic.Query, errChan chan error) {
	var innerwg sync.WaitGroup

	waitgroup.Add(1)

	for _, clause := range clauses {
		innerwg.Add(1)
		go func(clause *GenericClause, wg *sync.WaitGroup) {
			defer wg.Done()
			query, err := clause.Translate()
			if err != nil {
				errChan <- err
			} else {
				resultsChan <- query
			}
		}(clause, &innerwg)
	}

	go func(subpartswg *sync.WaitGroup, innerwg *sync.WaitGroup) {
		innerwg.Wait()
		subpartswg.Done()
	}(waitgroup, &innerwg)
}

// Translate turns a Query into an elastic.Query by way of translating everything contained within
func (q *Query) Translate() (elastic.Query, error) {
	baseQuery := elastic.NewBoolQuery()

	// Result channels
	allChan := make(chan elastic.Query, 10)
	anyChan := make(chan elastic.Query, 10)
	noneChan := make(chan elastic.Query, 10)

	// subpartswg tracks whether all three of the other waitgroups have completed
	var subpartswg sync.WaitGroup

	// errChan is used by everything to propagate errors
	errChan := make(chan error)

	launchClauseTranslators(q.All, &subpartswg, allChan, errChan)
	launchClauseTranslators(q.Any, &subpartswg, anyChan, errChan)
	launchClauseTranslators(q.None, &subpartswg, noneChan, errChan)

	// wait for all translators to be done, then send a nil error to signal completion
	go func() {
		subpartswg.Wait()
		errChan <- nil
	}()

	for {
		select {
		case query := <-allChan:
			baseQuery.Must(query)
		case query := <-anyChan:
			baseQuery.Should(query)
		case query := <-noneChan:
			baseQuery.MustNot(query)
		case err := <-errChan:
			if err != nil {
				return nil, err
			}
			return baseQuery, nil
		}
	}
}

func main() {
	var jsonBlob = []byte(`{
		"all": [{"type": "foo", "args": {}}],
		"any": [{"all": [], "any": [{"type": "foo", "args": {}}], "none": []}],
		"none": []
	}`)
	var query Query
	err := json.Unmarshal(jsonBlob, &query)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Printf("%+v\n", query)
	translated, err := query.Translate()
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Printf("%s\n", translated)
	querySource, err := translated.Source()
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Printf("%s\n", querySource)
	translatedJSON, err := json.Marshal(querySource)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Printf("%s\n", translatedJSON)
}
