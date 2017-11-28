package path

import (
	"context"
	"fmt"
	"testing"
)

func TestPathProcessor(t *testing.T) {
	cases := []struct {
		prefix        interface{}
		expectedQuery string
		shouldErr     bool
	}{
		{prefix: "/iplant/home/foo", expectedQuery: "/iplant/home/foo"},
		{shouldErr: true},              // empty prefix
		{prefix: 444, shouldErr: true}, // bad type
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%T(%+v)", c.prefix, c.prefix), func(t *testing.T) {
			args := make(map[string]interface{})

			args["prefix"] = c.prefix

			query, err := PathProcessor(context.Background(), args)
			if c.shouldErr && err == nil {
				t.Errorf("PathProcessor should have failed, instead returned nil error and query %+v", query)
			} else if !c.shouldErr && err != nil {
				t.Errorf("PathProcessor failed with error: %q", err)
			} else if !c.shouldErr {
				source, err := query.Source()
				if err != nil {
					t.Errorf("Source get failed with error: %q", err)
				}

				pQuery, ok := source.(map[string]interface{})["prefix"]
				if !ok {
					t.Error("Source did not contain 'prefix'")
				}

				pathQuery, ok := pQuery.(map[string]interface{})["path"]
				if !ok {
					t.Error("query did not contain 'path'")
				}

				if pathQuery.(string) != c.expectedQuery {
					t.Errorf("query %q did not match expected value %q", pathQuery, c.expectedQuery)
				}
			}
		})
	}
}
