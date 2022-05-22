package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"
)

// StringSlice wraps a []string to satisfy flag.Value
type StringSlice struct {
	slice      []string
	hasBeenSet bool
}

// NewStringSlice creates a *StringSlice with default values
func NewStringSlice(defaults ...string) *StringSlice {
	return &StringSlice{slice: append([]string{}, defaults...)}
}

// clone allocate a copy of self object
func (s *StringSlice) clone() *StringSlice {
	n := &StringSlice{
		slice:      make([]string, len(s.slice)),
		hasBeenSet: s.hasBeenSet,
	}
	copy(n.slice, s.slice)
	return n
}

// Set appends the string value to the list of values
func (s *StringSlice) Set(value string) error {
	if !s.hasBeenSet {
		s.slice = []string{}
		s.hasBeenSet = true
	}

	if strings.HasPrefix(value, slPfx) {
		// Deserializing assumes overwrite
		_ = json.Unmarshal([]byte(strings.Replace(value, slPfx, "", 1)), &s.slice)
		s.hasBeenSet = true
		return nil
	}

	for _, t := range flagSplitMultiValues(value) {
		s.slice = append(s.slice, strings.TrimSpace(t))
	}

	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (s *StringSlice) String() string {
	return fmt.Sprintf("%s", s.slice)
}

// Serialize allows StringSlice to fulfill Serializer
func (s *StringSlice) Serialize() string {
	jsonBytes, _ := json.Marshal(s.slice)
	return fmt.Sprintf("%s%s", slPfx, string(jsonBytes))
}

// Value returns the slice of strings set by this flag
func (s *StringSlice) Value() []string {
	return s.slice
}

// Get returns the slice of strings set by this flag
func (s *StringSlice) Get() interface{} {
	return *s
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *StringSliceFlag) String() string {
	return withEnvHint(f.GetEnvVars(), stringifyStringSliceFlag(f))
}

// TakesValue returns true of the flag takes a value, otherwise false
func (f *StringSliceFlag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (f *StringSliceFlag) GetUsage() string {
	return f.Usage
}

// GetCategory returns the category for the flag
func (f *StringSliceFlag) GetCategory() string {
	return f.Category
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *StringSliceFlag) GetValue() string {
	if f.Value != nil {
		return f.Value.String()
	}
	return ""
}

// GetDefaultText returns the default text for this flag
func (f *StringSliceFlag) GetDefaultText() string {
	if f.DefaultText != "" {
		return f.DefaultText
	}
	return f.GetValue()
}

// GetEnvVars returns the env vars for this flag
func (f *StringSliceFlag) GetEnvVars() []string {
	return f.EnvVars
}

// Apply populates the flag given the flag set and environment
func (f *StringSliceFlag) Apply(set *flag.FlagSet) error {

	if f.Destination != nil && f.Value != nil {
		f.Destination.slice = make([]string, len(f.Value.slice))
		copy(f.Destination.slice, f.Value.slice)

	}

	if val, source, found := flagFromEnvOrFile(f.EnvVars, f.FilePath); found {
		if f.Value == nil {
			f.Value = &StringSlice{}
		}
		destination := f.Value
		if f.Destination != nil {
			destination = f.Destination
		}

		for _, s := range flagSplitMultiValues(val) {
			if err := destination.Set(strings.TrimSpace(s)); err != nil {
				return fmt.Errorf("could not parse %q as string value from %s for flag %s: %s", val, source, f.Name, err)
			}
		}

		// Set this to false so that we reset the slice if we then set values from
		// flags that have already been set by the environment.
		destination.hasBeenSet = false
		f.HasBeenSet = true
	}

	if f.Value == nil {
		f.Value = &StringSlice{}
	}
	setValue := f.Destination
	if f.Destination == nil {
		setValue = f.Value.clone()
	}
	for _, name := range f.Names() {
		set.Var(setValue, name, f.Usage)
	}

	return nil
}

// Get returns the flag’s value in the given Context.
func (f *StringSliceFlag) Get(ctx *Context) []string {
	return ctx.StringSlice(f.Name)
}

// StringSlice looks up the value of a local StringSliceFlag, returns
// nil if not found
func (cCtx *Context) StringSlice(name string) []string {
	if fs := cCtx.lookupFlagSet(name); fs != nil {
		return lookupStringSlice(name, fs)
	}
	return nil
}

func lookupStringSlice(name string, set *flag.FlagSet) []string {
	f := set.Lookup(name)
	if f != nil {
		if slice, ok := f.Value.(*StringSlice); ok {
			return slice.Value()
		}
	}
	return nil
}
