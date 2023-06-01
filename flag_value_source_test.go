package cli

import (
	"fmt"
	"os"
	"testing"
)

func TestEnvSource(t *testing.T) {

	os.Setenv("foo", "bar")
	defer os.Unsetenv("foo")

	s := EnvSource("foo_1")
	_, ok := s.Get()
	expect(t, ok, false)

	s = EnvSource("foo")
	str, ok := s.Get()
	expect(t, ok, true)
	expect(t, str, "bar")

	os.Setenv("myfoo", "mybar")
	defer os.Unsetenv("myfoo")

	source := EnvVars("foo1", "myfoo")
	str, id, ok := source.Get()
	expect(t, ok, true)
	expect(t, str, "mybar")
	expect(t, id, fmt.Sprintf("environment variable %q", "myfoo"))
}

func TestFileSource(t *testing.T) {

	f := FileSource("junk_file_name")
	_, ok := f.Get()
	expect(t, ok, false)

	expect(t, os.WriteFile("some_file_name_1", []byte("Hello"), 0644), nil)
	defer os.Remove("some_file_name_1")

	sources := FilePaths("junk_file_name", "some_file_name_1")
	s, id, ok := sources.Get()
	expect(t, ok, true)
	expect(t, s, "Hello")
	expect(t, id, fmt.Sprintf("file %q", "some_file_name_1"))
}
