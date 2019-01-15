package altsrc

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"gopkg.in/urfave/cli.v1"
)

// MapInputSource implements InputSourceContext to return
// data from the map that is loaded.
type MapInputSource struct {
	valueMap map[interface{}]interface{}
}

// nestedVal checks if the name has '.' delimiters.
// If so, it tries to traverse the tree by the '.' delimited sections to find
// a nested value for the key.
func nestedVal(name string, tree map[interface{}]interface{}) (interface{}, bool) {
	if sections := strings.Split(name, "."); len(sections) > 1 {
		node := tree
		for _, section := range sections[:len(sections)-1] {
			child, ok := node[section]
			if !ok {
				return nil, false
			}
			ctype, ok := child.(map[interface{}]interface{})
			if !ok {
				return nil, false
			}
			node = ctype
		}
		if val, ok := node[sections[len(sections)-1]]; ok {
			return val, true
		}
	}
	return nil, false
}

func (fsm *MapInputSource) getValue(key string) (interface{}, bool) {
	parts := strings.Split(key, ",")
	var ret interface{}
	var exists bool = false
	for _, name := range parts {
		name = strings.Trim(name, " ")
		ret, exists = fsm.valueMap[name]
		if exists {
			break
		}
		ret, exists = nestedVal(name, fsm.valueMap)
		if exists {
			break
		}
	}
	return ret, exists
}

// Int returns an int from the map if it exists otherwise returns 0
func (fsm *MapInputSource) Int(name string) (int, error) {
	genericValue, exists := fsm.getValue(name)
	if exists {
		value, isType := genericValue.(int)
		if !isType {
			return 0, incorrectTypeForFlagError(name, "int", genericValue)
		}
		return value, nil
	}

	return 0, nil
}

// Duration returns a duration from the map if it exists otherwise returns 0
func (fsm *MapInputSource) Duration(name string) (time.Duration, error) {
	genericValue, exists := fsm.getValue(name)
	if exists {
		value, isType := genericValue.(time.Duration)
		if !isType {
			return 0, incorrectTypeForFlagError(name, "duration", genericValue)
		}
		return value, nil
	}

	return 0, nil
}

// Float64 returns an float64 from the map if it exists otherwise returns 0
func (fsm *MapInputSource) Float64(name string) (float64, error) {
	genericValue, exists := fsm.getValue(name)
	if exists {
		value, isType := genericValue.(float64)
		if !isType {
			return 0, incorrectTypeForFlagError(name, "float64", genericValue)
		}
		return value, nil
	}

	return 0, nil
}

// String returns a string from the map if it exists otherwise returns an empty string
func (fsm *MapInputSource) String(name string) (string, error) {
	genericValue, exists := fsm.getValue(name)
	if exists {
		value, isType := genericValue.(string)
		if !isType {
			return "", incorrectTypeForFlagError(name, "string", genericValue)
		}
		return value, nil
	}

	return "", nil
}

// StringSlice returns an []string from the map if it exists otherwise returns nil
func (fsm *MapInputSource) StringSlice(name string) ([]string, error) {
	genericValue, exists := fsm.getValue(name)
	if !exists {
		return nil, nil
	}

	value, isType := genericValue.([]interface{})
	if !isType {
		return nil, incorrectTypeForFlagError(name, "[]interface{}", genericValue)
	}

	var stringSlice = make([]string, 0, len(value))
	for i, v := range value {
		stringValue, isType := v.(string)

		if !isType {
			return nil, incorrectTypeForFlagError(fmt.Sprintf("%s[%d]", name, i), "string", v)
		}

		stringSlice = append(stringSlice, stringValue)
	}

	return stringSlice, nil
}

// IntSlice returns an []int from the map if it exists otherwise returns nil
func (fsm *MapInputSource) IntSlice(name string) ([]int, error) {
	genericValue, exists := fsm.getValue(name)
	if !exists {
		return nil, nil
	}

	value, isType := genericValue.([]interface{})
	if !isType {
		return nil, incorrectTypeForFlagError(name, "[]interface{}", genericValue)
	}

	var intSlice = make([]int, 0, len(value))
	for i, v := range value {
		intValue, isType := v.(int)

		if !isType {
			return nil, incorrectTypeForFlagError(fmt.Sprintf("%s[%d]", name, i), "int", v)
		}

		intSlice = append(intSlice, intValue)
	}

	return intSlice, nil
}

// Generic returns an cli.Generic from the map if it exists otherwise returns nil
func (fsm *MapInputSource) Generic(name string) (cli.Generic, error) {
	genericValue, exists := fsm.getValue(name)
	if exists {
		value, isType := genericValue.(cli.Generic)
		if !isType {
			return nil, incorrectTypeForFlagError(name, "cli.Generic", genericValue)
		}
		return value, nil
	}

	return nil, nil
}

// Bool returns an bool from the map otherwise returns false
func (fsm *MapInputSource) Bool(name string) (bool, error) {
	genericValue, exists := fsm.getValue(name)
	if exists {
		value, isType := genericValue.(bool)
		if !isType {
			return false, incorrectTypeForFlagError(name, "bool", genericValue)
		}
		return value, nil
	}

	return false, nil
}

// BoolT returns an bool from the map otherwise returns true
func (fsm *MapInputSource) BoolT(name string) (bool, error) {
	genericValue, exists := fsm.getValue(name)
	if exists {
		value, isType := genericValue.(bool)
		if !isType {
			return true, incorrectTypeForFlagError(name, "bool", genericValue)
		}
		return value, nil
	}

	return true, nil
}

func incorrectTypeForFlagError(name, expectedTypeName string, value interface{}) error {
	valueType := reflect.TypeOf(value)
	valueTypeName := ""
	if valueType != nil {
		valueTypeName = valueType.Name()
	}

	return fmt.Errorf("Mismatched type for flag '%s'. Expected '%s' but actual is '%s'", name, expectedTypeName, valueTypeName)
}
