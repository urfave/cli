package cli

import (
	"flag"
	"strconv"
)

// -- float64 Value
type float64Value float64

func (f float64Value) Create(val float64, p *float64, c FlagConfig) flag.Value {
	*p = val
	return (*float64Value)(p)
}

func (f *float64Value) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	*f = float64Value(v)
	return err
}

func (f *float64Value) Get() any { return float64(*f) }

func (f *float64Value) String() string { return strconv.FormatFloat(float64(*f), 'g', -1, 64) }

type Float64Flag = FlagBase[float64, float64Value]

// Int looks up the value of a local IntFlag, returns
// 0 if not found
func (cCtx *Context) Float64(name string) float64 {
	if v, ok := cCtx.Value(name).(float64); ok {
		return v
	}
	return 0
}
