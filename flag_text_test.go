package cli

import (
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// errText is a TextMarshalUnmarshaler whose MarshalText always fails, used to
// exercise the error branches of textValue.ToString and textValue.String.
type errText struct{}

func (errText) MarshalText() ([]byte, error) { return nil, errors.New("marshal boom") }

func (errText) UnmarshalText([]byte) error { return nil }

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

func TestTextFlagTextNotAvailable(t *testing.T) {
	cmd := &Command{
		Flags: []Flag{
			&StringFlag{Name: "str"},
		},
	}

	// The flag exists but its value is a string, not a
	// TextMarshalUnmarshaler, so Text returns nil.
	require.NoError(t, cmd.Set("str", "value"))
	assert.Nil(t, cmd.Text("str"))

	// An unknown flag name also returns nil.
	assert.Nil(t, cmd.Text("missing"))
}

func TestTextValueToString(t *testing.T) {
	var tv textValue

	// A nil value marshals to the empty string.
	assert.Equal(t, "", tv.ToString(nil))

	// A MarshalText error is swallowed and yields the empty string.
	assert.Equal(t, "", tv.ToString(errText{}))

	// A value that marshals cleanly is rendered.
	lv := &slog.LevelVar{}
	lv.Set(slog.LevelWarn)
	assert.Equal(t, "WARN", tv.ToString(lv))
}

func TestTextValueSetNilDestination(t *testing.T) {
	// Set is a no-op when there is no destination to unmarshal into.
	tv := &textValue{}
	require.NoError(t, tv.Set("anything"))
}

func TestTextValueString(t *testing.T) {
	// A nil destination stringifies to the empty string.
	tv := &textValue{}
	assert.Equal(t, "", tv.String())

	// A MarshalText error is swallowed and yields the empty string.
	tvErr := &textValue{destination: errText{}}
	assert.Equal(t, "", tvErr.String())

	// A destination that marshals cleanly is rendered.
	lv := &slog.LevelVar{}
	lv.Set(slog.LevelWarn)
	tvOK := &textValue{destination: lv}
	assert.Equal(t, "WARN", tvOK.String())
}
