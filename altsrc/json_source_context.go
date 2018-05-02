package altsrc

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"gopkg.in/urfave/cli.v1"
)

// NewJSONSourceFromFlagFunc returns a func that takes a cli.Context
// and returns an InputSourceContext suitable for retrieving config
// variables from a file containing JSON data with the file name defined
// by the given flag.
func NewJSONSourceFromFlagFunc(flag string) func(c *cli.Context) (InputSourceContext, error) {
	return func(context *cli.Context) (InputSourceContext, error) {
		return NewJSONSourceFromFile(context.String(flag))
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
	data, err := ioutil.ReadAll(r)
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
	case float64:
		return int(float64(v)), nil
	case float32:
		return int(float32(v)), nil
	}
}

func (x *jsonSource) Duration(name string) (time.Duration, error) {
	i, err := x.getValue(name)
	if err != nil {
		return 0, err
	}
	v, ok := (time.Duration)(0), false
	if v, ok = i.(time.Duration); !ok {
		return v, fmt.Errorf("unexpected type %T for %q", i, name)
	}
	return v, nil
}

func (x *jsonSource) Float64(name string) (float64, error) {
	i, err := x.getValue(name)
	if err != nil {
		return 0, err
	}
	v, ok := (float64)(0), false
	if v, ok = i.(float64); !ok {
		return v, fmt.Errorf("unexpected type %T for %q", i, name)
	}
	return v, nil
}

func (x *jsonSource) String(name string) (string, error) {
	i, err := x.getValue(name)
	if err != nil {
		return "", err
	}
	v, ok := "", false
	if v, ok = i.(string); !ok {
		return v, fmt.Errorf("unexpected type %T for %q", i, name)
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

func (x *jsonSource) Generic(name string) (cli.Generic, error) {
	i, err := x.getValue(name)
	if err != nil {
		return nil, err
	}
	v, ok := (cli.Generic)(nil), false
	if v, ok = i.(cli.Generic); !ok {
		return v, fmt.Errorf("unexpected type %T for %q", i, name)
	}
	return v, nil
}

func (x *jsonSource) Bool(name string) (bool, error) {
	i, err := x.getValue(name)
	if err != nil {
		return false, err
	}
	v, ok := false, false
	if v, ok = i.(bool); !ok {
		return v, fmt.Errorf("unexpected type %T for %q", i, name)
	}
	return v, nil
}

// since this source appears to require all configuration to be specified, the
// concept of a boolean defaulting to true seems inconsistent with no defaults
func (x *jsonSource) BoolT(name string) (bool, error) {
	return false, fmt.Errorf("unsupported type BoolT for JSONSource")
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
	deserialized map[string]interface{}
}
