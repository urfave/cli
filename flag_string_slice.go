package cli

type StringSlice = SliceBase[string, StringConfig, stringValue]
type StringSliceFlag = FlagBase[[]string, StringConfig, StringSlice]

var NewStringSlice = NewSliceBase[string, StringConfig, stringValue]

// StringSlice looks up the value of a local StringSliceFlag, returns
// nil if not found
func (cCtx *Context) StringSlice(name string) []string {
	if v, ok := cCtx.Value(name).([]string); ok {
		return v
	}
	return nil
}
