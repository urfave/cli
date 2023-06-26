package cli

type Uint64Slice = SliceBase[uint64, IntegerConfig, uint64Value]
type Uint64SliceFlag = FlagBase[[]uint64, IntegerConfig, Uint64Slice]

var NewUint64Slice = NewSliceBase[uint64, IntegerConfig, uint64Value]

// Uint64Slice looks up the value of a local Uint64SliceFlag, returns
// nil if not found
func (cCtx *Context) Uint64Slice(name string) []uint64 {
	if v, ok := cCtx.Value(name).([]uint64); ok {
		return v
	}
	return nil
}
