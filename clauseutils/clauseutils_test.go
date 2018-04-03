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

func TestDateToEpochMs(t *testing.T) {
	cases := []struct {
		input          string
		expectedOutput int64
		shouldErr      bool
	}{
		{"12345", 12345, false},
		{"2018-10-24T05:05:05.888-07:00", 1540382705888, false},
		{"not-a-date", 0, true},
		{"9223372036854775808", 0, true}, // bigger than int64
	}

	for _, c := range cases {
		t.Run(c.input, func(t *testing.T) {
			val, err := DateToEpochMs(c.input)
			if c.shouldErr && err == nil {
				t.Errorf("DateToEpochMs should have failed, instead returned %d", val)
			} else if !c.shouldErr && err != nil {
				t.Errorf("DateToEpochMs failed with error: %q", err)
			} else if !c.shouldErr && val != c.expectedOutput {
				t.Errorf("DateToEpochMs returned %d instead of expected %d", val, c.expectedOutput)
			}
		})
	}
}

func TestCreateRangeQuery(t *testing.T) {
	cases := []struct {
		field     string
		rangetype RangeType
		lower     int64
		upper     int64
		expected  elastic.Query
	}{
		{"meh", Both, 0, 10, elastic.NewRangeQuery("meh").Gte(int64(0)).Lte(int64(10))},
		{"meh", LowerOnly, 0, 10, elastic.NewRangeQuery("meh").Gte(int64(0))},
		{"meh", UpperOnly, 0, 10, elastic.NewRangeQuery("meh").Lte(int64(10))},
		{"meh", LowerOnly, 0, 1000, elastic.NewRangeQuery("meh").Gte(int64(0))},
		{"meh", Both, -3000000000, 3000000000, elastic.NewRangeQuery("meh").Gte(int64(-3000000000)).Lte(int64(3000000000))},
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
