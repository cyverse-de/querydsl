package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"github.com/cyverse-de/querydsl/v2"
	"github.com/cyverse-de/querydsl/v2/clause/created"
	"github.com/cyverse-de/querydsl/v2/clause/label"
	"github.com/cyverse-de/querydsl/v2/clause/metadata"
	"github.com/cyverse-de/querydsl/v2/clause/modified"
	"github.com/cyverse-de/querydsl/v2/clause/owner"
	"github.com/cyverse-de/querydsl/v2/clause/path"
	"github.com/cyverse-de/querydsl/v2/clause/permissions"
	"github.com/cyverse-de/querydsl/v2/clause/size"
	"github.com/cyverse-de/querydsl/v2/clause/tag"
)

func printDocumentation(qd *querydsl.QueryDSL) error {
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
	permissions.Register(qd)
	metadata.Register(qd)
	tag.Register(qd)
	created.Register(qd)
	modified.Register(qd)
	size.Register(qd)

	err := printDocumentation(qd)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	var jsonBlob = []byte(`{
		"all": [{"type": "path", "args": {"prefix": "/iplant/home"}}, {"type": "label", "args": {"label": "PDAP.fel.tree"}}, {"type": "permissions", "args": {"users": ["mian", "ipctest#iplant", "foo#bar", "baz"], "permission": "write"}}, {"type": "size", "args": {"from": "1KB", "to": "  4.8 GB  "}}],
		"any": [{"type": "owner", "args": {"owner": "ipctest"}},{"type": "metadata", "args": {"attribute": "foo", "value": "bar", "attribute_exact": true}},{"type": "metadata", "args": {"attribute": "foo", "value": "bar", "attribute_exact": true, "value_exact": true, "metadata_types": ["irods"]}}, {"type": "tag", "args": {"tags": ["dummy-tag-value"]}}, {"type": "created", "args": {"from": "2017-09-23T00:00:00.000Z"}}, {"type": "modified", "args": {"to": "2017-09-23T00:00:00.000-07:00"}}],
		"none": [{"type": "permissions", "args": {"permission": "read", "users": ["mian#iplant", "ipctest#iplant"]}}]
	}`)
	var query querydsl.Query
	err = json.Unmarshal(jsonBlob, &query)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Printf("%+v\n", query)
	translated, err := query.Translate(context.Background(), qd)
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
