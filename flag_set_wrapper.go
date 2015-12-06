package cli

import (
	"flag"
	"strconv"
	"time"
)

// FlagSetWrapper wraps the flag.FlagSet interface
// to allow customization of how data is selected and returned
// via the Context
type FlagSetWrapper interface {
	HasFlag(name string) bool
	IsSet(name string) bool
	NumFlags() int
	Args() Args
	Int(name string) int
	Duration(name string) time.Duration
	Float64(name string) float64
	String(name string) string
	StringSlice(name string) []string
	IntSlice(name string) []int
	Generic(name string) interface{}
	Bool(name string) bool
	BoolT(name string) bool
}

// DefaultFlagSetWrapper wraps the flag.FlagSet and provides
// implemention of the FlagSetWrapper interface
type DefaultFlagSetWrapper struct {
	set      *flag.FlagSet
	setFlags map[string]bool
}

func NewFlagSetWrapper(set *flag.FlagSet) FlagSetWrapper {
	return &DefaultFlagSetWrapper{set: set}
}

// Determines if the flag was actually set
func (fsm *DefaultFlagSetWrapper) HasFlag(name string) bool {
	if fsm.set.Lookup(name) != nil {
		return true
	}
	return false
}

// Determines if the flag was actually set
func (fsm *DefaultFlagSetWrapper) IsSet(name string) bool {
	if fsm.setFlags == nil {
		fsm.setFlags = make(map[string]bool)
		fsm.set.Visit(func(f *flag.Flag) {
			fsm.setFlags[f.Name] = true
		})
	}
	return fsm.setFlags[name] == true
}

// Returns the number of flags set
func (fsm *DefaultFlagSetWrapper) NumFlags() int {
	return fsm.set.NFlag()
}

// Returns the command line arguments associated with the context.
func (fsm *DefaultFlagSetWrapper) Args() Args {
	args := Args(fsm.set.Args())
	return args
}

func (fsm *DefaultFlagSetWrapper) Int(name string) int {
	f := fsm.set.Lookup(name)
	if f != nil {
		val, err := strconv.Atoi(f.Value.String())
		if err != nil {
			return 0
		}
		return val
	}

	return 0
}

func (fsm *DefaultFlagSetWrapper) Duration(name string) time.Duration {
	f := fsm.set.Lookup(name)
	if f != nil {
		val, err := time.ParseDuration(f.Value.String())
		if err == nil {
			return val
		}
	}

	return 0
}

func (fsm *DefaultFlagSetWrapper) Float64(name string) float64 {
	f := fsm.set.Lookup(name)
	if f != nil {
		val, err := strconv.ParseFloat(f.Value.String(), 64)
		if err != nil {
			return 0
		}
		return val
	}

	return 0
}

func (fsm *DefaultFlagSetWrapper) String(name string) string {
	f := fsm.set.Lookup(name)
	if f != nil {
		return f.Value.String()
	}

	return ""
}

func (fsm *DefaultFlagSetWrapper) StringSlice(name string) []string {
	f := fsm.set.Lookup(name)
	if f != nil {
		return (f.Value.(*StringSlice)).Value()

	}

	return nil
}

func (fsm *DefaultFlagSetWrapper) IntSlice(name string) []int {
	f := fsm.set.Lookup(name)
	if f != nil {
		return (f.Value.(*IntSlice)).Value()

	}

	return nil
}

func (fsm *DefaultFlagSetWrapper) Generic(name string) interface{} {
	f := fsm.set.Lookup(name)
	if f != nil {
		return f.Value
	}
	return nil
}

func (fsm *DefaultFlagSetWrapper) Bool(name string) bool {
	f := fsm.set.Lookup(name)
	if f != nil {
		val, err := strconv.ParseBool(f.Value.String())
		if err != nil {
			return false
		}
		return val
	}

	return false
}

func (fsm *DefaultFlagSetWrapper) BoolT(name string) bool {
	f := fsm.set.Lookup(name)
	if f != nil {
		val, err := strconv.ParseBool(f.Value.String())
		if err != nil {
			return true
		}
		return val
	}

	return false
}
