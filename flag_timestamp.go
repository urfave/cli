package cli

import (
	"flag"
	"fmt"
	"time"
)

type TimestampFlag = FlagBase[time.Time, TimestampConfig, timestampValue]

// TimestampConfig defines the config for timestamp flags
type TimestampConfig struct {
	Timezone *time.Location
	Layout   string
}

// timestampValue wrap to satisfy golang's flag interface.
type timestampValue struct {
	timestamp  *time.Time
	hasBeenSet bool
	layout     string
	location   *time.Location
	validator  func(time.Time) error
}

// Below functions are to satisfy the ValueCreator interface

func (i timestampValue) Create(val time.Time, p *time.Time, c TimestampConfig, validator func(time.Time) error) Value {
	*p = val
	return &timestampValue{
		timestamp: p,
		layout:    c.Layout,
		location:  c.Timezone,
		validator: validator,
	}
}

func (i timestampValue) ToString(b time.Time) string {
	if b.IsZero() {
		return ""
	}
	return fmt.Sprintf("%v", b)
}

// Timestamp constructor(for internal testing only)
func newTimestamp(timestamp time.Time) *timestampValue {
	return &timestampValue{timestamp: &timestamp}
}

// Below functions are to satisfy the flag.Value interface

// Parses the string value to timestamp
func (t *timestampValue) Set(value string) error {
	var timestamp time.Time
	var err error

	if t.location != nil {
		timestamp, err = time.ParseInLocation(t.layout, value, t.location)
	} else {
		timestamp, err = time.Parse(t.layout, value)
	}

	if err != nil {
		return err
	}

	if t.validator != nil {
		if err := t.validator(timestamp); err != nil {
			return err
		}
	}

	if t.timestamp != nil {
		*t.timestamp = timestamp
	}
	t.hasBeenSet = true
	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (t *timestampValue) String() string {
	return fmt.Sprintf("%#v", t.timestamp)
}

// Value returns the timestamp value stored in the flag
func (t *timestampValue) Value() *time.Time {
	return t.timestamp
}

// Get returns the flag structure
func (t *timestampValue) Get() any {
	return *t.timestamp
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
		return (f.Value.(*timestampValue)).Value()
	}
	return nil
}
