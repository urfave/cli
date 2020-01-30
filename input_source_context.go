package cli

import "time"

// InputSourceContext is an interface used to allow
// other input sources to be implemented as needed.
type InputSourceContext interface {
	// Source returns an identifier for the input source. In case of file source
	// it should return path to the file.
	Source() string

	Bool(name string) (bool, error)
	Duration(name string) (time.Duration, error)
	Float64(name string) (float64, error)
	Float64Slice(name string) ([]float64, error)
	Generic(name string) (Generic, error)
	Int(name string) (int, error)
	IntSlice(name string) ([]int, error)
	Int64(name string) (int64, error)
	Int64Slice(name string) ([]int64, error)
	String(name string) (string, error)
	StringSlice(name string) ([]string, error)
	Uint(name string) (uint, error)
	Uint64(name string) (uint64, error)
}
