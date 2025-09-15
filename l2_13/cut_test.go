package main

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

type testCase struct {
	name        string
	input       string
	args        []string
	expected    string
	shouldError bool
}

// runCutTest launches a test with given args
//
// 1. prepare command and readable stdin/stdout
//
// 2. exec
//
// 3. check expected stdout (no stderr)
func runCutTest(t *testing.T, input string, args []string, expected string) {
	//region  step 1. prepare command and readable stdin/stdout
	cmd := exec.Command("./cut.exe", args...)

	cmd.Stdin = strings.NewReader(input)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	//endregion

	//region  step 2. exec
	err := cmd.Run()
	if err != nil {
		t.Errorf("cmd exec error: %v, stderr: '%s'", err, stderr.String())
		return
	}
	//endregion

	//region  step 3. check stdout (no stderr)
	actual := stdout.String()
	if actual != expected {
		t.Errorf("expected: %q, got: %q", expected, actual)
	}
	//endregion
}

// TestBasicFunctionality has happy-way tests with inline testCases
func TestBasicFunctionality(t *testing.T) {
	tests := []testCase{
		{
			name:     "Basic field selection",
			input:    "aaa:bbb:ccc\n1:2:3\n",
			args:     []string{"-d", ":", "-f", "2"},
			expected: "bbb\n2\n",
		},
		{
			name:     "Multiple fields",
			input:    "aaa:bbb:cc:ddd\n1:2:3:4\n",
			args:     []string{"-d", ":", "-f", "1,3"},
			expected: "aaa:cc\n1:3\n",
		},
		{
			name:     "Suppress lines without delimiter",
			input:    "ab:bc:cd\nno_delimiter\n1:2:3\n",
			args:     []string{"-d", ":", "-f", "2", "-s"},
			expected: "bc\n2\n",
		},
		{
			name:     "Range of fields",
			input:    "a:b:c:d:e\n1:2:3:4:5\n",
			args:     []string{"-d", ":", "-f", "2-4"},
			expected: "b:c:d\n2:3:4\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runCutTest(t, tt.input, tt.args, tt.expected)
		})
	}
}

// TestEdgeCases edge inline tests: empty, out-of-range,
func TestEdgeCases(t *testing.T) {
	tests := []testCase{
		{
			name:     "Empty input",
			input:    "",
			args:     []string{"-d", ":", "-f", "1"},
			expected: "",
		},
		{
			name:     "Field out of range",
			input:    "a:b:c\n",
			args:     []string{"-d", ":", "-f", "5"},
			expected: "\n",
		},
		{
			name:     "Only delimiter characters",
			input:    ":::\n::::\n",
			args:     []string{"-d", ":", "-f", "2"},
			expected: "\n\n",
		},
		{
			name:     "Only delimiter characters with -s",
			input:    ":::\n::::\n",
			args:     []string{"-d", ":", "-f", "2", "-s"},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runCutTest(t, tt.input, tt.args, tt.expected)
		})
	}
}

// TestErrorCases tests with inline error-producing cases
func TestErrorCases(t *testing.T) {
	tests := []testCase{
		{
			name:        "No field specified",
			input:       "a:b:c\n",
			args:        []string{"-d", ":"},
			shouldError: true,
		},
		{
			name:        "Invalid field specification",
			input:       "a:b:c\n",
			args:        []string{"-d", ":", "-f", "abc"},
			shouldError: true,
		},
		{
			name:        "No delimiter specified",
			input:       "a:b:c\n",
			args:        []string{"-f", "1"},
			shouldError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command("./main", tt.args...)
			cmd.Stdin = strings.NewReader(tt.input)
			var stderr bytes.Buffer
			cmd.Stderr = &stderr

			err := cmd.Run()
			if tt.shouldError && err == nil {
				t.Errorf("expected err, got err = nil")
			}
			if !tt.shouldError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
