package cli

import (
	"flag"
	"fmt"
	"time"
)


// timestamp wrap to satisfy golang's flag interface.
type timestampWrap struct {
	timestamp *time.Time
	hasBeenSet bool
	layout 	   string
}

// Set the timestamp value directly
func (t *timestampWrap) SetTimestamp(value time.Time) {
	if !t.hasBeenSet {
		t.timestamp = &value
		t.hasBeenSet = true
	}
}
// Set the timestamp string layout for future parsing
func (t *timestampWrap) SetLayout(layout string) {
	t.layout = layout
}

// Parses the string value to timestamp
func (t *timestampWrap) Set(value string) error {
	timestamp, err := time.Parse(t.layout, value)
	if err != nil {
		return err
	}

	t.timestamp = &timestamp
	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (t *timestampWrap) String() string {
	return fmt.Sprintf("%#v", t.timestamp)
}

// Value returns the timestamp value stored in the flag
func (t *timestampWrap) Value() *time.Time {
	return t.timestamp
}

// Get returns the flag structure
func (t *timestampWrap) Get() interface{} {
	return *t
}

// TimestampFlag is a flag with type protobuf.timestamp
type TimestampFlag struct {
	Name        string
	Aliases     []string
	Usage       string
	EnvVars     []string
	FilePath    string
	Required    bool
	Hidden      bool
	Layout 	    string
	Value       timestampWrap
	DefaultText string
	HasBeenSet  bool
}

// IsSet returns whether or not the flag has been set through env or file
func (f *TimestampFlag) IsSet() bool {
	return f.HasBeenSet
}

// String returns a readable representation of this value
// (for usage defaults)
func (f *TimestampFlag) String() string {
	return FlagStringer(f)
}

// Names returns the names of the flag
func (f *TimestampFlag) Names() []string {
	return flagNames(f)
}

// IsRequired returns whether or not the flag is required
func (f *TimestampFlag) IsRequired() bool {
	return f.Required
}

// TakesValue returns true of the flag takes a value, otherwise false
func (f *TimestampFlag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (f *TimestampFlag) GetUsage() string {
	return f.Usage
}

// GetValue returns the flag value
func (f *TimestampFlag) GetValue() *time.Time {
	return f.Value.timestamp
}

// Apply populates the flag given the flag set and environment
func (f *TimestampFlag) Apply(set *flag.FlagSet) error {
	for _, name := range f.Names() {
		f.Value.SetLayout(f.Layout)
		set.Var(&f.Value, name, f.Usage)
	}
	return nil
}

// Timestamp gets the timestamp from a flag name
func (c *Context) Timestamp(name string) *time.Time {
	if fs := lookupFlagSet(name, c); fs != nil {
		return lookupTimestamp(name, fs)
	}
	return nil
}

// Fetches the timestamp value from the local timestampWrap
func lookupTimestamp(name string, set *flag.FlagSet) *time.Time {
	f := set.Lookup(name)
	if f != nil {
		return (f.Value.(*timestampWrap)).Value()
	}
	return nil
}
