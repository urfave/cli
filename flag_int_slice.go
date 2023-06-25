package cli

import "flag"

type IntSlice = SliceBase[int64, IntegerConfig, intValue]
type IntSliceFlag = FlagBase[[]int64, IntegerConfig, IntSlice]

var NewIntSlice = NewSliceBase[int64, IntegerConfig, intValue]

// IntSlice looks up the value of a local IntSliceFlag, returns
// nil if not found
func (cmd *Command) IntSlice(name string) []int64 {
	if fs := cmd.lookupFlagSet(name); fs != nil {
		return lookupIntSlice(name, fs)
	}
	return nil
}

func lookupIntSlice(name string, set *flag.FlagSet) []int64 {
	f := set.Lookup(name)
	if f != nil {
		if slice, ok := f.Value.(flag.Getter).Get().([]int64); ok {
			return slice
		}
	}
	return nil
}
