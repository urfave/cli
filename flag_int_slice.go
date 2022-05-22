package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

// IntSlice wraps []int to satisfy flag.Value
type IntSlice struct {
	slice      []int
	hasBeenSet bool
}

// NewIntSlice makes an *IntSlice with default values
func NewIntSlice(defaults ...int) *IntSlice {
	return &IntSlice{slice: append([]int{}, defaults...)}
}

// clone allocate a copy of self object
func (i *IntSlice) clone() *IntSlice {
	n := &IntSlice{
		slice:      make([]int, len(i.slice)),
		hasBeenSet: i.hasBeenSet,
	}
	copy(n.slice, i.slice)
	return n
}

// TODO: Consistently have specific Set function for Int64 and Float64 ?
// SetInt directly adds an integer to the list of values
func (i *IntSlice) SetInt(value int) {
	if !i.hasBeenSet {
		i.slice = []int{}
		i.hasBeenSet = true
	}

	i.slice = append(i.slice, value)
}

// Set parses the value into an integer and appends it to the list of values
func (i *IntSlice) Set(value string) error {
	if !i.hasBeenSet {
		i.slice = []int{}
		i.hasBeenSet = true
	}

	if strings.HasPrefix(value, slPfx) {
		// Deserializing assumes overwrite
		_ = json.Unmarshal([]byte(strings.Replace(value, slPfx, "", 1)), &i.slice)
		i.hasBeenSet = true
		return nil
	}

	for _, s := range flagSplitMultiValues(value) {
		tmp, err := strconv.ParseInt(strings.TrimSpace(s), 0, 64)
		if err != nil {
			return err
		}

		i.slice = append(i.slice, int(tmp))
	}

	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (i *IntSlice) String() string {
	return fmt.Sprintf("%#v", i.slice)
}

// Serialize allows IntSlice to fulfill Serializer
func (i *IntSlice) Serialize() string {
	jsonBytes, _ := json.Marshal(i.slice)
	return fmt.Sprintf("%s%s", slPfx, string(jsonBytes))
}

// Value returns the slice of ints set by this flag
func (i *IntSlice) Value() []int {
	return i.slice
}

// Get returns the slice of ints set by this flag
func (i *IntSlice) Get() interface{} {
	return *i
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *IntSliceFlag) String() string {
	return withEnvHint(f.GetEnvVars(), stringifyIntSliceFlag(f))
}

// TakesValue returns true of the flag takes a value, otherwise false
func (f *IntSliceFlag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (f *IntSliceFlag) GetUsage() string {
	return f.Usage
}

// GetCategory returns the category for the flag
func (f *IntSliceFlag) GetCategory() string {
	return f.Category
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *IntSliceFlag) GetValue() string {
	if f.Value != nil {
		return f.Value.String()
	}
	return ""
}

// GetDefaultText returns the default text for this flag
func (f *IntSliceFlag) GetDefaultText() string {
	if f.DefaultText != "" {
		return f.DefaultText
	}
	return f.GetValue()
}

// GetEnvVars returns the env vars for this flag
func (f *IntSliceFlag) GetEnvVars() []string {
	return f.EnvVars
}

// Apply populates the flag given the flag set and environment
func (f *IntSliceFlag) Apply(set *flag.FlagSet) error {
	if val, source, found := flagFromEnvOrFile(f.EnvVars, f.FilePath); found {
		f.Value = &IntSlice{}

		for _, s := range flagSplitMultiValues(val) {
			if err := f.Value.Set(strings.TrimSpace(s)); err != nil {
				return fmt.Errorf("could not parse %q as int slice value from %s for flag %s: %s", val, source, f.Name, err)
			}
		}

		// Set this to false so that we reset the slice if we then set values from
		// flags that have already been set by the environment.
		f.Value.hasBeenSet = false
		f.HasBeenSet = true
	}

	if f.Value == nil {
		f.Value = &IntSlice{}
	}
	copyValue := f.Value.clone()
	for _, name := range f.Names() {
		set.Var(copyValue, name, f.Usage)
	}

	return nil
}

// Get returns the flag’s value in the given Context.
func (f *IntSliceFlag) Get(ctx *Context) []int {
	return ctx.IntSlice(f.Name)
}

// IntSlice looks up the value of a local IntSliceFlag, returns
// nil if not found
func (cCtx *Context) IntSlice(name string) []int {
	if fs := cCtx.lookupFlagSet(name); fs != nil {
		return lookupIntSlice(name, fs)
	}
	return nil
}

func lookupIntSlice(name string, set *flag.FlagSet) []int {
	f := set.Lookup(name)
	if f != nil {
		if slice, ok := f.Value.(*IntSlice); ok {
			return slice.Value()
		}
	}
	return nil
}
