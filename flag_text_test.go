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

func (badMarshaller) UnmarshalText(_ []byte) error {
	return nil
}

func (badMarshaller) MarshalText() ([]byte, error) {
	return nil, errors.New("bad")
}

func ptr[T any](v T) *T {
	return &v
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
				Name:        "log-level",
				Destination: ptr[TextMarshalUnmarshaller](&slog.LevelVar{}),
			},
			want: "INFO",
		},
		{
			name: "info",
			flag: TextFlag{
				Name:        "log-level",
				Destination: ptr[TextMarshalUnmarshaller](&slog.LevelVar{}),
				Validator: func(v TextMarshalUnmarshaller) error {
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
				Name:        "log-level",
				Destination: ptr[TextMarshalUnmarshaller](&slog.LevelVar{}),
			},
			args: []string{"--log-level", "debug"},
			want: "DEBUG",
		},
		{
			name: "debug_with_trim",
			flag: TextFlag{
				Name:        "log-level",
				Destination: ptr[TextMarshalUnmarshaller](&slog.LevelVar{}),
				Config:      StringConfig{TrimSpace: true},
			},
			args: []string{"--log-level", " debug   "},
			want: "DEBUG",
		},
		{
			name: "invalid",
			flag: TextFlag{
				Name:        "log-level",
				Destination: ptr[TextMarshalUnmarshaller](&slog.LevelVar{}),
			},
			args:    []string{"--log-level", "invalid"},
			want:    "INFO",
			wantErr: true,
		},
		{
			name: "bad_marshaller",
			flag: TextFlag{
				Name:        "text",
				Value:       &badMarshaller{},
				Destination: ptr[TextMarshalUnmarshaller](&badMarshaller{}),
			},
			args:    []string{"--text", "foo"},
			wantErr: true,
		},
		{
			name: "default",
			flag: TextFlag{
				Name: "log-level",
				Value: func() *slog.LevelVar {
					var l slog.LevelVar

					l.Set(slog.LevelWarn)

					return &l
				}(),
				Destination: ptr[TextMarshalUnmarshaller](&slog.LevelVar{}),
			},
			args: []string{},
			want: "WARN",
		},
		{
			name: "override_default",
			flag: TextFlag{
				Name: "log-level",
				Value: func() *slog.LevelVar {
					var l slog.LevelVar

					l.Set(slog.LevelWarn)

					return &l
				}(),
				Destination: ptr[TextMarshalUnmarshaller](&slog.LevelVar{}),
			},
			args: []string{"--log-level", "error"},
			want: "ERROR",
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

			if tt.flag.Value != nil {
				assert.Equal(t, tt.want, tt.flag.GetDefaultText())
			}

			if tt.wantErr {
				return
			}

			assert.Equal(t, tt.want, set.Lookup(tt.flag.Name).Value.String())
		})
	}
}
