package cli

import (
	"flag"
)

type StringSlice = SliceBase[string, StringConfig, stringValue]
type StringSliceFlag = FlagBase[[]string, StringConfig, StringSlice]

var NewStringSlice = NewSliceBase[string, StringConfig, stringValue]

// StringSlice looks up the value of a local StringSliceFlag, returns
// nil if not found
func (cmd *Command) StringSlice(name string) []string {
	if fs := cmd.lookupFlagSet(name); fs != nil {
		return lookupStringSlice(name, fs)
	}
	return nil
}

func lookupStringSlice(name string, set *flag.FlagSet) []string {
	f := set.Lookup(name)
	if f != nil {
		if slice, ok := f.Value.(flag.Getter).Get().([]string); ok {
			return slice
		}
	}
	return nil
}
