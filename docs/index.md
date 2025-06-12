<!--
NOTE: This first section is intentionally identical to the top-level README.md at
https://github.com/urfave/cli/blob/main/README.md
-->
# Welcome to urfave/cli

[![Run Tests](https://github.com/urfave/cli/actions/workflows/test.yml/badge.svg)](https://github.com/urfave/cli/actions/workflows/test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/urfave/cli/v3.svg)](https://pkg.go.dev/github.com/urfave/cli/v3)
[![Go Report Card](https://goreportcard.com/badge/github.com/urfave/cli/v3)](https://goreportcard.com/report/github.com/urfave/cli/v3)
[![codecov](https://codecov.io/gh/urfave/cli/branch/main/graph/badge.svg?token=t9YGWLh05g)](https://codecov.io/gh/urfave/cli)

urfave/cli is a **declarative**, simple, fast, and fun package for building command line tools in Go featuring:

- commands and subcommands with alias and prefix match support
- flexible and permissive help system
- dynamic shell completion for `bash`, `zsh`, `fish`, and `powershell`
- `man` and markdown format documentation generation
- input flags for simple types, slices of simple types, time, duration, and others
- compound short flag support (`-a` `-b` `-c` :arrow_right: `-abc`)
- input lookup from:
    - environment variables
    - plain text files
    - [structured file formats supported via the `urfave/cli-altsrc` package](https://github.com/urfave/cli-altsrc)

<!--
/END first section that is identical to README.md first section
-->

These are the guides for each major version:

- [`v3`](./v3/getting-started.md)
- [`v2`](./v2/getting-started.md)
- [`v1`](./v1/getting-started.md)

In addition to the version-specific guides, these other documents are available:

- [CONTRIBUTING](./CONTRIBUTING.md)
- [CODE OF CONDUCT](./CODE_OF_CONDUCT.md)
- [RELEASING](./RELEASING.md)

## Installation

Using this package requires a working Go environment. [See the install instructions for Go](https://go.dev/doc/install).

Go Modules are required when using this package. [See the go blog guide on using Go Modules](https://blog.golang.org/using-go-modules).

### Using `v3` releases

The latest `v3` release may be installed via the `/v3` suffix. The state of the [`main`
branch](https://github.com/urfave/cli/tree/main) at any given time may correspond to a
`v3` series release or pre-release.  Please see the [`v3` migration
guide](./migrate-v2-to-v3.md) on using v3 if you are upgrading from v2.

```sh
go get github.com/urfave/cli/v3@latest
```

```go
import (
  "github.com/urfave/cli/v3" // imports as package "cli"
)
```

### Using `v2` releases

The `v2` series is the recommended version for new development. Ongoing
maintenance is done on the [`v2-maint`
branch](https://github.com/urfave/cli/tree/v2-maint) which receives **minor**
improvements, bug fixes, and security fixes.

```sh
go get github.com/urfave/cli/v2@latest
```

```go
import (
  "github.com/urfave/cli/v2" // imports as package "cli"
)
```

### Using `v1` releases

:warning: The `v1` series is receiving **security fixes only** via the
[`v1-maint`](https://github.com/urfave/cli/tree/v1-maint) branch and **should
not** be used in new development. Please see the [`v2` migration
guide](./migrate-v1-to-v2.md) and feel free to open an issue or discussion if
you need help with the migration to `v2`.

### Supported platforms

cli is tested against multiple versions of Go on Linux, and against the latest
released version of Go on OS X and Windows. This project uses GitHub Actions
for builds. To see our currently supported go versions and platforms, look at
the [github workflow
configuration](https://github.com/urfave/cli/blob/main/.github/workflows/test.yml).
