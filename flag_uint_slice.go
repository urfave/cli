package cli

type UintSlice = SliceBase[uint64, IntegerConfig, uintValue]
type UintSliceFlag = FlagBase[[]uint64, IntegerConfig, UintSlice]

var NewUintSlice = NewSliceBase[uint64, IntegerConfig, uintValue]

// UintSlice looks up the value of a local UintSliceFlag, returns
// nil if not found
func (cmd *Command) UintSlice(name string) []uint64 {
	if v, ok := cmd.Value(name).([]uint64); ok {
		return v
	}
	return nil
}
