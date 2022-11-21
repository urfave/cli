package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"reflect"
	"strings"
)

// SliceBase wraps []T to satisfy flag.Value
type SliceBase[T any, VC ValueCreator[T]] struct {
	slice      *[]T
	hasBeenSet bool
	value      flag.Value
}

func (i SliceBase[T, VC]) Create(val []T, p *[]T, c FlagConfig) flag.Value {
	*p = []T{}
	*p = append(*p, val...)
	var t T
	np := new(T)
	var vc VC
	return &SliceBase[T, VC]{
		slice: p,
		value: vc.Create(t, np, c),
	}
}

// NewIntSlice makes an *IntSlice with default values
func NewSliceBase[T any, VC ValueCreator[T]](defaults ...T) *SliceBase[T, VC] {
	return &SliceBase[T, VC]{
		slice: &defaults,
	}
}

// TODO: Consistently have specific Set function for Int64 and Float64 ?
// SetInt directly adds an integer to the list of values
func (i *SliceBase[T, VC]) SetOne(value T) {
	if !i.hasBeenSet {
		*i.slice = []T{}
		i.hasBeenSet = true
	}

	*i.slice = append(*i.slice, value)
}

// Set parses the value into an integer and appends it to the list of values
func (i *SliceBase[T, VC]) Set(value string) error {
	if !i.hasBeenSet {
		*i.slice = []T{}
		i.hasBeenSet = true
	}

	if strings.HasPrefix(value, slPfx) {
		// Deserializing assumes overwrite
		_ = json.Unmarshal([]byte(strings.Replace(value, slPfx, "", 1)), &i.slice)
		i.hasBeenSet = true
		return nil
	}

	for _, s := range flagSplitMultiValues(value) {
		if err := i.value.Set(strings.TrimSpace(s)); err != nil {
			return err
		}
		tmp, ok := i.value.(flag.Getter).Get().(T)
		if !ok {
			return fmt.Errorf("Unable to cast %v", i.value)
		}
		*i.slice = append(*i.slice, tmp)
	}

	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (i *SliceBase[T, VC]) String() string {
	v := *i.slice
	if v == nil {
		// treat nil the same as zero length non-nil
		v = make([]T, 0)
	}
	var t T
	if reflect.TypeOf(t).Kind() == reflect.String {
		return fmt.Sprintf("%v", v)
	}
	return fmt.Sprintf("%T{%s}", v, i.ToString(v))
}

// Serialize allows IntSlice to fulfill Serializer
func (i *SliceBase[T, VC]) Serialize() string {
	jsonBytes, _ := json.Marshal(i.slice)
	return fmt.Sprintf("%s%s", slPfx, string(jsonBytes))
}

// Value returns the slice of ints set by this flag
func (i *SliceBase[T, VC]) Value() []T {
	if i.slice == nil {
		return []T{}
	}
	return *i.slice
}

// Get returns the slice of ints set by this flag
func (i *SliceBase[T, VC]) Get() interface{} {
	return *i.slice
}

func (i SliceBase[T, VC]) ToString(t []T) string {
	var defaultVals []string
	var v VC
	for _, s := range t {
		defaultVals = append(defaultVals, v.ToString(s))
	}
	return strings.Join(defaultVals, ", ")
}
