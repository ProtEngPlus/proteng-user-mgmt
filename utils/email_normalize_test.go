package utils

import (
	"testing"
)

func TestNormalizeEmail(t *testing.T) {
	cases := []struct {
		name     string
		input    string
		expected string
	}{
		{"mixed case", "Foo@X.com", "foo@x.com"},
		{"leading/trailing space", "  foo@x.com  ", "foo@x.com"},
		{"already normalized", "foo@x.com", "foo@x.com"},
		{"upper case whole thing", "FOO@X.COM", "foo@x.com"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NormalizeEmail(tc.input)
			if got != tc.expected {
				t.Errorf("NormalizeEmail(%q) = %q, want %q", tc.input, got, tc.expected)
			}
		})
	}
}
