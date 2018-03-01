package metadata

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/mitchellh/mapstructure"
)

type MakeNested struct {
	Nested struct {
		Path      string
		ScoreMode string `mapstructure:"score_mode"`
		Query     struct {
			Bool struct {
				Must interface{}
			}
		}
	}
}

type QueryStringQ struct {
	QueryString struct {
		Query  string
		Fields []string
	} `mapstructure:"query_string"`
}

func TestNested(t *testing.T) {
	cases := []struct {
		attribute string
		value     string
		unit      string
	}{
		{attribute: "testa"},
		{value: "testv"},
		{unit: "testu"},
		{attribute: "testav-a", value: "testav-v"},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%+v,%+v,%+v", c.attribute, c.value, c.unit), func(t *testing.T) {
			v := makeNested(c.attribute, c.value, c.unit)

			source, err := v.Source()
			if err != nil {
				t.Errorf("Source get failed with error: %q", err)
			}

			var decoded MakeNested
			err = mapstructure.Decode(source, &decoded)

			if err != nil {
				t.Errorf("Decode failed with error: %q", err)
			}

			if decoded.Nested.Path != "metadata" {
				t.Error("Nested path is not 'metadata'")
			}

			if c.attribute != "" && c.value != "" {
				mq, ok := decoded.Nested.Query.Bool.Must.([]interface{})
				if !ok {
					t.Error("Could not grab an array from 'must'")
				}

				var first QueryStringQ
				err = mapstructure.Decode(mq[0], &first)
				if err != nil {
					t.Errorf("Decode failed with error: %q", err)
				}

				if first.QueryString.Fields[0] != "metadata.attribute" {
					t.Error("First nested query had wrong field")
				}

				if first.QueryString.Query != c.attribute {
					t.Error("First nested query had wrong query string")
				}

				var second QueryStringQ
				err = mapstructure.Decode(mq[1], &second)
				if err != nil {
					t.Errorf("Decode failed with error: %q", err)
				}

				if second.QueryString.Fields[0] != "metadata.value" {
					t.Error("Second nested query had wrong field")
				}

				if second.QueryString.Query != c.value {
					t.Error("Second nested query had wrong query string")
				}
			}
		})
	}
}

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
		{attribute: "test", metadataTypes: []string{"cyverse"}, attributeExact: "nil", valueExact: "nil", unitExact: "nil"},
		{attribute: "test", metadataTypes: []string{"irods"}, attributeExact: "nil", valueExact: "nil", unitExact: "nil"},
		{attribute: "test", metadataTypes: []string{"cyverse", "irods"}, attributeExact: "nil", valueExact: "nil", unitExact: "nil"},
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
				source, err := query.Source()
				if err != nil {
					t.Errorf("Source get failed with error: %q", err)
				}

				boolQuery, ok := source.(map[string]interface{})["bool"]
				if !ok {
					t.Error("Source did not contain 'bool'")
				}

				should, ok := boolQuery.(map[string]interface{})["should"]
				if !ok {
					t.Error("Bool query did not contain 'should'")
				}

				includeCyverse := len(c.metadataTypes) == 0 || len(c.metadataTypes) == 2 || c.metadataTypes[0] == "cyverse"
				includeIrods := len(c.metadataTypes) == 0 || len(c.metadataTypes) == 2 || c.metadataTypes[0] == "irods"

				if includeCyverse {
					shouldArr, ok := should.([]interface{})
					if !ok {
						t.Error("'should' was not an array, but should have been")
					}

					if includeIrods {
						if len(shouldArr) != 3 {
							t.Error("'should' should have three entries")
						}
					} else {
						if len(shouldArr) != 2 {
							t.Error("'should' should have two entries")
						}
					}
				} else if includeIrods {
					_, ok := should.([]interface{})
					if ok {
						t.Error("'should' was an array, but should not have been")
					}

				}

				// If the numbers look right, we'll figure that the query for the nested query creation is doing right due to the tests above.
			}
		})
	}
}
