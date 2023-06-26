package cli

type Float64Slice = SliceBase[float64, NoConfig, float64Value]
type Float64SliceFlag = FlagBase[[]float64, NoConfig, Float64Slice]

var NewFloat64Slice = NewSliceBase[float64, NoConfig, float64Value]

// Float64Slice looks up the value of a local Float64SliceFlag, returns
// nil if not found
func (cCtx *Context) Float64Slice(name string) []float64 {
	if v, ok := cCtx.Value(name).([]float64); ok {
		return v
	}
	return nil
}
