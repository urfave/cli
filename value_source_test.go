package cli

import (
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnvSource(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		t.Setenv("foo", "bar")

		s := &envVarValueSource{Key: "foo_1"}
		_, ok := s.Lookup()
		require.False(t, ok)
	})

	t.Run("found", func(t *testing.T) {
		t.Setenv("foo", "bar")

		s := &envVarValueSource{Key: "foo"}
		str, ok := s.Lookup()
		require.True(t, ok)
		require.Equal(t, str, "bar")
	})
}

func TestEnvVars(t *testing.T) {
	t.Setenv("myfoo", "mybar")

	source := EnvVars("foo1", "myfoo")
	str, src, ok := source.LookupWithSource()

	r := require.New(t)
	r.True(ok)
	r.Equal(str, "mybar")
	r.Contains(src.String(), "\"myfoo\"")
}

func TestFileSource(t *testing.T) {
	f := &fileValueSource{Path: "junk_file_name"}
	_, ok := f.Lookup()
	require.False(t, ok)
}

func TestFilePaths(t *testing.T) {
	r := require.New(t)

	fileName := fmt.Sprintf("some_file_name_%[1]v", rand.Int())
	t.Cleanup(func() { _ = os.Remove(fileName) })

	r.Nil(os.Chdir(t.TempDir()))
	r.Nil(os.WriteFile(fileName, []byte("Hello"), 0644))

	sources := Files("junk_file_name", fileName)
	str, src, ok := sources.LookupWithSource()
	r.True(ok)
	r.Equal(str, "Hello")
	r.Contains(src.String(), fmt.Sprintf("%[1]q", fileName))
}

func TestValueSourceChain(t *testing.T) {
	r := require.New(t)

	vsc := &ValueSourceChain{}

	r.Implements((*ValueSource)(nil), vsc)
	r.Equal("ValueSourceChain{Chain:{}}", vsc.GoString())
}
