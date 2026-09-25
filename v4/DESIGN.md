# urfave/cli v4 design

This is the proposed design for v4. The discussion is in [#2446](https://github.com/urfave/cli/discussions/2446) and the feature list is in [#2192](https://github.com/urfave/cli/issues/2192). v4 starts from an empty tree and is built outward from command actions. Errors and i18n come first, then one command, personality, help, flags and arguments, and finally subcommands.

The code in this directory is the first milestone, covering errors, i18n and exit handling. Everything else here is a proposal.

## What exists today

- Every error the library detects is a `*cli.Error` with a `Kind` and structured fields (`Flag`, `Value`, `Choices`, `Count`, `Err`). `errors.Is(err, cli.ErrUnknownFlag)` works through `%w` wrapping and `errors.Join`.
- Each kind is defined once with `cli.NewKind`, which gives it a name and a class (a failure or a usage error). Applications can define their own kinds and get matching, translation and exit codes the same way.
- Messages come from a catalogue keyed by stable strings such as `error.unknown_flag`, with `{flag}` style placeholders and plural forms. English is built in. German is a separate package. We can add other language packages as required.
- The locale comes from `LC_ALL`, `LC_MESSAGES`, `LANG` and `LANGUAGE`, following gettext rules.
- A `Personality` decides what an error looks like and which exit code it maps to. The POSIX and Git presets and the Agent and Quiet modifiers described below exist today.
- `Command.Run` never exits. `cli.Main` handles signals and calls `os.Exit`.

```
$ LANG=de_DE.UTF-8 hello a b c
hello: 3 unerwartete Argumente, beginnend mit „a“
„hello --help“ gibt weitere Informationen aus.
$ echo $?
2
```

## A modular library

Go only links packages a program imports, so we can split v4 into packages instead of build tags. The core never imports an optional package.

- **Core** (`cli/v4`): commands, flags, the parser, errors, the English catalogue, a help renderer without templates, exit handling. No dependencies. A test fails if the core pulls in `text/template`, `encoding/json`, `net` or `golang.org/x`.
- **Optional packages**: `help/tmpl` (v3 templates), `help/jsonhelp`, `complete`, `personality`, `i18n` (JSON catalogues), `i18n/locale/<tag>`, `clitest`. Language packages are plain Go data, so adding one doesn't add a JSON decoder to the binary.
- **Separate modules**: `docs` (markdown and man), `altsrc` (config files), `xtext` (adapter for `golang.org/x/text`) and `mcp` (serves commands as MCP tools).

`text/template` matters most here. Executing a template calls `reflect.Value.MethodByName`, which stops the linker from removing unused methods anywhere in the program. Moving template help out of the core removes that cost for everyone who doesn't use it.

## Personalities

A personality has two parts that change for different reasons. Presets set the policy, meaning the exit codes, the hints and what happens on an unknown command. Modifiers set the format. Any modifier works with any preset, and a modifier never changes exit codes, so a script, a person and an agent see the same code for the same failure.

Presets:

- **POSIX** (`ls`, `grep`): `prog: <message>` and "Try 'prog --help' for more information.", exit 2. This is the default.
- **BSD** (`cp`, `tar` on macOS): prints the usage synopsis on a usage error and uses the sysexits codes, so a usage error exits 64 (`EX_USAGE`). Scripts that follow `sysexits.h` can tell failures apart by code.
- **Git**: for tools that dispatch to subcommands. It suggests the most similar command and exits 1 for an unknown command and 129 for other usage errors, as git does. It could also run `prog-foo` from `PATH` for an unknown command `foo`.
- **V3**: reproduces v3 byte for byte, including "Incorrect Usage:" followed by the help, so apps can move to v4 without changing what users see. The v3 compatibility suite runs against it.

Modifiers:

- **Agent**: one JSON object on stderr with the errors and warnings, each with its kind, message, flag and value, plus the exit code. No colour, and never a prompt.
- **Quiet**: one line per error and nothing else, for CLIs that mostly run inside scripts.

```go
cmd := &cli.Command{
    Personality: personality.Auto(os.LookupEnv, personality.Git),
}
```

`Auto` returns the preset for people and `Agent(preset)` when `URFAVE_CLI_AGENT=1` is set. An agent, or the tool that runs it, can set that once for every command. An app can also offer a flag such as `--output=json`, which the library ships as a helper rather than reserving the name. Whether stderr is a terminal only decides colour. `prog 2>err.log` is still a person.

## Design for the remaining milestones

- **Help**: a `HelpRenderer` interface that takes a plain `HelpData` struct, never `*Command`. The default renderer uses `text/tabwriter`, and the Agent modifier renders the same data as JSON. The help flag becomes an ordinary flag that renders help and `Run` returns nil without running the action. That removes the special cases around the global `HelpFlag` that are left over from [#2176](https://github.com/urfave/cli/issues/2176). Template help ([#2075](https://github.com/urfave/cli/issues/2075)) moves to `help/tmpl`. `text/tabwriter` counts runes rather than screen width, so Chinese, Japanese and Korean text misaligns. The `xtext` module can supply a width-aware layout using `golang.org/x/text/width`.
- **App text**: the catalogue covers the library's strings, but an app's own command and flag descriptions need translating too, or a German user gets German errors and English help. The help renderer passes app text through the same Localizer, and apps add their own keys the way they add their own kinds.
- **Flags**: `Flag[T]` with a `Parser[T]`, keeping v3 names as generic type aliases. `TextFlag[T]` for any `encoding.TextUnmarshaler`, typed rather than the interface-based v3 version. `cmd.Source("name")` reports whether a value came from a default, the environment, a file or the command line. `EnvPrefix: "APP"` on the root gives every flag an `APP_FLAG_NAME` variable without listing each one.
- **Flag rules**: conflicts, requires, one-of and all-or-none groups. Each one fails with its own error kind.
- **Parse errors**: the parser reports every independent problem in one run, joined with `errors.Join`, instead of stopping at the first. A user fixes them all at once, and `prog --bogus --output=json` still reports in JSON.
- **Unknown flags** ([#2240](https://github.com/urfave/cli/issues/2240)): reject, pass through, or collect.
- **Numeric flags**: the parser checks declared flags before it treats `-4` as a positional argument, so `-4` and `-6` from [#2192](https://github.com/urfave/cli/issues/2192) can be flags.
- **Argument files**: opt-in response files, as in `prog @args.txt`, for long argument lists. They help on Windows, where the command line is capped at about 32,000 characters, and when an agent builds a large invocation. They stay opt-in because `@` can be a real argument.
- **Sensitive values**: a flag marked `Sensitive` is masked everywhere its value could appear, in error messages, help defaults, environment hints, Agent JSON and the MCP schema. Otherwise a mistyped `--token` ends up in an agent transcript or a CI log.
- **Prompts**: `cmd.Confirm()` asks a person in a terminal. When nobody can answer, because stdin isn't a terminal or the Agent modifier is on, it fails at once with a `ConfirmationRequired` kind and a hint to pass `--yes`. An agent never sits waiting on a prompt it can't see.
- **Warnings**: deprecated flags and commands produce structured warnings instead of loose text on stderr. Agent JSON carries them in a `warnings` array next to `errors`.
- **JSON formats**: Agent JSON, `jsonhelp` output and the MCP schema become an API as soon as an agent parses them. Each carries a `version` field and a documented format, and changes to them follow semver like the Go API. Agent JSON already carries one.
- **Commands**: `cmd.Walk()` and `cmd.AllFlags()` as iterators. `Run` checks the command tree for duplicate names and clashing short flags before it parses anything, and returns an error for a bad definition.
- **Annotations**: `ReadOnly`, `Destructive` and `Idempotent` on a command. They show in help and the JSON schema, and the `mcp` module maps them to the MCP tool hints, so a client can ask before it runs `prog delete`.
- **Version**: when an app doesn't set one, `--version` falls back to the module version and VCS revision from `runtime/debug.ReadBuildInfo`.
- **Exit status docs**: the `docs` module writes the man page's EXIT STATUS section from the personality's code table, so the documented codes can't drift from the real ones.
- **Globals**: `OsExiter`, `ErrWriter`, `HelpFlag` and the other package variables become fields on the root command, so parallel tests stop sharing state.
- **Windows**: `cli.Main` treats Ctrl-C the same on every platform. SIGTERM behaves differently on Windows, and colour needs the console's VT mode, so the docs will state what each platform gets.

## Migrating from v3

The aim is that most apps move with one command and no change in what their users see.

- **Same names where the meaning is the same.** `Command`, `Name`, `Usage`, `Action`, `Writer`, `ErrWriter`, the action signature, `ExitCoder` and `Exit(message, code)` carry over unchanged. v3 flag types become generic aliases, such as `type StringFlag = Flag[string]`.
- **`go fix` could do the mechanical work.** A final v3 release would mark each function that has a direct v4 equivalent with `// Deprecated:` and `//go:fix inline`, and `go fix ./...` would then rewrite callers and the import path. Go 1.26's `go fix` documents this use for moving to a higher major version, but nobody has tried it on v3 yet, so it needs a small experiment before we rely on it. The catch is that the final v3 release would have to import v4.
- **A fixer for what inlining can't reach.** Package variables that become fields (`cli.OsExiter`, `cli.HelpFlag`, `cli.HelpPrinter`) and custom templates need rewrites a function body can't express. We ship an analyzer for `go fix -fixtool`, and it lists anything it can't rewrite by file and line.
- **Output stays the same.** The V3 preset and `help/tmpl` keep output byte for byte, so an app can move its code first and change its output later.
- **One command at a time.** v3 and v4 have different module paths, so both can live in one binary. A small adapter module runs a v3 command as a v4 subcommand, so a large app can move its tree piece by piece.
- **A guide and a support window.** `docs/migrate-v3-to-v4.md` follows the v2 to v3 guide, and v3 gets bug and security fixes for a set period after v4.0.
- **v2 support ends with v4.0.** v2 has had fixes since November 2019, almost six years. The proposal is that v2 keeps getting fixes while v4 is in development, which gives v2 users that time to move to v3, and support ends when v4.0 ships. We should announce the date now so nobody is surprised.

## Building v4 with agents

Much of v4 can be written by coding agents, as long as the rules and the checks are clear before any code is.

- **An `AGENTS.md` in the repo** sets the rules every change follows. The core imports nothing on the dependency guard's list, there are no package globals, every user-facing string goes through the catalogue, and tests come before code.
- **One issue per milestone**, with its acceptance tests written first. An agent works from the issue, and the tests say when it is done.
- **The v3 compatibility suite and the v3 against v4 fuzzer** judge behaviour, so "compatible with v3" is something CI checks rather than something a reviewer has to spot.
- **A maintainer reviews every PR.** Agents write the code, and we decide what goes in.

## Build and CI

This covers "Rework build system" from [#2192](https://github.com/urfave/cli/issues/2192). CI runs plain `go` commands, so what passes on a laptop passes in CI.

- **No custom build runner.** `scripts/build.go` goes. Every check is a `go test`, `go vet` or `go tool` command. The Makefile keeps its one-line targets, which call `go` directly instead of `scripts/build.go`.
- **Pinned tools in their own module.** staticcheck, apidiff, govulncheck and gfmrun sit in `tools/go.mod` through the `tool` directive and run with `go tool`, instead of `scripts/build.go` downloading a gfmrun release binary. The core `go.mod` stays free of requirements.
- **One workspace, many modules.** A `go.work` ties the core, `docs`, `altsrc`, `xtext`, `mcp` and the v3 compatibility suite together for local work. CI tests each module on its own, and each is tagged on its own, such as `mcp/v0.3.0`.
- **Matrix.** The two supported Go releases on Linux, macOS and Windows, with `-race`.
- **Dependency guard.** The existing test that fails if the core imports `text/template`, `encoding/json`, `net` or `golang.org/x`.
- **Size budget.** CI builds `examples/hello` with `-ldflags='-s -w'`, fails above a set size, and reports the change in bytes on every PR.
- **API check.** `apidiff` against the last tag fails a PR that breaks the public API without a major version bump. It replaces the `godoc-current.txt` comparison.
- **Fuzzing.** The parser gets `testing.F` fuzz targets, with a short run on every PR and a long one nightly. Both [#2438](https://github.com/urfave/cli/issues/2438) and [#2419](https://github.com/urfave/cli/issues/2419) are the kind of edge case a fuzzer finds. A second target feeds the same arguments to the v3 and v4 parsers and fails when they disagree, so the compatibility promise is checked on every change. It lives in the compatibility module, so the core never requires v3.
- **Translations.** A test fails when a shipped language package misses a key the English catalogue has.
- **Docs.** gfmrun runs the code blocks in the README and docs, and `Example` tests cover the rest, so a doc example can't drift from the API.
- **Security.** `govulncheck` on every PR and on a weekly schedule.

## Ideas from other libraries

We looked at clap, argparse, click, yargs, oclif and cobra. Some ideas carry over to Go, some need changing, and some don't fit.

- **Typed errors with structured context** (clap): taken, but kinds stay open so applications can add their own, where clap's list is a closed enum.
- **One catalogue for every built-in string** (argparse, yargs): taken.
- **Where a value came from** (clap, click): taken as `cmd.Source`.
- **The `__complete` protocol** (cobra): taken, because existing shell glue and carapace already speak it.
- **A test runner** (click): taken as `clitest`, kept thin because Go already lets us inject writers.
- **Help and version as errors** (clap): rejected. In Go every `err != nil` check would treat `--help` as a failure.
- **Short help for `-h`, long help for `--help`** (clap): rejected. urfave/cli users expect them to be the same.
- **A global `--json` flag** (oclif): changed. A library shouldn't reserve a flag name, so the Agent modifier turns on through an environment variable or a flag the app chooses to add.
- **A separate definition check** (clap `debug_assert`): changed. `Run` validates the tree itself, so nobody has to remember to call it.
- **MCP servers built from the command tree** (ophis, mcp-cobra): taken, as a module we support. `cli/v4/mcp` serves each command as an MCP tool, using the JSON Schema from `help/jsonhelp`, so nobody has to write their own adapter. It has its own `go.mod` and version tags, so the MCP SDK never reaches the core and the module can follow the spec without a core release.
- **Manifests and lazy loading** (oclif, click): rejected. A Go binary has no startup cost to save.

## Standard library and toolchain we build on

- `errors.Is`, `errors.As` and `errors.Join` for matching and multiple errors.
- `signal.NotifyContext` for cancellation on Ctrl-C.
- `iter.Seq` for walking commands and flags.
- `encoding.TextUnmarshaler` for typed flags, as `flag.TextVar` does.
- `text/tabwriter` for help, and `go/doc/comment` for markdown and man output in `docs`.
- `embed` and `fs.FS` for catalogues and completion scripts.
- `runtime/debug.ReadBuildInfo` for a default version.
- `testing.F` for fuzzing the parser, and `testing/synctest` for signal and timeout tests.
- `go fix` with `//go:fix inline` for the v3 migration.
- The `tool` directive and `go tool` for pinned build tools, and `go.work` for the multi-module repo.

`golang.org/x/text` gives full CLDR plural rules and locale matching, but it adds roughly a megabyte, so it stays behind the `xtext` adapter.

## Open questions

- Is Go 1.25 the right minimum?
- Should the spike live in `v4/` on `main`, or on a long-lived `v4` branch?
- Which presets and modifiers should ship in the first release?
- Is `URFAVE_CLI_AGENT` the right name, and should `Auto` also recognise variables that agent tools may already set? Claude Code may set `CLAUDECODE`, for example, which needs confirming.
- How long should v3 get fixes after v4.0?
- Does ending v2 support at v4.0 work for everyone?
- What size budget should CI enforce for `examples/hello`?

## Release process

To be laid out by the maintainers in [#2446](https://github.com/urfave/cli/discussions/2446).
