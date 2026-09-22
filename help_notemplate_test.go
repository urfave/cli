//go:build urfave_cli_no_template

package cli

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNoTemplate_CustomHelpTemplate(t *testing.T) {
	var out bytes.Buffer
	cmd := &Command{
		Name:                          "foo",
		Writer:                        &out,
		CustomRootCommandHelpTemplate: "{{.Name}}",
	}

	require.NoError(t, cmd.Run(context.Background(), []string{"foo", "--help"}))
	assert.Empty(t, out.String())
}

func TestNoTemplate_CustomFishCompletionTemplate(t *testing.T) {
	defer func(old string) { FishCompletionTemplate = old }(FishCompletionTemplate)
	FishCompletionTemplate = "{{.Command.Name}}"

	_, err := (&Command{Name: "foo"}).ToFishCompletion()
	assert.ErrorIs(t, err, errNoTemplate)
}
