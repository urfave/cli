# Work in Progress
This package is not in a releasable state. Stay tuned as I try to crank this API out in the next week or so.

# cli.go
cli.go is simple, fast, and fun package for building command line apps in Go. The goal is to enable developers to write fast and distributable command line applications in an expressive way.

## Overview
Command line apps are usually so tiny that there is absolutely no reason why your code should *not* be self-documenting. Things like generating help text and parsing command flags/options should not hinder productivity when writing a command line app.

This is where cli.go comes into play. cli.go makes command line programming fun, organized, and expressive!

## Installation
Make sure you have the a working Go environment. [See the install instructions](http://golang.org/doc/install.html).

To install cli.go, simply run:
```
$ go get github.com/codegangsta/cli
```

Make sure your PATH includes to the `$GOPATH/bin` directory so your commands can be easily used:
```
export PATH=$PATH:$GOPATH/bin
```

## Example

Being a programmer can be a lonely job. Thankfully by the power of automation that is not the case! Let's create a greeter app to fend off our demons of loneliness!

``` go
/* greet.go */
package main

import "os"
import "github.com/codegangsta/cli"

func main() {
  cli.Name = "greet"
  cli.Usage = "fight the loneliness!"
  cli.Action = func(c cli.Context) {
    println("Hello friend!")
  }
  
  cli.Run(os.Args)
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
    --version	Shows version information
```

### Arguments
WIP

### Flags
WIP

### Subcommands
WIP


## About
cli.go is written by none other than the [Code Gangsta](http://codegangsta.io)
