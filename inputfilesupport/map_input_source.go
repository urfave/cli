package inputfilesupport

import (
	"time"

	"github.com/codegangsta/cli"
)

// MapInputSource implements InputSourceContext to return
// data from the map that is loaded.
// TODO: Didn't implement a way to write out various errors
// need to figure this part out.
type MapInputSource struct {
	valueMap map[string]interface{}
}

// Int returns an int from the map if it exists otherwise returns 0
func (fsm *MapInputSource) Int(name string) int {
	otherGenericValue, exists := fsm.valueMap[name]
	if exists {
		otherValue, isType := otherGenericValue.(int)
		if isType {
			return otherValue
		}
	}

	return 0
}

// Duration returns a duration from the map if it exists otherwise returns 0
func (fsm *MapInputSource) Duration(name string) time.Duration {
	otherGenericValue, exists := fsm.valueMap[name]
	if exists {
		otherValue, isType := otherGenericValue.(time.Duration)
		if isType {
			return otherValue
		}
	}

	return 0
}

// Float64 returns an float64 from the map if it exists otherwise returns 0
func (fsm *MapInputSource) Float64(name string) float64 {
	otherGenericValue, exists := fsm.valueMap[name]
	if exists {
		otherValue, isType := otherGenericValue.(float64)
		if isType {
			return otherValue
		}
	}

	return 0
}

// String returns a string from the map if it exists otherwise returns an empty string
func (fsm *MapInputSource) String(name string) string {
	otherGenericValue, exists := fsm.valueMap[name]
	if exists {
		otherValue, isType := otherGenericValue.(string)
		if isType {
			return otherValue
		}
	}

	return ""
}

// StringSlice returns an []string from the map if it exists otherwise returns nil
func (fsm *MapInputSource) StringSlice(name string) []string {
	otherGenericValue, exists := fsm.valueMap[name]
	if exists {
		otherValue, isType := otherGenericValue.([]string)
		if isType {
			return otherValue
		}
	}

	return nil
}

// IntSlice returns an []int from the map if it exists otherwise returns nil
func (fsm *MapInputSource) IntSlice(name string) []int {
	otherGenericValue, exists := fsm.valueMap[name]
	if exists {
		otherValue, isType := otherGenericValue.([]int)
		if isType {
			return otherValue
		}
	}

	return nil
}

// Generic returns an cli.Generic from the map if it exists otherwise returns nil
func (fsm *MapInputSource) Generic(name string) cli.Generic {
	otherGenericValue, exists := fsm.valueMap[name]
	if exists {
		otherValue, isType := otherGenericValue.(cli.Generic)
		if isType {
			return otherValue
		}
	}

	return nil
}

// Bool returns an bool from the map otherwise returns false
func (fsm *MapInputSource) Bool(name string) bool {
	otherGenericValue, exists := fsm.valueMap[name]
	if exists {
		otherValue, isType := otherGenericValue.(bool)
		if isType {
			return otherValue
		}
	}

	return false
}

// BoolT returns an bool from the map otherwise returns true
func (fsm *MapInputSource) BoolT(name string) bool {
	otherGenericValue, exists := fsm.valueMap[name]
	if exists {
		otherValue, isType := otherGenericValue.(bool)
		if isType {
			return otherValue
		}
	}

	return true
}
