package cli

import (
	"encoding"
	"errors"
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
		flag    *TextFlag
		args    []string
		want    string
		wantErr bool
	}{
		{
			name: "empty",
			flag: &TextFlag{
				Name:        "log-level",
				Value:       &slog.LevelVar{},
				Destination: ptr[TextMarshalUnmarshaler](&slog.LevelVar{}),
			},
			want: "INFO",
		},
		{
			name: "info",
			flag: &TextFlag{
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
			flag: &TextFlag{
				Name:        "log-level",
				Value:       &slog.LevelVar{},
				Destination: ptr[TextMarshalUnmarshaler](&slog.LevelVar{}),
			},
			args: []string{"--log-level", "debug"},
			want: "DEBUG",
		},
		{
			name: "invalid",
			flag: &TextFlag{
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
			flag: &TextFlag{
				Name:        "text",
				Value:       &badMarshaller{},
				Destination: ptr[TextMarshalUnmarshaler](&badMarshaller{}),
			},
			args:    []string{"--text", "foo"},
			wantErr: true,
		},
		{
			name: "default",
			flag: &TextFlag{
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
			flag: &TextFlag{
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
		t.Parallel()

		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Name:      tt.name,
				Flags:     []Flag{tt.flag},
				Writer:    io.Discard,
				ErrWriter: io.Discard,
			}

			err := cmd.Run(buildTestContext(t), append([]string{"mock"}, tt.args...))

			if err != nil && !tt.wantErr {
				require.NoError(t, err)

				return
			} else if err != nil {
				return
			}

			var got []byte

			got, err = tt.flag.Get().(encoding.TextMarshaler).MarshalText()
			if tt.wantErr {
				require.Error(t, err)

				return
			}

			assert.Equal(t, tt.want, string(got))
		})
	}
}
