package cli

import (
	"flag"
	"strconv"
)

// -- int Value
type intValue int

func (i intValue) Create(val int, p *int) flag.Value {
	*p = val
	return (*intValue)(p)
}

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, 0, strconv.IntSize)
	if err != nil {
		return err
	}
	*i = intValue(v)
	return err
}

func (i *intValue) Get() any { return int(*i) }

func (i *intValue) String() string { return strconv.Itoa(int(*i)) }

type IntFlag = flagImpl[int, intValue]

// Int looks up the value of a local IntFlag, returns
// 0 if not found
func (cCtx *Context) Int(name string) int {
	if v, ok := cCtx.Value(name).(int); ok {
		return v
	}
	return 0
}
