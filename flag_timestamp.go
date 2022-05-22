package cli

import (
	"flag"
	"fmt"
	"time"
)

// Timestamp wrap to satisfy golang's flag interface.
type Timestamp struct {
	timestamp  *time.Time
	hasBeenSet bool
	layout     string
}

// Timestamp constructor
func NewTimestamp(timestamp time.Time) *Timestamp {
	return &Timestamp{timestamp: &timestamp}
}

// Set the timestamp value directly
func (t *Timestamp) SetTimestamp(value time.Time) {
	if !t.hasBeenSet {
		t.timestamp = &value
		t.hasBeenSet = true
	}
}

// Set the timestamp string layout for future parsing
func (t *Timestamp) SetLayout(layout string) {
	t.layout = layout
}

// Parses the string value to timestamp
func (t *Timestamp) Set(value string) error {
	timestamp, err := time.Parse(t.layout, value)
	if err != nil {
		return err
	}

	t.timestamp = &timestamp
	t.hasBeenSet = true
	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (t *Timestamp) String() string {
	return fmt.Sprintf("%#v", t.timestamp)
}

// Value returns the timestamp value stored in the flag
func (t *Timestamp) Value() *time.Time {
	return t.timestamp
}

// Get returns the flag structure
func (t *Timestamp) Get() interface{} {
	return *t
}

// TakesValue returns true of the flag takes a value, otherwise false
func (f *TimestampFlag) TakesValue() bool {
	return true
}

// GetUsage returns the usage string for the flag
func (f *TimestampFlag) GetUsage() string {
	return f.Usage
}

// GetCategory returns the category for the flag
func (f *TimestampFlag) GetCategory() string {
	return f.Category
}

// GetValue returns the flags value as string representation and an empty
// string if the flag takes no value at all.
func (f *TimestampFlag) GetValue() string {
	if f.Value != nil {
		return f.Value.timestamp.String()
	}
	return ""
}

// GetDefaultText returns the default text for this flag
func (f *TimestampFlag) GetDefaultText() string {
	if f.DefaultText != "" {
		return f.DefaultText
	}
	return f.GetValue()
}

// GetEnvVars returns the env vars for this flag
func (f *TimestampFlag) GetEnvVars() []string {
	return f.EnvVars
}

// Apply populates the flag given the flag set and environment
func (f *TimestampFlag) Apply(set *flag.FlagSet) error {
	if f.Layout == "" {
		return fmt.Errorf("timestamp Layout is required")
	}
	if f.Value == nil {
		f.Value = &Timestamp{}
	}
	f.Value.SetLayout(f.Layout)

	if f.Destination != nil {
		f.Destination.SetLayout(f.Layout)
	}

	if val, source, found := flagFromEnvOrFile(f.EnvVars, f.FilePath); found {
		if err := f.Value.Set(val); err != nil {
			return fmt.Errorf("could not parse %q as timestamp value from %s for flag %s: %s", val, source, f.Name, err)
		}
		f.HasBeenSet = true
	}

	for _, name := range f.Names() {
		if f.Destination != nil {
			set.Var(f.Destination, name, f.Usage)
			continue
		}

		set.Var(f.Value, name, f.Usage)
	}
	return nil
}

// Get returns the flag’s value in the given Context.
func (f *TimestampFlag) Get(ctx *Context) *time.Time {
	return ctx.Timestamp(f.Name)
}

// Timestamp gets the timestamp from a flag name
func (cCtx *Context) Timestamp(name string) *time.Time {
	if fs := cCtx.lookupFlagSet(name); fs != nil {
		return lookupTimestamp(name, fs)
	}
	return nil
}

// Fetches the timestamp value from the local timestampWrap
func lookupTimestamp(name string, set *flag.FlagSet) *time.Time {
	f := set.Lookup(name)
	if f != nil {
		return (f.Value.(*Timestamp)).Value()
	}
	return nil
}
