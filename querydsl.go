// Package querydsl provides programmatic mapping from the CyVerse search DSL to Elasticsearch queries
//package querydsl
package main

import (
	"encoding/json"
	"fmt"
	"sync"

	"gopkg.in/olivere/elastic.v5"
)

type Query struct {
	All  []*Clause `json:"all,omitempty"`
	Any  []*Clause `json:"any,omitempty"`
	None []*Clause `json:"none,omitempty"`
}

type Clause struct {
	Type string                 `json:"type,omitempty"`
	Args map[string]interface{} `json:"args,omitempty"`
	Query
}

// ToQuery turns a Clause into an elastic.Query
func (c *Clause) ToQuery() (elastic.Query, error) {
	if len(c.All) > 0 || len(c.Any) > 0 || len(c.None) > 0 {
		// Looks like it's another nested query. Use the Translate method, then.
		return c.Translate()
	}
	return elastic.NewTermQuery("user", "olivere"), nil
}

func launchClauseTranslators(clauses []*Clause, waitgroup *sync.WaitGroup, resultsChan chan elastic.Query, errChan chan error) {
	var innerwg sync.WaitGroup

	waitgroup.Add(1)

	for _, clause := range clauses {
		innerwg.Add(1)
		go func(clause *Clause, wg *sync.WaitGroup) {
			defer wg.Done()
			query, err := clause.ToQuery()
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
