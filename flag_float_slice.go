package cli

type (
	FloatSlice       = SliceBase[float64, NoConfig, floatValue[float64]]
	Float32Slice     = SliceBase[float32, NoConfig, floatValue[float32]]
	Float64Slice     = SliceBase[float64, NoConfig, floatValue[float64]]
	FloatSliceFlag   = FlagBase[[]float64, NoConfig, FloatSlice]
	Float32SliceFlag = FlagBase[[]float32, NoConfig, Float32Slice]
	Float64SliceFlag = FlagBase[[]float64, NoConfig, Float64Slice]
)

var (
	NewFloatSlice   = NewSliceBase[float64, NoConfig, floatValue[float64]]
	NewFloat32Slice = NewSliceBase[float32, NoConfig, floatValue[float32]]
	NewFloat64Slice = NewSliceBase[float64, NoConfig, floatValue[float64]]
)

// FloatSlice looks up the value of a local FloatSliceFlag, returns
// nil if not found
func (cmd *Command) FloatSlice(name string) []float64 {
	return getFloatSlice[float64](cmd, name)
}

// Float32Slice looks up the value of a local Float32Slice, returns
// nil if not found
func (cmd *Command) Float32Slice(name string) []float32 {
	return getFloatSlice[float32](cmd, name)
}

// Float64Slice looks up the value of a local Float64SliceFlag, returns
// nil if not found
func (cmd *Command) Float64Slice(name string) []float64 {
	return getFloatSlice[float64](cmd, name)
}

func getFloatSlice[T float32 | float64](cmd *Command, name string) []T {
	if v, ok := cmd.Value(name).([]T); ok {
		tracef("float slice available for flag name %[1]q with value=%[2]v (cmd=%[3]q)", name, v, cmd.Name)

		return v
	}

	tracef("float slice NOT available for flag name %[1]q (cmd=%[2]q)", name, cmd.Name)
	return nil
}
