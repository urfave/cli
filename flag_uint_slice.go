package cli

import (
	"flag"
)

type UintSlice = SliceBase[uint, IntegerConfig, uintValue]
type UintSliceFlag = FlagBase[[]uint, IntegerConfig, UintSlice]

var NewUintSlice = NewSliceBase[uint, IntegerConfig, uintValue]

// UintSlice looks up the value of a local UintSliceFlag, returns
// nil if not found
func (cCtx *Context) UintSlice(name string) []uint {
	if fs := cCtx.lookupFlagSet(name); fs != nil {
		return lookupUintSlice(name, fs)
	}
	return nil
}

func lookupUintSlice(name string, set *flag.FlagSet) []uint {
	f := set.Lookup(name)
	if f != nil {
		if slice, ok := f.Value.(*UintSlice); ok {
			return slice.Value()
		}
	}
	return nil
}
