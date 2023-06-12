package cli

import (
	"fmt"
	"os"
	"strings"
)

// FlagValueSource encapsulates a source which can be used to
// fetch a value
type FlagValueSource interface {
	// Returns the value from the source and if it was found
	// otherwise returns an empty string & found is set to false
	Get() (string, bool)

	// The identifier for this source
	Identifier() string
}

// ValueSources encapsulates all value sources
type ValueSources []FlagValueSource

func (v ValueSources) Get() (string, string, bool) {
	for _, src := range v {
		if value, found := src.Get(); found {
			return value, src.Identifier(), true
		}
	}

	return "", "", false
}

// EnvSource encapsulates an env
type EnvSource string

func (e EnvSource) Get() (string, bool) {
	envVar := strings.TrimSpace(string(e))
	return os.LookupEnv(envVar)
}

func (e EnvSource) Identifier() string {
	return fmt.Sprintf("environment variable %q", string(e))
}

// EnvVars is a helper function to encapsulate a bunch
// of envs together as ValueSources
func EnvVars(envs ...string) ValueSources {
	vs := []FlagValueSource{}
	for _, env := range envs {
		vs = append(vs, EnvSource(env))
	}
	return vs
}

// FileSource encapsulates an file source
type FileSource string

func (f FileSource) Get() (string, bool) {
	data, err := os.ReadFile(string(f))
	return string(data), err == nil
}

func (f FileSource) Identifier() string {
	return fmt.Sprintf("file %q", string(f))
}

// FilePaths is a helper function to encapsulate a bunch
// of file paths together as ValueSources
func FilePaths(paths ...string) ValueSources {
	vs := []FlagValueSource{}
	for _, path := range paths {
		vs = append(vs, FileSource(path))
	}
	return vs
}
