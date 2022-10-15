# urfave-cli-genflags

This is a tool that is used internally by [urfave/cli] to generate
flag types and methods from a YAML input. It intentionally pins
usage of `github.com/urfave/cli/v2` to a *release* rather than
using the adjacent code so that changes don't result in *this* tool
refusing to compile. It's almost like dogfooding?

## support warning

This tool is maintained as a sub-project and is not covered by the
API and backward compatibility guaranteed by releases of
[urfave/cli].

[urfave/cli]: https://github.com/urfave/cli
