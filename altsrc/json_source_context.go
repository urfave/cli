package altsrc

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/urfave/cli/v2"
)

// NewJSONSourceFromFlagFunc returns a func that takes a cli.Context
// and returns an InputSourceContext suitable for retrieving config
// variables from a file containing JSON data with the file name defined
// by the given flag.
func NewJSONSourceFromFlagFunc(flag string) func(c *cli.Context) (InputSourceContext, error) {
	return func(cCtx *cli.Context) (InputSourceContext, error) {
		if cCtx.IsSet(flag) {
			return NewJSONSourceFromFile(cCtx.String(flag))
		}

		return defaultInputSource()
	}
}

// NewJSONSourceFromFile returns an InputSourceContext suitable for
// retrieving config variables from a file (or url) containing JSON
// data.
func NewJSONSourceFromFile(f string) (InputSourceContext, error) {
	data, err := loadDataFrom(f)
	if err != nil {
		return nil, err
	}

	return NewJSONSource(data)
}

// NewJSONSourceFromReader returns an InputSourceContext suitable for
// retrieving config variables from an io.Reader that returns JSON data.
func NewJSONSourceFromReader(r io.Reader) (InputSourceContext, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return NewJSONSource(data)
}

// NewJSONSource returns an InputSourceContext suitable for retrieving
// config variables from raw JSON data.
func NewJSONSource(data []byte) (InputSourceContext, error) {
	var deserialized map[string]interface{}
	if err := json.Unmarshal(data, &deserialized); err != nil {
		return nil, err
	}
	return &jsonSource{deserialized: deserialized}, nil
}

func (x *jsonSource) Source() string {
	return x.file
}

func (x *jsonSource) Int(name string) (int, error) {
	i, err := x.getValue(name)
	if err != nil {
		return 0, err
	}
	switch v := i.(type) {
	default:
		return 0, fmt.Errorf("unexpected type %T for %q", i, name)
	case int:
		return v, nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	}
}

func (x *jsonSource) Int64(name string) (int64, error) {
	i, err := x.getValue(name)
	if err != nil {
		return 0, err
	}
	switch v := i.(type) {
	default:
		return 0, fmt.Errorf("unexpected type %T for %q", i, name)
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	}
}

func (x *jsonSource) Uint(name string) (uint, error) {
	i, err := x.getValue(name)
	if err != nil {
		return 0, err
	}
	switch v := i.(type) {
	default:
		return 0, fmt.Errorf("unexpected type %T for %q", i, name)
	case uint:
		return v, nil
	case uint64:
		return uint(v), nil
	case float32:
		return uint(v), nil
	case float64:
		return uint(v), nil
	}
}

func (x *jsonSource) Uint64(name string) (uint64, error) {
	i, err := x.getValue(name)
	if err != nil {
		return 0, err
	}
	switch v := i.(type) {
	default:
		return 0, fmt.Errorf("unexpected type %T for %q", i, name)
	case uint64:
		return v, nil
	case uint:
		return uint64(v), nil
	case float32:
		return uint64(v), nil
	case float64:
		return uint64(v), nil
	}
}

func (x *jsonSource) Duration(name string) (time.Duration, error) {
	i, err := x.getValue(name)
	if err != nil {
		return 0, err
	}
	v, ok := i.(time.Duration)
	if !ok {
		return 0, fmt.Errorf("unexpected type %T for %q", i, name)
	}
	return v, nil
}

func (x *jsonSource) Float64(name string) (float64, error) {
	i, err := x.getValue(name)
	if err != nil {
		return 0, err
	}
	v, ok := i.(float64)
	if !ok {
		return 0, fmt.Errorf("unexpected type %T for %q", i, name)
	}
	return v, nil
}

func (x *jsonSource) String(name string) (string, error) {
	i, err := x.getValue(name)
	if err != nil {
		return "", err
	}
	v, ok := i.(string)
	if !ok {
		return "", fmt.Errorf("unexpected type %T for %q", i, name)
	}
	return v, nil
}

