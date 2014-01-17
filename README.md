[![Build Status](https://travis-ci.org/codegangsta/cli.png?branch=master)](https://travis-ci.org/codegangsta/cli)

# cli.go
cli.go is simple, fast, and fun package for building command line apps in Go. The goal is to enable developers to write fast and distributable command line applications in an expressive way.

You can view the API docs here:
http://godoc.org/github.com/codegangsta/cli

## Overview
Command line apps are usually so tiny that there is absolutely no reason why your code should *not* be self-documenting. Things like generating help text and parsing command flags/options should not hinder productivity when writing a command line app.

This is where cli.go comes into play. cli.go makes command line programming fun, organized, and expressive!

## Installation
Make sure you have the a working Go environment (go 1.1 is *required*). [See the install instructions](http://golang.org/doc/install.html).

To install cli.go, simply run:
```
$ go get github.com/codegangsta/cli
```

Make sure your PATH includes to the `$GOPATH/bin` directory so your commands can be easily used:
```
export PATH=$PATH:$GOPATH/bin
```

## Getting Started
One of the philosophies behind cli.go is that an API should be playful and full of discovery. So a cli.go app can be as little as one line of code in `main()`.

``` go
package main

import (
  "os"
  "github.com/codegangsta/cli"
)

func main() {
  cli.NewApp().Run(os.Args)
}
```

This app will run and show help text, but is not very useful. Let's give an action to execute and some help documentation:

``` go
package main

import (
  "os"
  "github.com/codegangsta/cli"
)

func main() {
  app := cli.NewApp()
  app.Name = "boom"
  app.Usage = "make an explosive entrance"
  app.Action = func(c *cli.Context) {
    println("boom! I say!")
  }

  app.Run(os.Args)
}
```

Running this already gives you a ton of functionality, plus support for things like subcommands and flags, which are covered below.

## Example

Being a programmer can be a lonely job. Thankfully by the power of automation that is not the case! Let's create a greeter app to fend off our demons of loneliness!

``` go
/* greet.go */
package main

import (
  "os"
  "github.com/codegangsta/cli"
)

func main() {
  app := cli.NewApp()
  app.Name = "greet"
  app.Usage = "fight the loneliness!"
  app.Author = "me"
  app.CopyrightHolder = app.Author
  app.Version = "0.0.1" // if ommited default is 0.0.0
  app.Copyright = "1973-2014" // if ommited default is current year
  app.License = "ASL-2 (Apache Software License version 2.0) <http://www.apache.org/licenses/LICENSE-2.0>"
  app.Reporting = "Report bugs to me, if you find any :-) ..."
  app.Action = func(c *cli.Context) {
    println("Hello friend!")
  }
  app.Run(os.Args)
}
```

Install our command to the `$GOPATH/bin` directory:

```
$ go install
```

Finally run our new command:

```
$ greet
Hello friend!
```

cli.go also generates some bitchass help text:
```
$ greet help
NAME:
    greet - fight the loneliness!

USAGE:
    greet [global options] command [command options] [arguments...]

VERSION:
    0.0.0

COMMANDS:
    help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS
   --version, -v print the version
   --help, -h    show help

AUTHOR:
    Written by me.

REPORTING BUGS:
    Report bugs to me, if you find any :-) ...

COPYRIGHT:
    Copyright © 1973-2014 me
    Licensed under the ASL-2 (Apache Software License version 2.0) <http://www.apache.org/licenses/LICENSE-2.0>
```

### Arguments
You can lookup arguments by calling the `Args` function on cli.Context.

``` go
...
app.Action = func(c *cli.Context) {
  println("Hello", c.Args()[0])
}
...
```

### Flags
Setting and querying flags is simple.
``` go
...
app.Flags = []cli.Flag {
  cli.StringFlag{"lang", "english", "language for the greeting"},
}
app.Action = func(c *cli.Context) {
  name := "someone"
  if len(c.Args()) > 0 {
    name = c.Args()[0]
  }
  if c.String("lang") == "spanish" {
    println("Hola", name)
  } else {
    println("Hello", name)
  }
}
...
```

#### Alternate Names

You can set alternate (or short) names for flags by providing a comma-delimited list for the Name. e.g.

``` go
app.Flags = []cli.Flag {
  cli.StringFlag{"lang, l", "english", "language for the greeting"},
}
```

That flag can then be set with `--lang spanish` or `-l spanish`. Note that giving two different forms of the same flag in the same command invocation is an error.

#### Boolean Flags

You can also set boolean flags e.g.
``` go
app.Flags = []cli.Flag{
    cli.BoolFlag{
      Name:  "debug",
      Usage: "enables debug mode",
    },
  }
app.Action = func(c *cli.Context) {
    if c.String("debug") == "true" {
      DEBUG = true
      log.Printf("DEBUG mode enabled.")
    }
  }

```

Any time an user invokes `--debug` the appropriate action will be run.

### Subcommands

Subcommands can be defined for a more git-like command line app.
```go
...
app.Commands = []cli.Command{
  {
    Name:      "add",
    ShortName: "a",
    Usage:     "add a task to the list",
    Action: func(c *cli.Context) {
      println("added task: ", c.Args().First())
    },
  },
  {
    Name:      "complete",
    ShortName: "c",
    Usage:     "complete a task on the list",
    Action: func(c *cli.Context) {
      println("completed task: ", c.Args().First())
    },
  },
}
...
```

## About
cli.go is written by none other than the [Code Gangsta](http://codegangsta.io)
