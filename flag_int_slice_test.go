package cli

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCommand_IntSlice(t *testing.T) {
	tests := []struct {
		name      string
		flag      Flag
		arguments []string
		expect    []int
		expectErr bool
	}{
		{
			flag: &IntSliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2,3,4"},
			expect:    []int{1, 2, 3, 4},
		},
		{
			flag: &IntSliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2", "--numbers", "3,4"},
			expect:    []int{1, 2, 3, 4},
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
				assert.Equalf(t, tt.expect, cmd.IntSlice(name), "IntSlice(%v)", name)
			}
		})
	}
}

func TestCommand_Int8Slice(t *testing.T) {
	tests := []struct {
		name      string
		flag      Flag
		arguments []string
		expect    []int8
		expectErr bool
	}{
		{
			flag: &Int8SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2,3,4"},
			expect:    []int8{1, 2, 3, 4},
		},
		{
			flag: &Int8SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2", "--numbers", "3,4"},
			expect:    []int8{1, 2, 3, 4},
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
				assert.Equalf(t, tt.expect, cmd.Int8Slice(name), "Int8Slice(%v)", name)
			}
		})
	}
}

func TestCommand_Int16Slice(t *testing.T) {
	tests := []struct {
		name      string
		flag      Flag
		arguments []string
		expect    []int16
		expectErr bool
	}{
		{
			flag: &Int16SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2,3,4"},
			expect:    []int16{1, 2, 3, 4},
		},
		{
			flag: &Int16SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2", "--numbers", "3,4"},
			expect:    []int16{1, 2, 3, 4},
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
				assert.Equalf(t, tt.expect, cmd.Int16Slice(name), "Int16Slice(%v)", name)
			}
		})
	}
}

func TestCommand_Int32Slice(t *testing.T) {
	tests := []struct {
		name      string
		flag      Flag
		arguments []string
		expect    []int32
		expectErr bool
	}{
		{
			flag: &Int32SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2,3,4"},
			expect:    []int32{1, 2, 3, 4},
		},
		{
			flag: &Int32SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2", "--numbers", "3,4"},
			expect:    []int32{1, 2, 3, 4},
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
				assert.Equalf(t, tt.expect, cmd.Int32Slice(name), "Int32Slice(%v)", name)
			}
		})
	}
}

func TestCommand_Int64Slice(t *testing.T) {
	tests := []struct {
		name      string
		flag      Flag
		arguments []string
		expect    []int64
		expectErr bool
	}{
		{
			flag: &Int64SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2,3,4"},
			expect:    []int64{1, 2, 3, 4},
		},
		{
			flag: &Int64SliceFlag{
				Name: "numbers",
			},
			arguments: []string{"--numbers", "1,2", "--numbers", "3,4"},
			expect:    []int64{1, 2, 3, 4},
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
				assert.Equalf(t, tt.expect, cmd.Int64Slice(name), "Int64Slice(%v)", name)
			}
		})
	}
}
