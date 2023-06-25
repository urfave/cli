package cli

import (
	"flag"
)

type UintSlice = SliceBase[uint64, IntegerConfig, uintValue]
type UintSliceFlag = FlagBase[[]uint64, IntegerConfig, UintSlice]

var NewUintSlice = NewSliceBase[uint64, IntegerConfig, uintValue]

// UintSlice looks up the value of a local UintSliceFlag, returns
// nil if not found
func (cmd *Command) UintSlice(name string) []uint64 {
	if fs := cmd.lookupFlagSet(name); fs != nil {
		return lookupUintSlice(name, fs)
	}
	return nil
}

func lookupUintSlice(name string, set *flag.FlagSet) []uint64 {
	f := set.Lookup(name)
	if f != nil {
		if slice, ok := f.Value.(flag.Getter).Get().([]uint64); ok {
			return slice
		}
	}
	return nil
}
