package cli

import (
	"testing"

	itesting "github.com/urfave/cli/v3/internal/testing"
)

var lexicographicLessTests = []struct {
	i        string
	j        string
	expected bool
}{
	{"", "a", true},
	{"a", "", false},
	{"a", "a", false},
	{"a", "A", false},
	{"A", "a", true},
	{"aa", "a", false},
	{"a", "aa", true},
	{"a", "b", true},
	{"a", "B", true},
	{"A", "b", true},
	{"A", "B", true},
}

func TestLexicographicLess(t *testing.T) {
	for _, test := range lexicographicLessTests {
		actual := lexicographicLess(test.i, test.j)
		itesting.Equal(t, test.expected, actual)
	}
}
