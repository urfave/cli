package cli

import (
	"flag"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntFlag(t *testing.T) {
	tests := []struct {
		name          string
		flag          Flag
		arguments     []string
		expectedValue int
		expectErr     bool
	}{
		{
			name: "valid",
			flag: &IntFlag{
				Name:    "number",
				Aliases: []string{"n"},
			},
			arguments:     []string{"--number", "-234567"},
			expectedValue: -234567,
		},
		{
			name: "invalid",
			flag: &IntFlag{
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
				assert.Equal(t, tt.expectedValue, cmd.Int(name))
			}
		})
	}
}

func TestInt8Flag(t *testing.T) {
	tests := []struct {
		name          string
		flag          Flag
		arguments     []string
		expectedValue int8
		expectErr     bool
	}{
		{
			name: "valid",
			flag: &Int8Flag{
				Name:    "number",
				Aliases: []string{"n"},
			},
			arguments:     []string{"--number", "127"},
			expectedValue: 127,
		},
		{
			name: "invalid",
			flag: &Int8Flag{
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
				assert.Equal(t, tt.expectedValue, cmd.Int8(name))
			}
		})
	}
}

func TestInt16Flag(t *testing.T) {
	tests := []struct {
		name          string
		flag          Flag
		arguments     []string
		expectedValue int16
		expectErr     bool
	}{
		{
			name: "valid",
			flag: &Int16Flag{
				Name:    "number",
				Aliases: []string{"n"},
			},
			arguments:     []string{"--number", "32767"},
			expectedValue: 32767,
		},
		{
			name: "invalid",
			flag: &Int16Flag{
				Name: "number",
			},
			arguments: []string{"--number", "gopher"},
			expectErr: true,
		},
		{
			name: "out of range",
			flag: &Int16Flag{
				Name: "number",
			},
			arguments: []string{"--number", "32768"},
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
				assert.Equal(t, tt.expectedValue, cmd.Int16(name))
			}
		})
	}
}

func TestInt32Flag(t *testing.T) {
	tests := []struct {
		name          string
		flag          Flag
		arguments     []string
		expectedValue int32
		expectErr     bool
	}{
		{
			name: "valid",
			flag: &Int32Flag{
				Name:    "number",
				Aliases: []string{"n"},
			},
			arguments:     []string{"--number", "2147483647"},
			expectedValue: 2147483647,
		},
		{
			name: "invalid",
			flag: &Int32Flag{
				Name: "number",
			},

			arguments: []string{"--number", "gopher"},
			expectErr: true,
		},
		{
			name: "out of range",
			flag: &Int32Flag{
				Name: "number",
			},
			arguments: []string{"--number", "2147483648"},
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
				assert.Equal(t, tt.expectedValue, cmd.Int32(name))
			}
		})
	}
}

func TestInt64Flag(t *testing.T) {
	tests := []struct {
		name          string
		flag          Flag
		arguments     []string
		expectedValue int64
		expectErr     bool
	}{
		{
			name: "valid",
			flag: &Int64Flag{
				Name:    "number",
				Aliases: []string{"n"},
			},
			arguments:     []string{"--number", "-2147483648"},
			expectedValue: -2147483648,
		},
		{
			name: "invalid",
			flag: &Int64Flag{
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
				assert.Equal(t, tt.expectedValue, cmd.Int64(name))
			}
		})
	}
}

func TestIntFlagExt(t *testing.T) {
	tests := []struct {
		name          string
		flag          *flag.Flag
		config        IntegerConfig
		arguments     []string
		flagName      string
		expectedValue string
		expectErr     bool
	}{
		{
			name: "valid",
			flag: &flag.Flag{
				Name: "number",
			},
			config:        IntegerConfig{},
			arguments:     []string{"--number", "234567"},
			flagName:      "number",
			expectedValue: "234567",
		},
		{
			name: "valid",
			flag: &flag.Flag{
				Name: "number",
			},
			config:        IntegerConfig{Base: 10},
			arguments:     []string{"--number", "234567"},
			flagName:      "number",
			expectedValue: "234567",
		},
		{
			name: "valid hex",
			flag: &flag.Flag{
				Name:     "number",
				DefValue: "FFFF",
			},
			config:        IntegerConfig{Base: 16},
			arguments:     []string{"--number", "39447"},
			flagName:      "number",
			expectedValue: "39447",
		},
		{
			name: "valid hex default",
			flag: &flag.Flag{
				Name:     "number",
				DefValue: "FFFF",
			},
			config:        IntegerConfig{Base: 16},
			expectedValue: "ffff",
		},
		{
			name: "invalid",
			flag: &flag.Flag{
				Name: "number",
			},
			arguments: []string{"--number", "gopher"},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var uValue intValue[int]
			var u int

			f := &extFlag{f: tt.flag}

			tt.flag.Value = uValue.Create(u, &u, tt.config)

			if tt.config.Base != 0 && tt.config.Base != 10 {
				t.Skipf("skipping %q with base %d, only base 10 is supported", tt.name, tt.config.Base)
			}

			cmd := &Command{
				Name:      "mock",
				Flags:     []Flag{f},
				Writer:    io.Discard,
				ErrWriter: io.Discard,
			}

			err := cmd.Run(buildTestContext(t), append([]string{"mock"}, tt.arguments...))

			if tt.expectErr {
				require.Error(t, err)

				return
			}

			require.NoError(t, err)

			assert.Equal(t, tt.expectedValue, f.GetValue())
		})
	}
}
