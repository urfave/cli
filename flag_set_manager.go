package cli

import (
	"flag"
	"strconv"
	"time"
)

type FlagSetManager interface {
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

// Context is a type that is passed through to
// each Handler action in a cli application. Context
// can be used to retrieve context-specific Args and
// parsed command-line options.
type DefaultFlagSetManager struct {
	set      *flag.FlagSet
	setFlags map[string]bool
}

func NewFlagSetManager(set *flag.FlagSet) FlagSetManager {
	return &DefaultFlagSetManager{set: set}
}

// Determines if the flag was actually set
func (fsm *DefaultFlagSetManager) HasFlag(name string) bool {
	if fsm.set.Lookup(name) != nil {
		return true
	}
	return false
}

// Determines if the flag was actually set
func (fsm *DefaultFlagSetManager) IsSet(name string) bool {
	if fsm.setFlags == nil {
		fsm.setFlags = make(map[string]bool)
		fsm.set.Visit(func(f *flag.Flag) {
			fsm.setFlags[f.Name] = true
		})
	}
	return fsm.setFlags[name] == true
}

// Returns the number of flags set
func (fsm *DefaultFlagSetManager) NumFlags() int {
	return fsm.set.NFlag()
}

// Returns the command line arguments associated with the context.
func (fsm *DefaultFlagSetManager) Args() Args {
	args := Args(fsm.set.Args())
	return args
}

func (fsm *DefaultFlagSetManager) Int(name string) int {
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

func (fsm *DefaultFlagSetManager) Duration(name string) time.Duration {
	f := fsm.set.Lookup(name)
	if f != nil {
		val, err := time.ParseDuration(f.Value.String())
		if err == nil {
			return val
		}
	}

	return 0
}

func (fsm *DefaultFlagSetManager) Float64(name string) float64 {
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

func (fsm *DefaultFlagSetManager) String(name string) string {
	f := fsm.set.Lookup(name)
	if f != nil {
		return f.Value.String()
	}

	return ""
}

func (fsm *DefaultFlagSetManager) StringSlice(name string) []string {
	f := fsm.set.Lookup(name)
	if f != nil {
		return (f.Value.(*StringSlice)).Value()

	}

	return nil
}

func (fsm *DefaultFlagSetManager) IntSlice(name string) []int {
	f := fsm.set.Lookup(name)
	if f != nil {
		return (f.Value.(*IntSlice)).Value()

	}

	return nil
}

func (fsm *DefaultFlagSetManager) Generic(name string) interface{} {
	f := fsm.set.Lookup(name)
	if f != nil {
		return f.Value
	}
	return nil
}

func (fsm *DefaultFlagSetManager) Bool(name string) bool {
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

func (fsm *DefaultFlagSetManager) BoolT(name string) bool {
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
