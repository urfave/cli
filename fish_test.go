package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFishCompletion(t *testing.T) {
	// Given
	cmd := buildExtendedTestCommand()
	cmd.Flags = append(cmd.Flags,
		&StringFlag{
			Name:      "logfile",
			TakesFile: true,
		},
		&StringSliceFlag{
			Name:      "foofile",
			TakesFile: true,
		})
	cmd.setupCommandGraph()

	oldTemplate := FishCompletionTemplate
	defer func() { FishCompletionTemplate = oldTemplate }()
	FishCompletionTemplate = "{{something"

	// test error case
	_, err1 := cmd.ToFishCompletion()
	assert.Error(t, err1)

	// reset the template
	FishCompletionTemplate = oldTemplate
	// When
	res, err := cmd.ToFishCompletion()

	// Then
	require.NoError(t, err)
	expectFileContent(t, "testdata/expected-fish-full.fish", res)
}
