package clauseutils

import (
	"testing"
)

func TestAddOrOperator(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"foo", "foo"},
		{"foo bar", "foo OR bar"},
		{"foo     ", "foo"},
		{"     foo", "foo"},
		{"  foo   ", "foo"},
		{"   foo bar ", "foo OR bar"},
		{"  foo  bar  ", "foo OR bar"},
		{" foo \n bar ", "foo OR bar"},
		{" foo \n\t\n bar ", "foo OR bar"},
		{" foo \nbar ", "foo OR bar"},
		{" foo\n bar ", "foo OR bar"},
		{" foo\n \nbar ", "foo OR bar"},
	}

	for _, c := range cases {
		gotValue := AddOrOperator(c.input)
		if gotValue != c.expected {
			t.Errorf("Got %q but expected %q", gotValue, c.expected)
		}
	}
}

func TestAddImplicitWildcard(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{"foo", "*foo*"},
		{"foo OR bar", "*foo* OR *bar*"},
		{"*foo OR bar", "*foo OR bar"},
		{"\\foo", "\\foo"},
		{"fo? OR x", "fo? OR x"},
	}

	for _, c := range cases {
		gotValue := AddImplicitWildcard(c.input)
		if gotValue != c.expected {
			t.Errorf("Got %q but expected %q", gotValue, c.expected)
		}
	}
}
