package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

// Float64Slice wraps []float64 to satisfy flag.Value
type Float64Slice struct {
	slice      []float64
	hasBeenSet bool
}

// NewFloat64Slice makes a *Float64Slice with default values
func NewFloat64Slice(defaults ...float64) *Float64Slice {
	return &Float64Slice{slice: append([]float64{}, defaults...)}
}

// clone allocate a copy of self object
func (f *Float64Slice) clone() *Float64Slice {
	n := &Float64Slice{
		slice:      make([]float64, len(f.slice)),
		hasBeenSet: f.hasBeenSet,
	}
	copy(n.slice, f.slice)
	return n
}

// Set parses the value into a float64 and appends it to the list of values
func (f *Float64Slice) Set(value string) error {
	if !f.hasBeenSet {
		f.slice = []float64{}
		f.hasBeenSet = true
	}

	if strings.HasPrefix(value, slPfx) {
		// Deserializing assumes overwrite
		_ = json.Unmarshal([]byte(strings.Replace(value, slPfx, "", 1)), &f.slice)
		f.hasBeenSet = true
		return nil
	}

	for _, s := range flagSplitMultiValues(value) {
		tmp, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
		if err != nil {
			return err
		}

		f.slice = append(f.slice, tmp)
	}
	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (f *Float64Slice) String() string {
	v := f.slice
	if v == nil {
		// treat nil the same as zero length non-nil
		v = make([]float64, 0)
	}
	return fmt.Sprintf("%#v", v)
}

// Serialize allows Float64Slice to fulfill Serializer
func (f *Float64Slice) Serialize() string {
	jsonBytes, _ := json.Marshal(f.slice)
	return fmt.Sprintf("%s%s", slPfx, string(jsonBytes))
}

// Value returns the slice of float64s set by this flag
func (f *Float64Slice) Value() []float64 {
	return f.slice
}

// Get returns the slice of float64s set by this flag
func (f *Float64Slice) Get() interface{} {
	return *f
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *Float64SliceFlag) String() string {
	return withEnvHint(f.GetEnvVars(), stringifyFloat64SliceFlag(f))
}

// TakesValue returns true if the flag takes a value, otherwise false
func (f *Float64SliceFlag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (f *Float64SliceFlag) GetUsage() string {
	return f.Usage
}

// GetCategory returns the category for the flag
func (f *Float64SliceFlag) GetCategory() string {
	return f.Category
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *Float64SliceFlag) GetValue() string {
	if f.Value != nil {
		return f.Value.String()
	}
	return ""
}

// GetDefaultText returns the default text for this flag
func (f *Float64SliceFlag) GetDefaultText() string {
	if f.DefaultText != "" {
		return f.DefaultText
	}
	return f.GetValue()
}

// GetEnvVars returns the env vars for this flag
func (f *Float64SliceFlag) GetEnvVars() []string {
	return f.EnvVars
}

// Apply populates the flag given the flag set and environment
func (f *Float64SliceFlag) Apply(set *flag.FlagSet) error {
	// apply any default
	if f.Destination != nil && f.Value != nil {
		f.Destination.slice = make([]float64, len(f.Value.slice))
		copy(f.Destination.slice, f.Value.slice)
	}

	// resolve setValue (what we will assign to the set)
	var setValue *Float64Slice
	switch {
	case f.Destination != nil:
		setValue = f.Destination
	case f.Value != nil:
		setValue = f.Value.clone()
	default:
		setValue = new(Float64Slice)
	}

	if val, source, found := flagFromEnvOrFile(f.EnvVars, f.FilePath); found {
		if val != "" {
			for _, s := range flagSplitMultiValues(val) {
				if err := setValue.Set(strings.TrimSpace(s)); err != nil {
					return fmt.Errorf("could not parse %q as float64 slice value from %s for flag %s: %s", val, source, f.Name, err)
				}
			}

			// Set this to false so that we reset the slice if we then set values from
			// flags that have already been set by the environment.
			setValue.hasBeenSet = false
			f.HasBeenSet = true
		}
	}

	for _, name := range f.Names() {
		set.Var(setValue, name, f.Usage)
	}

	return nil
}

// Get returns the flag’s value in the given Context.
func (f *Float64SliceFlag) Get(ctx *Context) []float64 {
	return ctx.Float64Slice(f.Name)
}

// Float64Slice looks up the value of a local Float64SliceFlag, returns
// nil if not found
func (cCtx *Context) Float64Slice(name string) []float64 {
	if fs := cCtx.lookupFlagSet(name); fs != nil {
		return lookupFloat64Slice(name, fs)
	}
	return nil
}

func lookupFloat64Slice(name string, set *flag.FlagSet) []float64 {
	f := set.Lookup(name)
	if f != nil {
		if slice, ok := unwrapFlagValue(f.Value).(*Float64Slice); ok {
			return slice.Value()
		}
	}
	return nil
}
