package cli

import (
	"flag"
	"io/ioutil"
	"os"
	"strconv"
	"time"
)

type FlagSet struct {
	Flags []Flag
	Args  []string

	index map[string]Flag

	s *flag.FlagSet
}

func NewFlagSet(name string, flags []Flag, args []string) *FlagSet {
	s := flag.NewFlagSet(name, flag.ContinueOnError)
	s.SetOutput(ioutil.Discard)

	if flags == nil {
		flags = []Flag{}
	}

	if args == nil {
		args = []string{}
	}

	fs := &FlagSet{
		Flags: flags,
		Args:  args,

		index: map[string]Flag{},

		s: s,
	}
	fs.apply()
	return fs
}

func (fs *FlagSet) apply() {
	for _, f := range fs.Flags {
		f.Apply(fs)
		for _, name := range FlagNames(f) {
			fs.index[name] = f
		}
	}
}

func (fs *FlagSet) Parse() error {
	err := fs.s.Parse(fs.Args)
	if err != nil {
		return err
	}

	return fs.normalize()
}

func (fs *FlagSet) RemainingArgs() []string {
	return fs.s.Args()
}

func (fs *FlagSet) NumFlags() int {
	return fs.s.NFlag()
}

func (fs *FlagSet) IsSet(name string) bool {
	cliFlag := fs.Lookup(name)
	if cliFlag == nil {
		return false
	}

	isSet := false

	fs.s.Visit(func(f *flag.Flag) {
		if f.Name == name {
			isSet = true
		}
	})

	if isSet {
		return isSet
	}

	for _, envVar := range flagStringSliceField(cliFlag, "EnvVars") {
		if s := os.Getenv(envVar); s != "" {
			return true
		}
	}

	return false
}

func (fs *FlagSet) Each(f func(fl Flag)) {
	fs.s.Visit(func(ff *flag.Flag) {
		cliFlag := fs.Lookup(ff.Name)
		if cliFlag != nil {
			f(cliFlag)
		}
	})
}

func (fs *FlagSet) Lookup(name string) Flag {
	if cliFlag, ok := fs.index[name]; ok {
		return cliFlag
	}

	return nil
}

func (fs *FlagSet) getValue(name string) (flag.Value, error) {
	if _, ok := fs.index[name]; ok {
		if ff := fs.s.Lookup(name); ff != nil {
			return ff.Value, nil
		}
	}

	return nil, &MissingFlagError{Name: name}
}

func (fs *FlagSet) GetInt(name string) int {
	if v, err := fs.getValue(name); err == nil {
		val, err := strconv.Atoi(v.String())
		if err != nil {
			return 0
		}
		return val
	}

	return 0
}

func (fs *FlagSet) GetDuration(name string) time.Duration {
	if v, err := fs.getValue(name); err == nil {
		val, err := time.ParseDuration(v.String())
		if err == nil {
			return val
		}
	}

	return 0
}

func (fs *FlagSet) GetFloat64(name string) float64 {
	if v, err := fs.getValue(name); err == nil {
		val, err := strconv.ParseFloat(v.String(), 64)
		if err != nil {
			return 0
		}
		return val
	}

	return 0
}

func (fs *FlagSet) GetBool(name string) bool {
	if v, err := fs.getValue(name); err == nil {
		val, err := strconv.ParseBool(v.String())
		if err != nil {
			return false
		}
		return val
	}

	return false
}

func (fs *FlagSet) GetString(name string) string {
	if v, err := fs.getValue(name); err == nil {
		return v.String()
	}

	return ""
}

func (fs *FlagSet) GetStringSlice(name string) []string {
	if v, err := fs.getValue(name); err == nil {
		return (v.(*StringSlice)).Value()
	}

	return nil
}

func (fs *FlagSet) GetIntSlice(name string) []int {
	if v, err := fs.getValue(name); err == nil {
		return (v.(*IntSlice)).Value()
	}

	return nil
}

func (fs *FlagSet) GetGeneric(name string) interface{} {
	if v, err := fs.getValue(name); err == nil {
		return v
	}

	return nil
}

func (fs *FlagSet) SetString(name, value string) error {
	return fs.s.Set(name, value)
}

func (fs *FlagSet) DefStringSliceVar(ss *StringSlice, name, usage string) {
	fs.s.Var(ss, name, usage)
}

func (fs *FlagSet) DefIntSliceVar(is *IntSlice, name, usage string) {
	fs.s.Var(is, name, usage)
}

func (fs *FlagSet) DefGenericVar(g Generic, name, usage string) {
	fs.s.Var(g, name, usage)
}

func (fs *FlagSet) DefBoolVar(dest *bool, name string, value bool, usage string) {
	fs.s.BoolVar(dest, name, value, usage)
}

func (fs *FlagSet) DefBool(name string, value bool, usage string) {
	fs.s.Bool(name, value, usage)
}

func (fs *FlagSet) DefStringVar(dest *string, name, value, usage string) {
	fs.s.StringVar(dest, name, value, usage)
}

func (fs *FlagSet) DefString(name, value, usage string) {
	fs.s.String(name, value, usage)
}

func (fs *FlagSet) DefIntVar(dest *int, name string, value int, usage string) {
	fs.s.IntVar(dest, name, value, usage)
}

func (fs *FlagSet) DefInt(name string, value int, usage string) {
	fs.s.Int(name, value, usage)
}

func (fs *FlagSet) DefDurationVar(dest *time.Duration, name string, value time.Duration, usage string) {
	fs.s.DurationVar(dest, name, value, usage)
}

func (fs *FlagSet) DefDuration(name string, value time.Duration, usage string) {
	fs.s.Duration(name, value, usage)
}

func (fs *FlagSet) DefFloat64Var(dest *float64, name string, value float64, usage string) {
	fs.s.Float64Var(dest, name, value, usage)
}

func (fs *FlagSet) DefFloat64(name string, value float64, usage string) {
	fs.s.Float64(name, value, usage)
}

func (fs *FlagSet) normalize() error {
	visited := make(map[string]bool)

	fs.s.Visit(func(f *flag.Flag) {
		visited[f.Name] = true
	})

	for _, f := range fs.Flags {
		parts := FlagNames(f)
		if len(parts) == 1 {
			continue
		}

		var ff *flag.Flag

		for _, name := range parts {
			if _, ok := visited[name]; ok {
				if ff != nil {
					return &flagConflictError{Name: name}
				}
				ff = fs.s.Lookup(name)
			}
		}

		if ff == nil {
			continue
		}

		for _, name := range parts {
			if !visited[name] {
				fs.copyFlag(name, ff)
			}
		}
	}

	return nil
}

func (fs *FlagSet) copyFlag(name string, ff *flag.Flag) {
	switch ff.Value.(type) {
	case Serializeder:
		fs.s.Set(name, ff.Value.(Serializeder).Serialized())
	default:
		fs.s.Set(name, ff.Value.String())
	}
}

// Serializeder is used to circumvent the limitations of flag.FlagSet.Set
type Serializeder interface {
	Serialized() string
}
