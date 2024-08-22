package cli

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	itesting "github.com/urfave/cli/v3/internal/testing"
)

func TestEnvVarValueSource(t *testing.T) {
	t.Run("implements ValueSource", func(t *testing.T) {
		src := EnvVar("foo")
		itesting.RequireImplements(t, (*ValueSource)(nil), src)

		t.Run("not found", func(t *testing.T) {
			t.Setenv("foo", "bar")

			src := EnvVar("foo_1")
			_, ok := src.Lookup()
			itesting.RequireFalse(t, ok)
		})

		t.Run("found", func(t *testing.T) {
			t.Setenv("foo", "bar")

			src := EnvVar("foo")

			str, ok := src.Lookup()
			itesting.RequireTrue(t, ok)
			itesting.RequireEqual(t, str, "bar")
		})
	})

	t.Run("implements fmt.Stringer", func(t *testing.T) {
		src := EnvVar("foo")

		itesting.RequireImplements(t, (*fmt.Stringer)(nil), src)
		itesting.RequireEqual(t, "environment variable \"foo\"", src.String())
	})

	t.Run("implements fmt.GoStringer", func(t *testing.T) {
		src := EnvVar("foo")

		itesting.RequireImplements(t, (*fmt.GoStringer)(nil), src)
		itesting.RequireEqual(t, "&envVarValueSource{Key:\"foo\"}", src.GoString())
	})
}

func TestEnvVars(t *testing.T) {
	t.Setenv("myfoo", "mybar")

	source := EnvVars("foo1", "myfoo")
	str, src, ok := source.LookupWithSource()

	itesting.RequireTrue(t, ok)
	itesting.RequireEqual(t, str, "mybar")
	itesting.RequireContains(t, src.String(), "\"myfoo\"")
}

func TestFileValueSource(t *testing.T) {
	t.Run("implements ValueSource", func(t *testing.T) {
		itesting.RequireImplements(t, (*ValueSource)(nil), &fileValueSource{})

		t.Run("not found", func(t *testing.T) {
			src := File(fmt.Sprintf("junk_file_name-%[1]v", rand.Int()))
			_, ok := src.Lookup()
			itesting.RequireFalse(t, ok)
		})

		fileName := filepath.Join(os.TempDir(), fmt.Sprintf("urfave-cli-testing-existing_file-%[1]v", rand.Int()))
		t.Cleanup(func() { _ = os.Remove(fileName) })

		itesting.RequireNil(t, os.WriteFile(fileName, []byte("pita"), 0o644))

		t.Run("found", func(t *testing.T) {
			src := File(fileName)
			str, ok := src.Lookup()
			itesting.RequireTrue(t, ok)
			itesting.RequireEqual(t, "pita", str)
		})
	})

	t.Run("implements fmt.Stringer", func(t *testing.T) {
		src := File("/dev/null")

		itesting.RequireImplements(t, (*ValueSource)(nil), src)
		itesting.RequireEqual(t, "file \"/dev/null\"", src.String())
	})

	t.Run("implements fmt.GoStringer", func(t *testing.T) {
		src := File("/dev/null")

		itesting.RequireImplements(t, (*ValueSource)(nil), src)
		itesting.RequireEqual(t, "&fileValueSource{Path:\"/dev/null\"}", src.GoString())
	})
}

func TestFilePaths(t *testing.T) {
	fileName := filepath.Join(os.TempDir(), fmt.Sprintf("urfave-cli-tests-some_file_name_%[1]v", rand.Int()))
	t.Cleanup(func() { _ = os.Remove(fileName) })

	itesting.RequireNil(t, os.WriteFile(fileName, []byte("Hello"), 0o644))

	sources := Files("junk_file_name", fileName)
	str, src, ok := sources.LookupWithSource()
	itesting.RequireTrue(t, ok)
	itesting.RequireEqual(t, str, "Hello")
	itesting.RequireContains(t, src.String(), fmt.Sprintf("%[1]q", fileName))
}

func TestValueSourceChainEnvKeys(t *testing.T) {
	chain := NewValueSourceChain(
		&staticValueSource{"hello"},
	)
	chain.Append(EnvVars("foo", "bar"))

	itesting.RequireEqual(t, []string{"foo", "bar"}, chain.EnvKeys())
}

func TestValueSourceChain(t *testing.T) {
	t.Run("implements ValueSource", func(t *testing.T) {
		vsc := &ValueSourceChain{}

		itesting.RequireImplements(t, (*ValueSource)(nil), vsc)

		_, ok := vsc.Lookup()
		itesting.RequireFalse(t, ok)
	})

	t.Run("implements fmt.GoStringer", func(t *testing.T) {
		vsc := &ValueSourceChain{}

		itesting.RequireImplements(t, (*fmt.GoStringer)(nil), vsc)
		itesting.RequireEqual(t, "&ValueSourceChain{Chain:{}}", vsc.GoString())

		vsc1 := NewValueSourceChain(&staticValueSource{v: "yahtzee"},
			&staticValueSource{v: "matzoh"},
		)
		itesting.RequireEqual(t, "&ValueSourceChain{Chain:{&staticValueSource{v:\"yahtzee\"},&staticValueSource{v:\"matzoh\"}}}", vsc1.GoString())
	})

	t.Run("implements fmt.Stringer", func(t *testing.T) {
		vsc := &ValueSourceChain{}

		itesting.RequireImplements(t, (*fmt.Stringer)(nil), vsc)
		itesting.RequireEqual(t, "", vsc.String())

		vsc1 := NewValueSourceChain(
			&staticValueSource{v: "soup"},
			&staticValueSource{v: "salad"},
			&staticValueSource{v: "pumpkins"},
		)
		itesting.RequireEqual(t, "soup,salad,pumpkins", vsc1.String())
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
