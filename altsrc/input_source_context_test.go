package altsrc

import (
	"time"

	"github.com/urfave/cli/v2"
)

type testInputSource struct {
	file     string
	valueMap map[interface{}]interface{}
}

func newTestInputSource(file string, valueMap map[interface{}]interface{}) *testInputSource {
	return &testInputSource{file: file, valueMap: valueMap}
}

func (tis testInputSource) Source() string {
	return tis.file
}

func (tis testInputSource) Int(name string) (int, error) {
	return 0, nil
}

func (tis testInputSource) Duration(name string) (time.Duration, error) {
	return 0, nil
}

func (tis testInputSource) Float64(name string) (float64, error) {
	return 0, nil
}

func (tis testInputSource) Int64(name string) (int64, error) {
	return 0, nil
}

func (tis testInputSource) Uint(name string) (uint, error) {
	return 0, nil
}

func (tis testInputSource) Uint64(name string) (uint64, error) {
	return 0, nil
}

func (tis testInputSource) String(name string) (string, error) {
	return "test", nil
}

func (tis testInputSource) StringSlice(name string) ([]string, error) {
	var stringSlice = make([]string, 0)
	return stringSlice, nil
}

func (tis testInputSource) IntSlice(name string) ([]int, error) {
	var intSlice = make([]int, 0)
	return intSlice, nil
}

func (tis testInputSource) Int64Slice(name string) ([]int64, error) {
	var int64Slice = make([]int64, 0)
	return int64Slice, nil
}

func (tis testInputSource) Float64Slice(name string) ([]float64, error) {
	var float64Slice = make([]float64, 0)
	return float64Slice, nil
}

func (tis testInputSource) Generic(name string) (cli.Generic, error) {
	return nil, nil
}

func (tis testInputSource) Bool(name string) (bool, error) {
	return false, nil
}

func (tis testInputSource) isSet(name string) bool {
	_, exists := tis.valueMap[name]
	return exists
}
