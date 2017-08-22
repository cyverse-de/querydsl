// Package querydsl provides programmatic mapping from the CyVerse search DSL to Elasticsearch queries
//package querydsl
package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/olivere/elastic.v5"
	"sync"
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

// Translate turns a Query into an elastic.Query by way of translating everything contained within
func (q *Query) Translate() (elastic.Query, error) {
	baseQuery := elastic.NewBoolQuery()

	// allwg, anywg, and nonewg track whether everything within each of the three respective lists has been processed
	// their associated channels send processed queries back
	var allwg sync.WaitGroup
	allChan := make(chan elastic.Query, 10)

	var anywg sync.WaitGroup
	anyChan := make(chan elastic.Query, 10)

	var nonewg sync.WaitGroup
	noneChan := make(chan elastic.Query, 10)

	// subpartswg tracks whether all three of the other waitgroups have completed
	var subpartswg sync.WaitGroup

	// errChan is used by everything to propagate errors
	errChan := make(chan error)

	for _, clause := range q.All {
		allwg.Add(1)
		go func(clause *Clause) {
			defer allwg.Done()
			query, err := clause.ToQuery()
			if err != nil {
				errChan <- err
			} else {
				allChan <- query
			}
		}(clause)
	}

	for _, clause := range q.Any {
		anywg.Add(1)
		go func(clause *Clause) {
			defer anywg.Done()
			query, err := clause.ToQuery()
			if err != nil {
				errChan <- err
			} else {
				anyChan <- query
			}
		}(clause)
	}

	for _, clause := range q.None {
		nonewg.Add(1)
		go func(clause *Clause) {
			defer nonewg.Done()
			query, err := clause.ToQuery()
			if err != nil {
				errChan <- err
			} else {
				noneChan <- query
			}
		}(clause)
	}

	subpartswg.Add(3)

	go func() {
		allwg.Wait()
		subpartswg.Done()
	}()
	go func() {
		anywg.Wait()
		subpartswg.Done()
	}()
	go func() {
		nonewg.Wait()
		subpartswg.Done()
	}()

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
	translatedJson, err := json.Marshal(querySource)
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Printf("%s\n", translatedJson)
}
