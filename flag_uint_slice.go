package cli

import (
	"flag"
)

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

type UintSlice = SliceBase[uint, uintValue]
type UintSliceFlag = FlagBase[[]uint, UintSlice]

var NewUintSlice = NewSliceBase[uint, uintValue]
