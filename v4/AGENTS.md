# Working on v4

These rules apply to every change under `v4/`, whether a person or a coding agent writes it. [DESIGN.md](DESIGN.md) explains the reasons behind them.

## Rules

- **The core stays small.** The `cli` package imports only the standard library, and never `text/template`, `encoding/json`, `net` or anything under `golang.org/x`. `TestCoreDependencies` enforces this. Features that need more go in their own package or module.
- **No package-level mutable state.** Settings live on the root `Command`. Error kinds are the one exception, and they are only created at package level with `NewKind`.
- **Every error the library detects is a `*cli.Error`.** Give it a `Kind` and fill the structured fields. Don't format values into the message.
- **Every user-facing string goes through the catalogue.** Add a key to `English` in `localize.go` and to every language package under `i18n/locale/`. `TestShippedLocalesAreComplete` fails if a language package misses a key.
- **Only `Main` exits the process.** Library code returns errors.
- **Help and version are not errors.** They render and `Run` returns nil.
- **Tests come first.** Write the test that describes the behaviour, watch it fail, then write the code. Tests are table-driven and use the standard `testing` package only.
- **v3 stays untouched.** Don't change code outside `v4/` unless the task says so.

## Checks

Run these from `v4/` before sending a change:

```sh
gofmt -l .
go vet ./...
go test ./...
GOOS=windows go vet ./...
```

`gofmt -l .` must print nothing.
