package cli

import "flag"

type StringMap = MapBase[string, StringConfig, stringValue]
type StringMapFlag = FlagBase[map[string]string, StringConfig, StringMap]

var NewStringMap = NewMapBase[string, StringConfig, stringValue]

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
		if mapping, ok := f.Value.(flag.Getter).Get().(map[string]string); ok {
			return mapping
		}
	}
	return nil
}
