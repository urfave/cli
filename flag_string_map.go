package cli

import "flag"

type StringMap = MapBase[string, NoConfig, stringValue]
type StringMapFlag = FlagBase[map[string]string, NoConfig, StringMap]

var NewStringMap = NewMapBase[string, NoConfig, stringValue]

// StringMap looks up the value of a local StringMapFlag, returns
// nil if not found
func (cCtx *Context) StringMap(name string) map[string]string {
	if fs := cCtx.lookupFlagSet(name); fs != nil {
		return lookupStringMap(name, fs)
	}
	return nil
}

func lookupStringMap(name string, set *flag.FlagSet) map[string]string {
	f := set.Lookup(name)
	if f != nil {
		if mapping, ok := f.Value.(*StringMap); ok {
			return mapping.Value()
		}
	}
	return nil
}
