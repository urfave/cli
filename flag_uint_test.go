package cli

import (
	"flag"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUintFlag(t *testing.T) {
	tests := []struct {
		name          string
		flag          Flag
		arguments     []string
		expectedValue uint
		expectErr     bool
	}{
		{
			name: "valid",
			flag: &UintFlag{
				Name:    "number",
				Aliases: []string{"n"},
			},
			arguments:     []string{"--number", "234567"},
			expectedValue: 234567,
		},
		{
			name: "invalid",
			flag: &UintFlag{
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
				assert.Equal(t, tt.expectedValue, cmd.Uint(name))
			}
		})
	}
}

func TestUint8Flag(t *testing.T) {
	tests := []struct {
		name          string
		flag          Flag
		arguments     []string
		expectedValue uint8
		expectErr     bool
	}{
		{
			name: "valid",
			flag: &Uint8Flag{
				Name:    "number",
				Aliases: []string{"n"},
			},
			arguments:     []string{"--number", "255"},
			expectedValue: 255,
		},
		{
			name: "invalid",
			flag: &Uint8Flag{
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
				assert.Equal(t, tt.expectedValue, cmd.Uint8(name))
			}
		})
	}
}

func TestUint16Flag(t *testing.T) {
	tests := []struct {
		name          string
		flag          Flag
		arguments     []string
		expectedValue uint16
		expectErr     bool
	}{
		{
			name: "valid",
			flag: &Uint16Flag{
				Name:    "number",
				Aliases: []string{"n"},
			},
			arguments:     []string{"--number", "65535"},
			expectedValue: 65535,
		},
		{
			name: "invalid",
			flag: &Uint16Flag{
				Name: "number",
			},
			arguments: []string{"--number", "gopher"},
			expectErr: true,
		},
		{
			name: "out of range",
			flag: &Uint16Flag{
				Name: "number",
			},
			arguments: []string{"--number", "65536"},
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
				assert.Equal(t, tt.expectedValue, cmd.Uint16(name))
			}
		})
	}
}

func TestUint32Flag(t *testing.T) {
	tests := []struct {
		name          string
		flag          Flag
		arguments     []string
		expectedValue uint32
		expectErr     bool
	}{
		{
			name: "valid",
			flag: &Uint32Flag{
				Name:    "number",
				Aliases: []string{"n"},
			},
			arguments:     []string{"--number", "2147483648"},
			expectedValue: 2147483648,
		},
		{
			name: "invalid",
			flag: &Uint32Flag{
				Name: "number",
			},
			arguments: []string{"--number", "gopher"},
			expectErr: true,
		},
		{
			name: "out of range",
			flag: &Uint32Flag{
				Name: "number",
			},
			arguments: []string{"--number", "4294967297"},
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
				assert.Equal(t, tt.expectedValue, cmd.Uint32(name))
			}
		})
	}
}

func TestUint64Flag(t *testing.T) {
	tests := []struct {
		name          string
		flag          Flag
		arguments     []string
		expectedValue uint64
		expectErr     bool
	}{
		{
			name: "valid",
			flag: &Uint64Flag{
				Name:    "number",
				Aliases: []string{"n"},
			},
			arguments:     []string{"--number", "21474836480"},
			expectedValue: 21474836480,
		},
		{
			name: "invalid",
			flag: &Uint64Flag{
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
				assert.Equal(t, tt.expectedValue, cmd.Uint64(name))
			}
		})
	}
}

func TestUintFlagExt(t *testing.T) {
	tests := []struct {
		name          string
		flag          *flag.Flag
		config        IntegerConfig
		arguments     []string
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
			expectedValue: "234567",
		},
		{
			name: "valid",
			flag: &flag.Flag{
				Name: "number",
			},
			config:        IntegerConfig{Base: 10},
			arguments:     []string{"--number", "234567"},
			expectedValue: "234567",
		},
		{
			name: "valid hex",
			flag: &flag.Flag{
				Name: "number",
			},
			config:        IntegerConfig{Base: 16},
			arguments:     []string{"--number", "39447"},
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
			var uValue uintValue[uint]
			var u uint

			f := &extFlag{f: tt.flag}

			tt.flag.Value = uValue.Create(u, &u, tt.config)

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
