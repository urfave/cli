//go:build !go1.18
// +build !go1.18

package cli

import (
	"flag"
)

func unwrapFlagValue(v flag.Value) flag.Value { return v }
