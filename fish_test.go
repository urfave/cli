package cli

import (
	"testing"

	itesting "github.com/urfave/cli/v3/internal/testing"
)

func TestFishCompletion(t *testing.T) {
	// Given
	cmd := buildExtendedTestCommand()
	cmd.Flags = append(cmd.Flags, &StringFlag{
		Name:      "logfile",
		TakesFile: true,
	})

	// When
	res, err := cmd.ToFishCompletion()

	// Then
	itesting.RequireNoError(t, err)
	expectFileContent(t, "testdata/expected-fish-full.fish", res)
}
