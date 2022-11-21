package cli

import (
	"flag"
)

// Uint64Slice looks up the value of a local Uint64SliceFlag, returns
// nil if not found
func (cCtx *Context) Uint64Slice(name string) []uint64 {
	if fs := cCtx.lookupFlagSet(name); fs != nil {
		return lookupUint64Slice(name, fs)
	}
	return nil
}

func lookupUint64Slice(name string, set *flag.FlagSet) []uint64 {
	f := set.Lookup(name)
	if f != nil {
		if slice, ok := f.Value.(*Uint64Slice); ok {
			return slice.Value()
		}
	}
	return nil
}

type Uint64Slice = SliceBase[uint64, uint64Value]
type Uint64SliceFlag = FlagBase[[]uint64, Uint64Slice]

var NewUint64Slice = NewSliceBase[uint64, uint64Value]
