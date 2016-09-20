package cli

import (
	"strings"
)

// Represents key=value environment variables
type Env map[string]string

// Transforms the os.Environ() format to Env.
//
// Theoretically the environ can contain duplicate keys but I never saw that
// in the wild.
func parseEnviron(environ []string) Env {
	env := make(Env, len(environ))
	for _, pair := range environ {
		kv := strings.SplitN(pair, "=", 2)
		env[kv[0]] = kv[1]
	}
	return env
}

// Get returns the value associated to `key'. If the key doesn't exists,
// returns an empty string.
func (e Env) Get(key string) string {
	return e[key]
}

// Lookup returns the value and true or empty string and false depending on if
// the environment variable exists or not.
func (e Env) Lookup(key string) (string, bool) {
	v, ok := e[key]
	return v, ok
}

// Has checks for the existence of the key in the data set
func (e Env) Has(key string) bool {
	_, ok := e[key]
	return ok
}

// Converst the env back to the os.Environ() format.
func (e Env) Environ() []string {
	environ := make([]string, 0, len(e))
	for k, v := range e {
		environ = append(environ, k+"="+v)
	}
	return environ
}
