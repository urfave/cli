package cli

type (
	IntSlice     = SliceBase[int64, IntegerConfig, intValue]
	IntSliceFlag = FlagBase[[]int64, IntegerConfig, IntSlice]
)

var NewIntSlice = NewSliceBase[int64, IntegerConfig, intValue]

// IntSlice looks up the value of a local IntSliceFlag, returns
// nil if not found
func (cmd *Command) IntSlice(name string) []int64 {
	if v, ok := cmd.Value(name).([]int64); ok {
		tracef("int slice available for flag name %[1]q with value=%[2]v (cmd=%[3]q)", name, v, cmd.Name)
		return v
	}

	tracef("int slice NOT available for flag name %[1]q (cmd=%[2]q)", name, cmd.Name)
	return nil
}
