package cli

import (
	"fmt"
	"regexp"
	"testing"

	osext "github.com/tmtk75/go-ext"
)

func argv(args string, argv []string, tt func(c *Context, out string)) {
	app := NewApp()
	var ctx *Context
	app.Commands = []Command{
		Command{
			Name: "hi",
			Args: args,
			Action: func(c *Context) {
				ctx = c
			},
		},
	}
	pipe, _ := osext.PipeStdout()
	app.Run(append([]string{"", "hi"}, argv...))
	tt(ctx, pipe.Close())
}

func TestUsage(t *testing.T) {
	argv("", []string{}, func(c *Context, out string) {
		if !(out == "") {
			t.Errorf("expect empty: %v", out)
		}
	})

	argv("<path>", []string{}, func(c *Context, out string) {
		a := regexp.MustCompile("(?m)command (.*)").FindStringSubmatch(out)
		if "hi <path>" != a[1] {
			t.Errorf("%v", a[1])
		}
	})

	argv("<path> [name]", []string{}, func(c *Context, out string) {
		a := regexp.MustCompile("(?m)command (.*)").FindStringSubmatch(out)
		if "hi <path> [name]" != a[1] {
			t.Errorf("%v", a[1])
		}
	})
}

func TestArgFor(t *testing.T) {
	// command runs
	argv("", []string{}, func(c *Context, out string) {
		if !(out == "") {
			t.Errorf("expect empty: %v", out)
		}
	})

	argv("<tall>", []string{}, func(c *Context, out string) {
		if !(out != "") {
			t.Errorf("expect help: %v", out)
		}
		if !(c == nil) {
			t.Errorf("expect nil: %v", c)
		}
	})

	argv("<name>", []string{"Lig"}, func(c *Context, out string) {
		if !(out == "") {
			t.Errorf("expect empty: %v", out)
		}
		if name, b := c.ArgFor("name"); !(name == "Lig" && b) {
			t.Errorf("expect Lig: %v, %v", name, b)
		}
	})

	argv("<age>", []string{""}, func(c *Context, out string) {
		if !(out == "") {
			t.Errorf("expect empty: %v", out)
		}
		if age, b := c.ArgFor("age"); !(age == "" && b) {
			t.Errorf("expect empty: %v, %v", age, b)
		}
	})

	argv("[path]", []string{""}, func(c *Context, out string) {
		if !(out == "") {
			t.Errorf("expect empty: %v", out)
		}
		if path, b := c.ArgFor("path"); !(path == "" && b) {
			t.Errorf("expect empty: %v, %v", path, b)
		}
	})

	argv("[path]", []string{}, func(c *Context, out string) {
		if !(out == "") {
			t.Errorf("expect empty: %v", out)
		}
		if path, b := c.ArgFor("path"); !(path == "" && !b) {
			t.Errorf("expect empty: %v, %v", path, b)
		}
	})

	argv("<path> [id]", []string{"a", "b"}, func(c *Context, out string) {
		if path, b := c.ArgFor("path"); !(path == "a" && b) {
			t.Errorf("expect a: %v, %v", path, b)
		}
		if id, b := c.ArgFor("id"); !(id == "b" && b) {
			t.Errorf("expect b: %v, %v", id, b)
		}
	})

	argv("[path] [id]", []string{}, func(c *Context, out string) {
		if path, b := c.ArgFor("id"); !(path == "" && !b) {
			t.Errorf("expect empty: %v, %v", path, b)
		}
	})

	argv("<path> [id]", []string{"", ""}, func(c *Context, out string) {
		if path, b := c.ArgFor("path"); !(path == "" && b) {
			t.Errorf("expect empty: %v, %v", path, b)
		}
		if id, b := c.ArgFor("id"); !(id == "" && b) {
			t.Errorf("expect '': %v, %v", id, b)
		}
	})

	argv("<date>", []string{"2014-09-22", ""}, func(c *Context, out string) {
		if day, b := c.ArgFor("day"); !(day == "" && !b) {
			t.Errorf("expect empty: %v, %v", day, b)
		}
	})

	// Ignored other than argument patterns
	argv("<id> [path](optional) ...", []string{"abc", ".", "_"}, func(c *Context, out string) {
		if path, b := c.ArgFor("path"); !(path == "." && b) {
			t.Errorf("expect .: %v, %v", path)
		}
		if any := c.Args().Get(2); !(any == "_") {
			t.Errorf("expect _: %v", any)
		}
	})
}

func TestValidateArgs(t *testing.T) {
	// Valid
	if v, err := validateArgs(""); !(v != nil && err == nil) {
		t.Errorf("expect valid")
	}

	if v, err := validateArgs("<name>"); !(v != nil && err == nil) {
		t.Errorf("expect valid")
	}

	if v, err := validateArgs("[name]"); !(v != nil && err == nil) {
		t.Errorf("expect valid")
	}

	if v, err := validateArgs("<name> [path]"); !(v != nil && err == nil) {
		t.Errorf("expect valid")
	}

	// Invalid
	if v, err := validateArgs("<path"); !(v == nil && err != nil) {
		t.Errorf("expect invalid, cannot parse")
	}

	if v, err := validateArgs("[name] <path>"); !(v == nil && err != nil) {
		t.Errorf("expect invalid, optional order")
	}

	if v, err := validateArgs("<name> [name]"); !(v == nil && err != nil) {
		t.Errorf("expect invalid, duplicated name")
	}
}

func run(args string, argv []string) error {
	app := NewApp()
	app.Commands = []Command{
		Command{
			Name:   "hello",
			Args:   args,
			Action: func(c *Context) {},
		},
	}
	return app.Run(append([]string{"", "hello"}, argv...))
}

func TestValidArgs(t *testing.T) {
	if err := run("<name", []string{}); !(err != nil && fmt.Sprintln(err) == "parse error for Args: <name\n") {
		t.Errorf("expect error: %v", err)
	}

	if err := run("<name> [name]", []string{}); !(err != nil && fmt.Sprintln(err) == "duplicated name\n") {
		t.Errorf("expect error: %v", err)
	}

	if err := run("<path>", []string{}); !(err != nil && fmt.Sprintln(err) == "insufficient args\n") {
		t.Errorf("expect help: %v", err)
	}
}
