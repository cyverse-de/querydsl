package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/cyverse-de/querydsl"
	"github.com/cyverse-de/querydsl/clause"
	"github.com/cyverse-de/querydsl/clause/label"
	"gopkg.in/olivere/elastic.v5"
)

func PrintDocumentation() error {
	tmpl, err := template.New("documentation").Parse(`Available clause types:
{{ range $k, $v := . }}{{ $k }}: {{ if $v.Summary }}{{ $v.Summary }}{{ else }}(no summary provided){{end}}
    Arguments:{{ if $v.Args }}{{ range $ak, $av := $v.Args }}
        {{ $ak }} ({{ $av.Type }}): {{ $av.Summary }}
{{ end }}{{ else }} (no arguments)
{{ end }}{{ end }}`)
	if err != nil {
		return err
	}

	err = tmpl.Execute(os.Stdout, querydsl.GetDocumentation())
	return err
}

func main() {
	querydsl.AddClauseType("foo", func(args map[string]interface{}) (elastic.Query, error) {
		return elastic.NewTermQuery("user", "olivere"), nil
	}, clause.ClauseDocumentation{})
	label.Register()

	err := PrintDocumentation()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var jsonBlob = []byte(`{
		"all": [{"type": "foo", "args": {}}],
		"any": [{"all": [], "any": [{"type": "foo", "args": {}}], "none": []}],
		"none": [{"type": "label", "args": {"label": "foo"}}]
	}`)
	var query querydsl.Query
	err = json.Unmarshal(jsonBlob, &query)
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
