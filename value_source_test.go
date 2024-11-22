package cli

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestZeroValueSourceChain(t *testing.T) {
	var vc ValueSourceChain
	assert.Empty(t, vc.EnvKeys())
	assert.NotEmpty(t, vc.GoString())
	assert.Empty(t, vc.Chain)
	assert.Empty(t, vc.String())
}

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
			src := File(fmt.Sprintf("junk_file_name-%[1]v", rand.Int()))
			_, ok := src.Lookup()
			r.False(ok)
		})

		fileName := filepath.Join(os.TempDir(), fmt.Sprintf("urfave-cli-testing-existing_file-%[1]v", rand.Int()))
		t.Cleanup(func() { _ = os.Remove(fileName) })

		r.Nil(os.WriteFile(fileName, []byte("pita"), 0o644))

		t.Run("found", func(t *testing.T) {
			src := File(fileName)
			str, ok := src.Lookup()
			r.True(ok)
			r.Equal("pita", str)
		})
	})

	t.Run("implements fmt.Stringer", func(t *testing.T) {
		src := File("/dev/null")
		r := require.New(t)

		r.Implements((*ValueSource)(nil), src)
		r.Equal("file \"/dev/null\"", src.String())
	})

	t.Run("implements fmt.GoStringer", func(t *testing.T) {
		src := File("/dev/null")
		r := require.New(t)

		r.Implements((*ValueSource)(nil), src)
		r.Equal("&fileValueSource{Path:\"/dev/null\"}", src.GoString())
	})
}

func TestFilePaths(t *testing.T) {
	r := require.New(t)

	fileName := filepath.Join(os.TempDir(), fmt.Sprintf("urfave-cli-tests-some_file_name_%[1]v", rand.Int()))
	t.Cleanup(func() { _ = os.Remove(fileName) })

	r.Nil(os.WriteFile(fileName, []byte("Hello"), 0o644))

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

func TestMapValueSource(t *testing.T) {
	tests := []struct {
		name  string
		m     map[any]any
		key   string
		val   string
		found bool
	}{
		{
			name: "No map no key",
		},
		{
			name: "No map with key",
			key:  "foo",
		},
		{
			name: "Empty map no key",
			m:    map[any]any{},
		},
		{
			name: "Empty map with key",
			key:  "foo",
			m:    map[any]any{},
		},
		{
			name: "Level 1 no key",
			key:  ".foob",
			m: map[any]any{
				"foo": 10,
			},
		},
		{
			name: "Level 2",
			key:  "foo.bar",
			m: map[any]any{
				"foo": map[any]any{
					"bar": 10,
				},
			},
			val:   "10",
			found: true,
		},
		{
			name: "Level 2 invalid key",
			key:  "foo.bar1",
			m: map[any]any{
				"foo": map[any]any{
					"bar": "10",
				},
			},
		},
		{
			name: "Level 3 no entry",
			key:  "foo.bar.t",
			m: map[any]any{
				"foo": map[any]any{
					"bar": "sss",
				},
			},
		},
		{
			name: "Level 3",
			key:  "foo.bar.t",
			m: map[any]any{
				"foo": map[any]any{
					"bar": map[any]any{
						"t": "sss",
					},
				},
			},
			val:   "sss",
			found: true,
		},
		{
			name: "Level 3 invalid key",
			key:  "foo.bar.t",
			m: map[any]any{
				"foo": map[any]any{
					"bar": map[any]any{
						"t1": 10,
					},
				},
			},
		},
		{
			name: "Level 4 no entry",
			key:  "foo.bar.t.gh",
			m: map[any]any{
				"foo": map[any]any{
					"bar": map[any]any{
						"t1": 10,
					},
				},
			},
		},
		{
			name: "Level 4 slice entry",
			key:  "foo.bar.t.gh",
			m: map[any]any{
				"foo": map[any]any{
					"bar": map[string]any{
						"t": map[any]any{
							"gh": []int{10},
						},
					},
				},
			},
			val:   "[10]",
			found: true,
		},
	}

	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			ms := NewMapSource("test", test.m)
			m := NewMapValueSource(test.key, ms)
			val, b := m.Lookup()
			if !test.found {
				assert.False(t, b)
			} else {
				assert.True(t, b)
				assert.Equal(t, val, test.val)
			}
		})
	}
}

func TestMapValueSourceStringer(t *testing.T) {
	m := map[any]any{
		"foo": map[any]any{
			"bar": 10,
		},
	}
	mvs := NewMapValueSource("bar", NewMapSource("test", m))

	assert.Equal(t, `&mapValueSource{key:"bar", src:&mapSource{name:"test"}}`, mvs.GoString())
	assert.Equal(t, `key "bar" from map source "test"`, mvs.String())
}
