package clauseutils

import (
	"testing"
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
