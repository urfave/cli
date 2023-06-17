package cli

import (
	"fmt"
	"os"
	"strings"
)

// ValueSource is a source which can be used to look up a value,
// typically for use with a cli.Flag
type ValueSource interface {
	fmt.Stringer
	fmt.GoStringer

	// Lookup returns the value from the source and if it was found
	// or returns an empty string and false
	Lookup() (string, bool)
}

// ValueSourceChain is a slice of ValueSource that allows for
// lookup where the first ValueSource to resolve is returned
type ValueSourceChain []ValueSource

func (v ValueSourceChain) Lookup() (string, ValueSource, bool) {
	for _, src := range v {
		if value, found := src.Lookup(); found {
			return value, src, true
		}
	}

	return "", nil, false
}

// envVarValueSource encapsulates a ValueSource from an environment variable
type envVarValueSource struct {
	Key string
}

func (e *envVarValueSource) Lookup() (string, bool) {
	return os.LookupEnv(strings.TrimSpace(string(e.Key)))
}

func (e *envVarValueSource) String() string   { return fmt.Sprintf("environment variable %[1]q", e.Key) }
func (e *envVarValueSource) GoString() string { return fmt.Sprintf("envVarValueSource(%[1]q)", e.Key) }

// EnvVars is a helper function to encapsulate a number of
// EnvSource together as ValueSources
func EnvVars(keys ...string) ValueSourceChain {
	vs := []ValueSource{}
	for _, key := range keys {
		vs = append(vs, &envVarValueSource{Key: key})
	}
	return vs
}

// fileValueSource encapsulates a ValueSource from a file
type fileValueSource struct {
	Path string
}

func (f *fileValueSource) Lookup() (string, bool) {
	data, err := os.ReadFile(string(f.Path))
	return string(data), err == nil
}

func (f *fileValueSource) String() string   { return fmt.Sprintf("file %[1]q", f.Path) }
func (f *fileValueSource) GoString() string { return fmt.Sprintf("fileValueSource(%[1]q)", f.Path) }

// Files is a helper function to encapsulate a number of
// FileSource together as ValueSources
func Files(paths ...string) ValueSourceChain {
	vs := []ValueSource{}
	for _, path := range paths {
		vs = append(vs, &fileValueSource{Path: path})
	}
	return vs
}
