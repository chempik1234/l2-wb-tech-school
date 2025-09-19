package main

import "testing"

func TestParseFields(t *testing.T) {
	cases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "cut -d \" \" -f1",
			expected: []string{"cut", "-d", " ", "-f1"},
		},
	}
	for _, c := range cases {
		result := parseFields(c.input)
		if len(result) != len(c.expected) {
			t.Errorf("parseFields(%q): expected %v, got %v", c.input, c.expected, result)
		}
		for i := range result {
			if result[i] != c.expected[i] {
				t.Errorf("parseFields(%q): expected %v, got %v", c.input, c.expected, result)
				break
			}
		}
	}
}
