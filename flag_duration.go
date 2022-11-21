package cli

import (
	"flag"
	"fmt"
	"time"
)

// -- time.Duration Value
type durationValue time.Duration

func (i durationValue) Create(val time.Duration, p *time.Duration, c FlagConfig) flag.Value {
	*p = val
	return (*durationValue)(p)
}

func (i durationValue) ToString(d time.Duration) string {
	return fmt.Sprintf("%v", d)
}

func (d *durationValue) Set(s string) error {
	v, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	*d = durationValue(v)
	return err
}

func (d *durationValue) Get() any { return time.Duration(*d) }

func (d *durationValue) String() string { return (*time.Duration)(d).String() }

type DurationFlag = FlagBase[time.Duration, durationValue]

func (cCtx *Context) Duration(name string) time.Duration {
	if v, ok := cCtx.Value(name).(time.Duration); ok {
		return v
	}
	return 0
}
