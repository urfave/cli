package cli

import (
	"fmt"
	"strconv"
)

type Float64Flag = FlagBase[float64, NoConfig, float64Value]

// -- float64 Value
type float64Value struct {
	val       *float64
	validator func(float64) error
}

// Below functions are to satisfy the ValueCreator interface

func (f float64Value) Create(val float64, p *float64, c NoConfig, validator func(float64) error) Value {
	*p = val
	return &float64Value{
		val:       p,
		validator: validator,
	}
}

func (f float64Value) ToString(b float64) string {
	return fmt.Sprintf("%v", b)
}

// Below functions are to satisfy the flag.Value interface

func (f *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	if f.validator != nil {
		if err := f.validator(v); err != nil {
			return err
		}
	}
	*f.val = v
	return err
}

func (f *float64Value) Get() any { return float64(*f.val) }

func (f *float64Value) String() string { return strconv.FormatFloat(float64(*f.val), 'g', -1, 64) }

// Int looks up the value of a local IntFlag, returns
// 0 if not found
func (cCtx *Context) Float64(name string) float64 {
	if v, ok := cCtx.Value(name).(float64); ok {
		return v
	}
	return 0
}
