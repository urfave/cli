package genflags_test

import (
	"fmt"
	"testing"

	"github.com/urfave/cli/v2/internal/genflags"
)

func TestTypeName(t *testing.T) {
	for _, tc := range []struct {
		gt       string
		fc       *genflags.FlagTypeConfig
		expected string
	}{
		{gt: "int", fc: nil, expected: "IntFlag"},
		{gt: "int", fc: &genflags.FlagTypeConfig{}, expected: "IntFlag"},
		{gt: "int", fc: &genflags.FlagTypeConfig{TypeName: "VeryIntyFlag"}, expected: "VeryIntyFlag"},
		{gt: "[]bool", fc: nil, expected: "BoolSliceFlag"},
		{gt: "[]bool", fc: &genflags.FlagTypeConfig{}, expected: "BoolSliceFlag"},
		{gt: "[]bool", fc: &genflags.FlagTypeConfig{TypeName: "ManyTruthsFlag"}, expected: "ManyTruthsFlag"},
		{gt: "time.Rumination", fc: nil, expected: "RuminationFlag"},
		{gt: "time.Rumination", fc: &genflags.FlagTypeConfig{}, expected: "RuminationFlag"},
		{gt: "time.Rumination", fc: &genflags.FlagTypeConfig{TypeName: "PonderFlag"}, expected: "PonderFlag"},
	} {
		t.Run(
			fmt.Sprintf("type=%s,cfg=%v", tc.gt, func() string {
				if tc.fc != nil {
					return tc.fc.TypeName
				}
				return "nil"
			}()),
			func(ct *testing.T) {
				actual := genflags.TypeName(tc.gt, tc.fc)
				if tc.expected != actual {
					ct.Errorf("expected %q, got %q", tc.expected, actual)
				}
			},
		)
	}
}
