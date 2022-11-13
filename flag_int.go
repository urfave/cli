package cli

import (
	"flag"
	"strconv"
)

// -- int Value
type intValue struct {
	val  *int
	base int
}

func (i intValue) Create(val int, p *int, c FlagConfig) flag.Value {
	*p = val
	return &intValue{
		val:  p,
		base: c.IntBase(),
	}
}

func (i *intValue) Set(s string) error {
	v, err := strconv.ParseInt(s, i.base, strconv.IntSize)
	if err != nil {
		return err
	}
	*i.val = int(v)
	return err
}

func (i *intValue) Get() any { return int(*i.val) }

func (i *intValue) String() string { return strconv.Itoa(int(*i.val)) }

type IntFlag = FlagBase[int, intValue]

// Int looks up the value of a local IntFlag, returns
// 0 if not found
func (cCtx *Context) Int(name string) int {
	if v, ok := cCtx.Value(name).(int); ok {
		return v
	}
	return 0
}
