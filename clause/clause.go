package clause

import (
	"gopkg.in/olivere/elastic.v5"
)

type ClauseType string

type ClauseProcessor func(args map[string]interface{}) (elastic.Query, error)
