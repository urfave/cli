// WARNING: this file is generated. DO NOT EDIT

package cli_test

import (
	"fmt"
	"testing"

	"github.com/urfave/cli/v3"
)

func TestFloat64SliceFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.Float64SliceFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestFloat64SliceFlag_SatisfiesRequiredFlagInterface(t *testing.T) {
	var f cli.RequiredFlag = &cli.Float64SliceFlag{}

	_ = f.IsRequired()
}

func TestFloat64SliceFlag_SatisfiesVisibleFlagInterface(t *testing.T) {
	var f cli.VisibleFlag = &cli.Float64SliceFlag{}

	_ = f.IsVisible()
}

func TestGenericFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.GenericFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestGenericFlag_SatisfiesRequiredFlagInterface(t *testing.T) {
	var f cli.RequiredFlag = &cli.GenericFlag{}

	_ = f.IsRequired()
}

func TestGenericFlag_SatisfiesVisibleFlagInterface(t *testing.T) {
	var f cli.VisibleFlag = &cli.GenericFlag{}

	_ = f.IsVisible()
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

func TestInt64SliceFlag_SatisfiesRequiredFlagInterface(t *testing.T) {
	var f cli.RequiredFlag = &cli.Int64SliceFlag{}

	_ = f.IsRequired()
}

func TestInt64SliceFlag_SatisfiesVisibleFlagInterface(t *testing.T) {
	var f cli.VisibleFlag = &cli.Int64SliceFlag{}

	_ = f.IsVisible()
}

func TestIntSliceFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.IntSliceFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestIntSliceFlag_SatisfiesRequiredFlagInterface(t *testing.T) {
	var f cli.RequiredFlag = &cli.IntSliceFlag{}

	_ = f.IsRequired()
}

func TestIntSliceFlag_SatisfiesVisibleFlagInterface(t *testing.T) {
	var f cli.VisibleFlag = &cli.IntSliceFlag{}

	_ = f.IsVisible()
}

func TestPathFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.PathFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestPathFlag_SatisfiesRequiredFlagInterface(t *testing.T) {
	var f cli.RequiredFlag = &cli.PathFlag{}

	_ = f.IsRequired()
}

func TestPathFlag_SatisfiesVisibleFlagInterface(t *testing.T) {
	var f cli.VisibleFlag = &cli.PathFlag{}

	_ = f.IsVisible()
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

func TestStringSliceFlag_SatisfiesRequiredFlagInterface(t *testing.T) {
	var f cli.RequiredFlag = &cli.StringSliceFlag{}

	_ = f.IsRequired()
}

func TestStringSliceFlag_SatisfiesVisibleFlagInterface(t *testing.T) {
	var f cli.VisibleFlag = &cli.StringSliceFlag{}

	_ = f.IsVisible()
}

func TestTimestampFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.TimestampFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestTimestampFlag_SatisfiesRequiredFlagInterface(t *testing.T) {
	var f cli.RequiredFlag = &cli.TimestampFlag{}

	_ = f.IsRequired()
}

func TestTimestampFlag_SatisfiesVisibleFlagInterface(t *testing.T) {
	var f cli.VisibleFlag = &cli.TimestampFlag{}

	_ = f.IsVisible()
}

func TestTimestampFlag_SatisfiesFmtStringerInterface(t *testing.T) {
	var f fmt.Stringer = &cli.TimestampFlag{}

	_ = f.String()
}

func TestUint64SliceFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.Uint64SliceFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestUintSliceFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.UintSliceFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestBoolFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.BoolFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestBoolFlag_SatisfiesRequiredFlagInterface(t *testing.T) {
	var f cli.RequiredFlag = &cli.BoolFlag{}

	_ = f.IsRequired()
}

func TestBoolFlag_SatisfiesVisibleFlagInterface(t *testing.T) {
	var f cli.VisibleFlag = &cli.BoolFlag{}

	_ = f.IsVisible()
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

func TestFloat64Flag_SatisfiesRequiredFlagInterface(t *testing.T) {
	var f cli.RequiredFlag = &cli.Float64Flag{}

	_ = f.IsRequired()
}

func TestFloat64Flag_SatisfiesVisibleFlagInterface(t *testing.T) {
	var f cli.VisibleFlag = &cli.Float64Flag{}

	_ = f.IsVisible()
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

func TestIntFlag_SatisfiesRequiredFlagInterface(t *testing.T) {
	var f cli.RequiredFlag = &cli.IntFlag{}

	_ = f.IsRequired()
}

func TestIntFlag_SatisfiesVisibleFlagInterface(t *testing.T) {
	var f cli.VisibleFlag = &cli.IntFlag{}

	_ = f.IsVisible()
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

func TestInt64Flag_SatisfiesRequiredFlagInterface(t *testing.T) {
	var f cli.RequiredFlag = &cli.Int64Flag{}

	_ = f.IsRequired()
}

func TestInt64Flag_SatisfiesVisibleFlagInterface(t *testing.T) {
	var f cli.VisibleFlag = &cli.Int64Flag{}

	_ = f.IsVisible()
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

func TestStringFlag_SatisfiesRequiredFlagInterface(t *testing.T) {
	var f cli.RequiredFlag = &cli.StringFlag{}

	_ = f.IsRequired()
}

func TestStringFlag_SatisfiesVisibleFlagInterface(t *testing.T) {
	var f cli.VisibleFlag = &cli.StringFlag{}

	_ = f.IsVisible()
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

func TestDurationFlag_SatisfiesRequiredFlagInterface(t *testing.T) {
	var f cli.RequiredFlag = &cli.DurationFlag{}

	_ = f.IsRequired()
}

func TestDurationFlag_SatisfiesVisibleFlagInterface(t *testing.T) {
	var f cli.VisibleFlag = &cli.DurationFlag{}

	_ = f.IsVisible()
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

func TestUintFlag_SatisfiesRequiredFlagInterface(t *testing.T) {
	var f cli.RequiredFlag = &cli.UintFlag{}

	_ = f.IsRequired()
}

func TestUintFlag_SatisfiesVisibleFlagInterface(t *testing.T) {
	var f cli.VisibleFlag = &cli.UintFlag{}

	_ = f.IsVisible()
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

func TestUint64Flag_SatisfiesRequiredFlagInterface(t *testing.T) {
	var f cli.RequiredFlag = &cli.Uint64Flag{}

	_ = f.IsRequired()
}

func TestUint64Flag_SatisfiesVisibleFlagInterface(t *testing.T) {
	var f cli.VisibleFlag = &cli.Uint64Flag{}

	_ = f.IsVisible()
}

func TestUint64Flag_SatisfiesFmtStringerInterface(t *testing.T) {
	var f fmt.Stringer = &cli.Uint64Flag{}

	_ = f.String()
}

// vim:ro
