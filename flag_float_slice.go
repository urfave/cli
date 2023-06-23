package cli

import (
	"flag"
)

type FloatSlice = SliceBase[float64, NoConfig, floatValue]
type FloatSliceFlag = FlagBase[[]float64, NoConfig, FloatSlice]

var NewFloatSlice = NewSliceBase[float64, NoConfig, floatValue]

// FloatSlice looks up the value of a local FloatSliceFlag, returns
// nil if not found
func (cCtx *Context) FloatSlice(name string) []float64 {
	if fs := cCtx.lookupFlagSet(name); fs != nil {
		return lookupFloatSlice(name, fs)
	}
	return nil
}

func lookupFloatSlice(name string, set *flag.FlagSet) []float64 {
	f := set.Lookup(name)
	if f != nil {
		if slice, ok := f.Value.(flag.Getter).Get().([]float64); ok {
			return slice
		}
	}
	return nil
}
