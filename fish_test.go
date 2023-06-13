package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
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
	require.NoError(t, err)
	expectFileContent(t, "testdata/expected-fish-full.fish", res)
}
