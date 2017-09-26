package label

import (
	"testing"
)

func TestLabelProcessor(t *testing.T) {
	cases := []struct {
		label         string
		exact         string // "true", "false", or "nil" to not set
		expectedQuery string
	}{
		{"foo bar", "nil", "*foo* OR *bar*"},
		{"foo bar", "false", "*foo* OR *bar*"},
		{"foo bar", "true", "foo bar"},
	}

	for _, c := range cases {
		args := make(map[string]interface{})

		args["label"] = c.label
		if c.exact == "true" {
			args["exact"] = true
		} else if c.exact == "false" {
			args["exact"] = false
		} else if c.exact != "nil" {
			t.Fatal("'exact' in a case was not set to one of 'true', 'false', or 'nil'")
		}

		query, err := LabelProcessor(args)
		if err != nil {
			t.Errorf("LabelProcessor failed with error: %q", err)
		}
		source, err := query.Source()
		if err != nil {
			t.Errorf("Source get failed with error: %q", err)
		}

		qsQuery, ok := source.(map[string]interface{})["query_string"]
		if !ok {
			t.Error("Source did not contain 'query_string'")
		}

		fields, ok := qsQuery.(map[string]interface{})["fields"]
		if !ok {
			t.Error("query did not contain 'fields'")
		}
		field, ok := fields.([]string)
		if !ok {
			t.Error("fields were not array")
		}

		if field[0] != "label" {
			t.Error("Specified field was not 'label'")
		}

		stringQuery, ok := qsQuery.(map[string]interface{})["query"]
		if !ok {
			t.Error("query did not contain 'query'")
		}

		if stringQuery.(string) != c.expectedQuery {
			t.Errorf("query %q did not match expected value %q", stringQuery, c.expectedQuery)
		}
	}
}
