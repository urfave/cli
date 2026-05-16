## CLI interface - greet

Description of the application.

Some app.

> app [first_arg] [second_arg]

Usage:

```bash
$ /usr/local/bin [GLOBAL FLAGS] [COMMAND] [COMMAND FLAGS] [ARGUMENTS...]
```

Global flags:

| Name                        | Description        | Default value |  Environment variables  |
|-----------------------------|--------------------|:-------------:|:-----------------------:|
| `--socket="…"` (`-s`)       | some 'usage' text  |    `value`    |         *none*          |
| `--flag="…"` (`--fl`, `-f`) |                    |               |         *none*          |
| `--another-flag` (`-b`)     | another usage text |    `false`    | `EXAMPLE_VARIABLE_NAME` |

### `config` command (aliases: `c`)

another usage test.

Usage:

```bash
$ /usr/local/bin [GLOBAL FLAGS] config [COMMAND FLAGS] [ARGUMENTS...]
```

The following flags are supported:

| Name                        | Description        | Default value | Environment variables |
|-----------------------------|--------------------|:-------------:|:---------------------:|
| `--flag="…"` (`--fl`, `-f`) |                    |               |        *none*         |
| `--another-flag` (`-b`)     | another usage text |    `false`    |        *none*         |

### `config sub-config` subcommand (aliases: `s`, `ss`)

another usage test.

Usage:

```bash
$ /usr/local/bin [GLOBAL FLAGS] config sub-config [COMMAND FLAGS] [ARGUMENTS...]
```

The following flags are supported:

| Name                                | Description     | Default value | Environment variables |
|-------------------------------------|-----------------|:-------------:|:---------------------:|
| `--sub-flag="…"` (`--sub-fl`, `-s`) |                 |               |        *none*         |
| `--sub-command-flag` (`-s`)         | some usage text |    `false`    |        *none*         |

### `info` command (aliases: `i`, `in`)

retrieve generic information.

Usage:

```bash
$ /usr/local/bin [GLOBAL FLAGS] info [ARGUMENTS...]
```

### `some-command` command

Usage:

```bash
$ /usr/local/bin [GLOBAL FLAGS] some-command [ARGUMENTS...]
```

### `usage` command (aliases: `u`)

standard usage text.

> Usage for the usage text
> - formatted:  Based on the specified ConfigMap and summon secrets.yml
> - list:       Inspect the environment for a specific process running on a Pod
> - for_effect: Compare 'namespace' environment with 'local'
> ```
> func() { ... }
> ```
> Should be a part of the same code block

Usage:

```bash
$ /usr/local/bin [GLOBAL FLAGS] usage [COMMAND FLAGS] [ARGUMENTS...]
```

The following flags are supported:

| Name                        | Description        | Default value | Environment variables |
|-----------------------------|--------------------|:-------------:|:---------------------:|
| `--flag="…"` (`--fl`, `-f`) |                    |               |        *none*         |
| `--another-flag` (`-b`)     | another usage text |    `false`    |        *none*         |

### `usage sub-usage` subcommand (aliases: `su`)

standard usage text.

> Single line of UsageText

Usage:

```bash
$ /usr/local/bin [GLOBAL FLAGS] usage sub-usage [COMMAND FLAGS] [ARGUMENTS...]
```

The following flags are supported:

| Name                        | Description     | Default value | Environment variables |
|-----------------------------|-----------------|:-------------:|:---------------------:|
| `--sub-command-flag` (`-s`) | some usage text |    `false`    |        *none*         |
