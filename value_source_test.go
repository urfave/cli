package cli

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnvVarValueSource(t *testing.T) {
	t.Run("implements ValueSource", func(t *testing.T) {
		src := EnvVar("foo")
		require.Implements(t, (*ValueSource)(nil), src)

		t.Run("not found", func(t *testing.T) {
			t.Setenv("foo", "bar")

			src := EnvVar("foo_1")
			_, ok := src.Lookup()
			require.False(t, ok)
		})

		t.Run("found", func(t *testing.T) {
			t.Setenv("foo", "bar")

			r := require.New(t)
			src := EnvVar("foo")

			str, ok := src.Lookup()
			r.True(ok)
			r.Equal(str, "bar")
		})

	})

	t.Run("implements fmt.Stringer", func(t *testing.T) {
		src := EnvVar("foo")
		r := require.New(t)

		r.Implements((*fmt.Stringer)(nil), src)
		r.Equal("environment variable \"foo\"", src.String())
	})

	t.Run("implements fmt.GoStringer", func(t *testing.T) {
		src := EnvVar("foo")
		r := require.New(t)

		r.Implements((*fmt.GoStringer)(nil), src)
		r.Equal("&envVarValueSource{Key:\"foo\"}", src.GoString())
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

func TestFileValueSource(t *testing.T) {
	t.Run("implements ValueSource", func(t *testing.T) {
		r := require.New(t)

		r.Implements((*ValueSource)(nil), &fileValueSource{})

		t.Run("not found", func(t *testing.T) {
			src := &fileValueSource{Path: fmt.Sprintf("junk_file_name-%[1]v", rand.Int())}
			_, ok := src.Lookup()
			r.False(ok)
		})

		fileName := filepath.Join(os.TempDir(), fmt.Sprintf("urfave-cli-testing-existing_file-%[1]v", rand.Int()))
		t.Cleanup(func() { _ = os.Remove(fileName) })

		r.Nil(os.WriteFile(fileName, []byte("pita"), 0644))

		t.Run("found", func(t *testing.T) {
			src := &fileValueSource{Path: fileName}
			str, ok := src.Lookup()
			r.True(ok)
			r.Equal("pita", str)
		})
	})

	t.Run("implements fmt.Stringer", func(t *testing.T) {
		src := &fileValueSource{Path: "/dev/null"}
		r := require.New(t)

		r.Implements((*ValueSource)(nil), src)
		r.Equal("file \"/dev/null\"", src.String())
	})

	t.Run("implements fmt.GoStringer", func(t *testing.T) {
		src := &fileValueSource{Path: "/dev/null"}
		r := require.New(t)

		r.Implements((*ValueSource)(nil), src)
		r.Equal("&fileValueSource{Path:\"/dev/null\"}", src.GoString())
	})
}

func TestFilePaths(t *testing.T) {
	r := require.New(t)

	fileName := filepath.Join(os.TempDir(), fmt.Sprintf("urfave-cli-tests-some_file_name_%[1]v", rand.Int()))
	t.Cleanup(func() { _ = os.Remove(fileName) })

	r.Nil(os.WriteFile(fileName, []byte("Hello"), 0644))

	sources := Files("junk_file_name", fileName)
	str, src, ok := sources.LookupWithSource()
	r.True(ok)
	r.Equal(str, "Hello")
	r.Contains(src.String(), fmt.Sprintf("%[1]q", fileName))
}

func TestValueSourceChainEnvKeys(t *testing.T) {
	chain := NewValueSourceChain(
		&staticValueSource{"hello"},
	)
	chain.Append(EnvVars("foo", "bar"))

	r := require.New(t)
	r.Equal([]string{"foo", "bar"}, chain.EnvKeys())
}

func TestValueSourceChain(t *testing.T) {
	t.Run("implements ValueSource", func(t *testing.T) {
		vsc := &ValueSourceChain{}
		r := require.New(t)

		r.Implements((*ValueSource)(nil), vsc)

		_, ok := vsc.Lookup()
		r.False(ok)
	})

	t.Run("implements fmt.GoStringer", func(t *testing.T) {
		vsc := &ValueSourceChain{}
		r := require.New(t)

		r.Implements((*fmt.GoStringer)(nil), vsc)
		r.Equal("&ValueSourceChain{Chain:{}}", vsc.GoString())

		vsc1 := NewValueSourceChain(&staticValueSource{v: "yahtzee"},
			&staticValueSource{v: "matzoh"},
		)
		r.Equal("&ValueSourceChain{Chain:{&staticValueSource{v:\"yahtzee\"},&staticValueSource{v:\"matzoh\"}}}", vsc1.GoString())
	})

	t.Run("implements fmt.Stringer", func(t *testing.T) {
		vsc := &ValueSourceChain{}
		r := require.New(t)

		r.Implements((*fmt.Stringer)(nil), vsc)
		r.Equal("", vsc.String())

		vsc1 := NewValueSourceChain(
			&staticValueSource{v: "soup"},
			&staticValueSource{v: "salad"},
			&staticValueSource{v: "pumpkins"},
		)
		r.Equal("soup,salad,pumpkins", vsc1.String())
	})
}

type staticValueSource struct {
	v string
}

func (svs *staticValueSource) GoString() string {
	return fmt.Sprintf("&staticValueSource{v:%[1]q}", svs.v)
}
func (svs *staticValueSource) String() string         { return svs.v }
func (svs *staticValueSource) Lookup() (string, bool) { return svs.v, true }
