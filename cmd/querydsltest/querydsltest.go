package main

import (
	"encoding/json"
	"fmt"

	"github.com/cyverse-de/querydsl"
	"github.com/cyverse-de/querydsl/clause/label"
	"gopkg.in/olivere/elastic.v5"
)

func main() {
	querydsl.AddClauseType("foo", func(args map[string]interface{}) (elastic.Query, error) {
		return elastic.NewTermQuery("user", "olivere"), nil
	}, nil)
	label.Register()
	var jsonBlob = []byte(`{
		"all": [{"type": "foo", "args": {}}],
		"any": [{"all": [], "any": [{"type": "foo", "args": {}}], "none": []}],
		"none": [{"type": "label", "args": {"label": "foo"}}]
	}`)
	var query querydsl.Query
	err := json.Unmarshal(jsonBlob, &query)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("%+v\n", query)
	translated, err := query.Translate()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("%s\n", translated)
	querySource, err := translated.Source()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("%s\n", querySource)
	translatedJSON, err := json.Marshal(querySource)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("%s\n", translatedJSON)
}
