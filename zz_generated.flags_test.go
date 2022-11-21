// WARNING: this file is generated. DO NOT EDIT

package cli_test

import (
	"fmt"
	"testing"

	"github.com/urfave/cli/v3"
)

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

// vim:ro
