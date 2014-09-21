package cli

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/tmtk75/cli"
	osext "github.com/tmtk75/go-ext"
)

func help(args string, tt func(out string)) {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		cli.Command{
			Name:   "hi",
			Args:   args,
			Action: func(c *cli.Context) {},
		},
	}
	cap, _ := osext.PipeStdout()
	app.Run([]string{"", "hi"})
	tt(cap.Close())
}

func TestUsage(t *testing.T) {
	help("", func(out string) {
		if !(out == "") {
			t.Errorf("expect empty: %v", out)
		}
	})

	help("<path>", func(out string) {
		a := regexp.MustCompile("(?m)command (.*)").FindStringSubmatch(out)
		if "hi <path>" != a[1] {
			t.Errorf("%v", a[1])
		}
	})

	help("<path> [name]", func(out string) {
		a := regexp.MustCompile("(?m)command (.*)").FindStringSubmatch(out)
		if "hi <path> [name]" != a[1] {
			t.Errorf("%v", a[1])
		}
	})
}

func argv(args string, argv []string, tt func(c *cli.Context, out string)) {
	app := cli.NewApp()
	var ctx *cli.Context
	app.Commands = []cli.Command{
		cli.Command{
			Name: "hi",
			Args: args,
			Action: func(c *cli.Context) {
				ctx = c
			},
		},
	}
	cap, _ := osext.PipeStdout()
	app.Run(append([]string{"", "hi"}, argv...))
	tt(ctx, cap.Close())
}

func TestArgFor(t *testing.T) {
	argv("", []string{}, func(c *cli.Context, out string) {
		if !(out == "") {
			t.Errorf("expect empty: %v", out)
		}
	})

	argv("<tall>", []string{}, func(c *cli.Context, out string) {
		if !(out != "") {
			t.Errorf("expect help: %v", out)
		}
		if !(c == nil) {
			t.Errorf("expect nil: %v", c)
		}
	})

	argv("<name>", []string{"Lig"}, func(c *cli.Context, out string) {
		if !(out == "") {
			t.Errorf("%v", out)
		}
		if name, b := c.ArgFor("name"); !(name == "Lig" && b) {
			t.Errorf("expect Lig: %v, %v", name, b)
		}
	})

	argv("<age>", []string{""}, func(c *cli.Context, out string) {
		if !(out == "") {
			t.Errorf("expect help: %v", out)
		}
		if age, b := c.ArgFor("age"); !(age == "" && b) {
			t.Errorf("expect empty: %v, %v", age, b)
		}
	})

	argv("[path]", []string{""}, func(c *cli.Context, out string) {
		if !(out == "") {
			t.Errorf("expect empty: %v", out)
		}
		if path, b := c.ArgFor("path"); !(path == "" && b) {
			t.Errorf("expect empty: %v, %v", path, b)
		}
	})

	argv("[path]", []string{}, func(c *cli.Context, out string) {
		if !(out == "") {
			t.Errorf("expect empty: %v", out)
		}
		if path, b := c.ArgFor("path"); !(path == "" && !b) {
			t.Errorf("expect empty: %v, %v", path, b)
		}
	})

	// variation
	argv("<path> [id]", []string{"a", "b"}, func(c *cli.Context, out string) {
		if path, b := c.ArgFor("path"); !(path == "a" && b) {
			t.Errorf("expect a: %v, %v", path, b)
		}
	})

	argv("[path] [id]", []string{}, func(c *cli.Context, out string) {
		if path, b := c.ArgFor("id"); !(path == "" && !b) {
			t.Errorf("expect empty: %v, %v", path, b)
		}
	})
}

func TestValidateArgs(t *testing.T) {
	// Valid
	if v, err := cli.ValidateArgs(""); !(v != nil && err == nil) {
		t.Errorf("expect valid")
	}

	if v, err := cli.ValidateArgs("<name>"); !(v != nil && err == nil) {
		t.Errorf("expect valid")
	}

	if v, err := cli.ValidateArgs("[name]"); !(v != nil && err == nil) {
		t.Errorf("expect valid")
	}

	if v, err := cli.ValidateArgs("<name> [path]"); !(v != nil && err == nil) {
		t.Errorf("expect valid")
	}

	// Invalid
	if v, err := cli.ValidateArgs("<path"); !(v == nil && err != nil) {
		t.Errorf("expect invalid, cannot parse")
	}

	if v, err := cli.ValidateArgs("[name] <path>"); !(v == nil && err != nil) {
		t.Errorf("expect invalid, optional order")
	}

	if v, err := cli.ValidateArgs("<name> [name]"); !(v == nil && err != nil) {
		t.Errorf("expect invalid, duplicated name")
	}
}

func run(args string, argv []string) error {
	app := cli.NewApp()
	app.Commands = []cli.Command{
		cli.Command{
			Name:   "hi",
			Args:   args,
			Action: func(c *cli.Context) {},
		},
	}
	return app.Run(append([]string{"", "hi"}, argv...))
}

func TestValidArgs(t *testing.T) {
	if err := run("<name", []string{}); !(err != nil && fmt.Sprintln(err) == "parse error for Args: <name\n") {
		t.Errorf("expect error: %v", err)
	}

	if err := run("<path>", []string{}); !(err != nil && fmt.Sprintln(err) == "insufficient args\n") {
		t.Errorf("expect help: %v", err)
	}
}
