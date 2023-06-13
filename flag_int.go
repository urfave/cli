package cli

import (
	"fmt"
	"strconv"
)

type IntFlag = FlagBase[int, IntegerConfig, intValue]

// IntegerConfig is the configuration for all integer type flags
type IntegerConfig struct {
	Base int
}

// -- int Value
type intValue struct {
	val       *int
	base      int
	validator func(int) error
}

// Below functions are to satisfy the ValueCreator interface

func (i intValue) Create(val int, p *int, config IntegerConfig, validator func(int) error) Value {
	*p = val
	return &intValue{
		val:       p,
		base:      config.Base,
		validator: validator,
	}
}

func (i intValue) ToString(b int) string {
	return fmt.Sprintf("%v", b)
}

// Below functions are to satisfy the flag.Value interface

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, i.base, strconv.IntSize)
	if err != nil {
		return err
	}
	if i.validator != nil {
		if err := i.validator(int(v)); err != nil {
			return err
		}
	}
	*i.val = int(v)
	return err
}

func (i *intValue) Get() any { return int(*i.val) }

func (i *intValue) String() string {
	if i == nil || i.val == nil {
		return ""
	}
	return strconv.Itoa(int(*i.val))
}

// Int looks up the value of a local IntFlag, returns
// 0 if not found
func (cCtx *Context) Int(name string) int {
	if v, ok := cCtx.Value(name).(int); ok {
		return v
	}
	return 0
}
