package cli

import (
	"strconv"
)

type FloatFlag = FlagBase[float64, NoConfig, floatValue]

// -- float64 Value
type floatValue float64

// Below functions are to satisfy the ValueCreator interface

func (f floatValue) Create(val float64, p *float64, c NoConfig) Value {
	*p = val
	return (*floatValue)(p)
}

func (f floatValue) ToString(b float64) string {
	return strconv.FormatFloat(b, 'g', -1, 64)
}

// Below functions are to satisfy the flag.Value interface

func (f *floatValue) Set(s string) error {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return err
	}
	*f = floatValue(v)
	return err
}

func (f *floatValue) Get() any { return float64(*f) }

func (f *floatValue) String() string { return strconv.FormatFloat(float64(*f), 'g', -1, 64) }

// Float looks up the value of a local FloatFlag, returns
// 0 if not found
func (cmd *Command) Float(name string) float64 {
	if v, ok := cmd.Value(name).(float64); ok {
		tracef("float available for flag name %[1]q with value=%[2]v (cmd=%[3]q)", name, v, cmd.Name)
		return v
	}

	tracef("float NOT available for flag name %[1]q (cmd=%[2]q)", name, cmd.Name)
	return 0
}
