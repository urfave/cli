//go:build !urfave_cli_no_template

package cli

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFishCompletionTemplateError(t *testing.T) {
	defer func(old string) { FishCompletionTemplate = old }(FishCompletionTemplate)
	FishCompletionTemplate = "{{something"

	_, err := buildExtendedTestCommand().ToFishCompletion()
	assert.Error(t, err)
}
