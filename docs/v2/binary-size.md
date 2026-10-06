# Binary Size

Go removes unreachable code during compilation. The first step in size optimization
is measuring the shipped executable rather than assuming a feature introduces overhead.

```sh-session
go build -trimpath -o myapp ./cmd/myapp
ls -lh myapp
```

## Compiler and Linker Flags

For release builds, combine path trimming with symbol table and debug metadata removal:

```sh-session
go build -trimpath -ldflags="-s -w" -o myapp ./cmd/myapp
ls -lh myapp
```

### Flag Breakdown

- `-trimpath`: Strips local file system paths from the compiled executable. This removes
  machine-specific directory paths from panic stack traces, produces reproducible
  binaries, and reduces string data size.
- `-ldflags="-s -w"`:
  - `-s`: Omits the symbol table.
  - `-w`: Omits DWARF debugging information.

Panic stack traces retain function names and line offsets because the Go runtime
maintains internal program counter tables in a distinct section.

## Inspecting Binary Contents

Use the Go toolchain to inspect embedded build metadata and symbol allocations on an unstripped binary:

```sh-session
go version -m myapp
go tool nm -size myapp | sort -k2,2nr | head -40
```

- `go version -m`: Displays module dependencies, Go toolchain version, build tags, and
  compiler flags embedded in the binary.
- `go tool nm -size`: Lists symbols sorted by size, identifying the largest functions
  and data structures. Run this command against an unstripped binary, as `-s` removes
  the symbol table.

## Build Tags in v2

In `urfave/cli/v2`, two build tags prune optional dependencies and features at compile time:

| Build Tag | Target Feature | Dependency Removed | Approximate Savings |
|-----------|----------------|--------------------|---------------------|
| `urfave_cli_no_docs` | Markdown and man page generation (`ToMarkdown`, `ToMan`) | Documentation generator packages | 300 to 400 KB |
| `urfave_cli_no_suggest` | Command and flag typo suggestions ("Did you mean?") | `github.com/xrash/smetrics` | Variable |

### Usage

To build without documentation generation methods:

```sh-session
go build -tags urfave_cli_no_docs -trimpath -ldflags="-s -w" -o myapp ./cmd/myapp
```

To build without typo suggestions:

```sh-session
go build -tags urfave_cli_no_suggest -trimpath -ldflags="-s -w" -o myapp ./cmd/myapp
```

To combine both build tags:

```sh-session
go build -tags "urfave_cli_no_docs urfave_cli_no_suggest" -trimpath -ldflags="-s -w" -o myapp ./cmd/myapp
```

## Transition to v3

These build tags are not present in `urfave/cli/v3` because v3 decouples these features by design:

- Documentation generation lives in the separate module `github.com/urfave/cli-docs/v3`.
- Typo suggestions use an internal Jaro-Winkler implementation with no external dependencies.
- v3 provides the `urfave_cli_no_template` build tag to restore linker dead code elimination.

See the [v2 to v3 Migration Guide](migrating-to-v3.md) and [v3 Binary Size](../v3/binary-size.md)
for details.
