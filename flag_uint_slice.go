package cli

<<<<<<< HEAD
import (
	"flag"
)

type UintSlice = SliceBase[uint64, IntegerConfig, uintValue]
type UintSliceFlag = FlagBase[[]uint64, IntegerConfig, UintSlice]
=======
type UintSlice = SliceBase[uint, IntegerConfig, uintValue]
type UintSliceFlag = FlagBase[[]uint, IntegerConfig, UintSlice]
>>>>>>> d16cd7e... Add new fv

var NewUintSlice = NewSliceBase[uint64, IntegerConfig, uintValue]

// UintSlice looks up the value of a local UintSliceFlag, returns
// nil if not found
<<<<<<< HEAD
func (cCtx *Context) UintSlice(name string) []uint64 {
	if fs := cCtx.lookupFlagSet(name); fs != nil {
		return lookupUintSlice(name, fs)
	}
	return nil
}

func lookupUintSlice(name string, set *flag.FlagSet) []uint64 {
	f := set.Lookup(name)
	if f != nil {
		if slice, ok := f.Value.(flag.Getter).Get().([]uint64); ok {
			return slice
		}
=======
func (cCtx *Context) UintSlice(name string) []uint {
	if v, ok := cCtx.Value(name).([]uint); ok {
		return v
>>>>>>> d16cd7e... Add new fv
	}
	return nil
}
