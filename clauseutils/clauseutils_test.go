package clauseutils

import (
	"fmt"
	"reflect"
	"testing"

	"gopkg.in/olivere/elastic.v5"
)

func TestAddImplicitWildcard(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"foo", "*foo*"},
		{"foo bar", "*foo* *bar*"},
		{"foo\tbar", "*foo* *bar*"},
		{"foo\n \nbar", "*foo* *bar*"},
		{"foo OR bar", "*foo* *bar*"},
		{"\"foo bar\"", "\"foo bar\""},
		{"*foo OR bar", "*foo OR bar"},
		{"\\foo", "\\foo"},
		{"fo? OR x", "fo? OR x"},
		{"", ""},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			gotValue := AddImplicitWildcard(c.input)
			if gotValue != c.expected {
				t.Errorf("Got %q but expected %q", gotValue, c.expected)
			}
		})
	}
}

func TestAddImplicitUsernameWildcard(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"foo", "foo#*"},
		{"foo#iplant", "foo#iplant"},
		{"", "#*"},
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			gotValue := AddImplicitUsernameWildcard(c.input)
			if gotValue != c.expected {
				t.Errorf("Got %q but expected %q", gotValue, c.expected)
			}
		})
	}
}

func TestCreateRangeQuery(t *testing.T) {
	cases := []struct {
		field     string
		rangetype RangeType
		lower     int
		upper     int
		expected  elastic.Query
	}{
		{"meh", Both, 0, 10, elastic.NewRangeQuery("meh").Gte(0).Lte(10)},
		{"meh", LowerOnly, 0, 10, elastic.NewRangeQuery("meh").Gte(0)},
		{"meh", UpperOnly, 0, 10, elastic.NewRangeQuery("meh").Lte(10)},
		{"meh", LowerOnly, 0, 1000, elastic.NewRangeQuery("meh").Gte(0)},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s:%d-%d(%d)", c.field, c.lower, c.upper, c.rangetype), func(t *testing.T) {
			val := CreateRangeQuery(c.field, c.rangetype, c.lower, c.upper)
			source, err := val.Source()
			if err != nil {
				t.Error("Source get on created range query failed")
			}
			expsource, err := c.expected.Source()
			if err != nil {
				t.Error("Source get on expected range query failed")
			}
			if !reflect.DeepEqual(source, expsource) {
				t.Errorf("Value %+v and expected value %+v were not deeply equal", source, expsource)
			}
		})
	}
}
