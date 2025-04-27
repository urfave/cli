package cli

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_getNumberSlice_int64(t *testing.T) {
	f := &Int64SliceFlag{Name: "numbers"}

	cmd := &Command{
		Name:      "mock",
		Flags:     []Flag{f},
		Writer:    io.Discard,
		ErrWriter: io.Discard,
	}

	err := f.Set("", "1,2,3")
	require.NoError(t, err)

	expected := []int64{1, 2, 3}

	assert.Equal(t, expected, getNumberSlice[int64](cmd, "numbers"))
}

func Test_getNumberSlice_float64(t *testing.T) {
	f := &Float64SliceFlag{Name: "numbers"}

	cmd := &Command{
		Name:      "mock",
		Flags:     []Flag{f},
		Writer:    io.Discard,
		ErrWriter: io.Discard,
	}

	err := f.Set("", "1,2,3")
	require.NoError(t, err)

	expected := []float64{1, 2, 3}

	assert.Equal(t, expected, getNumberSlice[float64](cmd, "numbers"))
}
