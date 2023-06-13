package cli

import (
	"fmt"
	"time"
)

type DurationFlag = FlagBase[time.Duration, NoConfig, durationValue]

// -- time.Duration Value
type durationValue struct {
	val       *time.Duration
	validator func(time.Duration) error
}

// Below functions are to satisfy the ValueCreator interface

func (i durationValue) Create(val time.Duration, p *time.Duration, c NoConfig, validator func(time.Duration) error) Value {
	*p = val
	return &durationValue{
		val:       p,
		validator: validator,
	}
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
	if d.validator != nil {
		if err := d.validator(v); err != nil {
			return err
		}
	}

	*d.val = v
	return err
}

func (d *durationValue) Get() any { return time.Duration(*d.val) }

func (d *durationValue) String() string { return (*time.Duration)(d.val).String() }

func (cCtx *Context) Duration(name string) time.Duration {
	if v, ok := cCtx.Value(name).(time.Duration); ok {
		return v
	}
	return 0
}
