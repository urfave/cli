package cli

import (
	"flag"
)

type Uint64Slice = SliceBase[uint64, IntegerConfig, uint64Value]
type Uint64SliceFlag = FlagBase[[]uint64, IntegerConfig, Uint64Slice]

var NewUint64Slice = NewSliceBase[uint64, IntegerConfig, uint64Value]

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
		if slice, ok := f.Value.(flag.Getter).Get().([]uint64); ok {
			return slice
		}
	}
	return nil
}
