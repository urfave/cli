# Releasing urfave/cli

Releasing small batches often is [backed by
research](https://itrevolution.com/accelerate-book/) as part of the
virtuous cycles that keep teams and products healthy.

To that end, the overall goal of the release process is to send
changes out into the world as close to the time the commits were
merged to the `main` branch as possible. In this way, the community
of humans depending on this library are able to make use of the
changes they need **quickly**, which means they shouldn't have to
maintain long-lived forks of the project, which means they can get
back to focusing on the work on which they want to focus. This also
means that the @urfave/cli team should be able to focus on
delivering a steadily improving product with significantly eased
ability to associate bugs and regressions with specific releases.

## Process

- Release versions follow [semantic versioning](https://semver.org/)
- Releases are associated with **signed, annotated git tags**[^1].
- Release notes are **automatically generated**[^2].

In the `main` or `v2-maint` branch, the current version is always
available via:

```sh
git describe --always --dirty --tags
```

**NOTE**: if the version reported contains `-dirty`, this is
indicative of a "dirty" work tree, which is not a great state for
creating a new release tag. Seek help from @urfave/cli teammates.

For example, given a described version of `v2.4.7-3-g68da1cd` and a
diff of `v2.4.7...` that contains only bug fixes, the next version
should be `v2.4.8`:

```sh
git tag -a -s -m 'Release 2.4.8' v2.4.8
git push origin v2.4.8
```

The tag push will trigger a GitHub Actions workflow and will be
**immediately available** to the [Go module mirror, index, and
checksum database](https://proxy.golang.org/). The remaining steps
require human intervention through the GitHub web view although
[automated solutions
exist](https://github.com/softprops/action-gh-release) that may be
adopted in the future.

- Open the [the new release page](https://github.com/urfave/cli/releases/new)
- At the top of the form, click on the `Choose a tag` select control and select `v2.4.8`
- In the `Write` tab below, click the `Auto-generate release notes` button
- At the bottom of the form, click the `Publish release` button
- :white_check_mark: you're done!

[^1]: This was not always true. There are many **lightweight git
  tags** present in the repository history.

[^2]: This was not always true. The
  [`docs/CHANGELOG.md`](./CHANGELOG.md) document used to be
  manually maintained. Relying on the automatic release notes
  generation requires the use of **merge commits** as opposed to
  squash merging or rebase merging.
