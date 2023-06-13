package cli

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// SliceBase wraps []T to satisfy flag.Value
type SliceBase[T any, C any, VC ValueCreator[T, C]] struct {
	slice      *[]T
	hasBeenSet bool
	value      Value
	validator  func([]T) error
}

func (i SliceBase[T, C, VC]) Create(val []T, p *[]T, config C, validator func([]T) error) Value {
	*p = []T{}
	*p = append(*p, val...)
	var t T
	np := new(T)
	var vc VC
	return &SliceBase[T, C, VC]{
		slice:     p,
		value:     vc.Create(t, np, config, nil),
		validator: validator,
	}
}

// NewIntSlice makes an *IntSlice with default values
func NewSliceBase[T any, C any, VC ValueCreator[T, C]](defaults ...T) *SliceBase[T, C, VC] {
	return &SliceBase[T, C, VC]{
		slice: &defaults,
	}
}

// SetOne directly adds a value to the list of values
func (i *SliceBase[T, C, VC]) SetOne(value T) {
	if !i.hasBeenSet {
		*i.slice = []T{}
		i.hasBeenSet = true
	}

	*i.slice = append(*i.slice, value)
}

// Set parses the value and appends it to the list of values
func (i *SliceBase[T, C, VC]) Set(value string) error {
	var tmpSlice []T

	if !i.hasBeenSet {
		*i.slice = []T{}
		i.hasBeenSet = true
	}

	if strings.HasPrefix(value, slPfx) {
		// Deserializing assumes overwrite
		_ = json.Unmarshal([]byte(strings.Replace(value, slPfx, "", 1)), &tmpSlice)
		if i.validator != nil {
			if err := i.validator(tmpSlice); err != nil {
				return err
			}
		}
		i.slice = &tmpSlice
		i.hasBeenSet = true
		return nil
	}

	for _, s := range flagSplitMultiValues(value) {
		if err := i.value.Set(strings.TrimSpace(s)); err != nil {
			return err
		}
		tmp, ok := i.value.Get().(T)
		if !ok {
			return fmt.Errorf("unable to cast %v", i.value)
		}
		tmpSlice = append(tmpSlice, tmp)
	}

	if i.validator != nil {
		if err := i.validator(tmpSlice); err != nil {
			return err
		}
	}

	if !i.hasBeenSet {
		*i.slice = []T{}
		i.hasBeenSet = true
	}

	*i.slice = append(*i.slice, tmpSlice...)
	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (i *SliceBase[T, C, VC]) String() string {
	v := i.Value()
	var t T
	if reflect.TypeOf(t).Kind() == reflect.String {
		return fmt.Sprintf("%v", v)
	}
	return fmt.Sprintf("%T{%s}", v, i.ToString(v))
}

// Serialize allows SliceBase to fulfill Serializer
func (i *SliceBase[T, C, VC]) Serialize() string {
	jsonBytes, _ := json.Marshal(i.slice)
	return fmt.Sprintf("%s%s", slPfx, string(jsonBytes))
}

// Value returns the slice of values set by this flag
func (i *SliceBase[T, C, VC]) Value() []T {
	if i.slice == nil {
		return []T{}
	}
	return *i.slice
}

// Get returns the slice of values set by this flag
func (i *SliceBase[T, C, VC]) Get() interface{} {
	return *i.slice
}

func (i SliceBase[T, C, VC]) ToString(t []T) string {
	var defaultVals []string
	var v VC
	for _, s := range t {
		defaultVals = append(defaultVals, v.ToString(s))
	}
	return strings.Join(defaultVals, ", ")
}
