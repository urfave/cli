package cli

import (
	"fmt"
	"strconv"
)

type Int64Flag = FlagBase[int64, IntegerConfig, int64Value]

// -- int64 Value
type int64Value struct {
	val       *int64
	base      int
	validator func(int64) error
}

// Below functions are to satisfy the ValueCreator interface

func (i int64Value) Create(val int64, p *int64, c IntegerConfig, validator func(int64) error) Value {
	*p = val
	return &int64Value{
		val:       p,
		base:      c.Base,
		validator: validator,
	}
}

func (i int64Value) ToString(b int64) string {
	return fmt.Sprintf("%d", b)
}

// Below functions are to satisfy the flag.Value interface

func (i *int64Value) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, 64)
	if err != nil {
		return err
	}
	if i.validator != nil {
		if err := i.validator(v); err != nil {
			return err
		}
	}
	*i.val = v
	return err
}

func (i *int64Value) Get() any { return int64(*i.val) }

func (i *int64Value) String() string { return strconv.FormatInt(int64(*i.val), 10) }

// Int64 looks up the value of a local Int64Flag, returns
// 0 if not found
func (cCtx *Context) Int64(name string) int64 {
	if v, ok := cCtx.Value(name).(int64); ok {
		return v
	}
	return 0
}
