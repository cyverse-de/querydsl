// Package querydsl provides programmatic mapping from the CyVerse search DSL to Elasticsearch queries
package querydsl

import (
	"context"
	"fmt"
	"sync"

	"github.com/cyverse-de/querydsl/clause"
	"gopkg.in/olivere/elastic.v5"
)

// QueryDSL represents a collection of processors and their documentation to use for translation
type QueryDSL struct {
	clauseProcessors    map[clause.ClauseType]clause.ClauseProcessor
	clauseDocumentation map[clause.ClauseType]clause.ClauseDocumentation
}

// Query represents a boolean query
type Query struct {
	All  []*GenericClause `json:"all,omitempty"`
	Any  []*GenericClause `json:"any,omitempty"`
	None []*GenericClause `json:"none,omitempty"`
}

// Clause represents a particular clause
type Clause struct {
	Type clause.ClauseType      `json:"type,omitempty"`
	Args map[string]interface{} `json:"args,omitempty"`
}

// GenericClause embeds both Query and Clause to represent the fact a clause
// can be a nested query.
type GenericClause struct {
	*Clause
	*Query
}

// IsQuery checks if a GenericClause has a valid Query part
func (c *GenericClause) IsQuery() bool {
	return c.Query != nil && (len(c.All) > 0 || len(c.Any) > 0 || len(c.None) > 0)
}

// IsClause checks if a GenericClause has a valid Clause part
func (c *GenericClause) IsClause() bool {
	return c.Clause != nil && len(c.Type) > 0
}

// Translate turns a GenericClause into an elastic.Query
func (c *GenericClause) Translate(ctx context.Context, qd *QueryDSL) (elastic.Query, error) {
	if c.IsQuery() {
		// Looks like it's another nested query.
		query := Query{All: c.All, Any: c.Any, None: c.None}
		return query.Translate(ctx, qd)
	} else if c.IsClause() {
		clause := Clause{Type: c.Type, Args: c.Args}
		return clause.Translate(ctx, qd)
	} else {
		return nil, fmt.Errorf("GenericClause %+v is neither a properly-formatted Query nor a Clause", c)
	}
}

// Translate turns a regular Clause into an elastic.Query
func (c *Clause) Translate(ctx context.Context, qd *QueryDSL) (elastic.Query, error) {
	clauseProcessors := qd.GetProcessors()
	if processor, exists := clauseProcessors[c.Type]; exists {
		return processor(ctx, c.Args)
	}
	return nil, fmt.Errorf("No processor found for type '%s'", c.Type)
}

// launchClauseTranslators launches a set of goroutines to translate a set of Clauses
// internally, it uses a WaitGroup to track when all of the goroutines it
// launched are finished, and once they are signals to the WaitGroup passed as
// an argument; this way if several calls to this function all pass the same
// WaitGroup, that WaitGroup only shows up finished when every clause across
// all of the several calls is processed
//
// This long comment brought to you by the author not wanting to forget how this works
func launchClauseTranslators(ctx context.Context, qd *QueryDSL, clauses []*GenericClause, waitgroup *sync.WaitGroup, resultsChan chan elastic.Query, errChan chan error) {
	var innerwg sync.WaitGroup

	waitgroup.Add(1)

	for _, clause := range clauses {
		innerwg.Add(1)
		go func(clause *GenericClause, wg *sync.WaitGroup) {
			defer wg.Done()
			query, err := clause.Translate(ctx, qd)
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
func (q *Query) Translate(ctx context.Context, qd *QueryDSL) (elastic.Query, error) {
	baseQuery := elastic.NewBoolQuery()

	// Result channels
	allChan := make(chan elastic.Query)
	anyChan := make(chan elastic.Query)
	noneChan := make(chan elastic.Query)

	// subpartswg tracks whether all three of the other waitgroups have completed
	var subpartswg sync.WaitGroup

	// errChan is used by everything to propagate errors
	errChan := make(chan error)

	launchClauseTranslators(ctx, qd, q.All, &subpartswg, allChan, errChan)
	launchClauseTranslators(ctx, qd, q.Any, &subpartswg, anyChan, errChan)
	launchClauseTranslators(ctx, qd, q.None, &subpartswg, noneChan, errChan)

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
			baseQuery.Should(query).MinimumNumberShouldMatch(1)
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

// New creates a new empty QueryDSL
func New() *QueryDSL {
	processors := make(map[clause.ClauseType]clause.ClauseProcessor)
	documentation := make(map[clause.ClauseType]clause.ClauseDocumentation)
	return &QueryDSL{clauseProcessors: processors, clauseDocumentation: documentation}
}

// AddClauseType takes a string (as clause.ClauseType), a function to process,
// and documentation, and registers them for use by the other functions in
// querydsl
func (qd *QueryDSL) AddClauseType(clausetype clause.ClauseType, processor clause.ClauseProcessor, documentation clause.ClauseDocumentation) {
	qd.clauseProcessors[clausetype] = processor
	qd.clauseDocumentation[clausetype] = documentation
}

// GetProcessors returns all the clause processors registered to a QueryDSL
func (qd *QueryDSL) GetProcessors() map[clause.ClauseType]clause.ClauseProcessor {
	return qd.clauseProcessors
}

// GetDocumentation returns documentation (if present) for all the clause processors registered to a QueryDSL
func (qd *QueryDSL) GetDocumentation() map[clause.ClauseType]clause.ClauseDocumentation {
	return qd.clauseDocumentation
}
