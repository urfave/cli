package cli

import (
	"flag"
)

type FloatSlice = SliceBase[float64, NoConfig, floatValue]
type FloatSliceFlag = FlagBase[[]float64, NoConfig, FloatSlice]

var NewFloatSlice = NewSliceBase[float64, NoConfig, floatValue]

// FloatSlice looks up the value of a local FloatSliceFlag, returns
// nil if not found
func (cmd *Command) FloatSlice(name string) []float64 {
	if flSet := cmd.lookupFlagSet(name); flSet != nil {
		return lookupFloatSlice(name, flSet, cmd.Name)
	}

	return nil
}

func lookupFloatSlice(name string, set *flag.FlagSet, cmdName string) []float64 {
	fl := set.Lookup(name)
	if fl != nil {
		if v, ok := fl.Value.(flag.Getter).Get().([]float64); ok {
			tracef("float slice available for flag name %[1]q with value=%[2]v (cmd=%[3]q)", name, v, cmdName)
			return v
		}
	}

	tracef("float slice NOT available for flag name %[1]q (cmd=%[2]q)", name, cmdName)
	return nil
}
