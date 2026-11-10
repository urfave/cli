package cli

import (
	"context"
	"fmt"
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
			src := File(filepath.Join(t.TempDir(), "junk_file_name"))
			_, ok := src.Lookup()
			r.False(ok)
		})

		fileName := filepath.Join(t.TempDir(), "existing_file")
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

	fileName := filepath.Join(t.TempDir(), "some_file_name")
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
			name: "Level 1",
			key:  "foobar",
			m: map[any]any{
				"foobar": 10,
			},
			val:   "10",
			found: true,
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
			name: "Level 2 string map type",
			key:  "foo.bar1",
			m: map[any]any{
				"foo": map[string]any{
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
			val:   slPfx() + "[10]",
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

func TestMapValueSourceLists(t *testing.T) {
	ms := NewMapSource("config", map[any]any{
		"ports": []any{80, 443},
		"hosts": []string{"a.example", "b.example"},
		"empty": []any{},
	})

	var ports []int
	var hosts []string
	cmd := &Command{
		Flags: []Flag{
			&IntSliceFlag{Name: "ports", Sources: NewValueSourceChain(NewMapValueSource("ports", ms))},
			&StringSliceFlag{Name: "hosts", Sources: NewValueSourceChain(NewMapValueSource("hosts", ms))},
		},
		Action: func(_ context.Context, cmd *Command) error {
			ports = cmd.IntSlice("ports")
			hosts = cmd.StringSlice("hosts")
			return nil
		},
	}

	if err := cmd.Run(context.Background(), []string{"app"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assert.Equal(t, []int{80, 443}, ports)
	assert.Equal(t, []string{"a.example", "b.example"}, hosts)

	var empty []string
	cmd = &Command{
		Flags: []Flag{
			&StringSliceFlag{Name: "empty", Sources: NewValueSourceChain(NewMapValueSource("empty", ms))},
		},
		Action: func(_ context.Context, cmd *Command) error {
			empty = cmd.StringSlice("empty")
			return nil
		},
	}
	if err := cmd.Run(context.Background(), []string{"app"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assert.Empty(t, empty)
}

// A value must survive the separator it is not split on: an element that
// contains a comma, and a command that changes or disables the separator.
func TestMapValueSourceListSeparators(t *testing.T) {
	ms := NewMapSource("config", map[any]any{
		"hosts": []string{"a,b.example", "c.example"},
		"ports": []any{80, 443},
	})

	tests := []struct {
		name  string
		cmd   *Command
		check func(t *testing.T, cmd *Command)
	}{
		{
			name: "element with a comma",
			cmd: &Command{
				Flags: []Flag{
					&StringSliceFlag{Name: "hosts", Sources: NewValueSourceChain(NewMapValueSource("hosts", ms))},
				},
			},
			check: func(t *testing.T, cmd *Command) {
				assert.Equal(t, []string{"a,b.example", "c.example"}, cmd.StringSlice("hosts"))
			},
		},
		{
			name: "custom separator",
			cmd: &Command{
				SliceFlagSeparator: ";",
				Flags: []Flag{
					&IntSliceFlag{Name: "ports", Sources: NewValueSourceChain(NewMapValueSource("ports", ms))},
				},
			},
			check: func(t *testing.T, cmd *Command) {
				assert.Equal(t, []int{80, 443}, cmd.IntSlice("ports"))
			},
		},
		{
			name: "separator disabled",
			cmd: &Command{
				DisableSliceFlagSeparator: true,
				Flags: []Flag{
					&IntSliceFlag{Name: "ports", Sources: NewValueSourceChain(NewMapValueSource("ports", ms))},
				},
			},
			check: func(t *testing.T, cmd *Command) {
				assert.Equal(t, []int{80, 443}, cmd.IntSlice("ports"))
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			test.cmd.Action = func(_ context.Context, cmd *Command) error {
				test.check(t, cmd)
				return nil
			}
			if err := test.cmd.Run(context.Background(), []string{"app"}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// A list from a map source must reach the flag through its own element parser,
// not through a JSON decode that silently produces zero values, and a scalar
// flag must never see the internal serialization prefix.
func TestMapValueSourceListsWithMismatchedTypes(t *testing.T) {
	ms := NewMapSource("config", map[any]any{
		"anyints":  []any{80, 443},
		"strings":  []string{"80", "443"},
		"mixed":    []any{"a", 2, true},
		"negative": []any{-1},
		"plain":    []string{"x", "y"},
	})

	for _, tc := range []struct {
		name    string
		flag    Flag
		want    any
		wantErr bool
	}{
		{
			name: "ints in a string slice read as their digits",
			flag: &StringSliceFlag{Name: "anyints", Sources: NewValueSourceChain(NewMapValueSource("anyints", ms))},
			want: []string{"80", "443"},
		},
		{
			name: "strings in an int slice are parsed",
			flag: &IntSliceFlag{Name: "strings", Sources: NewValueSourceChain(NewMapValueSource("strings", ms))},
			want: []int{80, 443},
		},
		{
			name: "mixed elements all arrive",
			flag: &StringSliceFlag{Name: "mixed", Sources: NewValueSourceChain(NewMapValueSource("mixed", ms))},
			want: []string{"a", "2", "true"},
		},
		{
			name:    "a negative into an unsigned slice is an error, not a zero",
			flag:    &UintSliceFlag{Name: "negative", Sources: NewValueSourceChain(NewMapValueSource("negative", ms))},
			wantErr: true,
		},
		{
			name: "a list into a scalar flag never carries the prefix",
			flag: &StringFlag{Name: "plain", Sources: NewValueSourceChain(NewMapValueSource("plain", ms))},
			want: "x,y",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var got any
			cmd := &Command{
				Flags: []Flag{tc.flag},
				Action: func(_ context.Context, cmd *Command) error {
					switch f := tc.flag.(type) {
					case *StringSliceFlag:
						got = cmd.StringSlice(f.Name)
					case *IntSliceFlag:
						got = cmd.IntSlice(f.Name)
					case *UintSliceFlag:
						got = cmd.UintSlice(f.Name)
					case *StringFlag:
						got = cmd.String(f.Name)
					}
					return nil
				},
			}
			err := cmd.Run(context.Background(), []string{"app"})
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected an error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			assert.Equal(t, tc.want, got)
		})
	}
}

// The serialized form is scanned without a JSON decoder, so the elements that
// a decoder would unescape have to arrive intact: the separator inside a
// value, quotes, backslashes, a newline, empty strings and brackets.
func TestMapValueSourceListElementsSurviveTheSerializedForm(t *testing.T) {
	want := []string{
		"a,b",
		`he said "hi"`,
		`back\slash`,
		"emoji 🎉",
		"",
		"[brackets]",
		"line\nbreak",
		"tab\there",
	}
	ms := NewMapSource("tricky", map[any]any{"tricky": want})

	var got []string
	cmd := &Command{
		Flags: []Flag{
			&StringSliceFlag{Name: "tricky", Sources: NewValueSourceChain(NewMapValueSource("tricky", ms))},
		},
		Action: func(_ context.Context, cmd *Command) error {
			got = cmd.StringSlice("tricky")
			return nil
		},
	}
	require.NoError(t, cmd.Run(context.Background(), []string{"app"}))
	assert.Equal(t, want, got)
}

// The serialized scanner is hand-written, so its refusals are pinned here: a
// value that is not the flat array Serialize writes must come back as an
// error, never as a silently truncated list.
func TestSerializedListElements(t *testing.T) {
	for _, tc := range []struct {
		name    string
		in      string
		want    []string
		wantErr bool
	}{
		{name: "strings", in: `["a","b"]`, want: []string{"a", "b"}},
		{name: "empty array", in: `[]`, want: []string{}},
		{name: "null", in: `null`},
		{name: "numbers and bools", in: `[80,-1,true]`, want: []string{"80", "-1", "true"}},
		{name: "escapes", in: `["a,b","he said \"hi\"","back\\slash",""]`, want: []string{"a,b", `he said "hi"`, `back\slash`, ""}},
		{name: "object", in: `{"a":1}`, wantErr: true},
		{name: "unterminated string", in: `["abc]`, wantErr: true},
		{name: "bad escape", in: `["\q"]`, wantErr: true},
		{name: "nested value", in: `[{"a":1}]`, wantErr: true},
		{name: "missing comma", in: `["a" "b"]`, wantErr: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := serializedListElements(tc.in)
			if tc.wantErr {
				require.Error(t, err, "got %q", got)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

// A malformed serialized value reaches the caller as an error, not as an empty
// list.
func TestSliceBaseRejectsAMalformedSerializedValue(t *testing.T) {
	var dest []string
	v := StringSlice{}.Create([]string{"old"}, &dest, StringConfig{})
	require.NoError(t, v.Set("a,b"))
	require.Error(t, v.Set(slPfx()+`["c"`))
	assert.Equal(t, []string{"a", "b"}, dest)
}
