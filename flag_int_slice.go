package cli

import "flag"

// IntSlice looks up the value of a local IntSliceFlag, returns
// nil if not found
func (cCtx *Context) IntSlice(name string) []int {
	if fs := cCtx.lookupFlagSet(name); fs != nil {
		return lookupIntSlice(name, fs)
	}
	return nil
}

func lookupIntSlice(name string, set *flag.FlagSet) []int {
	f := set.Lookup(name)
	if f != nil {
		if slice, ok := f.Value.(*IntSlice); ok {
			return slice.Value()
		}
	}
	return nil
}

type IntSlice = SliceBase[int, intValue]
type IntSliceFlag = FlagBase[[]int, IntSlice]

var NewIntSlice = NewSliceBase[int, intValue]
