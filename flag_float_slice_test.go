package cli

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommand_FloatSlice(t *testing.T) {
	tests := []struct {
		name      string
		flag      Flag
		arguments []string
		expect    []float64
		expectErr bool
	}{
		{
			flag: &FloatSliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2,3,4"},
			expect:    []float64{1, 2, 3, 4},
		},
		{
			flag: &FloatSliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2", "--numbers", "3,4"},
			expect:    []float64{1, 2, 3, 4},
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
				assert.Equalf(t, tt.expect, cmd.FloatSlice(name), "FloatSlice(%v)", name)
			}
		})
	}
}

func TestCommand_Float32Slice(t *testing.T) {
	tests := []struct {
		name      string
		flag      Flag
		arguments []string
		expect    []float32
		expectErr bool
	}{
		{
			flag: &Float32SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2,3,4"},
			expect:    []float32{1, 2, 3, 4},
		},
		{
			flag: &Float32SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2", "--numbers", "3,4"},
			expect:    []float32{1, 2, 3, 4},
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
				assert.Equalf(t, tt.expect, cmd.Float32Slice(name), "Float32Slice(%v)", name)
			}
		})
	}
}

func TestCommand_Float64Slice(t *testing.T) {
	tests := []struct {
		name      string
		flag      Flag
		arguments []string
		expect    []float64
		expectErr bool
	}{
		{
			flag: &Float64SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2,3,4"},
			expect:    []float64{1, 2, 3, 4},
		},
		{
			flag: &Float64SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2", "--numbers", "3,4"},
			expect:    []float64{1, 2, 3, 4},
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
				assert.Equalf(t, tt.expect, cmd.Float64Slice(name), "Float64Slice(%v)", name)
			}
		})
	}
}
