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
				Value:       &slog.LevelVar{},
				Destination: ptr[TextMarshalUnmarshaler](&slog.LevelVar{}),
			},
			want: "INFO",
		},
		{
			name: "info",
			flag: TextFlag{
				Name:        "log-level",
				Value:       &slog.LevelVar{},
				Destination: ptr[TextMarshalUnmarshaler](&slog.LevelVar{}),
				Validator: func(v TextMarshalUnmarshaler) error {
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
				Value:       &slog.LevelVar{},
				Destination: ptr[TextMarshalUnmarshaler](&slog.LevelVar{}),
			},
			args: []string{"--log-level", "debug"},
			want: "DEBUG",
		},
		{
			name: "debug_with_trim",
			flag: TextFlag{
				Name:        "log-level",
				Value:       &slog.LevelVar{},
				Destination: ptr[TextMarshalUnmarshaler](&slog.LevelVar{}),
				Config:      StringConfig{TrimSpace: true},
			},
			args: []string{"--log-level", " debug   "},
			want: "DEBUG",
		},
		{
			name: "invalid",
			flag: TextFlag{
				Name:        "log-level",
				Value:       &slog.LevelVar{},
				Destination: ptr[TextMarshalUnmarshaler](&slog.LevelVar{}),
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
				Destination: ptr[TextMarshalUnmarshaler](&badMarshaller{}),
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
				Destination: ptr[TextMarshalUnmarshaler](&slog.LevelVar{}),
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
				Destination: ptr[TextMarshalUnmarshaler](&slog.LevelVar{}),
			},
			args: []string{"--log-level", "error"},
			want: "ERROR",
		},
	}

	t.Parallel()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			set := flag.NewFlagSet(tt.name, flag.ContinueOnError)

			// If the test is expected to result in an error then the test needs to make sure that the test output is
			// not littered with output that's expected but not relevant.
			if tt.wantErr {
				set.SetOutput(io.Discard)
			}

			require.NoError(t, tt.flag.Apply(set))

			err := set.Parse(tt.args)

			// Ensure that there's only an error if we wanted an error.
			require.False(t, (err != nil) && !tt.wantErr)

			assert.Equal(t, tt.flag.GetDefaultText(), set.Lookup(tt.flag.Name).DefValue)

			// If the test is expected to fail and the code reaches this point then the test is successful and the test
			// must therefore conclude.
			if tt.wantErr {
				return
			}

			assert.Equal(t, tt.want, set.Lookup(tt.flag.Name).Value.String())
		})
	}
}
