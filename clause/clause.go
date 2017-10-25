package clause

import (
	"gopkg.in/olivere/elastic.v5"
)

type ClauseType string

type ClauseProcessor func(args map[string]interface{}) (elastic.Query, error)

type ClauseArgumentDocumentation struct {
	Type    string `json:"type"`
	Summary string `json:"summary"`
}

type ClauseDocumentation struct {
	Summary string                                 `json:"summary"`
	Args    map[string]ClauseArgumentDocumentation `json:"args"`
}
