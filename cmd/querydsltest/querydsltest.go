package main

import (
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/cyverse-de/querydsl"
	"github.com/cyverse-de/querydsl/clause/label"
	"github.com/cyverse-de/querydsl/clause/owner"
	"github.com/cyverse-de/querydsl/clause/path"
)

func PrintDocumentation(qd *querydsl.QueryDSL) error {
	tmpl, err := template.New("documentation").Parse(`Available clause types:
{{ range $k, $v := . }}{{ $k }}: {{ if $v.Summary }}{{ $v.Summary }}{{ else }}(no summary provided){{end}}
    Arguments:{{ if $v.Args }}{{ range $ak, $av := $v.Args }}
        {{ $ak }} ({{ $av.Type }}): {{ $av.Summary }}
{{ end }}{{ else }} (no arguments)
{{ end }}{{ end }}`)
	if err != nil {
		return err
	}

	err = tmpl.Execute(os.Stdout, qd.GetDocumentation())
	return err
}

func main() {
	qd := querydsl.New()
	label.Register(qd)
	path.Register(qd)
	owner.Register(qd)

	err := PrintDocumentation(qd)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var jsonBlob = []byte(`{
		"all": [{"type": "path", "args": {"prefix": "/iplant/home"}}, {"type": "label", "args": {"label": "PDAP.fel.tree"}}],
		"any": [{"type": "owner", "args": {"owner": "ipctest"}}]
	}`)
	var query querydsl.Query
	err = json.Unmarshal(jsonBlob, &query)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("%+v\n", query)
	translated, err := query.Translate(qd)
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
