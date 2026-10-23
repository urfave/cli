package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStringSlice_SerializedNil(t *testing.T) {
	source := NewStringSlice()
	target := NewStringSlice("old")
	require.NoError(t, target.Set(source.Serialize()))
	require.NotPanics(t, func() { require.Empty(t, target.Get()); require.Empty(t, target.String()) })
	require.NotPanics(t, func() { require.NoError(t, target.Set("new")) })
	require.Equal(t, []string{"new"}, target.Value())
}

func TestStringMap_SerializedNil(t *testing.T) {
	source := NewStringMap(nil)
	target := NewStringMap(map[string]string{"old": "value"})
	require.NoError(t, target.Set(source.Serialize()))
	require.NotPanics(t, func() { require.Empty(t, target.Get()); _ = target.String() })
	require.NotPanics(t, func() { require.NoError(t, target.Set("new=value")) })
	require.Equal(t, map[string]string{"new": "value"}, target.Value())
}

func TestCollectionSerializedNilPreservesDestination(t *testing.T) {
	t.Run("slice", func(t *testing.T) {
		destination := []string{"old"}
		var creator StringSlice
		value := creator.Create(destination, &destination, StringConfig{})
		require.NoError(t, value.Set(NewStringSlice().Serialize()))
		require.Nil(t, destination)
		require.NoError(t, value.Set("new"))
		require.Equal(t, []string{"new"}, destination)
		require.Equal(t, destination, value.Get())
	})
	t.Run("map", func(t *testing.T) {
		destination := map[string]string{"old": "value"}
		var creator StringMap
		value := creator.Create(destination, &destination, StringConfig{})
		require.NoError(t, value.Set(NewStringMap(nil).Serialize()))
		require.Nil(t, destination)
		require.NoError(t, value.Set("new=value"))
		require.Equal(t, map[string]string{"new": "value"}, destination)
		require.Equal(t, destination, value.Get())
	})
}
