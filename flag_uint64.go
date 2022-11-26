package cli

import (
	"fmt"
	"strconv"
)

type Uint64Flag = FlagBase[uint64, IntegerConfig, uint64Value]

// -- uint64 Value
type uint64Value struct {
	val  *uint64
	base int
}

// Below functions are to satisfy the ValueCreator interface

func (i uint64Value) Create(val uint64, p *uint64, c IntegerConfig) Value {
	*p = val
	return &uint64Value{
		val:  p,
		base: c.Base,
	}
}

func (i uint64Value) ToString(b uint64) string {
	return fmt.Sprintf("%d", b)
}

// Below functions are to satisfy the flag.Value interface

func (i *uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, i.base, 64)
	if err != nil {
		return err
	}
	*i.val = v
	return err
}

func (i *uint64Value) Get() any { return uint64(*i.val) }

func (i *uint64Value) String() string { return strconv.FormatUint(uint64(*i.val), 10) }

// Uint64 looks up the value of a local Uint64Flag, returns
// 0 if not found
func (cCtx *Context) Uint64(name string) uint64 {
	if v, ok := cCtx.GetValue(name).(uint64); ok {
		return v
	}
	return 0
}