func (x *jsonSource) StringSlice(name string) ([]string, error) {
	i, err := x.getValue(name)
	if err != nil {
		return nil, err
	}
	switch v := i.(type) {
	default:
		return nil, fmt.Errorf("unexpected type %T for %q", i, name)
	case []string:
		return v, nil
	case []interface{}:
		c := []string{}
		for _, s := range v {
			if str, ok := s.(string); ok {
				c = append(c, str)
			} else {
				return c, fmt.Errorf("unexpected item type %T in %T for %q", s, c, name)
			}
		}
		return c, nil
	}
}

func (x *jsonSource) IntSlice(name string) ([]int, error) {
	i, err := x.getValue(name)
	if err != nil {
		return nil, err
	}
	switch v := i.(type) {
	default:
		return nil, fmt.Errorf("unexpected type %T for %q", i, name)
	case []int:
		return v, nil
	case []interface{}:
		c := []int{}
		for _, s := range v {
			if i2, ok := s.(int); ok {
				c = append(c, i2)
			} else {
				return c, fmt.Errorf("unexpected item type %T in %T for %q", s, c, name)
			}
		}
		return c, nil
	}
}

func (x *jsonSource) Int64Slice(name string) ([]int64, error) {
	i, err := x.getValue(name)
	if err != nil {
		return nil, err
	}
	switch v := i.(type) {
	default:
		return nil, fmt.Errorf("unexpected type %T for %q", i, name)
	case []int64:
		return v, nil
	case []interface{}:
		c := []int64{}
		for _, s := range v {
			if i2, ok := s.(int64); ok {
				c = append(c, i2)
			} else {
				return c, fmt.Errorf("unexpected item type %T in %T for %q", s, c, name)
			}
		}
		return c, nil
	}
}

func (x *jsonSource) Float64Slice(name string) ([]float64, error) {
	i, err := x.getValue(name)
	if err != nil {
		return nil, err
	}
	switch v := i.(type) {
	default:
		return nil, fmt.Errorf("unexpected type %T for %q", i, name)
	case []float64:
		return v, nil
	case []interface{}:
		c := []float64{}
		for _, s := range v {
			if i2, ok := s.(float64); ok {
				c = append(c, i2)
			} else {
				return c, fmt.Errorf("unexpected item type %T in %T for %q", s, c, name)
			}
		}
		return c, nil
	}
}

func (x *jsonSource) Generic(name string) (cli.Generic, error) {
	i, err := x.getValue(name)
	if err != nil {
		return nil, err
	}
	v, ok := i.(cli.Generic)
	if !ok {
		return nil, fmt.Errorf("unexpected type %T for %q", i, name)
	}
	return v, nil
}

func (x *jsonSource) Bool(name string) (bool, error) {
	i, err := x.getValue(name)
	if err != nil {
		return false, err
	}
	v, ok := i.(bool)
	if !ok {
		return false, fmt.Errorf("unexpected type %T for %q", i, name)
	}
	return v, nil
}

func (x *jsonSource) isSet(name string) bool {
	_, err := x.getValue(name)
	return err == nil
}

func (x *jsonSource) getValue(key string) (interface{}, error) {
	return jsonGetValue(key, x.deserialized)
}

func jsonGetValue(key string, m map[string]interface{}) (interface{}, error) {
	var ret interface{}
	var ok bool
	working := m
	keys := strings.Split(key, ".")
	for ix, k := range keys {
		if ret, ok = working[k]; !ok {
			return ret, fmt.Errorf("missing key %q", key)
		}
		if working, ok = ret.(map[string]interface{}); !ok {
			if ix < len(keys)-1 {
				return ret, fmt.Errorf("unexpected intermediate value at %q segment of %q: %T", k, key, ret)
			}
		}
	}
	return ret, nil
}

type jsonSource struct {
	file         string
	deserialized map[string]interface{}
}
