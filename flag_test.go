package cli_test

import (
	"github.com/codegangsta/cli"
	"testing"
)

var boolFlagTests = []struct {
	name     string
	expected string
}{
	{"help", "--help\t"},
	{"h", "-h\t"},
}

func TestBoolFlagHelpOutput(t *testing.T) {

	for _, test := range boolFlagTests {
		flag := cli.BoolFlag{Name: test.name}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%s does not match %s", output, test.expected)
		}
	}
}

var stringFlagTests = []struct {
	name     string
	expected string
}{
	{"help", "--help ''\t"},
	{"h", "-h ''\t"},
}

func TestStringFlagHelpOutput(t *testing.T) {

	for _, test := range stringFlagTests {
		flag := cli.StringFlag{Name: test.name}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%s does not match %s", output, test.expected)
		}
	}
}

var intFlagTests = []struct {
	name     string
	expected string
}{
	{"help", "--help '0'\t"},
	{"h", "-h '0'\t"},
}

func TestIntFlagHelpOutput(t *testing.T) {

	for _, test := range intFlagTests {
		flag := cli.IntFlag{Name: test.name}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%s does not match %s", output, test.expected)
		}
	}
}

var float64FlagTests = []struct {
	name     string
	expected string
}{
	{"help", "--help '0'\t"},
	{"h", "-h '0'\t"},
}

func TestFloat64FlagHelpOutput(t *testing.T) {

	for _, test := range float64FlagTests {
		flag := cli.Float64Flag{Name: test.name}
		output := flag.String()

		if output != test.expected {
			t.Errorf("%s does not match %s", output, test.expected)
		}
	}
}

func TestParseMultiString(t *testing.T) {
	(&cli.App{
		Flags: []cli.Flag{
			cli.StringFlag{Name: "serve, s"},
		},
		Action: func(ctx *cli.Context) {
			if ctx.String("serve") != "10" {
				t.Errorf("main name not set")
			}
			if ctx.String("s") != "10" {
				t.Errorf("short name not set")
			}
		},
	}).Run([]string{"run", "-s", "10"})
}

func TestParseMultiInt(t *testing.T) {
	a := cli.App{
		Flags: []cli.Flag{
			cli.IntFlag{Name: "serve, s"},
		},
		Action: func(ctx *cli.Context) {
			if ctx.Int("serve") != 10 {
				t.Errorf("main name not set")
			}
			if ctx.Int("s") != 10 {
				t.Errorf("short name not set")
			}
		},
	}
	a.Run([]string{"run", "-s", "10"})
}

func TestParseMultiBool(t *testing.T) {
	a := cli.App{
		Flags: []cli.Flag{
			cli.BoolFlag{Name: "serve, s"},
		},
		Action: func(ctx *cli.Context) {
			if ctx.Bool("serve") != true {
				t.Errorf("main name not set")
			}
			if ctx.Bool("s") != true {
				t.Errorf("short name not set")
			}
		},
	}
	a.Run([]string{"run", "--serve"})
}
