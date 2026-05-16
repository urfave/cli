package cli

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommand_UintSlice(t *testing.T) {
	tests := []struct {
		name      string
		flag      Flag
		arguments []string
		expect    []uint
		expectErr bool
	}{
		{
			flag: &UintSliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2,3,4"},
			expect:    []uint{1, 2, 3, 4},
		},
		{
			flag: &UintSliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2", "--numbers", "3,4"},
			expect:    []uint{1, 2, 3, 4},
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
				assert.Equalf(t, tt.expect, cmd.UintSlice(name), "UintSlice(%v)", name)
			}
		})
	}
}

func TestCommand_Uint8Slice(t *testing.T) {
	tests := []struct {
		name      string
		flag      Flag
		arguments []string
		expect    []uint8
		expectErr bool
	}{
		{
			flag: &Uint8SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2,3,4"},
			expect:    []uint8{1, 2, 3, 4},
		},
		{
			flag: &Uint8SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2", "--numbers", "3,4"},
			expect:    []uint8{1, 2, 3, 4},
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
				assert.Equalf(t, tt.expect, cmd.Uint8Slice(name), "Uint8Slice(%v)", name)
			}
		})
	}
}

func TestCommand_Uint16Slice(t *testing.T) {
	tests := []struct {
		name      string
		flag      Flag
		arguments []string
		expect    []uint16
		expectErr bool
	}{
		{
			flag: &Uint16SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2,3,4"},
			expect:    []uint16{1, 2, 3, 4},
		},
		{
			flag: &Uint16SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2", "--numbers", "3,4"},
			expect:    []uint16{1, 2, 3, 4},
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
				assert.Equalf(t, tt.expect, cmd.Uint16Slice(name), "Uint16Slice(%v)", name)
			}
		})
	}
}

func TestCommand_Uint32Slice(t *testing.T) {
	tests := []struct {
		name      string
		flag      Flag
		arguments []string
		expect    []uint32
		expectErr bool
	}{
		{
			flag: &Uint32SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2,3,4"},
			expect:    []uint32{1, 2, 3, 4},
		},
		{
			flag: &Uint32SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2", "--numbers", "3,4"},
			expect:    []uint32{1, 2, 3, 4},
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
				assert.Equalf(t, tt.expect, cmd.Uint32Slice(name), "Uint32Slice(%v)", name)
			}
		})
	}
}

func TestCommand_Uint64Slice(t *testing.T) {
	tests := []struct {
		name      string
		flag      Flag
		arguments []string
		expect    []uint64
		expectErr bool
	}{
		{
			flag: &Uint64SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2,3,4"},
			expect:    []uint64{1, 2, 3, 4},
		},
		{
			flag: &Uint64SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2", "--numbers", "3,4"},
			expect:    []uint64{1, 2, 3, 4},
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
				assert.Equalf(t, tt.expect, cmd.Uint64Slice(name), "Uint64Slice(%v)", name)
			}
		})
	}
}
