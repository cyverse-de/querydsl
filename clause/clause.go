package clause

import (
	"context"

	"gopkg.in/olivere/elastic.v5"
)

// ClauseType is an alias for string used as the key for locating processors and documentation
type ClauseType string

// ClauseProcessor is a function taking a context and arguments for a given clause type and producing a Query
type ClauseProcessor func(ctx context.Context, args map[string]interface{}) (elastic.Query, error)

// ClauseArgumentDocumentation describes a single argument for a clause. The 'type' should look like a golang type, though this is not checked.
type ClauseArgumentDocumentation struct {
	Type    string `json:"type"`
	Summary string `json:"summary"`
}

// ClauseDocumentation describes a clause with an overall summary plus documentation of each argument.
type ClauseDocumentation struct {
	Summary string                                 `json:"summary"`
	Args    map[string]ClauseArgumentDocumentation `json:"args"`
}
