package altinputsource

import (
	"time"

	"github.com/codegangsta/cli"
)

type MapFlagSetWrapper struct {
	wrappedFsw cli.FlagSetWrapper
	valueMap   map[string]interface{}
}

func NewDefaultValuesFlagSetWrapper(wrappedFsw cli.FlagSetWrapper, flags []cli.Flag) cli.FlagSetWrapper {
	valueMap := map[string]interface{}{}
	canProcessFlag := map[string]bool{}
	for _, f := range flags {
		fise, implementsType := f.(FlagInputSourceExtension)
		if implementsType {
			valueMap[f.GetName()] = fise.getDefaultValue()
			canProcessFlag[f.GetName()] = true
		}
	}

	return &MapFlagSetWrapper{wrappedFsw: wrappedFsw, valueMap: valueMap}
}

// Determines if the flag was actually set
func (fsm *MapFlagSetWrapper) HasFlag(name string) bool {
	return fsm.wrappedFsw.HasFlag(name)
}

// Determines if the flag was actually set
func (fsm *MapFlagSetWrapper) IsSet(name string) bool {
	if fsm.wrappedFsw.IsSet(name) {
		return true
	}

	_, exists := fsm.valueMap[name]
	return exists
}

// Returns the number of flags set
func (fsm *MapFlagSetWrapper) NumFlags() int {
	return fsm.wrappedFsw.NumFlags()
}

// Returns the command line arguments associated with the context.
func (fsm *MapFlagSetWrapper) Args() cli.Args {
	return fsm.wrappedFsw.Args()
}

func (fsm *MapFlagSetWrapper) Int(name string) int {
	value := fsm.wrappedFsw.Int(name)
	if value == 0 {
		otherGenericValue, exists := fsm.valueMap[name]
		if exists {
			otherValue, isType := otherGenericValue.(int)
			if isType {
				return otherValue
			}
		}
	}

	return value
}

func (fsm *MapFlagSetWrapper) Duration(name string) time.Duration {
	value := fsm.wrappedFsw.Duration(name)
	if value == 0 {
		otherGenericValue, exists := fsm.valueMap[name]
		if exists {
			otherValue, isType := otherGenericValue.(time.Duration)
			if isType {
				return otherValue
			}
		}
	}

	return value
}

func (fsm *MapFlagSetWrapper) Float64(name string) float64 {
	value := fsm.wrappedFsw.Float64(name)
	if value == 0 {
		otherGenericValue, exists := fsm.valueMap[name]
		if exists {
			otherValue, isType := otherGenericValue.(float64)
			if isType {
				return otherValue
			}
		}
	}

	return value
}

func (fsm *MapFlagSetWrapper) String(name string) string {
	value := fsm.wrappedFsw.String(name)
	if value == "" {
		otherGenericValue, exists := fsm.valueMap[name]
		if exists {
			otherValue, isType := otherGenericValue.(string)
			if isType {
				return otherValue
			}
		}
	}

	return value
}

func (fsm *MapFlagSetWrapper) StringSlice(name string) []string {
	value := fsm.wrappedFsw.StringSlice(name)
	if value == nil {
		otherGenericValue, exists := fsm.valueMap[name]
		if exists {
			otherValue, isType := otherGenericValue.([]string)
			if isType {
				return otherValue
			}
		}
	}

	return value
}

func (fsm *MapFlagSetWrapper) IntSlice(name string) []int {
	value := fsm.wrappedFsw.IntSlice(name)
	if value == nil {
		otherGenericValue, exists := fsm.valueMap[name]
		if exists {
			otherValue, isType := otherGenericValue.([]int)
			if isType {
				return otherValue
			}
		}
	}

	return value
}

func (fsm *MapFlagSetWrapper) Generic(name string) interface{} {
	value := fsm.wrappedFsw.Generic(name)
	if value == nil {
		otherGenericValue, exists := fsm.valueMap[name]
		if exists {
			return otherGenericValue
		}
	}

	return value
}

func (fsm *MapFlagSetWrapper) Bool(name string) bool {
	value := fsm.wrappedFsw.Bool(name)
	if !value {
		otherGenericValue, exists := fsm.valueMap[name]
		if exists {
			otherValue, isType := otherGenericValue.(bool)
			if isType {
				return otherValue
			}
		}
	}

	return value
}

func (fsm *MapFlagSetWrapper) BoolT(name string) bool {
	value := fsm.wrappedFsw.BoolT(name)
	if value {
		otherGenericValue, exists := fsm.valueMap[name]
		if exists {
			otherValue, isType := otherGenericValue.(bool)
			if isType {
				return otherValue
			}
		}
	}

	return value
}
