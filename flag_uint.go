package cli

import (
	"flag"
	"strconv"
)

// -- uint Value
type uintValue uint

func (i uintValue) Create(val uint, p *uint) flag.Value {
	*p = val
	return (*uintValue)(p)
}

func (i *uintValue) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, strconv.IntSize)
	if err != nil {
		return err
	}
<<<<<<< HEAD
	return fmt.Sprintf("%d", f.defaultValue)
=======
	*i = uintValue(v)
	return err
>>>>>>> Add all flags
}

func (i *uintValue) Get() any { return uint(*i) }

func (i *uintValue) String() string { return strconv.FormatUint(uint64(*i), 10) }

type UintFlag = flagImpl[uint, uintValue]

// Int looks up the value of a local IntFlag, returns
// 0 if not found
func (cCtx *Context) Uint(name string) uint {
	if v, ok := cCtx.Value(name).(uint); ok {
		return v
	}
	return 0
}
