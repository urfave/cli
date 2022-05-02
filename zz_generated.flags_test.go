// WARNING: this file is generated. DO NOT EDIT

package cli_test

import (
	"fmt"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestFloat64SliceFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.Float64SliceFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestGenericFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.GenericFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestGenericFlag_SatisfiesFmtStringerInterface(t *testing.T) {
	var f fmt.Stringer = &cli.GenericFlag{}

	_ = f.String()
}

func TestInt64SliceFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.Int64SliceFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestIntSliceFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.IntSliceFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestPathFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.PathFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestPathFlag_SatisfiesFmtStringerInterface(t *testing.T) {
	var f fmt.Stringer = &cli.PathFlag{}

	_ = f.String()
}

func TestStringSliceFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.StringSliceFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestTimestampFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.TimestampFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestTimestampFlag_SatisfiesFmtStringerInterface(t *testing.T) {
	var f fmt.Stringer = &cli.TimestampFlag{}

	_ = f.String()
}

func TestBoolFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.BoolFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestBoolFlag_SatisfiesFmtStringerInterface(t *testing.T) {
	var f fmt.Stringer = &cli.BoolFlag{}

	_ = f.String()
}

func TestFloat64Flag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.Float64Flag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestFloat64Flag_SatisfiesFmtStringerInterface(t *testing.T) {
	var f fmt.Stringer = &cli.Float64Flag{}

	_ = f.String()
}

func TestIntFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.IntFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestIntFlag_SatisfiesFmtStringerInterface(t *testing.T) {
	var f fmt.Stringer = &cli.IntFlag{}

	_ = f.String()
}

func TestInt64Flag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.Int64Flag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestInt64Flag_SatisfiesFmtStringerInterface(t *testing.T) {
	var f fmt.Stringer = &cli.Int64Flag{}

	_ = f.String()
}

func TestStringFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.StringFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestStringFlag_SatisfiesFmtStringerInterface(t *testing.T) {
	var f fmt.Stringer = &cli.StringFlag{}

	_ = f.String()
}

func TestDurationFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.DurationFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestDurationFlag_SatisfiesFmtStringerInterface(t *testing.T) {
	var f fmt.Stringer = &cli.DurationFlag{}

	_ = f.String()
}

func TestUintFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.UintFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestUintFlag_SatisfiesFmtStringerInterface(t *testing.T) {
	var f fmt.Stringer = &cli.UintFlag{}

	_ = f.String()
}

func TestUint64Flag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.Uint64Flag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestUint64Flag_SatisfiesFmtStringerInterface(t *testing.T) {
	var f fmt.Stringer = &cli.Uint64Flag{}

	_ = f.String()
}

// vim:ro
