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
