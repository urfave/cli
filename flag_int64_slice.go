package cli

import (
	"flag"
)

type Int64Slice = SliceBase[int64, IntegerConfig, int64Value]
type Int64SliceFlag = FlagBase[[]int64, IntegerConfig, Int64Slice]

var NewInt64Slice = NewSliceBase[int64, IntegerConfig, int64Value]

// Int64Slice looks up the value of a local Int64SliceFlag, returns
// nil if not found
func (cCtx *Context) Int64Slice(name string) []int64 {
	if fs := cCtx.lookupFlagSet(name); fs != nil {
		return lookupInt64Slice(name, fs)
	}
	return nil
}

func lookupInt64Slice(name string, set *flag.FlagSet) []int64 {
	f := set.Lookup(name)
	if f != nil {
		if slice, ok := f.Value.(flag.Getter).Get().([]int64); ok {
			return slice
		}
	}
	return nil
}
