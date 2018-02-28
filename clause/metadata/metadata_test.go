package metadata

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

func TestMetadataProcessor(t *testing.T) {
	cases := []struct {
		attribute      interface{}
		attributeExact string // "true", "false", or "nil" to not set
		value          interface{}
		valueExact     string // "true", "false", or "nil" to not set
		unit           interface{}
		unitExact      string // "true", "false", or "nil" to not set

		metadataTypes []string

		expectedQuery string
		shouldErr     bool
	}{
		//{label: "foo bar", exact: "nil", expectedQuery: "*foo* *bar*"},
		//{label: "foo bar", exact: "false", expectedQuery: "*foo* *bar*"},
		//{label: "foo bar", exact: "true", expectedQuery: "foo bar"},
		//{label: "\"foo bar\"", exact: "false", expectedQuery: "\"foo bar\""},
		{attribute: "test", attributeExact: "nil", valueExact: "nil", unitExact: "nil"},
		{attribute: "test", metadataTypes: []string{"invalid"}, attributeExact: "nil", valueExact: "nil", unitExact: "nil", shouldErr: true}, // invalid type
		{attributeExact: "nil", valueExact: "nil", unitExact: "nil", shouldErr: true},                                                        // empty attr/value/unit
		{attribute: 444, attributeExact: "nil", valueExact: "nil", unitExact: "nil", shouldErr: true},                                        // bad type
		{value: 444, attributeExact: "nil", valueExact: "nil", unitExact: "nil", shouldErr: true},                                            // bad type
		{unit: 444, attributeExact: "nil", valueExact: "nil", unitExact: "nil", shouldErr: true},                                             // bad type
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%T(%+v)e:%s-%T(%+v)e:%s-%T(%+v)e:%s-t:[%s]", c.attribute, c.attribute, c.attributeExact, c.value, c.value, c.valueExact, c.unit, c.unit, c.unitExact, strings.Join(c.metadataTypes, ",")), func(t *testing.T) {
			args := make(map[string]interface{})

			args["attribute"] = c.attribute
			args["value"] = c.value
			args["unit"] = c.unit
			args["metadata_types"] = c.metadataTypes

			if c.attributeExact == "true" {
				args["attribute_exact"] = true
			} else if c.attributeExact == "false" {
				args["attribute_exact"] = false
			} else if c.attributeExact != "nil" {
				t.Fatal("'attributeExact' in a case was not set to one of 'true', 'false', or 'nil'")
			}

			if c.valueExact == "true" {
				args["value_exact"] = true
			} else if c.valueExact == "false" {
				args["value_exact"] = false
			} else if c.valueExact != "nil" {
				t.Fatal("'valueExact' in a case was not set to one of 'true', 'false', or 'nil'")
			}

			if c.unitExact == "true" {
				args["unit_exact"] = true
			} else if c.unitExact == "false" {
				args["unit_exact"] = false
			} else if c.unitExact != "nil" {
				t.Fatal("'unitExact' in a case was not set to one of 'true', 'false', or 'nil'")
			}

			query, err := MetadataProcessor(context.Background(), args)
			if c.shouldErr && err == nil {
				t.Errorf("MetadataProcessor should have failed, instead returned nil error and query %+v", query)
			} else if !c.shouldErr && err != nil {
				t.Errorf("MetadataProcessor failed with error: %q", err)
			} else if !c.shouldErr {
				_, err := query.Source()
				if err != nil {
					t.Errorf("Source get failed with error: %q", err)
				}

				//qsQuery, ok := source.(map[string]interface{})["query_string"]
				//if !ok {
				//	t.Error("Source did not contain 'query_string'")
				//}

				//fields, ok := qsQuery.(map[string]interface{})["fields"]
				//if !ok {
				//	t.Error("query did not contain 'fields'")
				//}
				//field, ok := fields.([]string)
				//if !ok {
				//	t.Error("fields were not array")
				//}

				//if field[0] != "label" {
				//	t.Error("Specified field was not 'label'")
				//}

				//stringQuery, ok := qsQuery.(map[string]interface{})["query"]
				//if !ok {
				//	t.Error("query did not contain 'query'")
				//}

				//if stringQuery.(string) != c.expectedQuery {
				//	t.Errorf("query %q did not match expected value %q", stringQuery, c.expectedQuery)
				//}
			}
		})
	}
}
