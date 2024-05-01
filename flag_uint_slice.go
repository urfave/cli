package cli

type (
	UintSlice     = SliceBase[uint64, IntegerConfig, uintValue]
	UintSliceFlag = FlagBase[[]uint64, IntegerConfig, UintSlice]
)

var NewUintSlice = NewSliceBase[uint64, IntegerConfig, uintValue]

// UintSlice looks up the value of a local UintSliceFlag, returns
// nil if not found
func (cmd *Command) UintSlice(name string) []uint64 {
	if v, ok := cmd.Value(name).([]uint64); ok {
		tracef("uint slice available for flag name %[1]q with value=%[2]v (cmd=%[3]q)", name, v, cmd.Name)
		return v
	}

	tracef("uint slice NOT available for flag name %[1]q (cmd=%[2]q)", name, cmd.Name)
	return nil
}
