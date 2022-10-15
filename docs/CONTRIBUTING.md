## Contributing

Welcome to the `urfave/cli` contributor docs! This goal of this document is to help those
interested in joining the 200+ humans who have contributed to this project over the years.

> As a general guiding principle, the current maintainers may be notified via the
> @urfave/cli GitHub team.

All of the current maintainers are *volunteers* who live in various timezones with
different scheduling needs, so please understand that your contribution or question may
not get a response for many days.

### semantic versioning adherence

The `urfave/cli` project strives to strictly adhere to [semantic
versioning](https://semver.org/spec/v2.0.0.html). The active development branches and the
milestones and import paths to which they correspond are:

#### `main` branch

<https://github.com/urfave/cli/tree/main>

The majority of active development and issue management is targeting the `main` branch,
which **MUST** *only* receive bug fixes and feature *additions*.

- :arrow_right: [`v2.x`](https://github.com/urfave/cli/milestone/16)
- :arrow_right: `github.com/urfave/cli/v2`

The `main` branch in particular includes tooling to help with keeping the `v2.x` series
backward compatible. More details on this process are in the development workflow section
below.

#### `v1` branch

<https://github.com/urfave/cli/tree/v1>

The `v1` branch **MUST** only receive bug fixes in the `v1.22.x` series. There is no
strict rule regarding bug fixes to the `v2.x` series being backported to the `v1.22.x`
series.

- :arrow_right: [`v1.22.x`](https://github.com/urfave/cli/milestone/11)
- :arrow_right: `github.com/urfave/cli`

#### `v3-dev-main` branch

<https://github.com/urfave/cli/tree/v3-dev-main>

The `v3-dev-branch` **MUST** receive all bug fixes and features added to the `main` branch
and **MAY** receive feature *removals* and other changes that are otherwise
*backward-incompatible* with the `v2.x` series.

- :arrow_right: [`v3.x`](https://github.com/urfave/cli/milestone/5)
- unreleased / unsupported

### development workflow

Most of the tooling around the development workflow strives for effective
[dogfooding](https://en.wikipedia.org/wiki/Eating_your_own_dog_food). There is a top-level
`Makefile` that is maintained strictly for the purpose of easing verification of one's
development environment and any changes one may have introduced:

```sh
make
```

Running the default `make` target (`all`) will ensure all of the critical steps are run to
verify one's changes are harmonious in nature. The same steps are also run during the
[continuous integration
phase](https://github.com/urfave/cli/blob/main/.github/workflows/cli.yml).

In the event that the `v2diff` target exits non-zero, this is a signal that the public API
surface area has changed. If the changes adhere to semantic versioning, meaning they are
*additions* or *bug fixes*, then manually running the approval step will "promote" the
current `go doc` output:

```sh
make v2approve
```

Because the `generate` step includes updating `godoc-current.txt` and
`testdata/godoc-v2.x.txt`, these changes *MUST* be part of any proposed pull request so
that reviewers have an opportunity to also make an informed decision about the "promotion"
step.

#### generated code

A significant portion of the project's source code is generated, with the goal being to
eliminate repetitive maintenance where other type-safe abstraction is impractical or
impossible with Go versions `< 1.18`. In a future where the eldest Go version supported is
`1.18.x`, there will likely be efforts to take advantage of
[generics](https://go.dev/doc/tutorial/generics).

The built-in `go generate` command is used to run the commands specified in
`//go:generate` directives. Each such command runs a file that also supports a command
line help system which may be consulted for further information, e.g.:

```sh
go run cmd/urfave-cli-genflags/main.go --help
```

#### docs output

The documentation in the `docs` directory is automatically built via `mkdocs` into a
static site and published when releases are pushed (see [RELEASING](./RELEASING/)). There
is no strict requirement to build the documentation when developing locally, but the
following `make` targets may be used if desired:

```sh
# install documentation dependencies with `pip`
make ensure-mkdocs
```

```sh
# build the static site in `./site`
make docs
```

```sh
# start an mkdocs development server
make serve-docs
```

### pull requests

Please feel free to open a pull request to fix a bug or add a feature. The @urfave/cli
team will review it as soon as possible, giving special attention to maintaining backward
compatibility. If the @urfave/cli team agrees that your contribution is in line with the
vision of the project, they will work with you to get the code into a mergeable state,
merged, and then released.

### granting of commit bit / admin mode

Those with a history of contributing to this project will likely be invited to join the
@urfave/cli team. As a member of the @urfave/cli team, you will have the ability to fully
administer pull requests, issues, and other repository bits.

If you feel that you should be a member of the @urfave/cli team but have not yet been
added, the most likely explanation is that this is an accidental oversight! :sweat_smile:.
Please open an issue!

<!--
vim:tw=90
-->
