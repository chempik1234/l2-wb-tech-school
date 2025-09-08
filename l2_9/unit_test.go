package l2_9

import (
	"testing"
)

type testCase struct {
	input          string
	expectedOutput string
	expectsError   bool
}

var basicStrings = []testCase{
	{"a4bc2d5e", "aaaabccddddde", false},
	{"abcd", "abcd", false},
	{"45", "", true},
	{"", "", false},
}

var escapingStrings = []testCase{
	{"qwe\\4\\5", "qwe45", false},
	{"qwe\\45", "qwe44444", false},
}

var complexStrings = []testCase{
	{"\\", "", true},
	{"\\45", "44444", false},
	{"asbc\\", "", true},
	{"09a12", "", true},
	{"\\11\\22", "122", false},
	{"a12\\a", "aaaaaaaaaaaaa", false},
}

func runTests(t *testing.T, testCases []testCase) {
	var result string
	var err error
	for _, v := range testCases {
		result, err = Unpack(v.input)
		if err != nil && !v.expectsError {
			t.Errorf("unexpected error (input: %s): %v", v.input, err)
			continue
		} else if err == nil && v.expectsError {
			t.Error("expected error, but didn't get one")
			continue
		}

		if result != v.expectedOutput {
			t.Errorf("unexpected result:\t%s\t(expected: %s, input: %s)", result, v.expectedOutput, v.input)
		}
	}
}

func TestBasic(t *testing.T) {
	runTests(t, basicStrings)
}

func TestEscaping(t *testing.T) {
	runTests(t, escapingStrings)
}

func TestComplex(t *testing.T) {
	runTests(t, complexStrings)
}
