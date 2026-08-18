package cli

import (
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTextFlagSetFromArg(t *testing.T) {
	lv := &slog.LevelVar{}
	cmd := &Command{
		Flags: []Flag{
			&TextFlag{Name: "level", Value: lv},
		},
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"", "--level", "WARN"}))
	assert.Equal(t, slog.LevelWarn, lv.Level())
}

func TestTextFlagDefaultValue(t *testing.T) {
	lv := &slog.LevelVar{}
	lv.Set(slog.LevelError)
	cmd := &Command{
		Flags: []Flag{
			&TextFlag{Name: "level", Value: lv},
		},
	}

	// Without the flag being passed, the value keeps its default.
	require.NoError(t, cmd.Run(buildTestContext(t), []string{""}))
	assert.Equal(t, slog.LevelError, lv.Level())
}

func TestTextFlagDestination(t *testing.T) {
	lv := &slog.LevelVar{}
	var dest TextMarshalUnmarshaler = lv
	cmd := &Command{
		Flags: []Flag{
			&TextFlag{Name: "level", Destination: &dest},
		},
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"", "--level", "DEBUG"}))
	assert.Equal(t, slog.LevelDebug, lv.Level())
}

func TestTextFlagTrimSpace(t *testing.T) {
	lv := &slog.LevelVar{}
	cmd := &Command{
		Flags: []Flag{
			&TextFlag{Name: "level", Value: lv, Config: StringConfig{TrimSpace: true}},
		},
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{"", "--level", "  INFO  "}))
	assert.Equal(t, slog.LevelInfo, lv.Level())
}

func TestTextFlagFromEnvSource(t *testing.T) {
	t.Setenv("LOG_LEVEL", "ERROR")
	lv := &slog.LevelVar{}
	cmd := &Command{
		Flags: []Flag{
			&TextFlag{Name: "level", Value: lv, Sources: EnvVars("LOG_LEVEL")},
		},
	}

	require.NoError(t, cmd.Run(buildTestContext(t), []string{""}))
	assert.Equal(t, slog.LevelError, lv.Level())
}

func TestTextFlagInvalidValue(t *testing.T) {
	lv := &slog.LevelVar{}
	cmd := &Command{
		Flags: []Flag{
			&TextFlag{Name: "level", Value: lv},
		},
	}

	err := cmd.Run(buildTestContext(t), []string{"", "--level", "NOPE"})
	require.Error(t, err)
}

func TestTextFlagValueFromCommand(t *testing.T) {
	lv := &slog.LevelVar{}
	f := &TextFlag{Name: "level", Value: lv}
	cmd := &Command{
		Flags: []Flag{f},
	}

	require.NoError(t, cmd.Set("level", "WARN"))
	require.Equal(t, lv, cmd.Text(f.Name))
}
