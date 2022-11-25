package cli

import (
	"fmt"
	"strconv"
)

type UintFlag = FlagBase[uint, IntegerConfig, uintValue]

// -- uint Value
type uintValue struct {
	val  *uint
	base int
}

// Below functions are to satisfy the ValueCreator interface

func (i uintValue) Create(val uint, p *uint, c IntegerConfig) Value {
	*p = val
	return &uintValue{
		val:  p,
		base: c.Base,
	}
}

func (i uintValue) ToString(b uint) string {
	return fmt.Sprintf("%v", b)
}

// Below functions are to satisfy the flag.Value interface

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

// Int looks up the value of a local IntFlag, returns
// 0 if not found
func (cCtx *Context) Uint(name string) uint {
	if v, ok := cCtx.Value(name).(uint); ok {
		return v
	}
	return 0
}
