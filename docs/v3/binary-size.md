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
