package cli

import (
	"flag"
	"strconv"
)

// -- uint Value
type uintValue struct {
	val  *uint
	base int
}

func (i uintValue) Create(val uint, p *uint, c FlagConfig) flag.Value {
	*p = val
	return &uintValue{
		val:  p,
		base: c.GetNumberBase(),
	}
}

func (i *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, i.base, strconv.IntSize)
	if err != nil {
		return err
	}
	*i.val = uint(v)
	return err
}

func (i *uintValue) Get() any { return uint(*i.val) }

func (i *uintValue) String() string { return strconv.FormatUint(uint64(*i.val), 10) }

type UintFlag = FlagBase[uint, uintValue]

// Int looks up the value of a local IntFlag, returns
// 0 if not found
func (cCtx *Context) Uint(name string) uint {
	if v, ok := cCtx.Value(name).(uint); ok {
		return v
	}
	return 0
}
