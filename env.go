package cli

import (
	"strings"
)

// Represents key=value environment variables
type Env []string

// Get returns the value associated to `key'. If the key doesn't exists,
// returns an empty string.
func (e Env) Get(name string) string {
	v, _ := e.Lookup(name)
	return v
}

// Has checks for the existence of the key in the data set
func (e Env) Has(name string) bool {
	_, ok := e.Lookup(name)
	return ok
}

// Lookup returns the value and true or empty string and false depending on if
// the environment variable exists or not.
func (e Env) Lookup(name string) (string, bool) {
	prefix := name + "="
	for _, pair := range e {
		if strings.HasPrefix(pair, prefix) {
			return pair[len(prefix):], true
		}
	}
	return "", false
}

// Converst the env back to the os.Environ() format.
func (e Env) Environ() []string {
	return []string(e)
}
