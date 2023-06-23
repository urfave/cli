package cli

import (
	"fmt"
	"strconv"
)

type UintFlag = FlagBase[uint64, IntegerConfig, uintValue]

// -- uint64 Value
type uintValue struct {
	val  *uint64
	base int
}

// Below functions are to satisfy the ValueCreator interface

func (i uintValue) Create(val uint64, p *uint64, c IntegerConfig) Value {
	*p = val
	return &uintValue{
		val:  p,
		base: c.Base,
	}
}

func (i uintValue) ToString(b uint64) string {
	return fmt.Sprintf("%d", b)
}

// Below functions are to satisfy the flag.Value interface

func (i *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, i.base, 64)
	if err != nil {
		return err
	}
	*i.val = v
	return err
}

func (i *uintValue) Get() any { return uint64(*i.val) }

func (i *uintValue) String() string { return strconv.FormatUint(uint64(*i.val), 10) }

// Uint looks up the value of a local Uint64Flag, returns
// 0 if not found
func (cCtx *Context) Uint(name string) uint64 {
	if v, ok := cCtx.Value(name).(uint64); ok {
		return v
	}
	return 0
}
