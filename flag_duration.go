package cli

import (
	"flag"
	"fmt"
	"time"
)

type DurationFlag = FlagBase[time.Duration, NoConfig, durationValue]

// -- time.Duration Value
type durationValue time.Duration

// Below functions are to satisfy the ValueCreator interface

func (i durationValue) Create(val time.Duration, p *time.Duration, c NoConfig) flag.Value {
	*p = val
	return (*durationValue)(p)
}

func (i durationValue) ToString(d time.Duration) string {
	return fmt.Sprintf("%v", d)
}

// Below functions are to satisfy the flag.Value interface

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

func (cCtx *Context) Duration(name string) time.Duration {
	if v, ok := cCtx.Value(name).(time.Duration); ok {
		return v
	}
	return 0
}
