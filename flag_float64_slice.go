package cli

import (
	"flag"
)

// Float64Slice looks up the value of a local Float64SliceFlag, returns
// nil if not found
func (cCtx *Context) Float64Slice(name string) []float64 {
	if fs := cCtx.lookupFlagSet(name); fs != nil {
		return lookupFloat64Slice(name, fs)
	}
	return nil
}

func lookupFloat64Slice(name string, set *flag.FlagSet) []float64 {
	f := set.Lookup(name)
	if f != nil {
		if slice, ok := f.Value.(*Float64Slice); ok {
			return slice.Value()
		}
	}
	return nil
}

type Float64Slice = SliceBase[float64, float64Value]
type Float64SliceFlag = FlagBase[[]float64, Float64Slice]

var NewFloat64Slice = NewSliceBase[float64, float64Value]
