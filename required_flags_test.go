package cli

import (
	"flag"
	"testing"
)

func TestContext_CheckRequiredFlagsSuccess(t *testing.T) {
	flags := []Flag{
		StringFlag{
			Name:     "required",
			Required: true,
		},
		StringFlag{
			Name: "optional",
		},
	}

	set := flag.NewFlagSet("test", 0)
	for _, f := range flags {
		f.Apply(set)
	}

	e := set.Parse([]string{"--required", "foo"})
	if e != nil {
		t.Errorf("Expected no error parsing but there was one: %s", e)
	}

	err := checkRequiredFlags(flags, set)
	if err != nil {
		t.Error("Expected flag parsing to be successful")
	}
}

func TestContext_CheckRequiredFlagsFailure(t *testing.T) {
	flags := []Flag{
		StringFlag{
			Name:     "required",
			Required: true,
		},
		StringFlag{
			Name: "optional",
		},
	}

	set := flag.NewFlagSet("test", 0)
	for _, f := range flags {
		f.Apply(set)
	}

	e := set.Parse([]string{"--optional", "foo"})
	if e != nil {
		t.Errorf("Expected no error parsing but there was one: %s", e)
	}

	err := checkRequiredFlags(flags, set)
	if err == nil {
		t.Error("Expected flag parsing to be unsuccessful")
	}
}
