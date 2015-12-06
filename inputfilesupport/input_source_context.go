package inputfilesupport

import (
	"time"

	"github.com/codegangsta/cli"
)

// InputSourceContext is an interface used to allow
// other input sources to be implemented as needed.
type InputSourceContext interface {
	Int(name string) int
	Duration(name string) time.Duration
	Float64(name string) float64
	String(name string) string
	StringSlice(name string) []string
	IntSlice(name string) []int
	Generic(name string) cli.Generic
	Bool(name string) bool
	BoolT(name string) bool
}
