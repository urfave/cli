# Welcome to urfave/cli

[![Go Reference][goreference_badge]][goreference_link]
[![Go Report Card][goreportcard_badge]][goreportcard_link]
[![codecov][codecov_badge]][codecov_link]
[![Tests status][test_badge]][test_link]

urfave/cli is a **declarative**, simple, fast, and fun package for building
command line tools in Go featuring:

- commands and subcommands with alias and prefix match support
- flexible and permissive help system
- dynamic shell completion for `bash`, `zsh`, `fish`, and `powershell`
- `man` and markdown format documentation generation
- input flags for simple types, slices of simple types, time, duration, and others
- compound short flag support (`-a` `-b` `-c` can be shortened to `-abc`)
- input lookup from:
  - environment variables
  - plain text files
  - [structured file formats supported via the `urfave/cli-altsrc` package](https://github.com/urfave/cli-altsrc)

## Documentation

See the hosted documentation website at <https://cli.urfave.org>. Contents of
this website are built from the [`./docs`](./docs) directory.

## Q&A

Please check the [Q&A discussions] or [ask a new question].

## License

See [`LICENSE`](./LICENSE).

[test_badge]: https://github.com/urfave/cli/actions/workflows/test.yml/badge.svg
[test_link]: https://github.com/urfave/cli/actions/workflows/test.yml
[goreference_badge]: https://pkg.go.dev/badge/github.com/urfave/cli/v3.svg
[goreference_link]: https://pkg.go.dev/github.com/urfave/cli/v3
[goreportcard_badge]: https://goreportcard.com/badge/github.com/urfave/cli/v3
[goreportcard_link]: https://goreportcard.com/report/github.com/urfave/cli/v3
[codecov_badge]: https://codecov.io/gh/urfave/cli/branch/main/graph/badge.svg?token=t9YGWLh05g
[codecov_link]: https://codecov.io/gh/urfave/cli
[Q&A discussions]: https://github.com/urfave/cli/discussions/categories/q-a
[ask a new question]: https://github.com/urfave/cli/discussions/new?category=q-a
