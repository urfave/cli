package cli

import (
	"flag"
	"strconv"
)

// -- uint64 Value
type uint64Value struct {
	val  *uint64
	base int
}

func (i uint64Value) Create(val uint64, p *uint64, c FlagConfig) flag.Value {
	*p = val
	return &uint64Value{
		val:  p,
		base: c.IntBase(),
	}
}

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

type Uint64Flag = FlagBase[uint64, uint64Value]

// Int64 looks up the value of a local Int64Flag, returns
// 0 if not found
func (cCtx *Context) Uint64(name string) uint64 {
	if v, ok := cCtx.Value(name).(uint64); ok {
		return v
	}
	return 0
}
