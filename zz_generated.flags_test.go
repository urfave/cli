// WARNING: this file is generated. DO NOT EDIT

package cli_test

import (
	"fmt"
	"testing"

	"github.com/urfave/cli/v2"
)

func TestFloat64SliceFlag_SatisfiesFlagInterface(t *testing.T) {
	var _ cli.Flag = &cli.Float64SliceFlag{}
}

func TestGenericFlag_SatisfiesFlagInterface(t *testing.T) {
	var _ cli.Flag = &cli.GenericFlag{}
}

func TestGenericFlag_SatisfiesFmtStringerInterface(t *testing.T) {
	var _ fmt.Stringer = &cli.GenericFlag{}
}

func TestInt64SliceFlag_SatisfiesFlagInterface(t *testing.T) {
	var _ cli.Flag = &cli.Int64SliceFlag{}
}

func TestIntSliceFlag_SatisfiesFlagInterface(t *testing.T) {
	var _ cli.Flag = &cli.IntSliceFlag{}
}

func TestPathFlag_SatisfiesFlagInterface(t *testing.T) {
	var _ cli.Flag = &cli.PathFlag{}
}

func TestPathFlag_SatisfiesFmtStringerInterface(t *testing.T) {
	var _ fmt.Stringer = &cli.PathFlag{}
}

func TestStringSliceFlag_SatisfiesFlagInterface(t *testing.T) {
	var _ cli.Flag = &cli.StringSliceFlag{}
}

func TestTimestampFlag_SatisfiesFlagInterface(t *testing.T) {
	var _ cli.Flag = &cli.TimestampFlag{}
}

func TestTimestampFlag_SatisfiesFmtStringerInterface(t *testing.T) {
	var _ fmt.Stringer = &cli.TimestampFlag{}
}

func TestBoolFlag_SatisfiesFlagInterface(t *testing.T) {
	var _ cli.Flag = &cli.BoolFlag{}
}

func TestBoolFlag_SatisfiesFmtStringerInterface(t *testing.T) {
	var _ fmt.Stringer = &cli.BoolFlag{}
}

func TestFloat64Flag_SatisfiesFlagInterface(t *testing.T) {
	var _ cli.Flag = &cli.Float64Flag{}
}

func TestFloat64Flag_SatisfiesFmtStringerInterface(t *testing.T) {
	var _ fmt.Stringer = &cli.Float64Flag{}
}

func TestIntFlag_SatisfiesFlagInterface(t *testing.T) {
	var _ cli.Flag = &cli.IntFlag{}
}

func TestIntFlag_SatisfiesFmtStringerInterface(t *testing.T) {
	var _ fmt.Stringer = &cli.IntFlag{}
}

func TestInt64Flag_SatisfiesFlagInterface(t *testing.T) {
	var _ cli.Flag = &cli.Int64Flag{}
}

func TestInt64Flag_SatisfiesFmtStringerInterface(t *testing.T) {
	var _ fmt.Stringer = &cli.Int64Flag{}
}

func TestStringFlag_SatisfiesFlagInterface(t *testing.T) {
	var _ cli.Flag = &cli.StringFlag{}
}

func TestStringFlag_SatisfiesFmtStringerInterface(t *testing.T) {
	var _ fmt.Stringer = &cli.StringFlag{}
}

func TestDurationFlag_SatisfiesFlagInterface(t *testing.T) {
	var _ cli.Flag = &cli.DurationFlag{}
}

func TestDurationFlag_SatisfiesFmtStringerInterface(t *testing.T) {
	var _ fmt.Stringer = &cli.DurationFlag{}
}

func TestUintFlag_SatisfiesFlagInterface(t *testing.T) {
	var _ cli.Flag = &cli.UintFlag{}
}

func TestUintFlag_SatisfiesFmtStringerInterface(t *testing.T) {
	var _ fmt.Stringer = &cli.UintFlag{}
}

func TestUint64Flag_SatisfiesFlagInterface(t *testing.T) {
	var _ cli.Flag = &cli.Uint64Flag{}
}

func TestUint64Flag_SatisfiesFmtStringerInterface(t *testing.T) {
	var _ fmt.Stringer = &cli.Uint64Flag{}
}

// vim:ro
