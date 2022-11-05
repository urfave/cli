package cli

import (
	"flag"
	"strconv"
)

<<<<<<< HEAD
// Apply populates the flag given the flag set and environment
func (f *Uint64Flag) Apply(set *flag.FlagSet) error {
	// set default value so that environment wont be able to overwrite it
	f.defaultValue = f.Value

	if val, source, found := flagFromEnvOrFile(f.EnvVars, f.FilePath); found {
		if val != "" {
			valInt, err := strconv.ParseUint(val, f.Base, 64)
			if err != nil {
				return fmt.Errorf("could not parse %q as uint64 value from %s for flag %s: %s", val, source, f.Name, err)
			}
=======
// -- uint64 Value
type uint64Value uint64
>>>>>>> Add all flags

func (i uint64Value) Create(val uint64, p *uint64) flag.Value {
	*p = val
	return (*uint64Value)(p)
}

func (i *uint64Value) Set(s string) error {
	v, err := strconv.ParseUint(s, 0, 64)
	if err != nil {
		return err
	}
	*i = uint64Value(v)
	return err
}

func (i *uint64Value) Get() any { return uint64(*i) }

func (i *uint64Value) String() string { return strconv.FormatUint(uint64(*i), 10) }

type Uint64Flag = flagImpl[uint64, uint64Value]

// Int64 looks up the value of a local Int64Flag, returns
// 0 if not found
func (cCtx *Context) Uint64(name string) uint64 {
	if v, ok := cCtx.Value(name).(uint64); ok {
		return v
	}
	return 0
}
