package cli

import (
	"strconv"
)

type IntFlag = FlagBase[int64, IntegerConfig, intValue]

// IntegerConfig is the configuration for all integer type flags
type IntegerConfig struct {
	Base int
}

// -- int64 Value
type intValue struct {
	val  *int64
	base int
}

// Below functions are to satisfy the ValueCreator interface

func (i intValue) Create(val int64, p *int64, c IntegerConfig) Value {
	*p = val
	return &intValue{
		val:  p,
		base: c.Base,
	}
}

func (i intValue) ToString(b int64) string {
	return strconv.FormatInt(b, 10)
}

// Below functions are to satisfy the flag.Value interface

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, i.base, 64)
	if err != nil {
		return err
	}
	*i.val = v
	return err
}

func (i *intValue) Get() any { return int64(*i.val) }

func (i *intValue) String() string { return strconv.FormatInt(int64(*i.val), 10) }

// Int looks up the value of a local Int64Flag, returns
// 0 if not found
func (cmd *Command) Int(name string) int64 {
	if v, ok := cmd.Value(name).(int64); ok {
		tracef("int available for flag name %[1]q with value=%[2]v (cmd=%[3]q)", name, v, cmd.Name)
		return v
	}

	tracef("int NOT available for flag name %[1]q (cmd=%[2]q)", name, cmd.Name)
	return 0
}
