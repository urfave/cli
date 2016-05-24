package cli

import "errors"

var (
	argsRangeErr = errors.New("index out of range")
)

// Args wraps a string slice with some convenience methods
type Args struct {
	slice []string
}

// Get returns the nth argument, or else a blank string
func (a *Args) Get(n int) string {
	if len(a.slice) > n {
		return a.slice[n]
	}
	return ""
}

// First returns the first argument, or else a blank string
func (a *Args) First() string {
	return a.Get(0)
}

// Tail returns the rest of the arguments (not the first one)
// or else an empty string slice
func (a *Args) Tail() []string {
	if a.Len() >= 2 {
		return a.slice[1:]
	}
	return []string{}
}

// Len returns the length of the wrapped slice
func (a *Args) Len() int {
	return len(a.slice)
}

// Present checks if there are any arguments present
func (a *Args) Present() bool {
	return a.Len() != 0
}

// Swap swaps arguments at the given indexes
func (a *Args) Swap(from, to int) error {
	if from >= a.Len() || to >= a.Len() {
		return argsRangeErr
	}
	a.slice[from], a.slice[to] = a.slice[to], a.slice[from]
	return nil
}

// Slice returns a copy of the internal slice
func (a *Args) Slice() []string {
	ret := make([]string, len(a.slice))
	copy(ret, a.slice)
	return ret
}
