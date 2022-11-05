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

func TestUint64SliceFlag_SatisfiesRequiredFlagInterface(t *testing.T) {
	var f cli.RequiredFlag = &cli.Uint64SliceFlag{}

	_ = f.IsRequired()
}

func TestUint64SliceFlag_SatisfiesVisibleFlagInterface(t *testing.T) {
	var f cli.VisibleFlag = &cli.Uint64SliceFlag{}

	_ = f.IsVisible()
}

func TestUintSliceFlag_SatisfiesFlagInterface(t *testing.T) {
	var f cli.Flag = &cli.UintSliceFlag{}

	_ = f.IsSet()
	_ = f.Names()
}

func TestUintSliceFlag_SatisfiesRequiredFlagInterface(t *testing.T) {
	var f cli.RequiredFlag = &cli.UintSliceFlag{}

	_ = f.IsRequired()
}

func TestUintSliceFlag_SatisfiesVisibleFlagInterface(t *testing.T) {
	var f cli.VisibleFlag = &cli.UintSliceFlag{}

	_ = f.IsVisible()
}

// vim:ro
