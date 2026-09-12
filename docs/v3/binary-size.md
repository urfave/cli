# Binary Size

Go removes unreachable code during compilation, so the first step is to
measure the binary you ship instead of assuming a specific feature is expensive.

```sh-session
go build -trimpath -o myapp ./cmd/myapp
ls -lh myapp
```

For a release-style build, combine reproducible paths with stripped symbol and
debug information:

```sh-session
go build -trimpath -ldflags="-s -w" -o myapp ./cmd/myapp
ls -lh myapp
```

Use the Go toolchain to inspect what is in the binary:

```sh-session
go version -m myapp
go tool nm -size myapp | sort -nr | head -40
```

## Practical Checks

- Run `make check-binary-size` in this repository to compare the current package
  contribution against the tracked binary-size budget.
- Build with `-trimpath` for reproducible paths.
- Use `-ldflags="-s -w"` for release builds when debug symbols are not needed.
- Keep optional integrations in separate packages when they bring large
  dependencies.
- Avoid adding reflection-heavy dependencies to the main command package unless
  they are required at runtime.
- Compare sizes from clean builds after each change.

If a specific `urfave/cli` feature appears to keep unexpected code reachable,
[open an issue](https://github.com/urfave/cli/issues/new) with the Go version,
build command, a minimal reproduction, and the `go tool nm -size` output that
shows the largest symbols.

## Current v3 Build Tags

The v3 module does not currently define build tags such as
`urfave_cli_no_docs`, `urfave_cli_no_completion`, or `urfave_cli_minimal`.
Documentation generation lives outside the core module in
[`urfave/cli-docs`](https://github.com/urfave/cli-docs), so applications that
only import `github.com/urfave/cli/v3` do not pull in that package.

Shell completion support is part of the core package. Leave
`EnableShellCompletion` disabled unless the application needs shell completion,
then measure the result with the commands above.


## Deeper Analysis with go-size-analyzer
 
The commands above show symbol-level sizes, but they do not clearly show which
packages contribute most to the binary. [go-size-analyzer](https://github.com/Zxilly/go-size-analyzer)
provides a package-level breakdown.
 
Install it with:
 
 
```sh-session
go install github.com/Zxilly/go-size-analyzer/cmd/gsa@latest
```
 
Then analyze a binary:
 
```sh-session
gsa myapp
```
 
For example, a minimal `urfave/cli/v3` binary showed:
 
```
21.58%  runtime                  1.0 MB
 6.57%  reflect                  315 kB
 6.25%  .rodata                  300 kB
 5.87%  github.com/urfave/cli/v3 282 kB
 5.74%  text/template            275 kB
```
 
This helps identify which packages are worth investigating. Packages such as
`runtime` are fundamental to a Go binary and are not normally removable.
For an interactive view, use:
 
```sh-session
gsa --tui myapp
```
 
> **Note:** analyze an unstripped binary when possible. Stripping debug
> information can make package attribution less accurate.
 
## Dependency Versions and Multiple Modules
 
A repository may contain multiple Go modules, each with its own `go.mod`.
Check the modules and their selected versions with:
 
```sh-session
go list -m all
```
 
For a specific dependency:
 
```sh-session
go list -m -json github.com/urfave/cli/v3
```
 
Different `go.mod` files do not automatically cause multiple versions of the
same module to be included in a binary. Go normally selects a single version
of a module path for a build.
 
However, when comparing builds across multiple modules, it is important to
check the actual selected version rather than relying only on the version
written in `go.mod`. For example, the documentation module uses:
 
```go
replace github.com/urfave/cli/v3 => ../
```
 
so it builds against the local `urfave/cli/v3` checkout.
 
## Reflection and Templates
 
Reflection and runtime type inspection can contribute to binary size because
additional type information and supporting code may remain reachable.
 
In the example binary, `go-size-analyzer` reported:
 
```
reflect       315 kB
text/template 275 kB
```
 
These numbers show their contribution to the analyzed binary, but do not by
themselves prove that all of this code comes from a single feature.
`urfave/cli/v3` uses templates for help rendering, making `text/template` one
of the packages worth investigating when analyzing binary size.
 
## Hiding Built-in Help
 
Disabling built-in help does not currently reduce the binary size.
 
Two otherwise identical binaries were built:
 
```go
&cli.Command{}
```
 
and:
 
```go
&cli.Command{
    HideHelp:        true,
    HideHelpCommand: true,
}
```
 
Both were approximately 4.6 MB, and `go-size-analyzer` reported the same
sizes for the relevant packages:
 
| Package                     | Normal | Hidden |
|------------------------------|--------|--------|
| `reflect`                    | 315 kB | 315 kB |
| `text/template`              | 275 kB | 275 kB |
| `github.com/urfave/cli/v3`   | 282 kB | 282 kB |
 
`HideHelp` and `HideHelpCommand` are runtime options, so hiding help does not
remove the help implementation from the compiled binary. If reducing this
overhead is important, a compile-time approach — such as separate build
configurations or changes to the library — would be required.




