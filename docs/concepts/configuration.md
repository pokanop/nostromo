---
title: Configuration and themes
---

# Configuration and themes

`nostromo` has a small set of settings, read and written with `get` and `set`:

```sh
nostromo get <key>
nostromo set <key> <value>
```

| Key           | Type                                          | Default       | Stored in                       | Effect                                                                                     |
| ------------- | --------------------------------------------- | ------------- | ------------------------------- | ------------------------------------------------------------------------------------------ |
| `verbose`     | `true` / `false`                              | `false`       | core manifest `config.verbose`  | Always print debug output (same as passing `-v` / `--verbose` to every command)             |
| `aliasesOnly` | `true` / `false`                              | `false`       | core manifest `config.aliasesonly` | Make every command added from now on an [alias-only](alias-only.md) command             |
| `mode`        | `concatenate` / `independent` / `exclusive`   | `concatenate` | core manifest `config.mode`     | Default [execution mode](modes.md) for commands added or updated without `-m`               |
| `backupCount` | number                                        | `10`          | core manifest `config.backupcount` | Maximum [backups](backups.md) kept per manifest; `0` disables backups                    |
| `theme`       | `default` / `grayscale` / `emoji`             | `emoji`       | `spaceport.yaml` `theme`        | Colour scheme and icons used for output                                                    |

Keys are case sensitive on the command line (`backupCount`, not `backupcount`). In the YAML files they're written lowercase by `nostromo`.

Examples:

```sh
nostromo set mode independent
nostromo set backupCount 20
nostromo set theme grayscale
nostromo get theme
```

`nostromo show` prints the current `[config]` values for each manifest.

## Verbose output

For a one-off look at what `nostromo` is doing, pass `-v` / `--verbose` to any command; it's a persistent flag available everywhere:

```sh
nostromo -v eval foo bar
```

Turning on the `verbose` setting makes this permanent and also makes `nostromo show` print more detail for every command.

## `NOSTROMO_HOME`

Everything `nostromo` writes lives under one folder, `~/.nostromo` by default. Set the `NOSTROMO_HOME` environment variable to relocate it:

```sh
export NOSTROMO_HOME="$XDG_CONFIG_HOME/nostromo"
```

Export it in your shell startup file *before* the `nostromo` block so completions and commands find the right manifests.

## Themes

`nostromo` supports themes to make it look even more neat. There are three:

- `default` — the basic theme and previous default
- `grayscale` — gray colored things are sometimes nice
- `emoji` — the new default, obviously :rocket:

```sh
nostromo set theme <name>
```

The theme is stored in `spaceport.yaml` (as `0`, `1` or `2` respectively) rather than in a manifest, so it's a per-machine preference that doesn't travel with shared manifests.

Enjoy!

🐳📑🍥🌞🍓🕖🕐💘🎵🌑🐻🐜📙💥👡🍈👝🎭🐄🌓🎏👔📁🍝🔼🕔💩🌒📥
