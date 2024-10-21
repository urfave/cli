package cli

type (
	FloatSlice     = SliceBase[float64, NoConfig, floatValue]
	FloatSliceFlag = FlagBase[[]float64, NoConfig, FloatSlice]
)

var NewFloatSlice = NewSliceBase[float64, NoConfig, floatValue]

// FloatSlice looks up the value of a local FloatSliceFlag, returns
// nil if not found
func (cmd *Command) FloatSlice(name string) []float64 {
	if v, ok := cmd.Value(name).([]float64); ok {
		tracef("float slice available for flag name %[1]q with value=%[2]v (cmd=%[3]q)", name, v, cmd.Name)
		return v
	}

	tracef("float slice NOT available for flag name %[1]q (cmd=%[2]q)", name, cmd.Name)
	return nil
}
