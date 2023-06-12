# Welcome to urfave/cli

[![Run Tests](https://github.com/urfave/cli/actions/workflows/cli.yml/badge.svg)](https://github.com/urfave/cli/actions/workflows/cli.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/urfave/cli/v3.svg)](https://pkg.go.dev/github.com/urfave/cli/v3)
[![Go Report Card](https://goreportcard.com/badge/github.com/urfave/cli/v3)](https://goreportcard.com/report/github.com/urfave/cli/v3)
[![codecov](https://codecov.io/gh/urfave/cli/branch/main/graph/badge.svg?token=t9YGWLh05g)](https://codecov.io/gh/urfave/cli)

urfave/cli is a **declarative**, simple, fast, and fun package for building command line tools in Go featuring:

- nestable commands with alias and prefix match support
- flexible and permissive help system
- dynamic shell completion for `bash`, `zsh`, `fish`, and `powershell`
- `man` and markdown format documentation generation
- input flag types for primitives, slices of primitives, time, duration, and others
- compound short flag support (`-a` `-b` `-c` :arrow_right: `-abc`)
- input value sources including:
    - environment variables
    - plain text files
    - [structured file formats supported via the `urfave/cli-altsrc` package](https://github.com/urfave/cli-altsrc)

## Documentation

More documentation is available in [`./docs`](./docs) or the hosted documentation site published from the latest release
at <https://cli.urfave.org>.

## Q&amp;A

Please check the [Q&amp;A discussions](https://github.com/urfave/cli/discussions/categories/q-a) or [ask a new
question](https://github.com/urfave/cli/discussions/new?category=q-a).

## License

See [`LICENSE`](./LICENSE)
