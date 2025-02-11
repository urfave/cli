package cli

import (
	"errors"
	"flag"
	"io"
	"log/slog"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type badMarshaller struct{}

func (_ badMarshaller) UnmarshalText(_ []byte) error {
	return nil
}

func (_ badMarshaller) MarshalText() ([]byte, error) {
	return nil, errors.New("bad")
}

func TestTextFlag(t *testing.T) {
	tests := []struct {
		name    string
		flag    TextFlag
		args    []string
		want    string
		wantErr bool
	}{
		{
			name: "empty",
			flag: TextFlag{
				Name:  "log-level",
				Value: &slog.LevelVar{},
			},
			want: "INFO",
		},
		{
			name: "info",
			flag: TextFlag{
				Name:  "log-level",
				Value: &slog.LevelVar{},
				Validator: func(v TextMarshalUnMarshaller) error {
					text, err := v.MarshalText()
					if err != nil {
						return err
					}

					if !slices.Equal(text, []byte("INFO")) {
						return errors.New("expected \"INFO\"")

					}

					return nil
				},
			},
			args: []string{"--log-level", "info"},
			want: "INFO",
		},
		{
			name: "debug",
			flag: TextFlag{
				Name:  "log-level",
				Value: &slog.LevelVar{},
			},
			args: []string{"--log-level", "debug"},
			want: "DEBUG",
		},
		{
			name: "debug_with_trim",
			flag: TextFlag{
				Name:   "log-level",
				Value:  &slog.LevelVar{},
				Config: StringConfig{TrimSpace: true},
			},
			args: []string{"--log-level", " debug   "},
			want: "DEBUG",
		},
		{
			name: "invalid",
			flag: TextFlag{
				Name:  "log-level",
				Value: &slog.LevelVar{},
			},
			args:    []string{"--log-level", "invalid"},
			want:    "INFO",
			wantErr: true,
		},
		{
			name: "bad_marshaller",
			flag: TextFlag{
				Name:  "text",
				Value: &badMarshaller{},
			},
			args:    []string{"--text", "foo"},
			wantErr: true,
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := flag.NewFlagSet(tt.name, flag.ContinueOnError)
			if tt.wantErr {
				set.SetOutput(io.Discard)
			}

			require.NoError(t, tt.flag.Apply(set))

			err := set.Parse(tt.args)
			require.False(t, (err != nil) && !tt.wantErr, tt.name)

			if tt.wantErr {
				require.Equal(t, tt.flag.GetDefaultText(), tt.want)
			}

			assert.Equal(t, set.Lookup(tt.flag.Name).Value.String(), tt.want)
			assert.Equal(t, tt.flag.GetDefaultText(), tt.want)
		})
	}
}
