package cli

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_FloatFlag(t *testing.T) {
	tests := []struct {
		name          string
		flag          Flag
		arguments     []string
		expectedValue float64
		expectErr     bool
	}{
		{
			name: "valid",
			flag: &FloatFlag{
				Name:    "number",
				Aliases: []string{"n"},
			},
			arguments:     []string{"--number", "-234567"},
			expectedValue: -234567,
		},
		{
			name: "invalid",
			flag: &FloatFlag{
				Name: "number",
			},
			arguments: []string{"--number", "gopher"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Name:      "mock",
				Flags:     []Flag{tt.flag},
				Writer:    io.Discard,
				ErrWriter: io.Discard,
			}

			err := cmd.Run(buildTestContext(t), append([]string{"mock"}, tt.arguments...))

			if tt.expectErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			for _, name := range tt.flag.Names() {
				assert.Equal(t, tt.expectedValue, cmd.Float(name))
			}
		})
	}
}

func Test_Float32Flag(t *testing.T) {
	tests := []struct {
		name          string
		flag          Flag
		arguments     []string
		expectedValue float32
		expectErr     bool
	}{
		{
			name: "valid",
			flag: &Float32Flag{
				Name:    "number",
				Aliases: []string{"n"},
			},
			arguments:     []string{"--number", "2147483647"},
			expectedValue: 2147483647,
		},
		{
			name: "invalid",
			flag: &Float32Flag{
				Name: "number",
			},

			arguments: []string{"--number", "gopher"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Name:      "mock",
				Flags:     []Flag{tt.flag},
				Writer:    io.Discard,
				ErrWriter: io.Discard,
			}

			err := cmd.Run(buildTestContext(t), append([]string{"mock"}, tt.arguments...))

			if tt.expectErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			for _, name := range tt.flag.Names() {
				assert.Equal(t, tt.expectedValue, cmd.Float32(name))
			}
		})
	}
}

func Test_Float64Flag(t *testing.T) {
	tests := []struct {
		name          string
		flag          Flag
		arguments     []string
		expectedValue float64
		expectErr     bool
	}{
		{
			name: "valid",
			flag: &Float64Flag{
				Name:    "number",
				Aliases: []string{"n"},
			},
			arguments:     []string{"--number", "-2147483648"},
			expectedValue: -2147483648,
		},
		{
			name: "invalid",
			flag: &Float64Flag{
				Name: "number",
			},
			arguments: []string{"--number", "gopher"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &Command{
				Name:      "mock",
				Flags:     []Flag{tt.flag},
				Writer:    io.Discard,
				ErrWriter: io.Discard,
			}

			err := cmd.Run(buildTestContext(t), append([]string{"mock"}, tt.arguments...))

			if tt.expectErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			for _, name := range tt.flag.Names() {
				assert.Equal(t, tt.expectedValue, cmd.Float64(name))
			}
		})
	}
}

func Test_floatValue_String(t *testing.T) {
	var f float64 = 100
	fv := floatValue[float64]{val: &f}

	assert.Equal(t, "100", fv.String())
}
