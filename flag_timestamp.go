package cli

import (
	"flag"
	"fmt"
	"time"
)

// timestampValue wrap to satisfy golang's flag interface.
type timestampValue struct {
	timestamp  *time.Time
	hasBeenSet bool
	layout     string
	location   *time.Location
}

func (i timestampValue) Create(val time.Time, p *time.Time, c FlagConfig) flag.Value {
	*p = val
	return &timestampValue{
		timestamp: p,
		layout:    c.GetLayout(),
		location:  c.GetTimezone(),
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
func (t *timestampValue) Get() interface{} {
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

type TimestampFlag = FlagBase[time.Time, timestampValue]
