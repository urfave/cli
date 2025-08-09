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

The majority of active development and issue management is targeting the `main` branch.

- :arrow_right: [`v3.x`](https://github.com/urfave/cli/milestone/5)
- :arrow_right: `github.com/urfave/cli/v3`

The `main` branch includes tooling to help with keeping track of `v3.x` series backward
compatibility. More details on this process are in the development workflow section below.

#### `v1-maint` branch

<https://github.com/urfave/cli/tree/v1-maint>

The `v1-maint` branch **MUST** only receive bug fixes in the `v1.22.x` series. There is no
strict rule regarding bug fixes to the `v3.x` or `v2.23.x` series being backported to the
`v1.22.x` series.

- :arrow_right: [`v1.22.x`](https://github.com/urfave/cli/milestone/11)
- :arrow_right: `github.com/urfave/cli`

#### `v2-maint` branch

<https://github.com/urfave/cli/tree/v2-maint>

The `v2-maint` branch **MUST** only receive bug fixes in the `v2.23.x` series. There is no
strict rule regarding bug fixes to the `v3.x` series being backported to the `v2.23.x`
series.

- :arrow_right: [`v2.23.x`](https://github.com/urfave/cli/milestone/16)
- :arrow_right: `github.com/urfave/cli/v2`

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
phase](https://github.com/urfave/cli/blob/main/.github/workflows/test.yml).

`gfmrun` is required to run the examples, and without it `make all` will fail.

You can find `gfmrun` here:

- [urfave/gfmrun](https://github.com/urfave/gfmrun)

To install `gfmrun`, you can use `go install`:

```
go install github.com/urfave/gfmrun/cmd/gfmrun@latest
```

In the event that the `v3diff` target exits non-zero, this is a signal that the public API
surface area has changed. If the changes are acceptable, then manually running the
approval step will "promote" the current `go doc` output:

```sh
make v3approve
```

Because the `generate` step includes updating `godoc-current.txt` and
`testdata/godoc-v3.x.txt`, these changes *MUST* be part of any proposed pull request so
that reviewers have an opportunity to also make an informed decision about the "promotion"
step.

#### docs output

The documentation in the `docs` directory is automatically built via `mkdocs` into a
static site and published when releases are pushed (see [RELEASING](./RELEASING.md)). There
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
