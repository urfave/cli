package cli

type Args interface {
	// Get returns the nth argument, or else a blank string
	Get(n int) string
	// First returns the first argument, or else a blank string
	First() string
	// Tail returns the rest of the arguments (not the first one)
	// or else an empty string slice
	Tail() []string
	// Len returns the length of the wrapped slice
	Len() int
	// Present checks if there are any arguments present
	Present() bool
	// Slice returns a copy of the internal slice
	Slice() []string
}

type stringSliceArgs struct {
	v []string
}

func (a *stringSliceArgs) Get(n int) string {
	if len(a.v) > n {
		return a.v[n]
	}
	return ""
}

func (a *stringSliceArgs) First() string {
	return a.Get(0)
}

func (a *stringSliceArgs) Tail() []string {
	if a.Len() >= 2 {
		tail := a.v[1:]
		ret := make([]string, len(tail))
		copy(ret, tail)
		return ret
	}

	return []string{}
}

func (a *stringSliceArgs) Len() int {
	return len(a.v)
}

func (a *stringSliceArgs) Present() bool {
	return a.Len() != 0
}

func (a *stringSliceArgs) Slice() []string {
	ret := make([]string, len(a.v))
	copy(ret, a.v)
	return ret
}
