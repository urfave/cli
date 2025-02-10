package cli

import (
	"errors"
	"flag"
	"io"
	"log/slog"
	"slices"
	"testing"
)

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

					if slices.Compare(text, []byte("INFO")) != 0 {
						return errors.New("expected empty string")
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

			if err := tt.flag.Apply(set); err != nil {
				t.Fatalf("Apply(%v) failed: %v", tt.args, err)
			}

			err := set.Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)

				return
			} else if (err != nil) == tt.wantErr {
				// Expected error.
				return
			}

			if got := tt.flag.GetValue(); got != tt.want {
				t.Errorf("Value = %v, want %v", got, tt.want)
			}
		})
	}
}
