package cli

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// MapBase wraps map[string]T to satisfy flag.Value
type MapBase[T any, C any, VC ValueCreator[T, C]] struct {
	dict             *map[string]T
	hasBeenSet       bool
	value            Value
	multiValueConfig multiValueParsingConfig
}

func (i MapBase[T, C, VC]) Create(val map[string]T, p *map[string]T, c C) Value {
	*p = map[string]T{}
	for k, v := range val {
		(*p)[k] = v
	}
	var t T
	np := new(T)
	var vc VC
	return &MapBase[T, C, VC]{
		dict:  p,
		value: vc.Create(t, np, c),
	}
}

// NewMapBase makes a *MapBase with default values
func NewMapBase[T any, C any, VC ValueCreator[T, C]](defaults map[string]T) *MapBase[T, C, VC] {
	return &MapBase[T, C, VC]{
		dict: &defaults,
	}
}

// configuration of slicing
func (i *MapBase[T, C, VC]) setMultiValueParsingConfig(c multiValueParsingConfig) {
	i.multiValueConfig = c
	mvc := &i.multiValueConfig
	tracef(
		"set map parsing config - keyValueSeparator '%s', slice separator '%s', disable separator:%v",
		mvc.MapFlagKeyValueSeparator,
		mvc.SliceFlagSeparator,
		mvc.DisableSliceFlagSeparator,
	)
}

// Set parses the value and appends it to the list of values
func (i *MapBase[T, C, VC]) Set(value string) error {
	if !i.hasBeenSet {
		*i.dict = map[string]T{}
		i.hasBeenSet = true
	}

	if strings.HasPrefix(value, slPfx) {
		// Deserializing assumes overwrite
		_ = json.Unmarshal([]byte(strings.Replace(value, slPfx, "", 1)), &i.dict)
		i.hasBeenSet = true
		return nil
	}

	mvc := &i.multiValueConfig
	keyValueSeparator := mvc.MapFlagKeyValueSeparator
	if len(keyValueSeparator) == 0 {
		keyValueSeparator = defaultMapFlagKeyValueSeparator
	}

	tracef(
		"splitting map value '%s', keyValueSeparator '%s', slice separator '%s', disable separator:%v",
		value,
		keyValueSeparator,
		mvc.SliceFlagSeparator,
		mvc.DisableSliceFlagSeparator,
	)
	for _, item := range flagSplitMultiValues(value, mvc.SliceFlagSeparator, mvc.DisableSliceFlagSeparator) {
		key, value, ok := strings.Cut(item, keyValueSeparator)
		if !ok {
			return fmt.Errorf("item %q is missing separator %q", item, keyValueSeparator)
		}
		if err := i.value.Set(value); err != nil {
			return err
		}
		(*i.dict)[key] = i.value.Get().(T)
	}

	return nil
}

// String returns a readable representation of this value (for usage defaults)
func (i *MapBase[T, C, VC]) String() string {
	v := i.Value()
	var t T
	if reflect.TypeOf(t).Kind() == reflect.String {
		return fmt.Sprintf("%v", v)
	}
	return fmt.Sprintf("%T{%s}", v, i.ToString(v))
}

// Serialize allows MapBase to fulfill Serializer
func (i *MapBase[T, C, VC]) Serialize() string {
	jsonBytes, _ := json.Marshal(i.dict)
	return fmt.Sprintf("%s%s", slPfx, string(jsonBytes))
}

// Value returns the mapping of values set by this flag
func (i *MapBase[T, C, VC]) Value() map[string]T {
	if i.dict == nil {
		return map[string]T{}
	}
	return *i.dict
}

// Get returns the mapping of values set by this flag
func (i *MapBase[T, C, VC]) Get() interface{} {
	return *i.dict
}

func (i MapBase[T, C, VC]) ToString(t map[string]T) string {
	var defaultVals []string
	var vc VC
	for _, k := range sortedKeys(t) {
		defaultVals = append(defaultVals, k+defaultMapFlagKeyValueSeparator+vc.ToString(t[k]))
	}
	return strings.Join(defaultVals, ", ")
}

func sortedKeys[T any](dict map[string]T) []string {
	keys := make([]string, 0, len(dict))
	for k := range dict {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
